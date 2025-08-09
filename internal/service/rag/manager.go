package rag

import (
    "context"
    "errors"
    "sync"
    "time"
)

// Manager 记忆系统入口，负责：
// - 路由写入：内存/磁盘（JSON），外部服务（预留）
// - 异步写入：降低写路径延迟
// - 检索：本地优先，后续可融合外部检索结果
// - 多租户隔离：通过 Tenant 实现

type Manager struct {
    opts RAGOptions

    mem  Store          // 可选内存后端
    disk Store          // 可选 JSONL 后端

    vec  VectorClient   // 预留：外部向量检索
    tri  TripleClient   // 预留：外部三元组检索

    // 异步写入
    asyncCh chan saveTask
    wg      sync.WaitGroup
    mu      sync.RWMutex
    closed  bool
}

type saveTask struct {
    ctx  context.Context
    item MemoryItem
    opt  SaveOptions
}

func NewManager(opts RAGOptions, vec VectorClient, tri TripleClient) (*Manager, error) {
    m := &Manager{opts: opts}

    // 存储后端
    if opts.InMemory.Enable {
        m.mem = NewMemoryStore(opts.InMemory)
    }
    if opts.DiskJSON.Enable {
        ds, err := NewDiskJSONStore(opts.Namespace, opts.DiskJSON)
        if err != nil {
            return nil, err
        }
        m.disk = ds
    }

    // 外部客户端
    if vec != nil {
        m.vec = vec
    } else {
        m.vec = NoopVectorClient{}
    }
    if tri != nil {
        m.tri = tri
    } else {
        m.tri = NoopTripleClient{}
    }

    // 异步写入
    if opts.Async.Enable {
        size := opts.Async.QueueSize
        if size <= 0 { size = 1024 }
        workers := opts.Async.Workers
        if workers <= 0 { workers = 1 }
        m.asyncCh = make(chan saveTask, size)
        for i := 0; i < workers; i++ {
            m.wg.Add(1)
            go m.worker()
        }
    }
    return m, nil
}

func (m *Manager) worker() {
    defer m.wg.Done()
    for task := range m.asyncCh {
        // 背景上下文保证落盘不被上游过早取消
        ctx := context.Background()
        if task.ctx != nil {
            // 设定短超时，避免卡死
            var cancel context.CancelFunc
            ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
            _ = m.saveSync(ctx, task.item, task.opt)
            cancel()
            continue
        }
        _ = m.saveSync(ctx, task.item, task.opt)
    }
}

// Close 关闭后台资源
func (m *Manager) Close(ctx context.Context) error {
    // 幂等关闭：仅第一次生效
    m.mu.Lock()
    if m.closed {
        m.mu.Unlock()
        return nil
    }
    m.closed = true
    ch := m.asyncCh
    m.mu.Unlock()

    if ch != nil {
        close(ch)
    }
    m.wg.Wait()
    if m.mem != nil { _ = m.mem.Close(ctx) }
    if m.disk != nil { _ = m.disk.Close(ctx) }
    return nil
}

// Save 写入记忆
// - 根据 SaveOptions 或默认配置路由到内存/磁盘
// - 异步模式：推送到队列
func (m *Manager) Save(ctx context.Context, item MemoryItem, opt SaveOptions) error {
    if item.Tenant.UserID == "" || item.Tenant.ArchiveID == "" {
        return errors.New("tenant(user_id, archive_id) 不能为空")
    }
    if item.CreatedAt.IsZero() {
        item.CreatedAt = time.Now()
    }

    if m.opts.Async.Enable {
        // 读取并在发送期间持有读锁，防止与 Close() 竞争引发向已关闭通道发送
        m.mu.RLock()
        ch := m.asyncCh
        closed := m.closed
        if ch != nil && !closed {
            select {
            case ch <- saveTask{ctx: ctx, item: item, opt: opt}:
                m.mu.RUnlock()
                return nil
            default:
                m.mu.RUnlock()
                // 队列满，降级为同步，避免丢失
                return m.saveSync(ctx, item, opt)
            }
        }
        m.mu.RUnlock()
    }
    return m.saveSync(ctx, item, opt)
}

func (m *Manager) saveSync(ctx context.Context, item MemoryItem, opt SaveOptions) error {
    // 默认路由：跟随全局配置
    toMem := opt.ToMemory || (!opt.ToDisk && !opt.ToVector && !opt.ToTriple && m.opts.InMemory.Enable)
    toDisk := opt.ToDisk || (!opt.ToMemory && !opt.ToVector && !opt.ToTriple && m.opts.DiskJSON.Enable)

    var firstErr error
    if toMem && m.mem != nil {
        if err := m.mem.Save(ctx, item); err != nil { firstErr = err }
    }
    if toDisk && m.disk != nil {
        if err := m.disk.Save(ctx, item); err != nil && firstErr == nil { firstErr = err }
    }
    // 预留：向量与三元组保存（通常需要 embedding/抽取，此处不主动调用）
    return firstErr
}

// Query 检索记忆（本地优先；外部检索预留）
func (m *Manager) Query(ctx context.Context, req QueryRequest) (QueryResult, error) {
    // TopK 默认
    topK := req.TopK
    if topK <= 0 { topK = 10 }

    var merged []MemoryItem

    // 1) 内存
    if m.mem != nil {
        r, err := m.mem.Query(ctx, req)
        if err == nil && len(r.Items) > 0 {
            merged = append(merged, r.Items...)
        }
    }

    // 2) 磁盘
    if len(merged) < topK && m.disk != nil {
        r, err := m.disk.Query(ctx, req)
        if err == nil && len(r.Items) > 0 {
            merged = append(merged, r.Items...)
        }
    }

    // 3) 预留：外部向量/三元组（按需启用并去重合并）
    // 若未来启用，可在此处调用 m.vec.Query / m.tri.QueryTriples，并按 Score 排序去重

    // 截断到 topK
    if len(merged) > topK { merged = merged[:topK] }

    return QueryResult{Items: merged}, nil
}

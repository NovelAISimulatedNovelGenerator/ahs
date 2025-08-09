package rag

import (
    "context"
    "strings"
    "sync"
    "time"
)

// memoryStore 进程内存存储（按租户隔离）
// - 简单子串匹配与标签/类型过滤
// - 基于配置的容量上限与可选 TTL 过滤

type memoryStore struct {
    mu         sync.RWMutex
    itemsByKey map[string][]MemoryItem // tenantKey -> items (按时间追加)

    maxEntries int
    ttl        time.Duration
}

func NewMemoryStore(opts InMemoryOptions) Store {
    return &memoryStore{
        itemsByKey: make(map[string][]MemoryItem),
        maxEntries: opts.MaxEntries,
        ttl:        opts.TTL,
    }
}

func (m *memoryStore) tenantKey(t Tenant) string {
    return t.UserID + "::" + t.ArchiveID
}

func (m *memoryStore) Save(ctx context.Context, item MemoryItem) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    key := m.tenantKey(item.Tenant)
    lst := append(m.itemsByKey[key], item)

    // 容量控制：超过上限时丢弃最旧的
    if m.maxEntries > 0 && len(lst) > m.maxEntries {
        lst = lst[len(lst)-m.maxEntries:]
    }
    m.itemsByKey[key] = lst
    return nil
}

func (m *memoryStore) Query(ctx context.Context, req QueryRequest) (QueryResult, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()

    key := m.tenantKey(req.Tenant)
    lst := m.itemsByKey[key]

    now := time.Now()
    res := make([]MemoryItem, 0, len(lst))

    for i := len(lst) - 1; i >= 0; i-- { // 从新到旧扫描，利于 TopK
        it := lst[i]

        // TTL 过滤（若启用）
        if m.ttl > 0 && it.CreatedAt.Add(m.ttl).Before(now) {
            continue
        }
        // 显式过期时间
        if it.ExpiresAt != nil && it.ExpiresAt.Before(now) {
            continue
        }
        // 类型过滤
        if len(req.Kinds) > 0 {
            ok := false
            for _, k := range req.Kinds {
                if it.Kind == k {
                    ok = true
                    break
                }
            }
            if !ok {
                continue
            }
        }
        // 标签过滤
        if len(req.Tags) > 0 {
            tagOK := true
            for _, want := range req.Tags {
                found := false
                for _, t := range it.Tags {
                    if t == want {
                        found = true
                        break
                    }
                }
                if !found {
                    tagOK = false
                    break
                }
            }
            if !tagOK {
                continue
            }
        }
        // 文本匹配（简单包含）
        if req.Query != "" && !strings.Contains(strings.ToLower(it.Content), strings.ToLower(req.Query)) {
            continue
        }

        res = append(res, it)
        if req.TopK > 0 && len(res) >= req.TopK {
            break
        }
    }

    return QueryResult{Items: res}, nil
}

func (m *memoryStore) Close(ctx context.Context) error { return nil }

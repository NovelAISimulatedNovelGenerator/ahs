package rag

import (
    "bufio"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "time"
)

// diskJSONStore 基于 JSONL 的本地持久化
// 每个租户一个 JSONL 文件：{RootPath}/{Namespace}/{user_id}/{archive_id}/data.jsonl
// 逐行追加；查询时顺序读取并在内存中过滤（中小规模适用）

type diskJSONStore struct {
    root      string
    namespace string
    maxBytes  int64
}

func NewDiskJSONStore(ns string, opts DiskJSONOptions) (Store, error) {
    if opts.RootPath == "" {
        return nil, errors.New("DiskJSON.RootPath 不能为空")
    }
    return &diskJSONStore{
        root:      opts.RootPath,
        namespace: ns,
        maxBytes:  opts.MaxFileBytes,
    }, nil
}

func (s *diskJSONStore) pathOf(t Tenant) string {
    // {RootPath}/{Namespace}/{user_id}/{archive_id}/data.jsonl
    dir := filepath.Join(s.root, s.namespace, safePath(t.UserID), safePath(t.ArchiveID))
    return filepath.Join(dir, "data.jsonl")
}

func safePath(p string) string {
    p = strings.TrimSpace(p)
    p = strings.ReplaceAll(p, "..", "_")
    p = strings.ReplaceAll(p, string(os.PathSeparator), "_")
    if p == "" { p = "_" }
    return p
}

func (s *diskJSONStore) ensureDir(filePath string) error {
    return os.MkdirAll(filepath.Dir(filePath), 0o755)
}

func (s *diskJSONStore) Save(ctx context.Context, item MemoryItem) error {
    fp := s.pathOf(item.Tenant)
    if err := s.ensureDir(fp); err != nil {
        return fmt.Errorf("ensure dir: %w", err)
    }

    // 如果未设置时间戳，自动填充
    if item.CreatedAt.IsZero() {
        item.CreatedAt = time.Now()
    }

    f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
    if err != nil {
        return fmt.Errorf("open file: %w", err)
    }
    defer f.Close()

    enc := json.NewEncoder(f)
    if err := enc.Encode(&item); err != nil {
        return fmt.Errorf("encode json: %w", err)
    }

    // 简单的大小限制（可选）：超过 maxBytes 时不处理（预留压缩策略）
    if s.maxBytes > 0 {
        if st, err := f.Stat(); err == nil && st.Size() > s.maxBytes {
            // 预留：触发后台压缩/轮转
        }
    }
    return nil
}

func (s *diskJSONStore) Query(ctx context.Context, req QueryRequest) (QueryResult, error) {
    fp := s.pathOf(req.Tenant)
    f, err := os.Open(fp)
    if err != nil {
        if os.IsNotExist(err) {
            return QueryResult{Items: nil}, nil
        }
        return QueryResult{}, fmt.Errorf("open file: %w", err)
    }
    defer f.Close()

    // 顺序读取后逆序筛选，保证新数据优先
    var all []MemoryItem
    sc := bufio.NewScanner(f)
    for sc.Scan() {
        var it MemoryItem
        if err := json.Unmarshal(sc.Bytes(), &it); err == nil {
            all = append(all, it)
        }
    }
    if err := sc.Err(); err != nil {
        return QueryResult{}, fmt.Errorf("scan jsonl: %w", err)
    }

    now := time.Now()
    res := make([]MemoryItem, 0, len(all))
    for i := len(all) - 1; i >= 0; i-- {
        it := all[i]
        // 过期过滤
        if it.ExpiresAt != nil && it.ExpiresAt.Before(now) {
            continue
        }
        // 类型过滤
        if len(req.Kinds) > 0 {
            ok := false
            for _, k := range req.Kinds {
                if it.Kind == k { ok = true; break }
            }
            if !ok { continue }
        }
        // 标签过滤（子集包含）
        if len(req.Tags) > 0 {
            tagOK := true
            for _, want := range req.Tags {
                found := false
                for _, t := range it.Tags { if t == want { found = true; break } }
                if !found { tagOK = false; break }
            }
            if !tagOK { continue }
        }
        // 文本包含
        if req.Query != "" && !strings.Contains(strings.ToLower(it.Content), strings.ToLower(req.Query)) {
            continue
        }
        res = append(res, it)
        if req.TopK > 0 && len(res) >= req.TopK { break }
    }
    return QueryResult{Items: res}, nil
}

func (s *diskJSONStore) Close(ctx context.Context) error { return nil }

package rag

import "time"

// Tenant 多租户隔离键，按 user_id + archive_id 双维度隔离
type Tenant struct {
    UserID    string `json:"user_id"`
    ArchiveID string `json:"archive_id"`
}

// MemoryKind 记忆类型（可扩展）
type MemoryKind string

const (
    KindShortTerm MemoryKind = "short_term" // 短期对话记忆
    KindLongTerm  MemoryKind = "long_term"  // 长期稳定记忆
    KindFact      MemoryKind = "fact"       // 事实性知识
    KindNote      MemoryKind = "note"       // 备注/笔记
)

// MemoryItem 记忆项
// 注意：Score 字段主要用于外部检索器返回打分结果，本地存储通常不设置。
type MemoryItem struct {
    ID        string                 `json:"id"`
    Tenant    Tenant                 `json:"tenant"`
    Content   string                 `json:"content"`
    Tags      []string               `json:"tags,omitempty"`
    Kind      MemoryKind             `json:"kind,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
    ExpiresAt *time.Time             `json:"expires_at,omitempty"`
    Meta      map[string]any         `json:"meta,omitempty"`
    Score     float64                `json:"score,omitempty"`
}

// QueryRequest 记忆检索请求
// Query: 文本查询（可为空，表示仅按标签/类型过滤）
// TopK: 期望返回条数，<=0 使用默认值
// Tags/Kinds: 过滤条件
// UseVector/UseTriple: 是否启用外部高级检索（由 Manager 决策）
type QueryRequest struct {
    Tenant    Tenant       `json:"tenant"`
    Query     string       `json:"query,omitempty"`
    TopK      int          `json:"top_k,omitempty"`
    Tags      []string     `json:"tags,omitempty"`
    Kinds     []MemoryKind `json:"kinds,omitempty"`
}

type QueryResult struct {
    Items []MemoryItem `json:"items"`
}

// SaveOptions 保存策略开关（由 Manager 解释并路由到底层后端）
type SaveOptions struct {
    ToMemory bool
    ToDisk   bool
    ToVector bool // 预留，向量后端
    ToTriple bool // 预留，三元组后端
}

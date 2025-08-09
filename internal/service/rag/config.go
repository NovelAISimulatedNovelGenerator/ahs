package rag

import "time"

// RAGOptions 记忆系统配置
// - InMemory: 进程内缓存
// - DiskJSON: 本地 JSONL 持久化
// - Vector/Triple: 外部服务（HTTP）占位，后续对接
// - Async: 异步写入配置
// - Retention: 预留压缩/保留策略
// - Namespace: 预留命名空间
// - ServiceMode: 偏好服务化/HTTP 对外
// 说明：遵循用户偏好，默认启用 JSON 持久化与异步写入。
type RAGOptions struct {
    InMemory    InMemoryOptions
    DiskJSON    DiskJSONOptions
    Vector      VectorOptions
    Triple      TripleOptions
    Async       AsyncOptions
    Retention   RetentionOptions
    Namespace   string
    ServiceMode bool // 预留：服务接口模式
}

type InMemoryOptions struct {
    Enable    bool
    MaxEntries int           // 每租户内存最大条目数
    TTL        time.Duration // 可选：查询时过滤过期
}

type DiskJSONOptions struct {
    Enable    bool
    RootPath  string // 数据根目录
    // 预留：按大小轮转、合并压缩
    MaxFileBytes int64
}

type VectorOptions struct {
    Enable   bool
    Endpoint string
    APIKey   string
    Index    string
    // Embedding 配置预留
    EmbeddingProvider string
    EmbeddingModel    string
    EmbeddingEndpoint string
    EmbeddingAPIKey   string
    Dim               int
}

type TripleOptions struct {
    Enable   bool
    Endpoint string
    APIKey   string
    SchemaVersion string
}

type AsyncOptions struct {
    Enable    bool
    QueueSize int
    Workers   int
}

type RetentionOptions struct {
    Enable  bool
    MaxDays int
    MaxBytes int64
}

// DefaultOptions 返回符合用户偏好的默认配置
func DefaultOptions() RAGOptions {
    return RAGOptions{
        InMemory: InMemoryOptions{
            Enable:     true,
            MaxEntries: 2048,
            TTL:        0,
        },
        DiskJSON: DiskJSONOptions{
            Enable:      true,
            RootPath:    "data/rag",
            MaxFileBytes: 0,
        },
        Vector: VectorOptions{Enable: false},
        Triple: TripleOptions{Enable: false},
        Async: AsyncOptions{
            Enable:    true,
            QueueSize: 1024,
            Workers:   1,
        },
        Retention: RetentionOptions{Enable: false},
        Namespace:   "default",
        ServiceMode: true,
    }
}

package rag

import (
    "context"
)

// Store 本地存储抽象
// 说明：实现需具备多租户隔离能力
// - Save: 保存/追加记忆
// - Query: 基于简单文本/标签/类型的过滤检索
// - Close: 释放资源
// 注意：打分由上游检索器或外部服务提供，本地存储可不负责评分

type Store interface {
    Save(ctx context.Context, item MemoryItem) error
    Query(ctx context.Context, req QueryRequest) (QueryResult, error)
    Close(ctx context.Context) error
}

// VectorClient 外部向量检索服务接口（HTTP 对接占位）
// - Query: 根据 QueryRequest 进行语义检索，返回打分的 MemoryItem 列表
// - Save: 可选能力，保存入库（通常需要 embedding），此处仅预留

type VectorClient interface {
    Query(ctx context.Context, req QueryRequest) (QueryResult, error)
    Save(ctx context.Context, item MemoryItem) error
}

// TripleClient 外部三元组检索/存储接口（HTTP 对接占位）
// - QueryTriples: 基于查询与过滤返回匹配条目
// - SaveTriples: 保存三元组（此处沿用 MemoryItem，后续可定义专用结构）

type TripleClient interface {
    QueryTriples(ctx context.Context, req QueryRequest) (QueryResult, error)
    SaveTriples(ctx context.Context, items []MemoryItem) error
}

// NoopVectorClient 默认空实现

type NoopVectorClient struct{}

func (NoopVectorClient) Query(ctx context.Context, req QueryRequest) (QueryResult, error) {
    return QueryResult{Items: nil}, nil
}
func (NoopVectorClient) Save(ctx context.Context, item MemoryItem) error { return nil }

// NoopTripleClient 默认空实现

type NoopTripleClient struct{}

func (NoopTripleClient) QueryTriples(ctx context.Context, req QueryRequest) (QueryResult, error) {
    return QueryResult{Items: nil}, nil
}
func (NoopTripleClient) SaveTriples(ctx context.Context, items []MemoryItem) error { return nil }

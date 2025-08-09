package rag

import (
    "sync"
)

var (
    mgrOnce sync.Once
    mgrInst *Manager
)

// InitDefault 使用给定配置初始化全局 Manager 实例。
// 若已初始化，则忽略后续调用。
func InitDefault(opts RAGOptions, vec VectorClient, tri TripleClient) error {
    var err error
    mgrOnce.Do(func() {
        var m *Manager
        m, err = NewManager(opts, vec, tri)
        if err == nil {
            mgrInst = m
        }
    })
    return err
}

// MustInitDefault 使用默认配置初始化全局 Manager。
func MustInitDefault() {
    _ = InitDefault(DefaultOptions(), nil, nil)
}

// Default 返回全局 Manager 实例；若未初始化，则按默认配置初始化。
func Default() *Manager {
    if mgrInst == nil {
        MustInitDefault()
    }
    return mgrInst
}

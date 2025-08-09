package rag

import (
    "context"
    "testing"
    "time"
)

func newTestManagerAsync(t *testing.T) *Manager {
    t.Helper()
    opts := DefaultOptions()
    opts.InMemory.Enable = true
    opts.DiskJSON.Enable = false
    opts.Async.Enable = true
    opts.Async.QueueSize = 64
    opts.Async.Workers = 1

    m, err := NewManager(opts, nil, nil)
    if err != nil { t.Fatalf("new manager: %v", err) }
    t.Cleanup(func() { _ = m.Close(context.Background()) })
    return m
}

func TestManager_AsyncBasic(t *testing.T) {
    m := newTestManagerAsync(t)
    ctx := context.Background()
    ten := Tenant{UserID: "au", ArchiveID: "aa"}

    // 异步写入 3 条
    for i := 0; i < 3; i++ {
        it := MemoryItem{ID: time.Now().Format("150405.000") + string(rune('a'+i)), Tenant: ten, Content: "v"}
        if err := m.Save(ctx, it, SaveOptions{}); err != nil {
            t.Fatalf("async save %d: %v", i, err)
        }
    }

    // 轮询等待直到可见
    deadline := time.Now().Add(500 * time.Millisecond)
    for {
        qr, err := m.Query(ctx, QueryRequest{Tenant: ten, TopK: 10})
        if err != nil { t.Fatalf("query: %v", err) }
        if len(qr.Items) >= 3 { break }
        if time.Now().After(deadline) {
            t.Fatalf("timeout waiting async items, got %d", len(qr.Items))
        }
        time.Sleep(10 * time.Millisecond)
    }
}

func TestManager_AsyncDegradeToSync_WhenNoChan(t *testing.T) {
    // 人工构造：启用 Async 但不初始化 asyncCh，以触发 select 的 default 分支（同步降级）
    m := &Manager{
        opts: RAGOptions{InMemory: InMemoryOptions{Enable: true}, Async: AsyncOptions{Enable: true}},
        mem:  NewMemoryStore(InMemoryOptions{Enable: true, MaxEntries: 100}),
        // asyncCh: nil
    }
    t.Cleanup(func() { _ = m.Close(context.Background()) })

    ten := Tenant{UserID: "du", ArchiveID: "da"}
    if err := m.Save(context.Background(), MemoryItem{ID: "x", Tenant: ten, Content: "c"}, SaveOptions{}); err != nil {
        t.Fatalf("save (degrade): %v", err)
    }
    qr, err := m.Query(context.Background(), QueryRequest{Tenant: ten, TopK: 10})
    if err != nil { t.Fatalf("query: %v", err) }
    if len(qr.Items) != 1 || qr.Items[0].ID != "x" {
        t.Fatalf("expect immediate visibility via sync degrade, got %+v", qr.Items)
    }
}

func TestManager_AsyncCloseDrainsAndCancelIgnored(t *testing.T) {
    m := newTestManagerAsync(t)
    ten := Tenant{UserID: "cu", ArchiveID: "ca"}

    // 上游已取消的 ctx
    canceled, cancel := context.WithCancel(context.Background())
    cancel()

    if err := m.Save(canceled, MemoryItem{ID: "c1", Tenant: ten, Content: "x"}, SaveOptions{}); err != nil {
        t.Fatalf("save canceled ctx: %v", err)
    }

    // 立即关闭，应等待 worker 处理完队列
    if err := m.Close(context.Background()); err != nil {
        t.Fatalf("close: %v", err)
    }

    // 关闭后查询（内存仍可用），应可见 c1
    qr, err := m.Query(context.Background(), QueryRequest{Tenant: ten, TopK: 10})
    if err != nil { t.Fatalf("query after close: %v", err) }
    if len(qr.Items) != 1 || qr.Items[0].ID != "c1" {
        t.Fatalf("expect c1 visible after close drain, got %+v", qr.Items)
    }
}

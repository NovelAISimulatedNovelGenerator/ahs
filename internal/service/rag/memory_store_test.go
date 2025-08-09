package rag

import (
    "context"
    "testing"
    "time"
)

func TestMemoryStore_SaveQuery_BasicFilters(t *testing.T) {
    st := NewMemoryStore(InMemoryOptions{Enable: true, MaxEntries: 100, TTL: 0})
    ctx := context.Background()

    tenA := Tenant{UserID: "u1", ArchiveID: "a1"}
    now := time.Now()

    // 准备数据（不同 kind/tags/content）
    i1 := MemoryItem{ID: "1", Tenant: tenA, Kind: KindShortTerm, Tags: []string{"t1", "t2"}, Content: "hello world", CreatedAt: now}
    i2 := MemoryItem{ID: "2", Tenant: tenA, Kind: KindLongTerm, Tags: []string{"t1"}, Content: "HELLO NOTE", CreatedAt: now.Add(time.Millisecond)}
    i3 := MemoryItem{ID: "3", Tenant: tenA, Kind: KindFact, Tags: []string{"t2"}, Content: "other content", CreatedAt: now.Add(2 * time.Millisecond)}

    if err := st.Save(ctx, i1); err != nil { t.Fatalf("save i1: %v", err) }
    if err := st.Save(ctx, i2); err != nil { t.Fatalf("save i2: %v", err) }
    if err := st.Save(ctx, i3); err != nil { t.Fatalf("save i3: %v", err) }

    // 过滤：query + tags + kinds
    qr, err := st.Query(ctx, QueryRequest{
        Tenant: tenA,
        Query:  "hello",          // 大小写不敏感
        Tags:   []string{"t1"},   // 必须包含 t1
        Kinds:  []MemoryKind{KindShortTerm, KindLongTerm},
        TopK:   10,
    })
    if err != nil { t.Fatalf("query: %v", err) }
    if len(qr.Items) != 2 { t.Fatalf("expect 2 items, got %d", len(qr.Items)) }
    // 最新在前（i2 比 i1 新）
    if qr.Items[0].ID != "2" || qr.Items[1].ID != "1" { t.Fatalf("order mismatch: %+v", qr.Items) }
}

func TestMemoryStore_TTL_and_Capacity(t *testing.T) {
    // TTL 100ms，容量 2
    st := NewMemoryStore(InMemoryOptions{Enable: true, MaxEntries: 2, TTL: 100 * time.Millisecond})
    ctx := context.Background()

    ten := Tenant{UserID: "u2", ArchiveID: "a2"}
    base := time.Now()

    // 旧数据（将被 TTL 过滤）
    old := MemoryItem{ID: "old", Tenant: ten, Content: "old", CreatedAt: base.Add(-200 * time.Millisecond)}
    // 新数据
    n1 := MemoryItem{ID: "n1", Tenant: ten, Content: "n1", CreatedAt: base}
    n2 := MemoryItem{ID: "n2", Tenant: ten, Content: "n2", CreatedAt: base.Add(1 * time.Millisecond)}

    if err := st.Save(ctx, old); err != nil { t.Fatalf("save old: %v", err) }
    if err := st.Save(ctx, n1); err != nil { t.Fatalf("save n1: %v", err) }
    if err := st.Save(ctx, n2); err != nil { t.Fatalf("save n2: %v", err) }

    // 容量 2，应只保留 n1, n2
    qr, err := st.Query(ctx, QueryRequest{Tenant: ten, TopK: 10})
    if err != nil { t.Fatalf("query: %v", err) }
    if len(qr.Items) != 2 { t.Fatalf("expect 2 items after capacity trim, got %d", len(qr.Items)) }
    if qr.Items[0].ID != "n2" || qr.Items[1].ID != "n1" { t.Fatalf("order/capacity mismatch: %+v", qr.Items) }
}

func TestMemoryStore_TopK_Order(t *testing.T) {
    st := NewMemoryStore(InMemoryOptions{Enable: true, MaxEntries: 100, TTL: 0})
    ctx := context.Background()

    ten := Tenant{UserID: "u3", ArchiveID: "a3"}
    base := time.Now()
    // 写入 5 条
    for i := 0; i < 5; i++ {
        it := MemoryItem{ID: string(rune('a'+i)), Tenant: ten, Content: "c", CreatedAt: base.Add(time.Duration(i) * time.Millisecond)}
        if err := st.Save(ctx, it); err != nil { t.Fatalf("save %d: %v", i, err) }
    }
    // TopK=3，应返回最近的 3 条，且最新在前
    qr, err := st.Query(ctx, QueryRequest{Tenant: ten, TopK: 3})
    if err != nil { t.Fatalf("query: %v", err) }
    if len(qr.Items) != 3 { t.Fatalf("expect 3 items, got %d", len(qr.Items)) }
    if !(qr.Items[0].ID == "e" && qr.Items[1].ID == "d" && qr.Items[2].ID == "c") {
        t.Fatalf("topk order mismatch: %+v", qr.Items)
    }
}

package rag

import (
    "context"
    "path/filepath"
    "testing"
    "time"
)

func TestDiskJSONStore_SaveQuery_BasicFilters(t *testing.T) {
    tmp := t.TempDir()
    stIface, err := NewDiskJSONStore("ns", DiskJSONOptions{Enable: true, RootPath: tmp})
    if err != nil { t.Fatalf("new disk store: %v", err) }
    st := stIface

    ctx := context.Background()
    ten := Tenant{UserID: "u1", ArchiveID: "a1"}

    // 按创建时间递增，便于校验逆序返回
    base := time.Now()
    i1 := MemoryItem{ID: "1", Tenant: ten, Kind: KindShortTerm, Tags: []string{"t1", "t2"}, Content: "hello world", CreatedAt: base}
    i2 := MemoryItem{ID: "2", Tenant: ten, Kind: KindLongTerm, Tags: []string{"t1"}, Content: "HELLO NOTE", CreatedAt: base.Add(1 * time.Millisecond)}
    i3 := MemoryItem{ID: "3", Tenant: ten, Kind: KindFact, Tags: []string{"t2"}, Content: "other content", CreatedAt: base.Add(2 * time.Millisecond)}

    if err := st.Save(ctx, i1); err != nil { t.Fatalf("save i1: %v", err) }
    if err := st.Save(ctx, i2); err != nil { t.Fatalf("save i2: %v", err) }
    if err := st.Save(ctx, i3); err != nil { t.Fatalf("save i3: %v", err) }

    qr, err := st.Query(ctx, QueryRequest{
        Tenant: ten,
        Query:  "hello",
        Tags:   []string{"t1"},
        Kinds:  []MemoryKind{KindShortTerm, KindLongTerm},
        TopK:   10,
    })
    if err != nil { t.Fatalf("query: %v", err) }
    if len(qr.Items) != 2 { t.Fatalf("expect 2 items, got %d", len(qr.Items)) }
    if qr.Items[0].ID != "2" || qr.Items[1].ID != "1" { t.Fatalf("order mismatch: %+v", qr.Items) }
}

func TestDiskJSONStore_MultiTenantIsolation(t *testing.T) {
    tmp := t.TempDir()
    st, err := NewDiskJSONStore("ns", DiskJSONOptions{Enable: true, RootPath: tmp})
    if err != nil { t.Fatalf("new disk store: %v", err) }

    ctx := context.Background()
    tenA := Tenant{UserID: "uA", ArchiveID: "aA"}
    tenB := Tenant{UserID: "uB", ArchiveID: "aB"}

    if err := st.Save(ctx, MemoryItem{ID: "A1", Tenant: tenA, Content: "foo", CreatedAt: time.Now()}); err != nil { t.Fatalf("save A1: %v", err) }
    if err := st.Save(ctx, MemoryItem{ID: "B1", Tenant: tenB, Content: "bar", CreatedAt: time.Now()}); err != nil { t.Fatalf("save B1: %v", err) }

    ra, err := st.Query(ctx, QueryRequest{Tenant: tenA, TopK: 10})
    if err != nil { t.Fatalf("query A: %v", err) }
    if len(ra.Items) != 1 || ra.Items[0].ID != "A1" { t.Fatalf("expect only A1, got %+v", ra.Items) }

    rb, err := st.Query(ctx, QueryRequest{Tenant: tenB, TopK: 10})
    if err != nil { t.Fatalf("query B: %v", err) }
    if len(rb.Items) != 1 || rb.Items[0].ID != "B1" { t.Fatalf("expect only B1, got %+v", rb.Items) }

    // 校验文件路径隔离（不同用户归档路径不同）
    pA := filepath.Clean(filepath.Join(tmp, "ns", safePath(tenA.UserID), safePath(tenA.ArchiveID), "data.jsonl"))
    pB := filepath.Clean(filepath.Join(tmp, "ns", safePath(tenB.UserID), safePath(tenB.ArchiveID), "data.jsonl"))
    if pA == pB { t.Fatalf("tenant files should differ: %s vs %s", pA, pB) }
}

func TestDiskJSONStore_AutoCreatedAt_Expiry_TopK(t *testing.T) {
    tmp := t.TempDir()
    st, err := NewDiskJSONStore("ns", DiskJSONOptions{Enable: true, RootPath: tmp})
    if err != nil { t.Fatalf("new disk store: %v", err) }

    ctx := context.Background()
    ten := Tenant{UserID: "uX", ArchiveID: "aX"}

    // 未设置 CreatedAt，应自动填充
    i1 := MemoryItem{ID: "1", Tenant: ten, Content: "keep"}
    // 过期数据：ExpiresAt 在过去
    past := time.Now().Add(-1 * time.Second)
    i2 := MemoryItem{ID: "2", Tenant: ten, Content: "expired", ExpiresAt: &past}
    // 最新数据
    i3 := MemoryItem{ID: "3", Tenant: ten, Content: "keep2", CreatedAt: time.Now().Add(1 * time.Millisecond)}

    if err := st.Save(ctx, i1); err != nil { t.Fatalf("save i1: %v", err) }
    if err := st.Save(ctx, i2); err != nil { t.Fatalf("save i2: %v", err) }
    if err := st.Save(ctx, i3); err != nil { t.Fatalf("save i3: %v", err) }

    // TopK=1，应只返回最新且未过期的 i3
    qr, err := st.Query(ctx, QueryRequest{Tenant: ten, TopK: 1})
    if err != nil { t.Fatalf("query: %v", err) }
    if len(qr.Items) != 1 || qr.Items[0].ID != "3" { t.Fatalf("expect only latest non-expired item 3, got %+v", qr.Items) }
}

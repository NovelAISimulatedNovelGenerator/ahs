package context

import (
	"context"
)

// TenantKey 是租户信息在 context 中的键类型
type TenantKey struct{}

// Tenant 租户信息结构
type Tenant struct {
	UserID    string `json:"user_id"`
	ArchiveID string `json:"archive_id"`
}

// WithTenant 将租户信息注入到 context 中
func WithTenant(ctx context.Context, userID, archiveID string) context.Context {
	tenant := Tenant{
		UserID:    userID,
		ArchiveID: archiveID,
	}
	return context.WithValue(ctx, TenantKey{}, tenant)
}

// GetTenant 从 context 中获取租户信息
func GetTenant(ctx context.Context) (Tenant, bool) {
	tenant, ok := ctx.Value(TenantKey{}).(Tenant)
	return tenant, ok
}

// GetTenantOrDefault 从 context 中获取租户信息，如果不存在则返回默认值
func GetTenantOrDefault(ctx context.Context, defaultUserID, defaultArchiveID string) Tenant {
	if tenant, ok := GetTenant(ctx); ok {
		return tenant
	}
	return Tenant{
		UserID:    defaultUserID,
		ArchiveID: defaultArchiveID,
	}
}

// MustGetTenant 从 context 中获取租户信息，如果不存在则 panic
func MustGetTenant(ctx context.Context) Tenant {
	tenant, ok := GetTenant(ctx)
	if !ok {
		panic("租户信息未在 context 中找到")
	}
	return tenant
}
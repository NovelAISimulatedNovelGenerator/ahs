package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"ahs/internal/config"
	"ahs/internal/context"
	"go.uber.org/zap"
)


// APIError 标准API错误响应
type APIError struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Details   map[string]string `json:"details,omitempty"`
	Timestamp string            `json:"timestamp"`
}

// Tenant 租户验证中间件
// 从 HTTP 头部提取租户信息并注入到 context 中
// 支持可选的租户验证和路径排除
func Tenant(cfg config.TenantConfig, logger *zap.Logger) Middleware {
	// 编译排除路径的正则表达式
	var excludePatterns []*regexp.Regexp
	for _, pattern := range cfg.ExcludePaths {
		if compiled, err := regexp.Compile(pattern); err == nil {
			excludePatterns = append(excludePatterns, compiled)
		} else if logger != nil {
			logger.Warn("租户中间件：无效的路径排除模式",
				zap.String("pattern", pattern),
				zap.Error(err),
			)
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否为排除路径
			for _, pattern := range excludePatterns {
				if pattern.MatchString(r.URL.Path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 从头部提取租户信息
			userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
			archiveID := strings.TrimSpace(r.Header.Get("X-Archive-ID"))

			// 验证租户信息
			if cfg.Required && (userID == "" || archiveID == "") {
				// 记录验证失败
				if logger != nil {
					logger.Warn("租户验证失败：缺少必需的租户头部",
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.String("remote_addr", r.RemoteAddr),
						zap.String("user_id", userID),
						zap.String("archive_id", archiveID),
					)
				}

				// 返回错误响应
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				
				apiErr := APIError{
					Code:      "MISSING_TENANT_HEADERS",
					Message:   "缺少必需的租户头部信息",
					Details: map[string]string{
						"required_headers": "X-User-ID, X-Archive-ID",
						"missing_headers":  getMissingHeaders(userID, archiveID),
					},
					Timestamp: time.Now().Format(time.RFC3339),
				}
				
				if err := json.NewEncoder(w).Encode(apiErr); err != nil && logger != nil {
					logger.Error("租户中间件：编码错误响应失败", zap.Error(err))
				}
				return
			}

			// 注入租户信息到 context
			var ctx = r.Context()
			if userID != "" && archiveID != "" {
				ctx = context.WithTenant(ctx, userID, archiveID)
				
				// 记录成功的租户注入
				if logger != nil {
					logger.Debug("租户信息已注入到context",
						zap.String("user_id", userID),
						zap.String("archive_id", archiveID),
						zap.String("path", r.URL.Path),
					)
				}
			}

			// 继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// getMissingHeaders 返回缺失的头部信息
func getMissingHeaders(userID, archiveID string) string {
	var missing []string
	if userID == "" {
		missing = append(missing, "X-User-ID")
	}
	if archiveID == "" {
		missing = append(missing, "X-Archive-ID")
	}
	return strings.Join(missing, ", ")
}

// NewDefaultTenantConfig 创建默认的租户配置
func NewDefaultTenantConfig() config.TenantConfig {
	return config.TenantConfig{
		Required:     true,
		ExcludePaths: []string{`^/health$`, `^/api/workflows$`}, // 健康检查和工作流列表不需要租户
	}
}

// NewOptionalTenantConfig 创建可选的租户配置（用于开发环境）
func NewOptionalTenantConfig() config.TenantConfig {
	return config.TenantConfig{
		Required:     false,
		ExcludePaths: []string{`^/health$`},
	}
}
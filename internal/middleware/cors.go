package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"ahs/internal/config"
	"go.uber.org/zap"
)

// ConfigurableCORS 可配置的跨域中间件
// 支持基于配置的域名白名单和安全策略
func ConfigurableCORS(cfg config.CORSConfig, logger *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否启用CORS
			if !cfg.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			origin := r.Header.Get("Origin")
			
			// 验证Origin是否在允许列表中
			allowedOrigin := ""
			if isOriginAllowed(origin, cfg.AllowedOrigins) {
				allowedOrigin = origin
			}

			// 设置CORS头部
			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}

			// 设置允许的方法
			if len(cfg.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
			}

			// 设置允许的头部
			if len(cfg.AllowedHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}

			// 设置暴露的头部
			if len(cfg.ExposedHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
			}

			// 设置是否允许凭证
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// 设置预检请求缓存时间
			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}

			// 处理预检请求(OPTIONS)
			if r.Method == "OPTIONS" {
				if logger != nil {
					logger.Debug("处理CORS预检请求",
						zap.String("origin", origin),
						zap.String("allowed_origin", allowedOrigin),
						zap.String("method", r.Header.Get("Access-Control-Request-Method")),
						zap.String("headers", r.Header.Get("Access-Control-Request-Headers")),
					)
				}
				
				// 如果Origin不被允许，返回403
				if allowedOrigin == "" && origin != "" {
					if logger != nil {
						logger.Warn("CORS预检请求被拒绝：不允许的Origin",
							zap.String("origin", origin),
							zap.Strings("allowed_origins", cfg.AllowedOrigins),
						)
					}
					w.WriteHeader(http.StatusForbidden)
					return
				}
				
				w.WriteHeader(http.StatusOK)
				return
			}

			// 记录跨域请求
			if origin != "" && logger != nil {
				if allowedOrigin != "" {
					logger.Debug("允许的跨域请求",
						zap.String("origin", origin),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
					)
				} else {
					logger.Warn("拒绝的跨域请求：不允许的Origin",
						zap.String("origin", origin),
						zap.String("method", r.Method),
						zap.String("path", r.URL.Path),
						zap.Strings("allowed_origins", cfg.AllowedOrigins),
					)
				}
			}

			// 继续处理请求
			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed 检查Origin是否在允许列表中
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		// 支持完全匹配
		if origin == allowed {
			return true
		}
		
		// 支持通配符 "*" (不推荐在生产环境使用)
		if allowed == "*" {
			return true
		}
		
		// 支持子域名通配符，如 "*.example.com"
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true
			}
		}
	}

	return false
}

// NewDefaultCORSConfig 创建默认的CORS配置（开发环境友好）
func NewDefaultCORSConfig() config.CORSConfig {
	return config.CORSConfig{
		Enabled:          true,
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-User-ID", "X-Archive-ID"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24小时
	}
}

// NewProductionCORSConfig 创建生产环境的CORS配置（安全）
func NewProductionCORSConfig(allowedDomains []string) config.CORSConfig {
	return config.CORSConfig{
		Enabled:          true,
		AllowedOrigins:   allowedDomains, // 必须明确指定允许的域名
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-User-ID", "X-Archive-ID"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           3600, // 1小时
	}
}
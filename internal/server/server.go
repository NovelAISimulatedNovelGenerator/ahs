package server

import (
	"context"
	"fmt"
	"net/http"

	//"time"

	"ahs/internal/config"
	"ahs/internal/handler"
	"ahs/internal/middleware"
	"ahs/internal/service"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Server HTTP服务器
type Server struct {
	config  *config.Config
	logger  *zap.Logger
	handler *handler.Handler
	server  *http.Server
}

// New 创建新的HTTP服务器
func New(cfg *config.Config, logger *zap.Logger, workflowManager service.WorkflowManager) *Server {
	// 创建工作流服务
	workflowService := service.NewWorkflowService(workflowManager)

	// 创建处理器
	h := handler.New(workflowService, logger)

	return &Server{
		config:  cfg,
		logger:  logger,
		handler: h,
	}
}

// Start 启动HTTP服务器
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 注册路由
	s.registerRoutes(mux)

	// 创建中间件链
	middlewareChain := s.createMiddlewareChain()

	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler:        middlewareChain.Then(mux),
		ReadTimeout:    s.config.Server.ReadTimeout,
		WriteTimeout:   s.config.Server.WriteTimeout,
		IdleTimeout:    s.config.Server.IdleTimeout,
		MaxHeaderBytes: s.config.Server.MaxHeaderBytes,
	}

	s.logger.Info("HTTP服务器启动",
		zap.String("地址", s.server.Addr),
		zap.Duration("读取超时", s.config.Server.ReadTimeout),
		zap.Duration("写入超时", s.config.Server.WriteTimeout),
	)

	return s.server.ListenAndServe()
}

// Stop 停止HTTP服务器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("正在关闭HTTP服务器")
	return s.server.Shutdown(ctx)
}

// registerRoutes 注册路由
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// 健康检查
	mux.HandleFunc("/health", s.handler.Health)

	// API路由
	mux.HandleFunc("/api/workflows", s.handler.ListWorkflows)
	mux.HandleFunc("/api/workflows/", s.handler.WorkflowInfo)
	mux.HandleFunc("/api/execute", s.handler.Execute)
	mux.HandleFunc("/api/stream", s.handler.ExecuteStream)
}

// createMiddlewareChain 创建中间件链
func (s *Server) createMiddlewareChain() middleware.Chain {
	var middlewares []middleware.Middleware

	// 恢复中间件 - 必须在最外层
	middlewares = append(middlewares, middleware.Recovery(s.logger))

	// 日志中间件
	middlewares = append(middlewares, middleware.Logger(s.logger))

	// CORS中间件 - 使用配置化的安全CORS策略
	middlewares = append(middlewares, middleware.ConfigurableCORS(s.config.CORS, s.logger))

	// 租户中间件 - 在业务逻辑之前验证租户信息
	middlewares = append(middlewares, middleware.Tenant(s.config.Tenant, s.logger))

	// 限流中间件
	if s.config.RateLimit.Enabled {
		limiter := rate.NewLimiter(rate.Limit(s.config.RateLimit.QPS), s.config.RateLimit.Burst)
		middlewares = append(middlewares, middleware.RateLimit(limiter))
	}

	// 内容类型中间件
	middlewares = append(middlewares, middleware.ContentType("application/json"))

	return middleware.New(middlewares...)
}

package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ahs/internal/config"
	"ahs/internal/server"
	"ahs/internal/workflow"

	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	// 初始化日志
	logger, err := initLogger(cfg)
	if err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("应用启动",
		zap.String("配置文件", configPath),
		zap.Int("服务端口", cfg.Server.Port),
		zap.Bool("限流启用", cfg.RateLimit.Enabled),
		zap.Int("QPS限制", cfg.RateLimit.QPS),
	)

	// 创建工作流管理器
	workflowManager := workflow.NewManager()
	logger.Info("工作流管理器初始化完成",
		zap.Strings("工作流列表", workflowManager.List()),
	)

	// 创建HTTP服务器
	srv := server.New(cfg, logger, workflowManager)

	// 启动服务器
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("服务器关闭失败", zap.Error(err))
	} else {
		logger.Info("服务器已优雅关闭")
	}
}

// initLogger 初始化日志
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	zapConfig := cfg.Log.BuildZapConfig()
	return zapConfig.Build()
}

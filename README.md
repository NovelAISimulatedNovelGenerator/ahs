# Agent HTTP Server

一个高性能、可扩展的Golang HTTP服务框架，专为eino项目提供LLM对外API服务。

## 特性

- 🚀 **高性能**: 支持50QPS峰值负载
- 🔧 **可扩展**: 预留中间件机制，支持功能扩展
- 📊 **限流保护**: 内置请求限流和连接池管理
- 📝 **结构化日志**: 基于zap的高性能日志系统
- ⚙️ **配置管理**: 使用viper进行配置管理
- 🐳 **容器化**: 支持Docker部署
- 🔄 **流式响应**: 支持SSE流式数据传输
- 💊 **健康检查**: 内置健康检查端点

## 快速开始

### 本地运行

1. 克隆项目
```bash
git clone <repository-url>
cd ahs
```

2. 安装依赖
```bash
go mod tidy
```

3. 运行服务
```bash
go run cmd/main.go
```

### Docker运行

1. 构建镜像
```bash
docker build -t ahs .
```

2. 运行容器
```bash
docker run -p 8080:8080 ahs
```

## API接口

### 健康检查
```bash
GET /health
```

### 工作流管理
```bash
# 列出所有工作流
GET /api/workflows

# 获取工作流信息
GET /api/workflows/{name}

# 执行工作流
POST /api/execute
{
  "workflow": "echo",
  "input": "hello world",
  "timeout": 30
}

# 流式执行工作流
POST /api/stream
{
  "workflow": "echo",
  "input": "hello world"
}
```

## 配置说明

配置文件 `config.yaml`:

```yaml
server:
  host: "0.0.0.0"         # 服务绑定地址
  port: 8080              # 服务端口
  read_timeout: 30s       # 读取超时
  write_timeout: 30s      # 写入超时
  idle_timeout: 60s       # 空闲超时
  max_header_bytes: 1048576 # 最大头部大小

rate_limit:
  enabled: true           # 是否启用限流
  qps: 50                # 每秒请求数限制
  burst: 100             # 突发请求限制

log:
  level: "info"          # 日志级别
  encoding: "json"       # 日志编码格式
  output_paths: ["stdout"] # 日志输出路径
  error_output_paths: ["stderr"] # 错误日志路径

worker_pool:
  workers: 8             # 工作器数量
  queue_size: 32         # 队列大小
```

## 架构设计

### 目录结构
```
├── cmd/                 # 主程序入口
├── internal/
│   ├── config/         # 配置管理
│   ├── handler/        # HTTP处理器
│   ├── middleware/     # 中间件系统
│   ├── server/         # HTTP服务器
│   ├── service/        # 业务服务层
│   └── workflow/       # 工作流管理
├── config.yaml         # 配置文件
├── Dockerfile          # Docker构建文件
├── go.mod              # Go模块文件
└── README.md           # 项目文档
```

### 组件说明

- **配置管理**: 使用viper加载和管理配置
- **中间件系统**: 支持日志、限流、CORS、恢复等中间件
- **工作流服务**: 抽象的工作流执行接口
- **HTTP处理器**: RESTful API和SSE流式接口
- **示例工作流**: 包含echo、time、calc等示例工作流

## 开发指南

### 添加新的工作流

1. 实现 `service.WorkflowProcessor` 接口:
```go
type MyProcessor struct{}

func (p *MyProcessor) Process(ctx context.Context, input string) (string, error) {
    // 实现同步处理逻辑
    return "result", nil
}

func (p *MyProcessor) ProcessStream(ctx context.Context, input string, callback service.StreamCallback) error {
    // 实现流式处理逻辑
    callback("data", false, nil)
    callback("final", true, nil)
    return nil
}
```

2. 在 `workflow.Manager` 中注册:
```go
manager.Register("my_workflow", &MyProcessor{})
```

### 添加新的中间件

```go
func MyMiddleware() middleware.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 中间件逻辑
            next.ServeHTTP(w, r)
        })
    }
}
```

## 性能指标

- **支持QPS**: 50 (配置可调)
- **并发连接**: ~10
- **内存使用**: 轻量级设计
- **启动时间**: < 1秒

## 监控和日志

### 日志格式
所有日志均为结构化JSON格式，包含：
- 请求方法和路径
- 响应状态码
- 处理时间
- 错误信息

### 健康检查
访问 `/health` 端点获取服务状态：
```json
{
  "status": "ok",
  "version": "1.0.0",
  "time": "2024-01-01T12:00:00Z",
  "workflows": 3,
  "workflow_names": ["echo", "time", "calc"]
}
```

## 故障排除

### 常见问题

1. **端口被占用**: 修改配置文件中的port设置
2. **内存不足**: 调整worker_pool配置
3. **请求被限流**: 调整rate_limit配置

### 调试模式
将日志级别设置为 `debug` 以获取详细信息：
```yaml
log:
  level: "debug"
```

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 创建Pull Request

## 许可证

MIT License

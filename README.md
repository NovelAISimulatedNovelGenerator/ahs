# agent_http_server (AHS)

一个面向工作流与智能体编排的轻量 HTTP 服务。内置 Eino 工作流适配、SSE 流式输出、结构化日志，以及可插拔的本地/外部 RAG 记忆系统（JSONL 持久化、内存缓存，向量/三元组预留）。

- 代码根路径：`ahs`
- 入口：`cmd/main.go`
- HTTP 服务：`internal/server/` + `internal/handler/`
- 工作流管理：`internal/workflow/`
- RAG 记忆系统：`internal/service/rag/`

> 采用“记忆系统驱动开发（MSDD）”。所有系统知识在代码与 docs/Task.md、README 中外部化，便于新成员快速接入与重置恢复。

---

## 核心特性

- 工作流 HTTP API：
  - 列表/详情/执行（同步与流式 SSE）。
  - 内建示例工作流：`echo`、`time`、`calc`，以及 Eino 集成示例：`agent`、`simple_example`。
- RAG 记忆系统：
  - 进程内存储 + 磁盘 JSONL 持久化，可异步写入、TopK 逆序返回、多租户隔离（`user_id`+`archive_id`）。
  - 预留向量库/三元组外部接口，后续可对接。
- 请求上下文透传：
  - `internal/handler/handler.go` 将原始请求 JSON 放入 `context`（`GetRequestBody(ctx)`）。
- 中间件：
  - 恢复、日志、CORS、限流（`golang.org/x/time/rate`）、默认 `Content-Type`。
- 结构化日志：
  - `zap`，中文字段；级别/编码/输出可配置。

---

## 快速开始

### 依赖

- Go ≥ 1.23（go.mod 指定 `go 1.23`，toolchain `1.24`）

### 本地构建与运行

```bash
# 构建（含 bytedance/sonic 链接规避）
./build.sh

# 或手动
go build -o agent_http_server -ldflags="-checklinkname=0" ./cmd

# 运行（默认读取 config.yaml）
./agent_http_server -config config.yaml
```

### Docker

```bash
# 构建镜像
docker build -t ahs:latest .

# 运行（注意端口映射与 config.yaml 中的端口）
docker run --rm -p 8081:8081 ahs:latest
```

> 注意：`Dockerfile` 当前 `EXPOSE 8080` 与 `HEALTHCHECK` 访问 `8080`，而默认配置 `config.yaml` 为 `8081`。建议统一到同一端口（见文末“问题与确认”）。

---

## 配置（`config.yaml`）

- `server`: `host`/`port`/超时/`max_header_bytes`
- `rate_limit`: `enabled`/`qps`/`burst`
- `log`: `level`/`encoding`/输出路径
- `worker_pool`: `workers`/`queue_size`
- `llm_configs`: 示例（请替换示例 API Key 与模型）

示例片段：见根目录 `config.yaml`。

---

## HTTP API

- 健康检查
  - GET `/health`
  - 响应：`status`、`version`、`time`、`workflows`、`workflow_names`

- 列出工作流
  - GET `/api/workflows`
  - 响应：`{"workflows": [..], "count": N}`

- 获取工作流信息
  - GET `/api/workflows/{name}`
  - 响应：`WorkflowInfo{ name, description, version, status }`

- 执行（同步）
  - POST `/api/execute`
  - 请求体 `WorkflowRequest`：
    - `workflow`(string, 必填)
    - `input`(string)
    - `user_id`(string, 可选)
    - `archive_id`(string, 可选)
    - `timeout`(int, 秒, 可选)
  - 响应 `WorkflowResponse`：`{ status: success|error, result?, error? }`

- 执行（SSE 流式）
  - POST `/api/stream`
  - Header：`Content-Type: text/event-stream`
  - 事件：`data`（分片）、`done`（完成）、`error`（错误）

> 处理器会把原始 JSON 请求体放入 `context`：`handler.GetRequestBody(ctx)`。

---

## 工作流开发

- 代码结构：`internal/workflow/`
  - 管理器：`manager.go`
  - 自动注册：`register.go`（由脚本生成/更新）
  - Eino 实现：`internal/workflow/eino_imp/`

- 生成新工作流

```bash
./scripts/generate_workflow.sh my_workflow "我的工作流描述"
./scripts/scaffold_eino_imp_to_register.sh   # 扫描并写入 register.go
```

> `scaffold_eino_imp_to_register.sh` 会：
> - 仅追加新导入，不覆盖旧导入。
> - 优先识别具有 `Process(context.Context, ...)` 方法的结构体，并注册到 `Manager`。

---

## RAG 记忆系统（`internal/service/rag/`）

- 能力
  - 后端：内存（`memory_store.go`）、磁盘 JSONL（`disk_json_store.go`）。
  - 异步写入队列（`manager.go`：`AsyncOptions`、`worker()`）。
  - 多租户：`Tenant{UserID, ArchiveID}`。
  - 过滤：标签/类型/文本包含、TTL 过期、TopK 逆序。
  - 单例：`rag.Default()`（`service.go`）。

- 使用示例（代码内使用）：

```go
import (
    rag "ahs/internal/service/rag"
    "context"
    "time"
)

func demo(ctx context.Context) {
    m := rag.Default() // 默认 InMemory+DiskJSON+Async
    _ = m.Save(ctx, rag.MemoryItem{
        Tenant:   rag.Tenant{UserID: "u1", ArchiveID: "a1"},
        Kind:     "note",
        Content:  "hello world",
        Tags:     []string{"k1"},
        CreatedAt: time.Now(),
    }, rag.SaveOptions{})

    res, _ := m.Query(ctx, rag.QueryRequest{
        Tenant: rag.Tenant{UserID: "u1", ArchiveID: "a1"},
        Query:  "hello",
        TopK:   5,
    })
    _ = res
}
```

- 测试

```bash
go test -v ./internal/service/rag
# 或仅异步/管理器相关
 go test -v ./internal/service/rag -run TestManager_
```

---

## 目录速览

```
cmd/                      # 程序入口
internal/server/          # HTTP 服务装配、路由与中间件链
internal/handler/         # 具体 HTTP 处理器（含上下文原始请求透传）
internal/workflow/        # 工作流管理器与示例
internal/workflow/eino_imp/ # Eino 工作流实现与示例
internal/service/rag/     # 记忆系统（store、manager、types、tests）
scripts/                  # 代码生成与注册脚本
configs 与 Dockerfile     # 配置与镜像
```

---

## 开发约定（MSDD）

- 所有重要知识外部化：`docs/Task.md`、`README.md`、注释。
- 状态与路径分离记录：
  - 当前状态（进展、活跃任务）。
  - 实现路径（系统模式、技术上下文）。
- 触发更新：
  - 发现新模式、实施重大变更、澄清上下文、定期审查。
- 日志中文、可审计；新成员以记忆系统为唯一参考点。

---

## 常见问题（FAQ）

- bytedance/sonic 链接问题？
  - 使用 `-ldflags="-checklinkname=0"`（见 `build.sh`）。
- SSE 调试方法？
  - 使用 `curl -N -H "Content-Type: application/json" -X POST --data '{"workflow":"echo","input":"hi"}' http://localhost:8081/api/stream`
- API 报 429？
  - 调整 `config.yaml` 的 `rate_limit` 或关闭 `enabled`。

---

## 规划与 TODO（摘）

- [ ] 暴露 RAG 工具化接口（memory_save / memory_query）到工作流工具集。
- [ ] 对接外部向量/三元组检索服务，融合召回与排序。
- [ ] OpenAPI/Swagger 文档与 SDK。
- [ ] 完善权限与多租户校验策略。

---

## 问题与确认（需要你的决定）

1. 端口统一：`Dockerfile` 使用 8080 健康检查与暴露端口，而 `config.yaml` 默认 8081，是否统一到同一端口（建议 8081 或 8080）？
2. Go 版本：`go.mod` 为 Go 1.23 + toolchain 1.24，但 Docker 基础镜像是 `golang:1.21-alpine`，是否升级镜像至 1.23+/1.24 以避免编译不一致？
3. LLM 配置：`config.yaml` 中 `llm_configs.local.api_key` 为示例，是否改为从环境变量读取并在 README 强制说明替换？
4. RAG 对外 API：是否需要在 HTTP 层新增 `POST /api/memory/save` 与 `POST /api/memory/query` 便于工作流外部直接调用？
5. 稳定工作流清单：目前注册了 `agent` 与 `simple_example`，以及示例 `echo/time/calc`，哪些属于对外可见/稳定 API？
6. OpenAPI：是否需要生成 swagger.json 并纳入 CI（用于前端/第三方集成）？

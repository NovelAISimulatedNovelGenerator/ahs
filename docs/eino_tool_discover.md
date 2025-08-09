# Eino Tool 组件用法与集成指南

本文总结 CloudWeGo Eino `components/tool` 的接口与工具化实践，并给出与本项目 `internal/service/rag` 记忆系统的对接范式。

## 核心接口与选项

* __BaseTool__
  - 接口：`Info(ctx) (*schema.ToolInfo, error)`。
  - 作用：向 ChatModel 提供工具元信息（名称、描述、参数 JSONSchema）。

* __InvokableTool__
  - 接口：`InvokableRun(ctx, argumentsInJSON string, opts ...Option) (string, error)`。
  - 用于 ToolsNode 的一次性调用。

* __StreamableTool__
  - 接口：`StreamableRun(ctx, argumentsInJSON string, opts ...Option) (*schema.StreamReader[string], error)`。
  - 用于需要流式输出的工具。

* __Option（统一选项包装）__
  - `tool.Option`：用于在统一签名下透传实现特有的配置。
  - `tool.WrapImplSpecificOptFn[T any](func(*T)) Option`：将具体实现的 option 包装为统一 `Option`。
  - `tool.GetImplSpecificOptions(base *T, opts ...Option) *T`：在工具实现内部提取自定义选项。

## 从函数推断工具（utils 层）

基于泛型与 OpenAPI 反射，将 Go 函数快速包装为 Tool：

* __InferTool / NewTool__（一次性调用）
  - `utils.InferTool[T, D](name, desc string, fn InvokeFunc[T,D], opts ...Option) (tool.InvokableTool, error)`
  - `utils.NewTool[T, D](info *schema.ToolInfo, fn InvokeFunc[T,D], opts ...Option) tool.InvokableTool`

* __InferStreamTool / NewStreamTool__（流式调用）
  - `utils.InferStreamTool[T, D](...) (tool.StreamableTool, error)`
  - `utils.NewStreamTool[T, D](...) tool.StreamableTool`

* __结构与 Schema__
  - `utils.GoStruct2ToolInfo[T](toolName, toolDesc string, opts ...Option) (*schema.ToolInfo, error)`
  - `utils.GoStruct2ParamsOneOf[T](opts ...Option) (*schema.ParamsOneOf, error)`
  - `utils.WithSchemaCustomizer(sc SchemaCustomizerFn)`：自定义结构体标签解析为 OpenAPI 字段。
  - 默认支持标签：
    - `jsonschema:"description=..."`
    - `jsonschema:"enum=..."`（自动按字段类型解析为 string/int/float/bool 等）
    - `jsonschema:"required"` 或通过 `json:"field,omitempty"` 控制必填/可选

* __编解码自定义__
  - `utils.WithUnmarshalArguments(um UnmarshalArguments)`：自定义入参反序列化。
  - `utils.WithMarshalOutput(m MarshalOutput)`：自定义出参序列化（默认 `utils.marshalString` 基于 `bytedance/sonic`）。

## 错误处理包装（utils/error_handler.go）

* __WrapToolWithErrorHandler(t tool.BaseTool, h ErrorHandler) tool.BaseTool__
  - 自动识别 `InvokableTool` / `StreamableTool` 并包裹。
  - 将错误通过 `h(context, err) string` 转为字符串结果返回，避免向上抛错。

* __WrapInvokableToolWithErrorHandler / WrapStreamableToolWithErrorHandler__
  - 分别针对单一类型进行包装。

## 链接清单（来自 Task.md）

源文件（components/tool/ 根目录）：
* callback_extra.go（refactor: move callback template to utils, Dec 26, 2024）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/callback_extra.go
* callback_extra_test.go（feat: manually mirror, Dec 6, 2024）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/callback_extra_test.go
* doc.go（feat: manually mirror, Dec 6, 2024）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/doc.go
* interface.go（fix(lint), Aug 5, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/interface.go
* option.go（feat: manually mirror, Dec 6, 2024）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/option.go
* option_test.go
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/option_test.go

子目录 components/tool/utils/：
* common.go（feat: optimize format, Jun 18, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/common.go
* common_test.go（feat: optimize format, Jun 18, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/common_test.go
* create_options.go（fix: enum parse, May 21, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/create_options.go
* doc.go（feat: manually mirror, Dec 6, 2024）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/doc.go
* error_handler.go（feat: add tool error wrapper, Apr 24, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/error_handler.go
* error_handler_test.go（feat: add tool error wrapper, Apr 24, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/error_handler_test.go
* invokable_func.go（feat: optimize format, Jun 18, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/invokable_func.go
* invokable_func_test.go（fix: enum parse, May 21, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/invokable_func_test.go
* streamable_func.go（feat: optimize format, Jun 18, 2025）
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/streamable_func.go
* streamable_func_test.go
  - https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/utils/streamable_func_test.go

模板：`https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/{$file}`

## 与本项目记忆系统（RAG）集成

入口：`internal/service/rag/manager.go`；全局单例：`internal/service/rag/service.go` 的 `rag.Default()`。

建议将记忆能力以工具形式注册到 Eino Workflow 的 ToolsNode：

* __写入工具（示意）__
  - 请求：`SaveMemoryReq{ Tenant{user_id, archive_id}, Item{kind,tags,content,...}, SaveOptions }`
  - 调用：`rag.Default().Save(ctx, tenant, item, opts)`
  - 封装：`utils.InferTool("memory_save", "保存记忆", func(ctx context.Context, req SaveMemoryReq) (string, error) { /* ... */ })`

* __查询工具（示意）__
  - 请求：`QueryReq{ Tenant, QueryRequest{query,kinds,tags,limit,ttlFilter,...} }`
  - 调用：`rag.Default().Query(ctx, tenant, query)` 返回 `QueryResult`。
  - 封装：`utils.InferTool("memory_query", "检索记忆", func(ctx context.Context, req QueryReq) (any, error) { /* ... */ })`

* __可选：错误包装__
  - `utils.WrapToolWithErrorHandler(tool, func(ctx context.Context, err error) string { return fmt.Sprintf("调用失败: %v", err) })`

## 标签与 Schema 约定

* 必填：缺少 `omitempty` 视为必填；或 `jsonschema:"required"`。
* 枚举：支持 `jsonschema:"enum=..."` 多次出现（自动转为字段类型）。
* 描述：`jsonschema:"description=..."`。
* 编解码：可用 `WithUnmarshalArguments` / `WithMarshalOutput` 定制。

## 最佳实践

* __多租户隔离__：所有工具入参必须包含 `user_id` + `archive_id`，服务端校验。
* __本地优先__：查询先走内存/磁盘，外部向量/三元组服务为占位实现。
* __异步写入__：`rag.Manager` 队列写入；工具返回“已入队/已完成”。
* __保留策略__：当前仅预留接口，后续对接压缩/淘汰。
* __选项透传__：按工具自定义 option struct，使用 `tool.WrapImplSpecificOptFn` / `tool.GetImplSpecificOptions`。


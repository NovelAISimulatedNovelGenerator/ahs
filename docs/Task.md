User: 需要查看的网址：https://github.com/cloudwego/eino/blob/4c89a9c3/compose/graph.go 了解构建RAG系统所需的核心组件抽象

DeepWiki: 根据代码库分析，Eino框架本身没有内置的记忆系统或RAG（检索增强生成）实现，但提供了构建RAG系统所需的核心组件抽象。<cite/>

## RAG相关组件

Eino提供了构建RAG系统的基础组件：

### 检索器组件
框架提供了`Retriever`接口抽象，可以在Chain和Graph中使用： [1](#1-0) 

在Graph系统中也有对应的检索器节点： [2](#1-1) 

### 高级检索器实现
框架在`flow`包中提供了一些高级检索器实现：

1. **MultiQuery检索器** - 支持多查询检索： [3](#1-2) 

2. **Router检索器** - 支持多检索器路由： [4](#1-3) 

### 状态管理和检查点
框架提供了状态管理和检查点功能，可以用于实现记忆系统： [5](#1-4) 

检查点系统支持序列化自定义类型： [6](#1-5) 

## 扩展仓库中的实现

根据文档说明，具体的RAG和记忆系统实现位于扩展仓库中： [7](#1-6) 

## Notes

Eino核心框架提供了构建RAG系统的所有必要抽象（Retriever、Embedding、DocumentLoader等），以及状态管理和检查点功能来支持记忆系统。具体的RAG实现和记忆系统需要查看`cloudwego/eino-ext`扩展仓库。框架的设计允许你通过组合这些组件来构建自定义的RAG和记忆系统。

Wiki pages you might want to explore:
- [Overview (cloudwego/eino)](/wiki/cloudwego/eino#1)
- [Chain Framework (cloudwego/eino)](/wiki/cloudwego/eino#4)

eino-ext 库所包含的：
根据代码库分析，eino-ext 拓展库中确实包含了记忆系统，主要体现在以下几个方面：<cite/>

## 嵌入向量缓存系统

eino-ext 提供了一个完整的嵌入向量缓存系统，用于存储和检索已计算的嵌入向量，避免重复计算： [1](#0-0) 

该系统包含以下核心组件：

- **缓存接口**：定义了 `Cacher` 接口用于存储和检索向量数据 [2](#0-1) 
- **Redis 实现**：提供基于 Redis 的缓存后端实现 [3](#0-2) 
- **缓存嵌入器**：包装现有的嵌入器以提供透明的缓存功能 [4](#0-3) 

## ARK 模型前缀缓存

ARK 聊天模型组件提供了前缀缓存功能，可以在服务器端创建前缀上下文缓存： [5](#0-4) 

这个功能通过 `CreatePrefixCache` 方法实现，能够缓存初始消息作为前缀上下文，提高后续调用的效率： [6](#0-5) 

## 向量数据库存储

Milvus 索引器组件提供了向量存储和检索功能，可以看作是一种持久化的记忆系统： [7](#0-6) 

该系统能够将文档转换为向量并存储在 Milvus 向量数据库中，支持语义搜索和文档检索。 [8](#0-7) 

## Notes

虽然这些组件提供了不同形式的"记忆"功能（缓存、存储、检索），但它们主要专注于性能优化和数据持久化，而不是传统意义上的对话记忆或上下文记忆系统。如果你需要的是对话历史记忆或会话状态管理，可能需要在应用层面结合这些基础组件来实现。

Wiki pages you might want to explore:
- [Embedding Cache System (cloudwego/eino-ext)](/wiki/cloudwego/eino-ext#3.2)
- [Storage and Retrieval (cloudwego/eino-ext)](/wiki/cloudwego/eino-ext#4)

User: 勿重复实现

User: 以下是对于问题的回复:

本地持久化我倾向JSON，较为直观
向量搜索和三元组我希望是外部部署服务，暂时预留接口和框架，待后续开发暂定
多租户需要支持user_id+archive_id同时隔离
记录内容可选，异步写入
需要自动压缩，但是我还没想好相应策略，预留框架接口，待后续开发
我倾向服务+接口，通过http调用接口，分布式系统架构

更新url links:

files: 
callback_extra.go
	
refactor: move callback template to utils
	
Dec 26, 2024
callback_extra_test.go
	
feat: manually mirror eino's code from bytedance
	
Dec 6, 2024
doc.go
	
feat: manually mirror eino's code from bytedance
	
Dec 6, 2024
interface.go
	
fix(lint): remove nolint comment (#380)
	
Aug 5, 2025
option.go
	
feat: manually mirror eino's code from bytedance
	
Dec 6, 2024
option_test.go

directories:
/utils

utils/files:

common.go
	
feat: optimize tool output format with marshalString refactoring (#287)
	
Jun 18, 2025
common_test.go
	
feat: optimize tool output format with marshalString refactoring (#287)
	
Jun 18, 2025
create_options.go
	
fix: defaultSchemaCustomizer now correctly parses enum tag values to …
	
May 21, 2025
doc.go
	
feat: manually mirror eino's code from bytedance
	
Dec 6, 2024
error_handler.go
	
feat: add tool error wrapper (#183)
	
Apr 24, 2025
error_handler_test.go
	
feat: add tool error wrapper (#183)
	
Apr 24, 2025
invokable_func.go
	
feat: optimize tool output format with marshalString refactoring (#287)
	
Jun 18, 2025
invokable_func_test.go
	
fix: defaultSchemaCustomizer now correctly parses enum tag values to …
	
May 21, 2025
streamable_func.go
	
feat: optimize tool output format with marshalString refactoring (#287)
	
Jun 18, 2025
streamable_func_test.go
https://raw.githubusercontent.com/cloudwego/eino/main/components/tool/{$file}


RAG 对外 API：是否需要在 HTTP 层新增 `POST /api/memory/save` 与 `POST /api/memory/query` 便于工作流外部直接调用？

---

# RAG 工具化与 HTTP API 实现计划

## 一、RAG 工具包装实现（internal/workflow/tools/rag_tool/）

### 1.1 文件结构
```
internal/workflow/tools/rag_tool/
├── memory_save.go      # 记忆保存工具
├── memory_query.go     # 记忆查询工具
├── types.go           # 工具专用类型定义
└── tool_test.go       # 单元测试
```

#### 1.1.1 ToolArgumentsHandler 注入租户信息（推荐）

为避免在 Agent 可见的工具参数中暴露敏感租户信息（user_id、archive_id），采用 ToolArgumentsHandler 在工具执行前对参数进行“服务端注入/合并”：

```go
// 伪代码，仅展示核心思路
ToolArgumentsHandler: func(ctx context.Context, name string, arguments string) (string, error) {
    // 1) 从受信任来源获取租户信息（例如：鉴权中间件注入的 ctx 值、请求头、JWT claims 等）
    t, _ := ctx.Value("tenant").(rag.Tenant)

    // 2) 解析 LLM 生成的原始 arguments（其中不包含 user_id/archive_id）
    var m map[string]any
    _ = json.Unmarshal([]byte(arguments), &m)

    // 3) 由后端注入敏感字段，避免经由 Agent 传递
    m["user_id"] = t.UserID
    m["archive_id"] = t.ArchiveID

    // 4) 返回合并后的参数字符串
    b, _ := json.Marshal(m)
    return string(b), nil
}
```

安全性说明：
- 工具入参结构体可包含 `user_id`/`archive_id` 字段，但这些字段由 ToolArgumentsHandler 在服务端注入，LLM 侧无需、也不应提供。
- 这样既避免在提示词/对话中暴露租户信息，又保证工具函数拿到完整参数。

### 1.2 memory_save.go 实现细节
```go
package ragtool

import (
    "context"
    "fmt"
    "time"
    
    "ahs/internal/handler"
    "ahs/internal/service/rag"
    
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
    "github.com/bytedance/sonic"
)

// MemorySaveInput 保存记忆输入参数
type MemorySaveInput struct {
    // 注意：租户字段由 ToolArgumentsHandler 服务端注入，LLM 无需填写
    UserID    string   `json:"user_id" jsonschema:"required,description=用户ID"`
    ArchiveID string   `json:"archive_id" jsonschema:"required,description=档案ID"`
    Content   string   `json:"content" jsonschema:"required,description=记忆内容"`
    Tags      []string `json:"tags,omitempty" jsonschema:"description=标签列表"`
    Kind      string   `json:"kind,omitempty" jsonschema:"description=记忆类型,enum=short_term|long_term|fact|note"`
    TTL       int      `json:"ttl_seconds,omitempty" jsonschema:"description=过期时间(秒)"`
}

// MemorySaveOutput 保存结果
type MemorySaveOutput struct {
    Success bool   `json:"success"`
    ID      string `json:"id,omitempty"`
    Message string `json:"message"`
}

// GetMemorySaveTool 创建记忆保存工具
func GetMemorySaveTool() (tool.InvokableTool, error) {
    // 使用 utils.InferTool 推断工具
    t, err := utils.InferTool(
        "memory_save",
        "保存记忆到系统，支持标签、类型、TTL等",
        memorySaveFunc,
        utils.WithParametersSchema(generateSaveSchema()),
    )
    if err != nil {
        return nil, err
    }
    
    // 包装错误处理
    return utils.WrapToolWithErrorHandler(t, func(ctx context.Context, err error) string {
        return fmt.Sprintf("记忆保存失败: %v", err)
    }).(tool.InvokableTool), nil
}

func memorySaveFunc(ctx context.Context, input *MemorySaveInput) (*MemorySaveOutput, error) {
    // 获取 RAG Manager 单例
    mgr := rag.Default()
    
    // 构建记忆项
    item := rag.MemoryItem{
        Tenant: rag.Tenant{
            UserID:    input.UserID,
            ArchiveID: input.ArchiveID,
        },
        Content:   input.Content,
        Tags:      input.Tags,
        Kind:      rag.MemoryKind(input.Kind),
        CreatedAt: time.Now(),
    }
    
    // 处理 TTL
    if input.TTL > 0 {
        expireAt := time.Now().Add(time.Duration(input.TTL) * time.Second)
        item.ExpiresAt = &expireAt
    }
    
    // 保存（异步/同步由 Manager 内部决定）
    err := mgr.Save(ctx, item, rag.SaveOptions{
        ToMemory: true,
        ToDisk:   true,
    })
    
    if err != nil {
        return &MemorySaveOutput{
            Success: false,
            Message: fmt.Sprintf("保存失败: %v", err),
        }, nil // 错误已转为消息，不再向上抛
    }
    
    return &MemorySaveOutput{
        Success: true,
        ID:      item.ID,
        Message: "记忆保存成功",
    }, nil
}
```

### 1.3 memory_query.go 实现细节
```go
package ragtool

import (
    "context"
    "fmt"
    
    "ahs/internal/service/rag"
    
    "github.com/cloudwego/eino/components/tool"
    "github.com/cloudwego/eino/components/tool/utils"
)

// MemoryQueryInput 查询记忆输入参数
type MemoryQueryInput struct {
    // 注意：租户字段由 ToolArgumentsHandler 服务端注入，LLM 无需填写
    UserID    string   `json:"user_id" jsonschema:"required,description=用户ID"`
    ArchiveID string   `json:"archive_id" jsonschema:"required,description=档案ID"`
    Query     string   `json:"query,omitempty" jsonschema:"description=查询文本"`
    TopK      int      `json:"top_k,omitempty" jsonschema:"description=返回条数,默认10"`
    Tags      []string `json:"tags,omitempty" jsonschema:"description=标签过滤"`
    Kinds     []string `json:"kinds,omitempty" jsonschema:"description=类型过滤"`
}

// MemoryQueryOutput 查询结果
type MemoryQueryOutput struct {
    Success bool              `json:"success"`
    Items   []MemoryItemView  `json:"items"`
    Count   int               `json:"count"`
    Message string            `json:"message"`
}

// MemoryItemView 记忆项视图（简化版）
type MemoryItemView struct {
    ID        string         `json:"id"`
    Content   string         `json:"content"`
    Tags      []string       `json:"tags,omitempty"`
    Kind      string         `json:"kind,omitempty"`
    CreatedAt string         `json:"created_at"`
    Score     float64        `json:"score,omitempty"`
}

// GetMemoryQueryTool 创建记忆查询工具
func GetMemoryQueryTool() (tool.InvokableTool, error) {
    t, err := utils.InferTool(
        "memory_query",
        "查询记忆系统，支持文本搜索、标签过滤、TopK等",
        memoryQueryFunc,
        utils.WithParametersSchema(generateQuerySchema()),
    )
    if err != nil {
        return nil, err
    }
    
    return utils.WrapToolWithErrorHandler(t, func(ctx context.Context, err error) string {
        return fmt.Sprintf("记忆查询失败: %v", err)
    }).(tool.InvokableTool), nil
}

func memoryQueryFunc(ctx context.Context, input *MemoryQueryInput) (*MemoryQueryOutput, error) {
    mgr := rag.Default()
    
    // 构建查询请求
    req := rag.QueryRequest{
        Tenant: rag.Tenant{
            UserID:    input.UserID,
            ArchiveID: input.ArchiveID,
        },
        Query: input.Query,
        TopK:  input.TopK,
        Tags:  input.Tags,
    }

    // 转换 Kinds
    if len(input.Kinds) > 0 {
        req.Kinds = make([]rag.MemoryKind, len(input.Kinds))
        for i, k := range input.Kinds {
            req.Kinds[i] = rag.MemoryKind(k)
        }
    }
    
    // 执行查询
    result, err := mgr.Query(ctx, req)
    if err != nil {
        return &MemoryQueryOutput{
            Success: false,
            Message: fmt.Sprintf("查询失败: %v", err),
        }, nil
    }
    
    // 转换结果
    items := make([]MemoryItemView, len(result.Items))
    for i, item := range result.Items {
        items[i] = MemoryItemView{
            ID:        item.ID,
            Content:   item.Content,
            Tags:      item.Tags,
            Kind:      string(item.Kind),
            CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
            Score:     item.Score,
        }
    }
    
    return &MemoryQueryOutput{
        Success: true,
        Items:   items,
        Count:   len(items),
        Message: fmt.Sprintf("查询成功，返回 %d 条记忆", len(items)),
    }, nil
}
```

### 1.4 工具注册（集成到 register.go）
需要在 `internal/workflow/register.go` 中添加：
```go
import (
    ragtool "ahs/internal/workflow/tools/rag_tool"
)

func init() {
    // 注册 RAG 工具
    if saveTool, err := ragtool.GetMemorySaveTool(); err == nil {
        manager.RegisterTool("memory_save", saveTool)
    }
    if queryTool, err := ragtool.GetMemoryQueryTool(); err == nil {
        manager.RegisterTool("memory_query", queryTool)
    }
}
```

### 1.5 ToolsNode 配置（ToolArgumentsHandler 示例）
在构建 `ToolsNode` 时，通过 `ToolArgumentsHandler` 在服务端注入敏感的多租户字段：
```go
package agent

import (
    "context"
    "encoding/json"

    "ahs/internal/service/rag"
    "github.com/cloudwego/eino/compose"
)

func newToolsConfig(tools []compose.Tool) *compose.ToolsNodeConfig {
    return &compose.ToolsNodeConfig{
        Tools: tools,
        ToolArgumentsHandler: func(ctx context.Context, name string, arguments string) (string, error) {
            // 解析 LLM 生成的原始参数
            var p map[string]any
            if err := json.Unmarshal([]byte(arguments), &p); err != nil {
                return "", fmt.Errorf("解析工具参数失败: %w", err)
            }
            if p == nil { 
                p = map[string]any{} 
            }
            
            // 服务端注入租户信息（从请求头或其他可信源获取）
            p["user_id"] = "default_user"    // 实际应从请求头等获取
            p["archive_id"] = "default_archive" // 实际应从请求头等获取
            
            // 返回有效的 JSON 字符串
            b, err := json.Marshal(p)
            if err != nil {
                return "", fmt.Errorf("序列化工具参数失败: %w", err)
            }
            return string(b), nil
        },
    }
}
```

## 二、HTTP API 实现（internal/handler/memory.go）

### 2.1 Memory Handler 实现（多租户从鉴权上下文/请求头注入）
```go
package handler

import (
    "encoding/json"
    "net/http"
    "time"
    
    "ahs/internal/service/rag"
    "go.uber.org/zap"
)

// MemorySaveRequest HTTP 保存请求（不包含敏感租户字段）
type MemorySaveRequest struct {
    Content   string                 `json:"content"`
    Tags      []string               `json:"tags,omitempty"`
    Kind      string                 `json:"kind,omitempty"`
    TTL       int                    `json:"ttl_seconds,omitempty"`
    Meta      map[string]interface{} `json:"meta,omitempty"`
}

// MemorySaveResponse HTTP 保存响应
type MemorySaveResponse struct {
    Status  string `json:"status"`
    ID      string `json:"id,omitempty"`
    Message string `json:"message,omitempty"`
    Error   string `json:"error,omitempty"`
}

// MemoryQueryRequest HTTP 查询请求（直接复用 rag.QueryRequest）
type MemoryQueryRequest rag.QueryRequest

// MemoryQueryResponse HTTP 查询响应
type MemoryQueryResponse struct {
    Status  string            `json:"status"`
    Items   []rag.MemoryItem  `json:"items,omitempty"`
    Count   int               `json:"count"`
    Message string            `json:"message,omitempty"`
    Error   string            `json:"error,omitempty"`
}

// MemorySave 处理记忆保存请求
func (h *Handler) MemorySave(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // 解析请求
    var req MemorySaveRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("解析保存请求失败", zap.Error(err))
        respondJSON(w, http.StatusBadRequest, MemorySaveResponse{
            Status: "error",
            Error:  "无效的请求格式",
        })
        return
    }
    // 从鉴权上下文/请求头注入多租户
    // 示例：从自定义头获取，实际可由鉴权中间件注入到 ctx
    userID := r.Header.Get("X-User-ID")
    archiveID := r.Header.Get("X-Archive-ID")

    // 验证必填字段
    if userID == "" || archiveID == "" || req.Content == "" {
        respondJSON(w, http.StatusBadRequest, MemorySaveResponse{
            Status: "error",
            Error:  "缺少必填字段：X-User-ID、X-Archive-ID 或 content",
        })
        return
    }
    
    // 构建记忆项
    item := rag.MemoryItem{
        Tenant: rag.Tenant{
            UserID:    userID,
            ArchiveID: archiveID,
        },
        Content:   req.Content,
        Tags:      req.Tags,
        Kind:      rag.MemoryKind(req.Kind),
        CreatedAt: time.Now(),
        Meta:      req.Meta,
    }
    
    // 处理 TTL
    if req.TTL > 0 {
        expireAt := time.Now().Add(time.Duration(req.TTL) * time.Second)
        item.ExpiresAt = &expireAt
    }
    
    // 获取 RAG Manager 并保存
    mgr := rag.Default()
    ctx := r.Context()
    
    err := mgr.Save(ctx, item, rag.SaveOptions{
        ToMemory: true,
        ToDisk:   true,
    })
    
    if err != nil {
        h.logger.Error("保存记忆失败", 
            zap.String("user_id", userID),
            zap.String("archive_id", archiveID),
            zap.Error(err))
        respondJSON(w, http.StatusInternalServerError, MemorySaveResponse{
            Status: "error",
            Error:  "保存失败",
        })
        return
    }
    
    // 成功响应
    h.logger.Info("记忆保存成功",
        zap.String("id", item.ID),
        zap.String("user_id", userID),
        zap.String("archive_id", archiveID))
    
    respondJSON(w, http.StatusOK, MemorySaveResponse{
        Status:  "success",
        ID:      item.ID,
        Message: "记忆保存成功",
    })
}

// MemoryQueryRequest HTTP 查询请求（不包含敏感租户字段）
type MemoryQueryRequest struct {
    Query string   `json:"query,omitempty"`
    TopK  int      `json:"top_k,omitempty"`
    Tags  []string `json:"tags,omitempty"`
    Kinds []string `json:"kinds,omitempty"`
}

// MemoryQuery 处理记忆查询请求
func (h *Handler) MemoryQuery(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    // 解析请求
    var req MemoryQueryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.logger.Error("解析查询请求失败", zap.Error(err))
        respondJSON(w, http.StatusBadRequest, MemoryQueryResponse{
            Status: "error",
            Error:  "无效的请求格式",
        })
        return
    }
    
    // 从鉴权上下文/请求头注入多租户
    userID := r.Header.Get("X-User-ID")
    archiveID := r.Header.Get("X-Archive-ID")

    // 验证必填字段
    if userID == "" || archiveID == "" {
        respondJSON(w, http.StatusBadRequest, MemoryQueryResponse{
            Status: "error",
            Error:  "缺少必填字段：X-User-ID、X-Archive-ID",
        })
        return
    }
    
    // 默认 TopK
    if req.TopK <= 0 {
        req.TopK = 10
    }
    
    // 获取 RAG Manager 并查询
    mgr := rag.Default()
    ctx := r.Context()
    
    // 组装查询请求（服务端合入租户信息）
    query := rag.QueryRequest{
        Tenant: rag.Tenant{UserID: userID, ArchiveID: archiveID},
        Query:  req.Query,
        TopK:   req.TopK,
        Tags:   req.Tags,
    }
    if len(req.Kinds) > 0 {
        query.Kinds = make([]rag.MemoryKind, len(req.Kinds))
        for i, k := range req.Kinds {
            query.Kinds[i] = rag.MemoryKind(k)
        }
    }

    result, err := mgr.Query(ctx, query)
    
    if err != nil {
        h.logger.Error("查询记忆失败",
            zap.String("user_id", userID),
            zap.String("archive_id", archiveID),
            zap.Error(err))
        respondJSON(w, http.StatusInternalServerError, MemoryQueryResponse{
            Status: "error",
            Error:  "查询失败",
        })
        return
    }
    
    // 成功响应
    h.logger.Info("记忆查询成功",
        zap.String("user_id", userID),
        zap.String("archive_id", archiveID),
        zap.Int("count", len(result.Items)))
    
    respondJSON(w, http.StatusOK, MemoryQueryResponse{
        Status:  "success",
        Items:   result.Items,
        Count:   len(result.Items),
        Message: fmt.Sprintf("查询成功，返回 %d 条记忆", len(result.Items)),
    })
}

// respondJSON 辅助函数发送 JSON 响应
func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    json.NewEncoder(w).Encode(payload)
}
```

### 2.2 路由注册（internal/server/server.go）
在 `registerRoutes` 方法中添加：
```go
func (s *Server) registerRoutes(mux *http.ServeMux) {
    // ... 现有路由 ...
    
    // 记忆系统 API
    mux.HandleFunc("/api/memory/save", s.handler.MemorySave)
    mux.HandleFunc("/api/memory/query", s.handler.MemoryQuery)
}
```

## 三、测试计划

### 3.1 工具单元测试（internal/workflow/tools/rag_tool/tool_test.go）
```go
package ragtool_test

import (
    "context"
    "testing"
    "time"
    
    "ahs/internal/service/rag"
    "ahs/internal/workflow/tools/rag_tool"
)

func TestMemorySaveTool(t *testing.T) {
    // 测试用例：
    // 1. 正常保存（有效输入）
    // 2. 缺失必填字段（X-User-ID/X-Archive-ID/content）
    // 3. TTL 过期处理
    // 4. 标签和类型过滤
    // 5. 并发保存
}

func TestMemoryQueryTool(t *testing.T) {
    // 测试用例：
    // 1. 正常查询（文本匹配）
    // 2. 标签过滤查询
    // 3. 类型过滤查询
    // 4. TopK 限制
    // 5. 多租户隔离验证
    // 6. 空查询（返回所有）
}
```

### 3.2 HTTP API 集成测试（internal/handler/memory_test.go）
```go
package handler_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestMemorySaveAPI(t *testing.T) {
    // 测试用例：
    // 1. POST 正常保存
    // 2. 非 POST 方法拒绝
    // 3. 无效 JSON 格式
    // 4. 缺失必填字段
    // 5. TTL 设置验证
    // 6. 响应格式验证
}

func TestMemoryQueryAPI(t *testing.T) {
    // 测试用例：
    // 1. POST 正常查询
    // 2. 分页与 TopK
    // 3. 过滤条件组合
    // 4. 空结果处理
    // 5. 多租户隔离
}
```

## 四、接口文档与使用示例

### 4.1 记忆保存 API
```
POST /api/memory/save
Content-Type: application/json
X-User-ID: user123
X-Archive-ID: session456

请求体:
{
    "content": "用户喜欢吃川菜，特别是麻辣火锅",
    "tags": ["preference", "food"],
    "kind": "long_term",
    "ttl_seconds": 86400,
    "meta": {
        "source": "conversation",
        "confidence": 0.95
    }
}

响应:
{
    "status": "success",
    "id": "mem_abc123",
    "message": "记忆保存成功"
}
```

### 4.2 记忆查询 API
```
POST /api/memory/query
Content-Type: application/json
X-User-ID: user123
X-Archive-ID: session456

请求体:
{
    "query": "用户饮食偏好",
    "top_k": 5,
    "tags": ["preference"],
    "kinds": ["long_term", "fact"]
}

响应:
{
    "status": "success",
    "items": [
        {
            "id": "mem_abc123",
            "tenant": {
                "user_id": "user123",
                "archive_id": "session456"
            },
            "content": "用户喜欢吃川菜，特别是麻辣火锅",
            "tags": ["preference", "food"],
            "kind": "long_term",
            "created_at": "2025-08-09T19:30:00Z",
            "score": 0.92
        }
    ],
    "count": 1,
    "message": "查询成功，返回 1 条记忆"
}
```

## 五、关键实现约束

1. **多租户隔离**：所有操作必须验证 user_id 和 archive_id，避免跨租户数据泄露；且这两个字段不得经由 Agent 可见的工具参数或 HTTP 请求体传递，必须由服务端（ToolArgumentsHandler/鉴权中间件/请求头）注入。

2. **错误处理**：
   - 工具层：使用 `utils.WrapToolWithErrorHandler` 将错误转为中文字符串
   - HTTP 层：统一返回 JSON 格式错误，状态码遵循 RESTful 规范

3. **异步写入**：
   - 工具和 HTTP API 都通过 `rag.Manager.Save()` 保存
   - Manager 内部根据配置决定同步/异步
   - 异步失败不阻塞响应，但需记录日志

4. **性能考虑**：
   - 查询默认 TopK=10，最大不超过 100
   - 支持标签/类型过滤减少扫描范围
   - 本地存储优先，外部服务作为补充

5. **幂等性**：
   - 保存操作通过 ID 字段支持幂等（可选）
   - 查询操作天然幂等

6. **日志规范**：
   - 所有日志使用中文
   - 包含 user_id、archive_id 便于追踪
   - 错误日志包含完整 error 信息

## 六、实施步骤

1. **第一阶段：工具实现**
   - 创建 `internal/workflow/tools/rag_tool/` 目录
   - 实现 `memory_save.go` 和 `memory_query.go`
   - 编写工具单元测试
   - 更新 `register.go` 注册工具

2. **第二阶段：HTTP API**
   - 创建 `internal/handler/memory.go`
   - 实现 `MemorySave` 和 `MemoryQuery` handler
   - 更新 `server.go` 注册路由
   - 编写 HTTP 集成测试

3. **第三阶段：文档与示例**
   - 更新 README.md 添加 API 文档
   - 添加 curl/httpie 调用示例
   - 更新 OpenAPI 规范（如有）

4. **第四阶段：验证与优化**
   - 运行所有单元测试和集成测试
   - 使用 race detector 验证并发安全
   - 压力测试验证异步队列性能
   - 根据测试结果优化参数配置

## 七、依赖关系

- `github.com/cloudwego/eino/components/tool`: 工具接口与 utils
- `github.com/cloudwego/eino/components/tool/utils`: InferTool、错误包装
- `github.com/bytedance/sonic`: JSON 序列化
- `ahs/internal/service/rag`: RAG Manager 和类型定义
- `ahs/internal/handler`: HTTP handler 基础设施
- `go.uber.org/zap`: 结构化日志

## 八、监控与运维

1. **指标收集**：
   - 保存/查询 QPS
   - 响应时间 P50/P95/P99
   - 异步队列长度
   - 错误率

2. **告警规则**：
   - 异步队列满（需扩容或降级）
   - 查询超时（可能数据量过大）
   - 多租户数据异常（可能隔离失效）

3. **容灾方案**：
   - 异步写入失败降级为同步
   - 磁盘满时仅写内存
   - 外部服务不可用时使用本地缓存

---

## Answer

您询问的是 `ToolArgumentsHandler` 的全面揭示。这是 Eino 工具系统中 `ToolsNodeConfig` 的一个重要配置字段，用于在工具执行前预处理工具参数。<cite/>

### ToolArgumentsHandler 定义

`ToolArgumentsHandler` 在 `ToolsNodeConfig` 中定义为一个函数类型： [1](#0-0) 

该处理器接受三个参数：
- `ctx`: 工具调用的上下文
- `name`: 被调用工具的名称  
- `arguments`: 工具的原始参数字符串

返回处理后的参数字符串和可能的错误。<cite/>

### 执行流程中的作用

在 `genToolCallTasks` 方法中，`ToolArgumentsHandler` 会在工具执行前被调用： [2](#0-1) 

如果配置了 `ToolArgumentsHandler`，系统会：
1. 调用处理器处理原始参数
2. 如果处理失败，返回包含工具名称和原始参数的错误信息
3. 如果处理成功，使用处理后的参数替代原始参数

如果未配置处理器，则直接使用原始参数。<cite/>

### 在工具节点中的集成

`ToolArgumentsHandler` 作为 `ToolsNode` 结构体的字段存储： [3](#0-2) 

在创建 `ToolsNode` 时，该处理器从配置中传入： [4](#0-3) 

### 使用场景

`ToolArgumentsHandler` 的典型使用场景包括：
- 参数格式转换和标准化
- 参数验证和清理
- 添加默认值或上下文信息
- 参数加密或脱敏处理

这个机制为工具系统提供了灵活的参数预处理能力，确保工具接收到正确格式的参数。<cite/>

## Notes

`ToolArgumentsHandler` 是 Eino 工具系统架构中的一个可选但重要的组件，它在工具调用链路中起到参数预处理的关键作用。该处理器在 `genToolCallTasks` 阶段执行，确保所有工具都能接收到经过适当处理的参数。

Wiki pages you might want to explore:
- [Tool System (cloudwego/eino)](/wiki/cloudwego/eino#5.1)
- [Creating Tools (cloudwego/eino)](/wiki/cloudwego/eino#5.2)
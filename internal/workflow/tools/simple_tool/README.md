这里写着eino tool框架的研究发现：

创建工具的方法
1. 使用 InferTool 自动推断模式

这是最简单的方式，通过 Go 结构体自动生成参数模式
```go
// 定义输入和输出结构体  
type User struct {  
    Name string `json:"name"`  
}  
  
type UserResult struct {  
    Code int    `json:"code"`  
    Msg  string `json:"msg"`  
}  
  
// 实现工具函数  
func updateUserInfo(ctx context.Context, input *User) (*UserResult, error) {  
    return &UserResult{  
        Code: 200,  
        Msg:  fmt.Sprintf("update %v success", input.Name),  
    }, nil  
}  
  
// 创建工具  
tool, err := utils.InferTool("update_user_info", "更新用户信息", updateUserInfo)
```

2. 使用 InferStreamTool 创建流式工具 
```go
func streamFunction(ctx context.Context, input *InputType) (*schema.StreamReader[OutputType], error) {  
    // 返回流式数据  
}  
  
streamTool, err := utils.InferStreamTool("stream_tool", "流式工具", streamFunction)
```

3. 手动实现工具接口

您也可以直接实现工具接口，如测试代码中的示例
(没有深入研究，DeepWiki自己查)
```go
type mockTool struct{}  
  
func (m *mockTool) Info(ctx context.Context) (*schema.ToolInfo, error) {  
    return &schema.ToolInfo{  
        Name: "mock_tool",  
        Desc: "mock tool",  
        ParamsOneOf: schema.NewParamsOneOfByParams(/* 参数定义 */),  
    }, nil  
}  
  
func (m *mockTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {  
    // 工具执行逻辑  
}
```
核心工具包
```go
import (  
    "context"  
    "github.com/cloudwego/eino/components/tool"  
    "github.com/cloudwego/eino/schema"  
)
```

工具创建工具包
```go
import (  
    "github.com/cloudwego/eino/components/tool/utils"  
)
```


您询问的是 toolcall function 的返回值定义要求以及错误处理机制<cite/>。根据 Eino 框架的代码，工具函数的返回值定义和错误处理有明确的规范。

## 工具函数返回值定义

### InvokableTool 返回值

对于可调用工具，`InvokableRun` 方法必须返回 `(string, error)` [1](#2-0) ：

```go
InvokableRun(ctx context.Context, argumentsInJSON string, opts ...Option) (string, error)
```

### StreamableTool 返回值

对于流式工具，`StreamableRun` 方法必须返回 `(*schema.StreamReader[string], error)` [2](#2-1) ：

```go
StreamableRun(ctx context.Context, argumentsInJSON string, opts ...Option) (*schema.StreamReader[string], error)
```

## 错误处理机制

### 1. 工具执行中的错误处理

在工具执行过程中，错误会被包装并添加工具名称上下文 [3](#2-2) ：

```go
resp, err := i.Fn(ctx, inst, opts...)
if err != nil {
    return "", fmt.Errorf("[LocalFunc] failed to invoke tool, toolName=%s, err=%w", i.getToolName(), err)
}
```

### 2. 参数解析错误处理

JSON 参数解析失败时的错误处理 [4](#2-3) ：

```go
err = sonic.UnmarshalString(arguments, &inst)
if err != nil {
    return "", fmt.Errorf("[LocalFunc] failed to unmarshal arguments in json, toolName=%s, err=%w", i.getToolName(), err)
}
```

### 3. 输出序列化错误处理

返回值序列化失败时的错误处理 [5](#2-4) ：

```go
output, err = marshalString(resp)
if err != nil {
    return "", fmt.Errorf("[LocalFunc] failed to marshal output in json, toolName=%s, err=%w", i.getToolName(), err)
}
```

### 4. 流式工具的错误处理

流式工具在转换过程中的错误处理 [6](#2-5) ：

```go
out, e = marshalString(d)
if e != nil {
    return "", fmt.Errorf("[LocalStreamFunc] failed to marshal output in json, toolName=%s, err=%w", s.getToolName(), e)
}
```

### 5. 特殊错误：InterruptAndRerunErr

工具可以返回特殊的 `InterruptAndRerunErr` 来触发中断和重新运行机制 [7](#2-6) ：

```go
func (m *myTool1) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    if m.times == 0 {
        m.times++
        return "", NewInterruptAndRerunErr("tool1 rerun extra")
    }
    return "tool1 input: " + argumentsInJSON, nil
}
```

### 6. 错误包装器

Eino 还提供了错误处理包装器，可以将工具错误转换为字符串结果而不是返回错误 [8](#2-7) ：

```go
func (s *errorHelper) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    result, err := s.i(ctx, argumentsInJSON, opts...)
    if err != nil {
        return s.h(ctx, err), nil
    }
    return result, nil
}
```

## Notes

工具函数的返回值必须严格遵循接口定义：InvokableTool 返回 `(string, error)`，StreamableTool 返回 `(*schema.StreamReader[string], error)`。错误处理机制包括参数解析错误、执行错误、序列化错误等多个层面，所有错误都会被包装并添加工具名称上下文以便调试。特别地，`InterruptAndRerunErr` 提供了状态恢复机制，而错误包装器可以将错误转换为正常的字符串输出。

Wiki pages you might want to explore:
- [Tool System (cloudwego/eino)](/wiki/cloudwego/eino#5.1)
- [Creating Tools (cloudwego/eino)](/wiki/cloudwego/eino#5.2)


您询问的是 Eino 如何处理工具定义的结构体输入和输出，以及空结构体的解析机制<cite/>。

## 结构体处理机制

Eino 通过反射和 OpenAPI 模式生成来处理工具的结构体输入和输出。核心处理流程在 `goStruct2ParamsOneOf` 函数中： [1](#6-0) 

该函数使用 `generic.NewInstance[T]()` 创建类型 T 的零值实例，然后通过 `openapi3gen.NewSchemaRefForValue` 生成 OpenAPI v3 模式。

## 输入参数解析

在工具执行时，JSON 参数会被解析为对应的结构体类型： [2](#6-1) 

这里使用 `generic.NewInstance[T]()` 创建输入类型的零值实例，然后通过 `sonic.UnmarshalString` 将 JSON 参数解析到该实例中。

## 输出结果序列化

工具的输出结果通过 `marshalString` 函数进行序列化： [3](#6-2) 

## 空结构体的特殊处理

### 1. 零值实例创建

对于空结构体，`generic.NewInstance[T]()` 会创建一个零值实例。在内部序列化系统中，可以看到对空值的处理： [4](#6-3) 

### 2. JSON 解析处理

当输入是空结构体时，JSON 解析仍然会正常工作。即使传入空的 JSON 对象 `{}`，`sonic.UnmarshalString` 也能正确解析到空结构体实例。

### 3. 模式生成

空结构体在 OpenAPI 模式生成时会被识别为对象类型，但没有属性字段。这在 `defaultSchemaCustomizer` 中可以看到相关处理： [5](#6-4) 

## 实际使用示例

在测试代码中可以看到各种结构体类型的处理示例： [6](#6-5) 

这个例子展示了指针类型结构体的处理，对于空结构体也是类似的处理方式。

## 流式工具的结构体处理

流式工具的结构体处理机制与普通工具相同： [7](#6-6) 

## Notes

Eino 的结构体处理机制基于 Go 的反射系统和泛型，能够自动处理各种类型的结构体，包括空结构体。空结构体会被正确解析为零值实例，并在 OpenAPI 模式中表示为没有属性的对象类型。这种设计确保了类型安全性，同时提供了灵活的工具定义方式。

Wiki pages you might want to explore:
- [Creating Tools (cloudwego/eino)](/wiki/cloudwego/eino#5.2)
- [Infrastructure and Utilities (cloudwego/eino)](/wiki/cloudwego/eino#9)
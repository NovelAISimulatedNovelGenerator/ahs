## Answer

您询问的是如何实现一个同时支持不需要处理的输入和 agent 输入的方法。基于 Eino 工具系统的架构，您可以通过组合使用 `ToolArgumentsHandler` 和工具内部的自定义参数处理来实现这个需求。<cite/>

### 方案一：使用 WithUnmarshalArguments 在工具层面处理

您可以在工具创建时使用 `WithUnmarshalArguments` 选项来实现混合参数处理：

```go
func GetMemorySaveTool() (tool.InvokableTool, error) {
    t, err := utils.InferTool(
        "memory_save",
        "保存记忆到系统，支持标签、类型、TTL等",
        memorySaveFunc,
        utils.WithUnmarshalArguments(func(ctx context.Context, arguments string) (interface{}, error) {
            // 解析 agent 输入的参数
            var agentInput MemorySaveInput
            if err := sonic.UnmarshalString(arguments, &agentInput); err != nil {
                return nil, err
            }
            
            // 注入系统级参数（不需要 agent 处理的）
            agentInput.UserID = getUserIDFromContext(ctx)
            agentInput.ArchiveID = getArchiveIDFromContext(ctx)
            
            return &agentInput, nil
        }),
    )
    if err != nil {
        return nil, err
    }
    
    return t, nil
}
``` [1](#16-0) 

### 方案二：结合 ToolArgumentsHandler 和工具选项

您也可以在 ToolsNode 层面使用 `ToolArgumentsHandler` 进行预处理，同时通过工具选项传递额外参数：

```go
func CreateMemoryToolsNode(ctx context.Context) (*compose.ToolsNode, error) {
    memorySaveTool, err := GetMemorySaveTool()
    if err != nil {
        return nil, err
    }
    
    config := &compose.ToolsNodeConfig{
        Tools: []tool.BaseTool{memorySaveTool},
        
        // 只处理需要注入的敏感信息
        ToolArgumentsHandler: func(ctx context.Context, name, arguments string) (string, error) {
            if name == "memory_save" {
                var input map[string]interface{}
                if err := sonic.UnmarshalString(arguments, &input); err != nil {
                    return "", err
                }
                
                // 只注入敏感信息，保留 agent 的原始输入
                input["user_id"] = getUserIDFromContext(ctx)
                input["archive_id"] = getArchiveIDFromContext(ctx)
                
                return sonic.MarshalString(input)
            }
            // 其他工具直接返回原始参数
            return arguments, nil
        },
    }
    
    return compose.NewToolNode(ctx, config)
}
``` [2](#16-1) 

### 方案三：使用工具选项系统传递运行时参数

您还可以通过工具选项系统在运行时传递不需要 agent 处理的参数：

```go
// 定义工具特定的选项
type MemoryToolOptions struct {
    SystemConfig string
    InternalFlag bool
}

func WithMemoryConfig(config string) tool.Option {
    return tool.WrapImplSpecificOptFn(func(o *MemoryToolOptions) {
        o.SystemConfig = config
    })
}

// 在工具实现中获取选项
func memorySaveFunc(ctx context.Context, input *MemorySaveInput) (*MemorySaveOutput, error) {
    // 获取工具选项（不来自 agent）
    opts := tool.GetImplSpecificOptions(&MemoryToolOptions{}, 
        tool.GetImplSpecificOptions(ctx)...)
    
    // 使用 agent 输入和系统选项
    // ...
}
``` [3](#16-2) 

### 参数流程对比

不同方案的参数处理流程：

1. **方案一**：Agent Arguments → Tool WithUnmarshalArguments → Mixed Input Struct → Tool Function
2. **方案二**：Agent Arguments → ToolArgumentsHandler → Tool Interface → Input Struct → Tool Function  
3. **方案三**：Agent Arguments + Tool Options → Tool Interface → Input Struct + Options → Tool Function [4](#16-3) 

## Notes

这三种方案都可以实现您的需求，选择哪种取决于您的具体场景：
- 方案一适合单个工具的复杂参数处理
- 方案二适合系统级的统一参数注入
- 方案三适合运行时动态配置

所有方案都保持了 agent 输入的透明性，同时允许注入系统级参数。

Wiki pages you might want to explore:
- [Tool System (cloudwego/eino)](/wiki/cloudwego/eino#5.1)
- [Creating Tools (cloudwego/eino)](/wiki/cloudwego/eino#5.2)

Wiki pages you might want to explore:
- [Tool System (cloudwego/eino)](/wiki/cloudwego/eino#5.1)
- [Creating Tools (cloudwego/eino)](/wiki/cloudwego/eino#5.2)
- [ReAct Agent (cloudwego/eino)](/wiki/cloudwego/eino#8.1)
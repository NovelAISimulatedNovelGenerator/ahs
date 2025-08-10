基于代码库分析，我来总结Eino框架中Graph系统的四个关键方面：

## 1. Graph的全流程调用机制

Graph的调用机制基于有向图的节点执行模式 [1](#0-0) 。核心流程包括：

**编译阶段**：Graph通过`Compile()`方法将节点和边转换为可执行的运行时 [2](#0-1) 。编译过程中会验证类型匹配、构建数据流通道，并生成`composableRunnable`对象。

**执行阶段**：支持四种执行范式 [3](#0-2) ：
- `Invoke`: 非流式输入→非流式输出
- `Stream`: 非流式输入→流式输出  
- `Collect`: 流式输入→非流式输出
- `Transform`: 流式输入→流式输出

**节点触发模式**：支持两种触发模式 [4](#0-3) ：
- `runTypePregel`: 默认模式，任意前驱节点完成即可触发
- `runTypeDAG`: 所有前驱节点完成才触发

## 2. Graph如何分支

Graph通过`AddBranch()`方法实现条件分支 [5](#0-4) 。分支机制包括：

**分支类型**：
- `GraphBranch`: 基于输入条件决定下一个执行节点
- `StreamGraphBranch`: 支持流式数据的分支判断 [6](#0-5) 

**分支逻辑**：分支函数接收当前节点输出，返回目标节点名称 [7](#0-6) 。框架支持复杂的嵌套分支和交叉分支模式 [8](#0-7) 。

## 3. Graph如何使用工具

Graph通过`AddToolsNode()`集成工具调用能力 [9](#0-8) ：

**ToolsNode配置**：支持多种工具类型和执行模式 [10](#0-9) ：
- 并行或顺序执行工具调用
- 未知工具处理器处理LLM幻觉
- 工具参数预处理

**工具集成示例**：在Graph中添加工具节点并连接到聊天模型 [11](#0-10) 。工具节点可以处理来自ChatModel的工具调用请求，执行相应工具并返回结果。

**ReAct模式**：框架提供完整的ReAct Agent实现 [12](#0-11) ，展示了ChatModel和ToolsNode的循环调用模式。

## 4. Graph如何嵌入parallel和chain以及互相嵌入

**Chain嵌入Graph**：Chain本质上是Graph的特殊形式 [13](#0-12) 。Chain通过`AppendGraph()`方法嵌入其他Graph [14](#0-13) 。

**Graph嵌入Graph**：通过`AddGraphNode()`实现Graph的嵌套 [15](#0-14) 。子Graph作为单个节点运行，支持独立的编译选项 [16](#0-15) 。

**Parallel集成**：Chain支持通过`AppendParallel()`添加并行执行块 [17](#0-16) 。Parallel内部可以嵌套Chain和Graph [18](#0-17) 。

**实际应用示例**：MultiAgent系统展示了复杂的嵌套模式 [19](#0-18) ，其中MultiAgent作为子图嵌入到更大的Chain中，实现了多层次的组合。

## Notes

所有这些组合模式都基于统一的`compose.Runnable`接口，确保了类型安全和流式处理的一致性。Graph系统的设计允许任意深度的嵌套和组合，为构建复杂的AI应用提供了强大的编排能力。

Wiki pages you might want to explore:
- [Overview (cloudwego/eino)](/wiki/cloudwego/eino#1)
- [Agent Implementations (cloudwego/eino)](/wiki/cloudwego/eino#8)
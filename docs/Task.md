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
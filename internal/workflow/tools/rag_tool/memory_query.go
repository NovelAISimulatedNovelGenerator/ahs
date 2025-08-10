package ragtool

import (
	"ahs/internal/service/rag"
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GetMemoryQueryTool 创建记忆查询工具
func GetMemoryQueryTool() (tool.InvokableTool, error) {
	t, err := utils.InferTool(
		"memory_query",
		"查询记忆系统，支持文本搜索、标签过滤、TopK等",
		memoryQueryFunc,
		/*
			WithUnmarshalArguments 中的匿名函数会在工具运行时被调用，
			此时会接收到包含租户信息的运行时上下文，
			因此 parseTenantFromContext(ctx) 能够正常工作。
			这种设计允许在工具创建时配置参数处理逻辑，而在运行时动态注入上下文相关的参数。
		*/
		utils.WithUnmarshalArguments(func(ctx context.Context, arguments string) (interface{}, error) {
			// 解析 agent 输入的参数
			var agentInput MemoryQueryInput
			if err := sonic.UnmarshalString(arguments, &agentInput); err != nil {
				return nil, fmt.Errorf("参数解析失败: %w", err)
			}

			// 从上下文注入租户信息
			userID, archiveID, err := parseTenantFromContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("租户信息解析失败: %w", err)
			}

			agentInput.UserID = userID
			agentInput.ArchiveID = archiveID

			// 设置默认值
			if agentInput.TopK <= 0 {
				agentInput.TopK = 10
			}

			return &agentInput, nil
		}),
	)
	if err != nil {
		return nil, err
	}

	// 包装错误处理
	return utils.WrapToolWithErrorHandler(t, func(ctx context.Context, err error) string {
		return fmt.Sprintf("记忆查询失败: %v", err)
	}).(tool.InvokableTool), nil
}

func memoryQueryFunc(ctx context.Context, input *MemoryQueryInput) (*MemoryQueryOutput, error) {
	// 获取 RAG Manager 单例
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
		}, nil // 错误已转为消息，不再向上抛
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

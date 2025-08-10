package ragtool

import (
	"context"
	"fmt"
	"time"

	"ahs/internal/handler"
	"ahs/internal/service/rag"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// 创建工具
func GetMemorySaveTool() (tool.InvokableTool, error) {
	t, err := utils.InferTool(
		"memory_save",
		"保存记忆到系统，支持标签、类型、TTL等",
		memorySaveFunc,
		utils.WithUnmarshalArguments(func(ctx context.Context, arguments string) (interface{}, error) {
			// 解析 agent 输入的参数
			var agentInput MemorySaveInput
			if err := sonic.UnmarshalString(arguments, &agentInput); err != nil {
				return nil, fmt.Errorf("参数解析失败: %w", err)
			}

			// 注入系统级参数（不需要 agent 处理的）
			userID, archiveID, err := parseTenantFromContext(ctx)
			if err != nil {
				return nil, fmt.Errorf("参数解析失败: %w", err)
			}
			agentInput.UserID = userID
			agentInput.ArchiveID = archiveID

			return &agentInput, nil
		}),
	)
	if err != nil {
		return nil, err
	}

	return utils.WrapToolWithErrorHandler(t, func(ctx context.Context, err error) string {
		return fmt.Sprintf("记忆保存失败: %v", err)
	}).(tool.InvokableTool), nil
}

func parseTenantFromContext(ctx context.Context) (string, string, error) {
	// 从上下文中获取原始请求体并解析 user_id / archive_id
	b, ok := handler.GetRequestBody(ctx)
	if !ok || len(b) == 0 {
		return "", "", fmt.Errorf("上下文中没有原始请求体")
	}
	var payload struct {
		UserID    string `json:"user_id"`
		ArchiveID string `json:"archive_id"`
	}
	if err := sonic.Unmarshal(b, &payload); err != nil {
		return "", "", fmt.Errorf("解析上下文请求体失败: %w", err)
	}
	return payload.UserID, payload.ArchiveID, nil
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

package simpletool

import (
	"context"
	"fmt"

	"ahs/internal/handler"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/bytedance/sonic"  
)

const (
	toolName = "simple_extractor"
	toolDesc = "从请求中提取user_id和archive_id的工具"
	userId = "user_id"
	archiveId = "archive_id"
)

type SimpleToolInput struct {
}

// ResponseData 定义工具的输出数据结构
type ResponseData struct {
	UserID    string `json:"user_id"`
	ArchiveID string `json:"archive_id"`
	Message   string `json:"message"`
}

// GetTool 创建并返回工具实例
func GetTool() (tool.InvokableTool, error) {
	return utils.InferTool(toolName, toolDesc, SimpleTool)
}

// SimpleTool 实际的工具函数
func SimpleTool(ctx context.Context, _ *SimpleToolInput) (*ResponseData, error) {
	// 从context获取原始请求body（可选，用于调试或额外验证）
	body, ok := handler.GetRequestBody(ctx)
	if !ok {
		return nil, fmt.Errorf("无法从context获取请求body")
	}

	// 记录原始请求用于调试
	fmt.Printf("原始请求: %s\n", string(body))

	var req map[string]interface{}
	if err := sonic.UnmarshalString(string(body), &req); err != nil {
		return nil, fmt.Errorf("unmarshal request body failed: %w", err)
	}

	if req[userId] == "" || req[archiveId] == "" {
		return nil, fmt.Errorf("user_id和archive_id不能为空")
	}
	
	// 类型断言：将interface{}转换为string
	userID, ok1 := req[userId].(string)
	archiveID, ok2 := req[archiveId].(string)
	
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("user_id或archive_id类型错误，期望string类型")
	}

	// 构建响应
	return &ResponseData{
		UserID:    userID,
		ArchiveID: archiveID,
		Message:   fmt.Sprintf("成功提取数据 - 用户ID: %s, 档案ID: %s", userID, archiveID),
	}, nil
}

package service

import (
	"context"
	"errors"
)

var (
	ErrWorkflowNotFound = errors.New("工作流未找到")
	ErrInvalidRequest   = errors.New("无效的请求")
)

// WorkflowRequest 工作流请求结构
type WorkflowRequest struct {
	Workflow  string `json:"workflow"`
	Input     string `json:"input"`
	UserID    string `json:"user_id,omitempty"`
	ArchiveID string `json:"archive_id,omitempty"`
	Timeout   int    `json:"timeout,omitempty"`
}

// WorkflowResponse 工作流响应结构
type WorkflowResponse struct {
	Status string `json:"status"`
	Result string `json:"result"`
	Error  string `json:"error,omitempty"`
}

// WorkflowInfo 工作流信息结构
type WorkflowInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Status      string `json:"status"`
}

// StreamCallback 流式处理回调函数类型
type StreamCallback func(data string, done bool, err error)

// WorkflowManager 工作流管理器接口
type WorkflowManager interface {
	List() []string
	Get(name string) (WorkflowProcessor, bool)
	GetInfo(name string) (*WorkflowInfo, error)
}

// WorkflowProcessor 工作流处理器接口
type WorkflowProcessor interface {
	Process(ctx context.Context, input string) (string, error)
	ProcessStream(ctx context.Context, input string, callback StreamCallback) error
}

// WorkflowService 工作流服务
type WorkflowService struct {
	manager WorkflowManager
}

// NewWorkflowService 创建工作流服务
func NewWorkflowService(manager WorkflowManager) *WorkflowService {
	return &WorkflowService{
		manager: manager,
	}
}

// ListWorkflows 列出所有工作流
func (s *WorkflowService) ListWorkflows() []string {
	return s.manager.List()
}

// GetWorkflowInfo 获取工作流信息
func (s *WorkflowService) GetWorkflowInfo(name string) (*WorkflowInfo, error) {
	return s.manager.GetInfo(name)
}

// Execute 执行工作流
func (s *WorkflowService) Execute(ctx context.Context, req WorkflowRequest) (*WorkflowResponse, error) {
	// 验证请求
	if req.Workflow == "" {
		return nil, ErrInvalidRequest
	}

	// 获取工作流处理器
	processor, ok := s.manager.Get(req.Workflow)
	if !ok {
		return nil, ErrWorkflowNotFound
	}

	// 处理请求
	result, err := processor.Process(ctx, req.Input)
	if err != nil {
		return &WorkflowResponse{
			Status: "error",
			Error:  err.Error(),
		}, nil
	}

	return &WorkflowResponse{
		Status: "success",
		Result: result,
	}, nil
}

// ExecuteStream 执行流式工作流
func (s *WorkflowService) ExecuteStream(ctx context.Context, req WorkflowRequest, callback StreamCallback) error {
	// 验证请求
	if req.Workflow == "" {
		return ErrInvalidRequest
	}

	// 获取工作流处理器
	processor, ok := s.manager.Get(req.Workflow)
	if !ok {
		return ErrWorkflowNotFound
	}

	// 执行流式处理
	return processor.ProcessStream(ctx, req.Input, callback)
}

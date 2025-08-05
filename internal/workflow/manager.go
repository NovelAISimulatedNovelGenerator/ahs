package workflow

import (
	"context"
	"fmt"
	"time"

	"ahs/internal/service"
)

// Manager 示例工作流管理器
type Manager struct {
	workflows map[string]service.WorkflowProcessor
}

// NewManager 创建新的工作流管理器
func NewManager() *Manager {
	m := &Manager{
		workflows: make(map[string]service.WorkflowProcessor),
	}

	// 注册示例工作流
	m.registerExampleWorkflows()

	return m
}

// List 列出所有工作流
func (m *Manager) List() []string {
	var names []string
	for name := range m.workflows {
		names = append(names, name)
	}
	return names
}

// Get 获取工作流处理器
func (m *Manager) Get(name string) (service.WorkflowProcessor, bool) {
	processor, ok := m.workflows[name]
	return processor, ok
}

// GetInfo 获取工作流信息
func (m *Manager) GetInfo(name string) (*service.WorkflowInfo, error) {
	_, ok := m.workflows[name]
	if !ok {
		return nil, service.ErrWorkflowNotFound
	}

	// 返回示例信息
	return &service.WorkflowInfo{
		Name:        name,
		Description: fmt.Sprintf("示例工作流: %s", name),
		Version:     "1.0.0",
		Status:      "active",
	}, nil
}

// Register 注册工作流处理器
func (m *Manager) Register(name string, processor service.WorkflowProcessor) {
	m.workflows[name] = processor
}

// registerExampleWorkflows 注册示例工作流
func (m *Manager) registerExampleWorkflows() {
	// 注册简单的回声工作流
	m.Register("echo", &EchoProcessor{})

	// 注册时间工作流
	m.Register("time", &TimeProcessor{})

	// 注册计算工作流
	m.Register("calc", &CalcProcessor{})

}

// EchoProcessor 回声处理器
type EchoProcessor struct{}

func (p *EchoProcessor) Process(ctx context.Context, input string) (string, error) {
	return fmt.Sprintf("回声: %s", input), nil
}

func (p *EchoProcessor) ProcessStream(ctx context.Context, input string, callback service.StreamCallback) error {
	// 模拟流式响应
	words := []string{"回", "声", ":", " ", input}

	for i, word := range words {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(100 * time.Millisecond) // 模拟处理时间
			callback(word, i == len(words)-1, nil)
		}
	}

	return nil
}

// TimeProcessor 时间处理器
type TimeProcessor struct{}

func (p *TimeProcessor) Process(ctx context.Context, input string) (string, error) {
	now := time.Now()
	return fmt.Sprintf("当前时间: %s", now.Format("2006-01-02 15:04:05")), nil
}

func (p *TimeProcessor) ProcessStream(ctx context.Context, input string, callback service.StreamCallback) error {
	// 流式输出当前时间
	now := time.Now()
	timeStr := now.Format("2006-01-02 15:04:05")

	for i, char := range timeStr {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(50 * time.Millisecond)
			callback(string(char), i == len(timeStr)-1, nil)
		}
	}

	return nil
}

// CalcProcessor 计算处理器
type CalcProcessor struct{}

func (p *CalcProcessor) Process(ctx context.Context, input string) (string, error) {
	// 简单的计算示例
	if input == "" {
		return "请提供计算表达式", nil
	}

	// 这里可以集成实际的计算逻辑
	return fmt.Sprintf("计算结果: %s = 42", input), nil
}

func (p *CalcProcessor) ProcessStream(ctx context.Context, input string, callback service.StreamCallback) error {
	// 模拟计算过程
	steps := []string{
		"开始计算...",
		"解析表达式...",
		"执行计算...",
		fmt.Sprintf("计算结果: %s = 42", input),
	}

	for i, step := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			time.Sleep(200 * time.Millisecond)
			callback(step, i == len(steps)-1, nil)
		}
	}

	return nil
}

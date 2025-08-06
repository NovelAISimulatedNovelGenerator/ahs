package agent

import (
	"context"
	"fmt"
	"io"

	"ahs/internal/service"
)

// agentConfig 配置结构体
type agentConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

// agentProcessor workflow 处理器
type agentProcessor struct {
	config    *agentConfig
	ctx       context.Context
	userInput string
}

// NewagentProcessor 创建新的处理器实例
func NewagentProcessorWithDefaults() *agentProcessor {
	return &agentProcessor{
		config: &agentConfig{
			APIKey:  "your-api-key-here",
			Model:   "gpt-3.5-turbo",
			BaseURL: "https://api.openai.com/v1",
		},
	}
}

// Process 执行工作流
func (p *agentProcessor) Process(ctx context.Context, input string) (string, error) {
	p.userInput = input
	p.ctx = ctx
	
	// 构建图
	graph, err := p.buildGraph()
	if err != nil {
		return "", fmt.Errorf("build graph failed: %w", err)
	}
	
	// 创建可运行实例
	runnable, err := graph.Compile(ctx)
	if err != nil {
		return "", fmt.Errorf("compile graph failed: %w", err)
	}
	
	// 执行图
	result, err := runnable.Invoke(ctx, map[string]any{
		"input": input,
	})
	if err != nil {
		return "", fmt.Errorf("invoke graph failed: %w", err)
	}
	
	if result == nil {
		return "", nil
	}
	
	return result.Content, nil
}

// ProcessStream 流式执行工作流
func (p *agentProcessor) ProcessStream(
	ctx context.Context,
	input string,
	callback service.StreamCallback,
) error {
	p.userInput = input
	p.ctx = ctx
	
	// 构建图
	graph, err := p.buildGraph()
	if err != nil {
		return fmt.Errorf("build graph failed: %w", err)
	}
	
	// 创建可运行实例
	runnable, err := graph.Compile(ctx)
	if err != nil {
		return fmt.Errorf("compile graph failed: %w", err)
	}
	
	// 流式执行
	stream, err := runnable.Stream(ctx, map[string]any{
		"input": input,
	})
	if err != nil {
		return fmt.Errorf("stream graph failed: %w", err)
	}
	
	defer stream.Close()
	
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("receive stream failed: %w", err)
		}
		
		if chunk != nil {
			callback(chunk.Content, false, nil)
		}
	}
	
	callback("", true, nil)
	return nil
}

// SetConfig 设置配置
func (p *agentProcessor) SetConfig(config *agentConfig) {
	p.config = config
}

// GetConfig 获取当前配置
func (p *agentProcessor) GetConfig() *agentConfig {
	return p.config
}

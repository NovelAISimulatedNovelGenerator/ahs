package agent

import (
	"context"
	"fmt"
	"io"

	"ahs/internal/config"
	"ahs/internal/handler"
	"ahs/internal/service"
)

// agentConfig 配置结构体
type agentConfig struct {
	APIKey  string
	Model   string
	BaseURL string
}

// AgentProcessor workflow 处理器
type AgentProcessor struct {
	config    *agentConfig
	ctx       context.Context
	userInput string
}

// NewAgentProcessor 创建新的处理器实例，使用统一配置系统
func NewAgentProcessorWithDefaults() *AgentProcessor {
	return &AgentProcessor{
		config: loadAgentConfig(),
	}
}

// Process 执行工作流
func (p *AgentProcessor) Process(ctx context.Context, input string) (string, error) {
	p.userInput = input
	p.ctx = ctx

	if _, ok := handler.GetRequestBody(ctx); !ok {
		return "", fmt.Errorf("no request body")
	}

	// 确保配置已初始化
	if p.config == nil {
		p.config = loadAgentConfig()
	}

	// 构建图
	graph, err := p.buildGraph(ctx)
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
func (p *AgentProcessor) ProcessStream(
	ctx context.Context,
	input string,
	callback service.StreamCallback,
) error {
	p.userInput = input
	p.ctx = ctx

	// 确保配置已初始化
	if p.config == nil {
		p.config = loadAgentConfig()
	}

	// 构建图
	graph, err := p.buildGraph(ctx)
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
func (p *AgentProcessor) SetConfig(config *agentConfig) {
	p.config = config
}

// GetConfig 获取当前配置
func (p *AgentProcessor) GetConfig() *agentConfig {
	return p.config
}

// loadAgentConfig 从统一配置系统加载agent配置
func loadAgentConfig() *agentConfig {
	globalConfig := config.GetGlobalConfig()
	if globalConfig == nil || globalConfig.LLMConfigs == nil {
		// 如果配置未加载，使用默认值
		return &agentConfig{
			APIKey:  "your-api-key-here",
			Model:   "gpt-3.5-turbo",
			BaseURL: "https://api.openai.com/v1",
		}
	}

	// 使用 local 配置
	if localCfg, ok := globalConfig.LLMConfigs["local"]; ok {
		return &agentConfig{
			APIKey:  localCfg.APIKey,
			Model:   localCfg.Model,
			BaseURL: localCfg.APIBaseURL,
		}
	}

	// 如果没有 local 配置，使用第一个可用的配置
	for _, c := range globalConfig.LLMConfigs {
		return &agentConfig{
			APIKey:  c.APIKey,
			Model:   c.Model,
			BaseURL: c.APIBaseURL,
		}
	}

	// 确保配置不为空
	return &agentConfig{
		APIKey:  "your-api-key-here",
		Model:   "gpt-3.5-turbo",
		BaseURL: "https://api.openai.com/v1",
	}
}

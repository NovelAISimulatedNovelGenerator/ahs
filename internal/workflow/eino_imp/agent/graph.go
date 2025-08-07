package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

/*
用户输入  -> 选择调用工具 -> 记入记忆 -> 输出
*/

const (
	nodePrompt    = "prompt"
	nodeChatModel = "chat_model"
	nodeTools     = "tools"

	systemMessage = ``
)

func (p *AgentProcessor) buildGraph() (*compose.Graph[map[string]any, *schema.Message], error) {
	g := compose.NewGraph[map[string]any, *schema.Message]()

	// 添加提示模板节点
	if err := g.AddChatTemplateNode(nodePrompt, p.newChatTemplate()); err != nil {
		return nil, fmt.Errorf("add prompt node failed: %w", err)
	}

	// 添加聊天模型节点
	chatModel, err := p.newChatModel(p.ctx)
	if err != nil {
		return nil, fmt.Errorf("create chat model failed: %w", err)
	}
	if err := g.AddChatModelNode(nodeChatModel, chatModel); err != nil {
		return nil, fmt.Errorf("add chat model node failed: %w", err)
	}

	// 添加工具节点 - 暂时留空，后续添加工具配置
	// if err := g.AddToolsNode(nodeTools, p.newToolsConfig()); err != nil {
	// 	return nil, fmt.Errorf("add tools node failed: %w", err)
	// }

	// 连接节点
	if err := g.AddEdge(compose.START, nodePrompt); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodePrompt, nodeChatModel); err != nil {
		return nil, err
	}
	/*if err := g.AddEdge(nodeChatModel, nodeTools); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeTools, nodeChatModel); err != nil {
		return nil, err
	}*/

	if err := g.AddEdge(nodeChatModel, compose.END); err != nil {
		return nil, err
	}

	return g, nil
}

func (p *AgentProcessor) newChatTemplate() prompt.ChatTemplate {
	return prompt.FromMessages(
		schema.FString,
		schema.SystemMessage(systemMessage),
		schema.UserMessage(p.userInput),
	)
}

func (p *AgentProcessor) newChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	var temp float32 = 0
	cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:      p.config.APIKey,
		BaseURL:     p.config.BaseURL,
		Model:       p.config.Model,
		Temperature: &temp,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model failed: %w", err)
	}
	return cm, nil
}

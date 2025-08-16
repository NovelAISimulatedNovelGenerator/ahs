package agent

import (
	"context"
	"fmt"

	pb "ahs/internal/service/prompt_builder"
	rt "ahs/internal/workflow/tools/rag_tool"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/components/tool"
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
)

// 本地状态：用于保留会话历史，确保在工具调用往返后
// 再次调用 ChatModel 时，历史中包含上一条包含 tool_calls 的助手消息，
// 以避免 OpenAI 对话接口出现 tool_call_id 不匹配错误。
type agentState struct {
	History []*schema.Message
}

func (p *AgentProcessor) buildGraph(ctx context.Context) (*compose.Graph[map[string]any, *schema.Message], error) {
	g := compose.NewGraph[map[string]any, *schema.Message](
		compose.WithGenLocalState(func(ctx context.Context) *agentState {
			return &agentState{History: []*schema.Message{}}
		}),
	)

	// 添加提示模板节点
	if err := g.AddChatTemplateNode(nodePrompt, p.newChatTemplate()); err != nil {
		return nil, fmt.Errorf("add prompt node failed: %w", err)
	}

	mst, err := rt.GetMemorySaveTool()
	if err != nil {
		return nil, fmt.Errorf("create memory save tool failed: %w", err)
	}
	mqt, err := rt.GetMemoryQueryTool()
	if err != nil {
		return nil, fmt.Errorf("create memory query tool failed: %w", err)
	}

	// 绑定工具到 ChatModel
	toolsList := []tool.BaseTool{mst, mqt}
	infos := make([]*schema.ToolInfo, 0, len(toolsList))
	for _, t := range toolsList {
		info, err := t.Info(ctx)
		if err != nil {
			return nil, fmt.Errorf("get tool info failed: %w", err)
		}
		infos = append(infos, info)
	}

	// 创建聊天模型并绑定工具
	chatModel, err := p.newChatModel(ctx)
	if err != nil {
		return nil, fmt.Errorf("create chat model failed: %w", err)
	}
	chatModel, err = chatModel.WithTools(infos)
	if err != nil {
		return nil, fmt.Errorf("bind tools to chat model failed: %w", err)
	}
	if err := g.AddChatModelNode(
		nodeChatModel,
		chatModel,
		// 在喂给 ChatModel 之前，将输入消息追加到状态历史，并把完整历史作为输入
		compose.WithStatePreHandler(func(ctx context.Context, in []*schema.Message, st *agentState) ([]*schema.Message, error) {
			st.History = append(st.History, in...)
			return st.History, nil
		}),
		// 将 ChatModel 的输出（可能包含 tool_calls 的助手消息）也放入历史
		compose.WithStatePostHandler(func(ctx context.Context, out *schema.Message, st *agentState) (*schema.Message, error) {
			st.History = append(st.History, out)
			return out, nil
		}),
	); err != nil {
		return nil, fmt.Errorf("add chat model node failed: %w", err)
	}

	// 创建并添加 ToolsNode
	toolsNode, err := compose.NewToolNode(ctx,
		&compose.ToolsNodeConfig{
			Tools: toolsList,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create NewToolNode failed: %w", err)
	}
	if err := g.AddToolsNode(nodeTools, toolsNode); err != nil {
		return nil, fmt.Errorf("add tools node failed: %w", err)
	}

	// 连接节点
	if err := g.AddEdge(compose.START, nodePrompt); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodePrompt, nodeChatModel); err != nil {
		return nil, err
	}
	// 根据是否存在工具调用进行分支：有工具调用 -> tools；否则 -> END
	if err := g.AddBranch(nodeChatModel, compose.NewGraphBranch(
		func(ctx context.Context, in *schema.Message) (endNode string, err error) {
			if len(in.ToolCalls) > 0 {
				return nodeTools, nil
			}
			return compose.END, nil
		},
		map[string]bool{
			nodeTools:   true,
			compose.END: true,
		},
	)); err != nil {
		return nil, err
	}
	if err := g.AddEdge(nodeTools, nodeChatModel); err != nil {
		return nil, err
	}

	return g, nil
}

func (p *AgentProcessor) newChatTemplate() prompt.ChatTemplate {
	template, _ := pb.GetSimpleManager().GetTemplate("general_assistant")

	systemMessage, _ := pb.PromptBuilderSugar(
		template,
	)

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

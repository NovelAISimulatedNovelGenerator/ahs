package simpleexample

import (
	"context"
	"log"
	"sync"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"ahs/internal/config"
	"ahs/internal/service"
)

var (
	cfg     *config.LLMConfig
	cfgOnce sync.Once
	input   string
)

type SimpleProcessor struct{}

func (p *SimpleProcessor) Process(ctx context.Context, input string) (string, error) {
	s, err := gen()
	if err != nil {
		return "", err
	}
	return s.Content, nil
}
func (p *SimpleProcessor) ProcessStream(
	ctx context.Context, input string, callback service.StreamCallback) error {
	//TODO:
	return nil
}

func init() {
	input = "hello"
}

// initConfig 初始化配置，使用懒加载模式
func initConfig() {
	cfgOnce.Do(func() {
		// 获取全局配置
		globalConfig := config.GetGlobalConfig()
		if globalConfig == nil || globalConfig.LLMConfigs == nil {
			// 如果配置未加载，使用默认值
			cfg = &config.LLMConfig{
				APIBaseURL: "http://127.0.0.1:3000/v1",
				APIKey:     "sk-rac1XoSpt3eESULMNGKxAvBQq2WwcqIoSJMhsg2ubOU6tiJQ",
				Model:      "kimi-k2-turbo-preview",
			}
			return
		}
		
		// 使用 local 配置
		if localCfg, ok := globalConfig.LLMConfigs["local"]; ok {
			cfg = &localCfg
		} else {
			// 如果没有 local 配置，使用第一个可用的配置
			for _, c := range globalConfig.LLMConfigs {
				cfg = &c
				break
			}
		}
		
		// 确保配置不为空
		if cfg == nil {
			cfg = &config.LLMConfig{
				APIBaseURL: "http://127.0.0.1:3000/v1",
				APIKey:     "sk-rac1XoSpt3eESULMNGKxAvBQq2WwcqIoSJMhsg2ubOU6tiJQ",
				Model:      "kimi-k2-turbo-preview",
			}
		}
	})
}

func newChatModel(ctx context.Context) model.ToolCallingChatModel {
	// 确保配置已初始化
	initConfig()
	
	var cm model.ToolCallingChatModel
	var err error

	cm, err = openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.APIBaseURL,
		Model:   cfg.Model,
	})
	if err != nil {
		log.Fatal(err)
	}
	return cm
}

func gen() (*schema.Message, error) {
	ctx := context.Background()
	messages := []*schema.Message{{Role: schema.User, Content: input}}
	cm := newChatModel(ctx)
	ret, err := cm.Generate(ctx, messages)
	if err != nil {
		log.Fatalf("Generate failed, err=%v", err)
	}
	return ret, err
}

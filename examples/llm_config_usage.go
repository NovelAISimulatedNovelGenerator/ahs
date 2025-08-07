package main

import (
	"fmt"
	"log"

	"ahs/internal/config"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 使用示例1: 获取特定LLM配置
	if openaiConfig, ok := cfg.LLMConfigs["openai"]; ok {
		fmt.Printf("OpenAI 配置: URL=%s, Model=%s\n", 
			openaiConfig.APIBaseURL, openaiConfig.Model)
	}

	// 使用示例2: 遍历所有LLM配置
	fmt.Println("\n所有LLM配置:")
	for name, llmConfig := range cfg.LLMConfigs {
		fmt.Printf("  %s: URL=%s, Model=%s\n", 
			name, llmConfig.APIBaseURL, llmConfig.Model)
	}

	// 使用示例3: 动态选择LLM提供商
	provider := "claude" // 可以从请求参数中获取
	if llmConfig, ok := cfg.LLMConfigs[provider]; ok {
		fmt.Printf("\n使用 %s 提供商: %s\n", provider, llmConfig.Model)
		// 这里可以初始化对应的 LLM 客户端
	} else {
		fmt.Printf("\n未找到 %s 提供商配置\n", provider)
	}
}

package promptbuilder

import (
	"context"
	"testing"
	"time"
)

func TestPromptBuilder_BasicFunctionality(t *testing.T) {
	// 创建构建器
	builder := NewPromptBuilder()
	
	// 创建简单测试模板
	template := &PromptTemplate{
		ID:          "test_template",
		Name:        "测试模板",
		Description: "用于测试的简单模板",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Variables: map[string]interface{}{
			"default_name": "测试用户",
		},
		Layers: map[PromptLayer]Layer{
			CoreDefinition: {
				Name:        "核心定义",
				Description: "测试核心定义层",
				Enabled:     true,
				Template:    "你是{{.ai_name | default \"AI助手\"}}，为{{.user_name | default .default_name}}服务。",
			},
			InteractionInterface: {
				Name:        "交互接口",
				Description: "测试交互接口层",
				Enabled:     true,
				Template:    "当前时间：{{now}}，模板版本：{{.template_version}}",
			},
		},
	}
	
	// 注册模板
	err := builder.RegisterTemplate(template)
	if err != nil {
		t.Fatalf("注册模板失败: %v", err)
	}
	
	// 构建提示词
	ctx := context.Background()
	variables := map[string]interface{}{
		"ai_name":   "智能助手",
		"user_name": "张三",
	}
	
	prompt, err := builder.BuildPrompt(ctx, "test_template", variables)
	if err != nil {
		t.Fatalf("构建提示词失败: %v", err)
	}
	
	t.Logf("生成的提示词:\n%s", prompt)
	
	// 验证提示词包含预期内容
	if !contains(prompt, "智能助手") {
		t.Error("提示词应该包含'智能助手'")
	}
	
	if !contains(prompt, "张三") {
		t.Error("提示词应该包含'张三'")
	}
	
	if !contains(prompt, "核心定义") {
		t.Error("提示词应该包含层级标题'核心定义'")
	}
}

func TestManager_DefaultTemplates(t *testing.T) {
	// 获取管理器实例
	manager := GetSimpleManager()
	
	// 测试默认模板是否已注册
	templates := manager.ListTemplates()
	expectedTemplates := []string{"data_analyst", "code_assistant", "general_assistant"}
	
	for _, expected := range expectedTemplates {
		found := false
		for _, actual := range templates {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("默认模板 %s 未找到", expected)
		}
	}
	
	// 测试构建数据分析师提示词
	ctx := context.Background()
	variables := map[string]interface{}{
		"company":        "测试公司",
		"database_type":  "MySQL",
		"analysis_focus": "销售数据分析",
	}
	
	prompt, err := manager.BuildPrompt(ctx, "data_analyst", variables)
	if err != nil {
		t.Fatalf("构建数据分析师提示词失败: %v", err)
	}
	
	t.Logf("数据分析师提示词:\n%s", prompt)
	
	// 验证关键内容
	if !contains(prompt, "测试公司") {
		t.Error("提示词应该包含公司名称")
	}
	
	if !contains(prompt, "MySQL") {
		t.Error("提示词应该包含数据库类型")
	}
}

func TestManager_PreviewPrompt(t *testing.T) {
	manager := GetSimpleManager()
	
	variables := map[string]interface{}{
		"programming_lang": "Python",
		"code_style":       "PEP8",
	}
	
	preview, err := manager.PreviewPrompt("code_assistant", variables)
	if err != nil {
		t.Fatalf("预览提示词失败: %v", err)
	}
	
	t.Logf("提示词预览: %+v", preview)
	
	// 验证预览内容
	if preview["template_id"] != "code_assistant" {
		t.Error("预览应该包含正确的模板ID")
	}
	
	if preview["template_name"] != "智能代码助手" {
		t.Error("预览应该包含正确的模板名称")
	}
	
	// 验证变量合并
	variables_map, ok := preview["variables"].(map[string]interface{})
	if !ok {
		t.Fatal("预览变量应该是map类型")
	}
	
	if variables_map["programming_lang"] != "Python" {
		t.Error("预览应该包含用户提供的变量")
	}
}

func TestPromptBuilder_TemplateValidation(t *testing.T) {
	builder := NewPromptBuilder()
	
	// 测试无效模板（缺少ID）
	invalidTemplate := &PromptTemplate{
		Name: "无效模板",
	}
	
	err := builder.RegisterTemplate(invalidTemplate)
	if err == nil {
		t.Error("应该拒绝没有ID的模板")
	}
	
	// 测试无效模板（缺少层级名称）
	invalidTemplate2 := &PromptTemplate{
		ID:   "invalid_template_2",
		Name: "无效模板2",
		Layers: map[PromptLayer]Layer{
			CoreDefinition: {
				// 缺少Name字段
				Description: "测试层级",
				Enabled:     true,
			},
		},
	}
	
	err = builder.RegisterTemplate(invalidTemplate2)
	if err == nil {
		t.Error("应该拒绝层级名称为空的模板")
	}
}

func TestPromptBuilder_VariableMerging(t *testing.T) {
	builder := NewPromptBuilder()
	
	template := &PromptTemplate{
		ID:   "variable_test",
		Name: "变量测试模板",
		Variables: map[string]interface{}{
			"default_value": "默认值",
			"template_var":  "模板变量",
		},
		Layers: map[PromptLayer]Layer{
			CoreDefinition: {
				Name:     "测试层",
				Enabled:  true,
				Template: "默认值: {{.default_value}}, 用户值: {{.user_value}}, 模板变量: {{.template_var}}",
			},
		},
	}
	
	builder.RegisterTemplate(template)
	
	ctx := context.Background()
	userVariables := map[string]interface{}{
		"user_value":     "用户提供的值",
		"default_value":  "用户覆盖的默认值", // 应该覆盖模板默认值
	}
	
	prompt, err := builder.BuildPrompt(ctx, "variable_test", userVariables)
	if err != nil {
		t.Fatalf("构建提示词失败: %v", err)
	}
	
	t.Logf("变量合并测试结果:\n%s", prompt)
	
	// 验证变量合并逻辑
	if !contains(prompt, "用户覆盖的默认值") {
		t.Error("用户变量应该覆盖模板默认值")
	}
	
	if !contains(prompt, "用户提供的值") {
		t.Error("应该包含用户提供的变量")
	}
	
	if !contains(prompt, "模板变量") {
		t.Error("应该包含模板变量")
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

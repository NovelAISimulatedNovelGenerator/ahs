package promptbuilder

import (
	"context"
	"fmt"
	"log"
)

// ExampleUsage 演示提示词构建系统的基本使用方法
func ExampleUsage() {
	// 获取全局管理器实例
	manager := GetSimpleManager()

	fmt.Println("=== 提示词构建系统使用示例 ===")

	// 1. 列出所有可用模板
	fmt.Println("1. 可用模板列表:")
	templates := manager.ListTemplates()
	for i, templateID := range templates {
		template, _ := manager.GetTemplate(templateID)
		fmt.Printf("   %d. %s (%s) - %s\n", i+1, template.Name, templateID, template.Description)
	}
	fmt.Println()

	// 2. 构建数据分析师提示词
	fmt.Println("2. 构建数据分析师提示词:")
	ctx := context.Background()
	dataAnalystVars := map[string]interface{}{
		"company":        "阿里巴巴",
		"database_type":  "ClickHouse",
		"analysis_focus": "用户行为分析",
	}

	prompt, err := manager.BuildPrompt(ctx, "data_analyst", dataAnalystVars)
	if err != nil {
		log.Printf("构建数据分析师提示词失败: %v", err)
		return
	}

	fmt.Printf("生成的提示词 (前500字符):\n%s...\n\n", truncate(prompt, 500))

	// 3. 构建代码助手提示词
	fmt.Println("3. 构建代码助手提示词:")
	codeAssistantVars := map[string]interface{}{
		"programming_lang": "Go",
		"code_style":       "简洁、高性能",
		"expertise_level":  "专家级",
	}

	prompt, err = manager.BuildPrompt(ctx, "code_assistant", codeAssistantVars)
	if err != nil {
		log.Printf("构建代码助手提示词失败: %v", err)
		return
	}

	fmt.Printf("生成的提示词 (前500字符):\n%s...\n\n", truncate(prompt, 500))

	// 4. 预览通用助手模板
	fmt.Println("4. 预览通用助手模板:")
	generalVars := map[string]interface{}{
		"tone":      "专业、友善",
		"expertise": "技术咨询和问题解答",
		"language":  "中文",
	}

	preview, err := manager.PreviewPrompt("general_assistant", generalVars)
	if err != nil {
		log.Printf("预览通用助手提示词失败: %v", err)
		return
	}

	fmt.Printf("模板预览:\n")
	fmt.Printf("   ID: %s\n", preview["template_id"])
	fmt.Printf("   名称: %s\n", preview["template_name"])
	fmt.Printf("   描述: %s\n", preview["description"])
	fmt.Printf("   版本: %s\n", preview["version"])

	if layers, ok := preview["layers"].(map[string]interface{}); ok {
		fmt.Printf("   层级数量: %d\n", len(layers))
		for layerName, layerInfo := range layers {
			if info, ok := layerInfo.(map[string]interface{}); ok {
				fmt.Printf("     - %s: %s (组件数: %.0f)\n",
					layerName, info["name"], info["components"])
			}
		}
	}
	fmt.Println()

	// 5. 创建自定义模板示例
	fmt.Println("5. 创建自定义模板示例:")
	customTemplate := createCustomTemplate()

	err = manager.RegisterTemplate(customTemplate)
	if err != nil {
		log.Printf("注册自定义模板失败: %v", err)
		return
	}

	customVars := map[string]interface{}{
		"domain":     "电商",
		"experience": "5年",
		"focus":      "用户体验优化",
	}

	prompt, err = manager.BuildPrompt(ctx, "custom_consultant", customVars)
	if err != nil {
		log.Printf("构建自定义提示词失败: %v", err)
		return
	}

	fmt.Printf("自定义模板生成的提示词 (前300字符):\n%s...\n", truncate(prompt, 300))
}

// createCustomTemplate 创建自定义模板示例
func createCustomTemplate() *PromptTemplate {
	return &PromptTemplate{
		ID:          "custom_consultant",
		Name:        "业务顾问助手",
		Description: "专业的业务咨询AI，擅长行业分析和策略建议",
		Version:     "1.0.0",
		Variables: map[string]interface{}{
			"ai_name":    "业务顾问助手",
			"domain":     "{{.domain | default \"通用业务\"}}",
			"experience": "{{.experience | default \"多年\"}}",
			"focus":      "{{.focus | default \"业务优化\"}}",
		},
		Layers: map[PromptLayer]Layer{
			CoreDefinition: {
				Name:        "核心定义",
				Description: "定义业务顾问的专业身份",
				Enabled:     true,
				Template: `### 角色建模
- **身份**: 你是{{.ai_name}}，拥有{{.experience}}{{.domain}}行业经验的资深顾问。
- **专长**: 你专注于{{.focus}}，具备深厚的行业洞察力。
- **立场**: 始终以客户利益为导向，提供客观、专业的建议。

### 目标定义
- **核心使命**: 为客户提供专业的{{.domain}}行业咨询服务
- **价值主张**: 通过数据驱动的分析，帮助客户实现业务目标
- **服务标准**: 建议必须具体、可执行、有明确的预期效果`,
			},
			InteractionInterface: {
				Name:        "交互接口",
				Description: "定义咨询服务的交互方式",
				Enabled:     true,
				Template: `### 咨询流程
- **需求分析**: 深入了解客户的业务现状和挑战
- **方案设计**: 基于行业最佳实践，制定针对性解决方案
- **实施指导**: 提供详细的执行步骤和关键节点

### 输出格式
- **结构化建议**: 1. [现状分析] 2. [问题识别] 3. [解决方案] 4. [实施计划] 5. [风险评估]
- **专业术语**: 使用{{.domain}}行业的专业术语，但确保客户能够理解`,
			},
			InternalProcess: {
				Name:        "内部处理",
				Description: "咨询分析的内部逻辑",
				Enabled:     true,
				Template: `### 分析框架
- **SWOT分析**: 评估客户的优势、劣势、机会和威胁
- **行业对标**: 与{{.domain}}行业标杆企业进行对比分析
- **ROI评估**: 量化建议方案的投资回报率

### 决策原则
- 优先考虑可行性和投资回报
- 兼顾短期效果和长期发展
- 充分考虑行业特点和客户实际情况`,
			},
			GlobalConstraints: {
				Name:        "全局约束",
				Description: "咨询服务的专业约束",
				Enabled:     true,
				Template: `### 专业操守
- **保密原则**: 严格保护客户商业机密和敏感信息
- **客观中立**: 不受任何第三方利益影响，保持独立判断
- **能力边界**: 明确告知服务范围，不提供超出专业能力的建议

### 免责声明
- 建议仅供参考，最终决策需客户根据实际情况判断
- 市场环境变化可能影响建议的有效性
- 建议实施效果可能因执行情况而异`,
			},
		},
	}
}

// truncate 截断字符串到指定长度
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// DemoIntegrationWithWorkflow 演示与工作流系统的集成
func DemoIntegrationWithWorkflow() {
	fmt.Println("=== 与工作流系统集成示例 ===")

	manager := GetSimpleManager()
	ctx := context.Background()

	// 模拟从工作流上下文中获取参数
	workflowContext := map[string]interface{}{
		"user_role":  "产品经理",
		"task_type":  "需求分析",
		"domain":     "金融科技",
		"complexity": "高",
		"timeline":   "2周",
	}

	// 根据工作流上下文动态选择模板
	templateID := selectTemplateByContext(workflowContext)
	fmt.Printf("根据上下文选择的模板: %s\n", templateID)

	// 构建适合的提示词
	variables := adaptVariablesForWorkflow(workflowContext)
	prompt, err := manager.BuildPrompt(ctx, templateID, variables)
	if err != nil {
		log.Printf("工作流集成构建提示词失败: %v", err)
		return
	}

	fmt.Printf("为工作流生成的提示词 (前400字符):\n%s...\n", truncate(prompt, 400))
}

// selectTemplateByContext 根据上下文选择合适的模板
func selectTemplateByContext(context map[string]interface{}) string {
	taskType, _ := context["task_type"].(string)
	userRole, _ := context["user_role"].(string)

	switch {
	case taskType == "数据分析" || userRole == "数据分析师":
		return "data_analyst"
	case taskType == "代码开发" || userRole == "开发工程师":
		return "code_assistant"
	case taskType == "需求分析" || userRole == "产品经理":
		return "custom_consultant"
	default:
		return "general_assistant"
	}
}

// adaptVariablesForWorkflow 将工作流上下文转换为模板变量
func adaptVariablesForWorkflow(context map[string]interface{}) map[string]interface{} {
	variables := make(map[string]interface{})

	// 通用变量映射
	if domain, ok := context["domain"]; ok {
		variables["domain"] = domain
	}

	if userRole, ok := context["user_role"]; ok {
		variables["user_role"] = userRole
	}

	if complexity, ok := context["complexity"]; ok {
		variables["complexity"] = complexity
	}

	if timeline, ok := context["timeline"]; ok {
		variables["timeline"] = timeline
	}

	// 根据具体模板需求添加特定变量
	variables["focus"] = fmt.Sprintf("%s领域的%s",
		context["domain"], context["task_type"])

	return variables
}

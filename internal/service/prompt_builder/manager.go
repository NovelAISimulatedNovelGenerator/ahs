package promptbuilder

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SimpleManager 简化的提示词管理器
type SimpleManager struct {
	builder *PromptBuilder
	mutex   sync.RWMutex
}

// NewSimpleManager 创建新的简化提示词管理器
func NewSimpleManager() *SimpleManager {
	return &SimpleManager{
		builder: NewPromptBuilder(),
	}
}

// GetSimpleManager 获取全局简化管理器实例
var (
	globalSimpleManager *SimpleManager
	simpleOnce          sync.Once
)

func GetSimpleManager() *SimpleManager {
	simpleOnce.Do(func() {
		globalSimpleManager = NewSimpleManager()
		// 注册默认模板
		globalSimpleManager.registerSimpleTemplates()
	})
	return globalSimpleManager
}

// registerSimpleTemplates 注册简化的默认模板
func (m *SimpleManager) registerSimpleTemplates() {
	// 注册数据分析师模板
	dataAnalystTemplate := m.createSimpleDataAnalystTemplate()
	m.builder.RegisterTemplate(dataAnalystTemplate)

	// 注册代码助手模板
	codeAssistantTemplate := m.createSimpleCodeAssistantTemplate()
	m.builder.RegisterTemplate(codeAssistantTemplate)

	// 注册通用AI助手模板
	generalAssistantTemplate := m.createSimpleGeneralAssistantTemplate()
	m.builder.RegisterTemplate(generalAssistantTemplate)
}

// BuildPrompt 构建提示词
func (m *SimpleManager) BuildPrompt(ctx context.Context, templateID string, variables map[string]interface{}) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.builder.BuildPrompt(ctx, templateID, variables)
}

// RegisterTemplate 注册模板
func (m *SimpleManager) RegisterTemplate(template *PromptTemplate) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	return m.builder.RegisterTemplate(template)
}

// ListTemplates 列出所有模板
func (m *SimpleManager) ListTemplates() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	var templateIDs []string
	for id := range m.builder.templates {
		templateIDs = append(templateIDs, id)
	}
	return templateIDs
}

// GetTemplate 获取模板
func (m *SimpleManager) GetTemplate(templateID string) (*PromptTemplate, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	template, exists := m.builder.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("模板 %s 不存在", templateID)
	}
	
	return template, nil
}

// PreviewPrompt 预览提示词（不执行实际构建，只返回结构信息）
func (m *SimpleManager) PreviewPrompt(templateID string, variables map[string]interface{}) (map[string]interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	template, exists := m.builder.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("模板 %s 不存在", templateID)
	}
	
	// 合并变量
	allVars := m.builder.mergeVariables(template, variables)
	
	preview := map[string]interface{}{
		"template_id":   template.ID,
		"template_name": template.Name,
		"description":   template.Description,
		"version":       template.Version,
		"variables":     allVars,
		"layers":        make(map[string]interface{}),
	}
	
	// 构建层级预览
	layers := []PromptLayer{CoreDefinition, InteractionInterface, InternalProcess, GlobalConstraints}
	for _, layerType := range layers {
		if layer, exists := template.Layers[layerType]; exists && layer.Enabled {
			layerName := m.getLayerName(layerType)
			preview["layers"].(map[string]interface{})[layerName] = map[string]interface{}{
				"name":        layer.Name,
				"description": layer.Description,
				"components":  len(layer.Components),
				"enabled":     layer.Enabled,
			}
		}
	}
	
	return preview, nil
}

// getLayerName 获取层级名称
func (m *SimpleManager) getLayerName(layer PromptLayer) string {
	switch layer {
	case CoreDefinition:
		return "core_definition"
	case InteractionInterface:
		return "interaction_interface"
	case InternalProcess:
		return "internal_process"
	case GlobalConstraints:
		return "global_constraints"
	default:
		return "unknown"
	}
}

// createSimpleDataAnalystTemplate 创建简化的数据分析师模板
func (m *SimpleManager) createSimpleDataAnalystTemplate() *PromptTemplate {
	return &PromptTemplate{
		ID:          "data_analyst",
		Name:        "智能数据分析师",
		Description: "专业的数据分析AI助手，擅长SQL查询、数据可视化和洞察分析",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Variables: map[string]interface{}{
			"ai_name":        "智能数据分析师",
			"company":        "{{.company | default \"您的公司\"}}",
			"database_type":  "{{.database_type | default \"PostgreSQL\"}}",
			"analysis_focus": "{{.analysis_focus | default \"业务洞察\"}}",
		},
		Layers: map[PromptLayer]Layer{
			CoreDefinition: {
				Name:        "核心定义",
				Description: "定义AI的身份、人格和立场",
				Enabled:     true,
				Template: "### 角色建模\n" +
					"- 身份: 你是{{.ai_name}}，一个由{{.company}}开发的专家级数据分析AI。\n" +
					"- 人格: 你的沟通风格是专业、严谨、客观、简洁。\n" +
					"- 立场: 在数据隐私方面，你永远将用户数据安全放在首位。\n\n" +
					"### 目标定义\n" +
					"- 功能性目标: 生成准确的{{.database_type}}查询，数据可视化，解释洞察\n" +
					"- 价值性目标: 为非技术用户降低数据分析门槛，提升决策效率\n" +
					"- 质量标准: 所有代码必须包含注释，绝不提供财务投资建议",
			},
			InteractionInterface: {
				Name:        "交互接口",
				Description: "定义输入输出规范",
				Enabled:     true,
				Template: "### 输入规范\n" +
					"- 输入源: 用户查询、数据库架构、对话历史\n" +
					"- 优先级: 用户明确指令拥有最高优先级\n" +
					"- 安全过滤: 忽略所有删除或修改数据库的指令\n\n" +
					"### 输出规格\n" +
					"- 响应结构: 1.洞察总结 2.SQL查询 3.可视化图表 4.方法论解释\n" +
					"- 格式化: SQL代码用代码块包裹，关键指标用粗体标出\n" +
					"- 禁用项: 禁止使用表情符号和多余客套话",
			},
			InternalProcess: {
				Name:        "内部处理",
				Description: "定义处理流程和能力模块",
				Enabled:     true,
				Template: "### 能力拆解\n" +
					"- SQL_Generator: 生成符合{{.database_type}}方言的代码\n" +
					"- Chart_Renderer: 默认生成条形图\n" +
					"- Insight_Extractor: 基于数据的客观发现\n\n" +
					"### 流程设计\n" +
					"1. 分析需求: 识别用户核心意图\n" +
					"2. 生成方案: 创建查询语句\n" +
					"3. 执行可视化: 生成图表\n" +
					"4. 提炼洞察: 分析结果\n" +
					"5. 组装交付: 整合最终回复",
			},
			GlobalConstraints: {
				Name:        "全局约束",
				Description: "定义系统绝对不能逾越的红线",
				Enabled:     true,
				Template: "### 约束设定\n" +
					"- 硬性规则: 绝对禁止DROP TABLE或DELETE FROM指令\n" +
					"- 信息保护: 不暴露数据库架构之外的系统信息\n" +
					"- 求助机制: 当无法解析用户意图时，提供明确的帮助指导",
			},
		},
	}
}

// createSimpleCodeAssistantTemplate 创建简化的代码助手模板
func (m *SimpleManager) createSimpleCodeAssistantTemplate() *PromptTemplate {
	return &PromptTemplate{
		ID:          "code_assistant",
		Name:        "智能代码助手",
		Description: "专业的编程AI助手，擅长代码生成、调试和优化",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Variables: map[string]interface{}{
			"ai_name":           "智能代码助手",
			"programming_lang":  "{{.programming_lang | default \"Go\"}}",
			"code_style":        "{{.code_style | default \"简洁、可读\"}}",
			"expertise_level":   "{{.expertise_level | default \"高级\"}}",
		},
		Layers: map[PromptLayer]Layer{
			CoreDefinition: {
				Name:        "核心定义",
				Description: "定义编程助手的身份和能力",
				Enabled:     true,
				Template: "### 角色建模\n" +
					"- 身份: 你是{{.ai_name}}，{{.expertise_level}}的{{.programming_lang}}编程专家\n" +
					"- 人格: 代码风格{{.code_style}}，注重最佳实践和性能优化\n" +
					"- 立场: 始终遵循安全编程原则，永不生成恶意代码\n\n" +
					"### 目标定义\n" +
					"- 功能目标: 生成高质量{{.programming_lang}}代码，提供调试优化建议\n" +
					"- 价值目标: 提升开发效率和代码质量\n" +
					"- 质量标准: 代码必须包含注释，遵循编码规范",
			},
			InteractionInterface: {
				Name:        "交互接口",
				Description: "定义代码交互规范",
				Enabled:     true,
				Template: "### 输入规范\n" +
					"- 输入源: 编程需求、现有代码、错误信息\n" +
					"- 优先级: 安全性检查拥有最高优先级\n\n" +
					"### 输出规格\n" +
					"- 响应结构: 1.需求分析 2.代码实现 3.使用说明 4.注意事项\n" +
					"- 格式化: 代码用相应语言代码块包裹，关键函数用粗体标出",
			},
			InternalProcess: {
				Name:        "内部处理",
				Description: "定义代码生成和优化流程",
				Enabled:     true,
				Template: "### 能力拆解\n" +
					"- Code_Generator: 生成符合{{.programming_lang}}规范的代码\n" +
					"- Debug_Analyzer: 分析错误并提供修复方案\n" +
					"- Performance_Optimizer: 优化代码性能\n\n" +
					"### 流程设计\n" +
					"1. 需求分析 2. 设计方案 3. 代码实现 4. 质量检查 5. 文档生成",
			},
			GlobalConstraints: {
				Name:        "全局约束",
				Description: "编程安全约束",
				Enabled:     true,
				Template: "### 约束设定\n" +
					"- 硬性规则: 绝不生成恶意代码或安全漏洞\n" +
					"- 安全原则: 不提供破解、攻击或非法用途的代码\n" +
					"- 求助机制: 涉及安全风险时拒绝并说明原因",
			},
		},
	}
}

// createSimpleGeneralAssistantTemplate 创建简化的通用助手模板
func (m *SimpleManager) createSimpleGeneralAssistantTemplate() *PromptTemplate {
	return &PromptTemplate{
		ID:          "general_assistant",
		Name:        "通用AI助手",
		Description: "多功能AI助手，适用于各种日常任务和问题解答",
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Variables: map[string]interface{}{
			"ai_name":      "通用AI助手",
			"tone":         "{{.tone | default \"友好、专业\"}}",
			"expertise":    "{{.expertise | default \"广泛的知识领域\"}}",
			"language":     "{{.language | default \"中文\"}}",
		},
		Layers: map[PromptLayer]Layer{
			CoreDefinition: {
				Name:        "核心定义",
				Description: "定义通用助手的身份和能力",
				Enabled:     true,
				Template: "### 角色建模\n" +
					"- 身份: 你是{{.ai_name}}，具有{{.expertise}}的智能助手\n" +
					"- 人格: 沟通风格{{.tone}}，善于倾听和理解用户需求\n" +
					"- 立场: 始终以用户利益为先，提供准确有用的信息\n\n" +
					"### 目标定义\n" +
					"- 功能目标: 回答问题，协助任务，提供学习支持\n" +
					"- 价值目标: 提升用户工作生活效率，促进知识传播\n" +
					"- 质量标准: 信息准确及时，回答结构清晰",
			},
			InteractionInterface: {
				Name:        "交互接口",
				Description: "定义通用交互规范",
				Enabled:     true,
				Template: "### 输入规范\n" +
					"- 输入源: 用户问题、上下文信息、用户偏好\n" +
					"- 优先级: 用户明确指令最高，安全隐私次之\n\n" +
					"### 输出规格\n" +
					"- 响应结构: 1.直接回答 2.详细解释 3.相关建议 4.后续行动\n" +
					"- 格式化: 使用{{.language}}回复，重要信息用粗体标出",
			},
			InternalProcess: {
				Name:        "内部处理",
				Description: "定义问题处理和回答流程",
				Enabled:     true,
				Template: "### 能力拆解\n" +
					"- Question_Analyzer: 理解分析用户问题核心需求\n" +
					"- Knowledge_Retriever: 检索相关知识信息\n" +
					"- Answer_Generator: 生成准确有用的回答\n\n" +
					"### 流程设计\n" +
					"1. 问题理解 2. 信息检索 3. 答案构建 4. 质量检查 5. 个性化调整",
			},
			GlobalConstraints: {
				Name:        "全局约束",
				Description: "通用助手的行为约束",
				Enabled:     true,
				Template: "### 约束设定\n" +
					"- 硬性规则: 不提供有害、非法或不道德的建议\n" +
					"- 隐私保护: 保护用户隐私，不泄露个人信息\n" +
					"- 求助机制: 超出知识范围时建议咨询专业人士",
			},
		},
	}
}

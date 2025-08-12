package promptbuilder

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// PromptLayer 定义提示词的四层架构
type PromptLayer int

const (
	CoreDefinition       PromptLayer = iota // 第一层：核心定义
	InteractionInterface                    // 第二层：交互接口
	InternalProcess                         // 第三层：内部处理
	GlobalConstraints                       // 第四层：全局约束
)

// PromptTemplate 提示词模板结构
type PromptTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Layers      map[PromptLayer]Layer  `json:"layers"`
	Variables   map[string]interface{} `json:"variables"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Layer 表示提示词的一个层级
type Layer struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Components  []Component            `json:"components"`
	Template    string                 `json:"template"`
	Variables   map[string]interface{} `json:"variables"`
	Enabled     bool                   `json:"enabled"`
}

// Component 表示层级中的组件
type Component struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Template    string                 `json:"template"`
	Variables   map[string]interface{} `json:"variables"`
	Required    bool                   `json:"required"`
	Validation  *ValidationRule        `json:"validation,omitempty"`
}

// ValidationRule 验证规则
type ValidationRule struct {
	Type     string      `json:"type"`     // "regex", "length", "enum", "custom"
	Rule     string      `json:"rule"`     // 具体规则
	Message  string      `json:"message"`  // 错误消息
	Required bool        `json:"required"` // 是否必需
	Value    interface{} `json:"value"`    // 验证值
}

// PromptBuilder 提示词构建器
type PromptBuilder struct {
	templates map[string]*PromptTemplate
	funcMap   template.FuncMap
}

func PromptBuilderSugar(template *PromptTemplate) (string, error) {
	pb := NewPromptBuilder()
	pb.RegisterTemplate(template)
	return pb.BuildPrompt(context.Background(), template.ID, map[string]interface{}{})
}

// NewPromptBuilder 创建新的提示词构建器
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		templates: make(map[string]*PromptTemplate),
		funcMap:   getDefaultFuncMap(),
	}
}

// RegisterTemplate 注册提示词模板
func (pb *PromptBuilder) RegisterTemplate(template *PromptTemplate) error {
	if template.ID == "" {
		return fmt.Errorf("模板ID不能为空")
	}

	// 验证模板结构
	if err := pb.validateTemplate(template); err != nil {
		return fmt.Errorf("模板验证失败: %w", err)
	}

	pb.templates[template.ID] = template
	return nil
}

// BuildPrompt 构建提示词
func (pb *PromptBuilder) BuildPrompt(ctx context.Context, templateID string, variables map[string]interface{}) (string, error) {
	template, exists := pb.templates[templateID]
	if !exists {
		return "", fmt.Errorf("模板 %s 不存在", templateID)
	}

	// 合并变量
	allVars := pb.mergeVariables(template, variables)

	// 验证必需变量
	if err := pb.validateVariables(template, allVars); err != nil {
		return "", fmt.Errorf("变量验证失败: %w", err)
	}

	// 按层级构建提示词
	var promptParts []string

	// 按顺序构建四层
	layers := []PromptLayer{CoreDefinition, InteractionInterface, InternalProcess, GlobalConstraints}
	for _, layerType := range layers {
		if layer, exists := template.Layers[layerType]; exists && layer.Enabled {
			part, err := pb.buildLayer(ctx, &layer, allVars)
			if err != nil {
				return "", fmt.Errorf("构建层级 %v 失败: %w", layerType, err)
			}
			if part != "" {
				promptParts = append(promptParts, part)
			}
		}
	}

	return strings.Join(promptParts, "\n\n---\n\n"), nil
}

// buildLayer 构建单个层级
func (pb *PromptBuilder) buildLayer(ctx context.Context, layer *Layer, variables map[string]interface{}) (string, error) {
	var parts []string

	// 添加层级标题
	if layer.Name != "" {
		parts = append(parts, fmt.Sprintf("## %s", layer.Name))
	}

	// 添加层级描述
	if layer.Description != "" {
		parts = append(parts, layer.Description)
	}

	// 构建组件
	for _, component := range layer.Components {
		if component.Required || pb.shouldIncludeComponent(&component, variables) {
			part, err := pb.buildComponent(ctx, &component, variables)
			if err != nil {
				return "", fmt.Errorf("构建组件 %s 失败: %w", component.Name, err)
			}
			if part != "" {
				parts = append(parts, part)
			}
		}
	}

	// 构建层级模板
	if layer.Template != "" {
		part, err := pb.renderTemplate(ctx, layer.Template, variables)
		if err != nil {
			return "", fmt.Errorf("渲染层级模板失败: %w", err)
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, "\n\n"), nil
}

// buildComponent 构建单个组件
func (pb *PromptBuilder) buildComponent(ctx context.Context, component *Component, variables map[string]interface{}) (string, error) {
	if component.Template == "" {
		return "", nil
	}

	// 合并组件变量
	componentVars := pb.mergeComponentVariables(component, variables)

	// 验证组件
	if component.Validation != nil {
		if err := pb.validateComponent(component, componentVars); err != nil {
			return "", fmt.Errorf("组件验证失败: %w", err)
		}
	}

	return pb.renderTemplate(ctx, component.Template, componentVars)
}

// renderTemplate 渲染模板
func (pb *PromptBuilder) renderTemplate(ctx context.Context, templateStr string, variables map[string]interface{}) (string, error) {
	tmpl, err := template.New("prompt").Funcs(pb.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("解析模板失败: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("执行模板失败: %w", err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// 辅助方法实现

// validateTemplate 验证模板结构
func (pb *PromptBuilder) validateTemplate(template *PromptTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("模板名称不能为空")
	}

	// 验证层级结构
	for layerType, layer := range template.Layers {
		if err := pb.validateLayer(layerType, &layer); err != nil {
			return fmt.Errorf("层级 %v 验证失败: %w", layerType, err)
		}
	}

	return nil
}

// validateLayer 验证层级结构
func (pb *PromptBuilder) validateLayer(layerType PromptLayer, layer *Layer) error {
	if layer.Name == "" {
		return fmt.Errorf("层级名称不能为空")
	}

	// 验证组件
	for i, component := range layer.Components {
		if err := pb.validateComponentStructure(&component); err != nil {
			return fmt.Errorf("组件 %d 验证失败: %w", i, err)
		}
	}

	return nil
}

// validateComponentStructure 验证组件结构
func (pb *PromptBuilder) validateComponentStructure(component *Component) error {
	if component.Name == "" {
		return fmt.Errorf("组件名称不能为空")
	}

	if component.Type == "" {
		return fmt.Errorf("组件类型不能为空")
	}

	return nil
}

// mergeVariables 合并变量
func (pb *PromptBuilder) mergeVariables(template *PromptTemplate, variables map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 首先添加模板默认变量
	for k, v := range template.Variables {
		result[k] = v
	}

	// 然后添加用户提供的变量（覆盖默认值）
	for k, v := range variables {
		result[k] = v
	}

	// 添加系统变量
	result["timestamp"] = time.Now().Format("2006-01-02 15:04:05")
	result["template_id"] = template.ID
	result["template_version"] = template.Version

	return result
}

// mergeComponentVariables 合并组件变量
func (pb *PromptBuilder) mergeComponentVariables(component *Component, variables map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 首先添加全局变量
	for k, v := range variables {
		result[k] = v
	}

	// 然后添加组件特定变量（覆盖全局变量）
	for k, v := range component.Variables {
		result[k] = v
	}

	return result
}

// validateVariables 验证必需变量
func (pb *PromptBuilder) validateVariables(template *PromptTemplate, variables map[string]interface{}) error {
	// 检查所有层级的必需变量
	for _, layer := range template.Layers {
		for _, component := range layer.Components {
			if component.Required {
				if err := pb.validateComponent(&component, variables); err != nil {
					return fmt.Errorf("必需组件 %s 验证失败: %w", component.Name, err)
				}
			}
		}
	}

	return nil
}

// validateComponent 验证组件
func (pb *PromptBuilder) validateComponent(component *Component, variables map[string]interface{}) error {
	if component.Validation == nil {
		return nil
	}

	rule := component.Validation

	// 检查必需字段
	if rule.Required {
		for k := range component.Variables {
			if _, exists := variables[k]; !exists {
				return fmt.Errorf("缺少必需变量: %s", k)
			}
		}
	}

	// 根据验证类型进行验证
	switch rule.Type {
	case "regex":
		return pb.validateRegex(rule, variables)
	case "length":
		return pb.validateLength(rule, variables)
	case "enum":
		return pb.validateEnum(rule, variables)
	default:
		return nil
	}
}

// validateRegex 正则表达式验证
func (pb *PromptBuilder) validateRegex(rule *ValidationRule, variables map[string]interface{}) error {
	regex, err := regexp.Compile(rule.Rule)
	if err != nil {
		return fmt.Errorf("无效的正则表达式: %w", err)
	}

	for k, v := range variables {
		if str, ok := v.(string); ok {
			if !regex.MatchString(str) {
				return fmt.Errorf("变量 %s 不匹配正则表达式 %s: %s", k, rule.Rule, rule.Message)
			}
		}
	}

	return nil
}

// validateLength 长度验证
func (pb *PromptBuilder) validateLength(rule *ValidationRule, variables map[string]interface{}) error {
	// 这里可以根据需要实现长度验证逻辑
	return nil
}

// validateEnum 枚举验证
func (pb *PromptBuilder) validateEnum(rule *ValidationRule, variables map[string]interface{}) error {
	// 这里可以根据需要实现枚举验证逻辑
	return nil
}

// shouldIncludeComponent 判断是否应该包含组件
func (pb *PromptBuilder) shouldIncludeComponent(component *Component, variables map[string]interface{}) bool {
	// 如果组件是必需的，总是包含
	if component.Required {
		return true
	}

	// 检查组件的变量是否都有值
	for k := range component.Variables {
		if _, exists := variables[k]; !exists {
			return false
		}
	}

	return true
}

// getDefaultFuncMap 获取默认的模板函数映射
func getDefaultFuncMap() template.FuncMap {
	return template.FuncMap{
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		"trim":  strings.TrimSpace,
		"join": func(sep string, elems []string) string {
			return strings.Join(elems, sep)
		},
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"contains": strings.Contains,
		"replace": func(old, new, s string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"now": func() string {
			return time.Now().Format("2006-01-02 15:04:05")
		},
		"date": func(format string) string {
			return time.Now().Format(format)
		},
		"default": func(defaultVal interface{}, val interface{}) interface{} {
			if val == nil || val == "" {
				return defaultVal
			}
			return val
		},
	}
}

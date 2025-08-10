package ragtool

// MemorySaveInput 保存记忆输入参数
type MemorySaveInput struct {
	Content string   `json:"content" jsonschema:"required,description=记忆内容"`
	Tags    []string `json:"tags,omitempty" jsonschema:"description=标签列表"`
	Kind    string   `json:"kind,omitempty" jsonschema:"description=记忆类型,enum=short_term|long_term|fact|note"`
	TTL     int      `json:"ttl_seconds,omitempty" jsonschema:"description=过期时间(秒)"`

	// 这些字段不会出现在工具的 schema 中，agent 无法直接设置
	UserID    string `json:"user_id,omitempty"`
	ArchiveID string `json:"archive_id,omitempty"`
}

// MemorySaveOutput 保存结果
type MemorySaveOutput struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message"`
}

// MemoryQueryInput 查询记忆输入参数
type MemoryQueryInput struct {
	// 注意：租户字段由 ToolArgumentsHandler 服务端注入，LLM 无需填写
	Query string   `json:"query,omitempty" jsonschema:"description=查询文本"`
	TopK  int      `json:"top_k,omitempty" jsonschema:"description=返回条数,默认10"`
	Tags  []string `json:"tags,omitempty" jsonschema:"description=标签过滤"`
	Kinds []string `json:"kinds,omitempty" jsonschema:"description=类型过滤"`

	// 这些字段不会出现在工具的 schema 中，agent 无法直接设置
	UserID    string `json:"user_id,omitempty"`
	ArchiveID string `json:"archive_id,omitempty"`
}

// MemoryQueryOutput 查询结果
type MemoryQueryOutput struct {
	Success bool             `json:"success"`
	Items   []MemoryItemView `json:"items"`
	Count   int              `json:"count"`
	Message string           `json:"message"`
}

// MemoryItemView 记忆项视图（简化版）
type MemoryItemView struct {
	ID        string   `json:"id"`
	Content   string   `json:"content"`
	Tags      []string `json:"tags,omitempty"`
	Kind      string   `json:"kind,omitempty"`
	CreatedAt string   `json:"created_at"`
	Score     float64  `json:"score,omitempty"`
}

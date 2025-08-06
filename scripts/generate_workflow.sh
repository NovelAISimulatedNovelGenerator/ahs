#!/bin/sh

# generate_workflow.sh
# 基于 simple_example 模板生成新的 workflow 实现
# 兼容 sh (POSIX shell)

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 默认目录
WORKFLOW_DIR="$PROJECT_ROOT/internal/workflow/eino_imp"

# 颜色输出函数
red() {
    printf '\033[31m%s\033[0m\n' "$1"
}

green() {
    printf '\033[32m%s\033[0m\n' "$1"
}

yellow() {
    printf '\033[33m%s\033[0m\n' "$1"
}

# 检查参数
if [ $# -lt 1 ]; then
    red "用法: $0 <workflow_name> [workflow_description]"
    red "示例: $0 my_workflow \"我的自定义工作流\""
    exit 1
fi

WORKFLOW_NAME="$1"
WORKFLOW_DESC="${2:-$1 workflow}"
WORKFLOW_DIR_PATH="$WORKFLOW_DIR/$WORKFLOW_NAME"

# 检查是否已存在
if [ -d "$WORKFLOW_DIR_PATH" ]; then
    red "错误: workflow '$WORKFLOW_NAME' 已存在"
    exit 1
fi

# 创建目录
mkdir -p "$WORKFLOW_DIR_PATH"



# 生成主实现文件
cat > "$WORKFLOW_DIR_PATH/${WORKFLOW_NAME}.go" << EOF
package $WORKFLOW_NAME

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"ahs/internal/service"
)

// ${WORKFLOW_NAME}Config 配置结构体
type ${WORKFLOW_NAME}Config struct {
	APIKey  string
	Model   string
	BaseURL string
}

// ${WORKFLOW_NAME}Processor workflow 处理器
type ${WORKFLOW_NAME}Processor struct {
	config *${WORKFLOW_NAME}Config
}

// New${WORKFLOW_NAME}Processor 创建新的处理器实例
func New${WORKFLOW_NAME}Processor() *${WORKFLOW_NAME}Processor {
	return &${WORKFLOW_NAME}Processor{
		config: &${WORKFLOW_NAME}Config{
			APIKey:  "your-api-key-here",
			Model:   "gpt-3.5-turbo",
			BaseURL: "https://api.openai.com/v1",
		},
	}
}

// Process 执行工作流
func (p *${WORKFLOW_NAME}Processor) Process(ctx context.Context, input string) (string, error) {
	// TODO: 实现你的业务逻辑

	return "", nil
}

// ProcessStream 流式执行工作流
func (p *${WORKFLOW_NAME}Processor) ProcessStream(
	ctx context.Context, 
	input string, 
	callback service.StreamCallback,
) error {
	// TODO: 实现流式处理逻辑

	return nil
}

// newChatModel 创建聊天模型
func (p *${WORKFLOW_NAME}Processor) newChatModel(ctx context.Context) (*openai.ChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  p.config.APIKey,
		Model:   p.config.Model,
		BaseURL: p.config.BaseURL,
	})
}

// SetConfig 设置配置
func (p *${WORKFLOW_NAME}Processor) SetConfig(config *${WORKFLOW_NAME}Config) {
	p.config = config
}

// GetConfig 获取当前配置
func (p *${WORKFLOW_NAME}Processor) GetConfig() *${WORKFLOW_NAME}Config {
	return p.config
}
EOF



# 设置文件权限
chmod +x "$WORKFLOW_DIR_PATH"

# 格式化代码
cd "$PROJECT_ROOT" && go fmt "$WORKFLOW_DIR_PATH"/*.go

green "✅ 成功生成 workflow '$WORKFLOW_NAME'"
green "📁 目录: $WORKFLOW_DIR_PATH"
green "📄 文件:"
green "  - ${WORKFLOW_NAME}.go (主实现)"

yellow "📝 下一步:"
yellow "  1. 修改配置中的 API 密钥"
yellow "  2. 实现具体的业务逻辑"
yellow "  3. 运行 ./scripts/scaffold_eino_imp_to_register.sh 注册 workflow"
yellow "  4. 运行测试: go test -v ./$WORKFLOW_DIR_PATH"

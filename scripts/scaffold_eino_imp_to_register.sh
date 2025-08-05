#!/bin/sh

# scaffold_eino_imp_to_register.sh
# 自动扫描 eino_imp 目录下的 workflow 实现并生成注册代码
# 兼容 sh (POSIX shell)

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 默认目录，可通过参数覆盖
EINO_IMP_DIR="${1:-$PROJECT_ROOT/internal/workflow/eino_imp}"
REGISTER_FILE="$PROJECT_ROOT/internal/workflow/register.go"

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

# 检查文件是否存在
if [ ! -d "$EINO_IMP_DIR" ]; then
    red "错误: 目录 $EINO_IMP_DIR 不存在"
    exit 1
fi

if [ ! -f "$REGISTER_FILE" ]; then
    red "错误: 注册文件 $REGISTER_FILE 不存在"
    exit 1
fi

# 创建临时文件
TEMP_FILE="/tmp/register.go.tmp"

# 开始生成新的 register.go
cat > "$TEMP_FILE" << 'EOF'
package workflow

import (
EOF

# 收集所有包名和注册信息
PACKAGES=""
REGISTERS=""

# 遍历目录
for dir in "$EINO_IMP_DIR"/*/; do
    [ ! -d "$dir" ] && continue
    
    dir_name=$(basename "$dir")
    
    # 查找包含 Process 方法的 .go 文件
    processor_file=""
    for go_file in "$dir"*.go; do
        [ ! -f "$go_file" ] && continue
        if grep -q "func.*Process.*context\.Context" "$go_file"; then
            processor_file="$go_file"
            break
        fi
    done
    
    [ -z "$processor_file" ] && continue
    
    # 提取类型名
    type_name=$(grep -E 'type.*struct' "$processor_file" | head -1 | sed 's/type \([A-Za-z0-9_]*\).*/\1/')
    [ -z "$type_name" ] && continue
    
    # 检查是否已注册
    register_line="m.Register(\"${dir_name}\", &${dir_name}.${type_name}{})"
    if grep -q "$register_line" "$REGISTER_FILE"; then
        green "已注册: $dir_name ($type_name)"
        continue
    fi
    
    PACKAGES="$PACKAGES $dir_name"
    REGISTERS="$REGISTERS $dir_name|$type_name"
    green "发现新 workflow: $dir_name ($type_name)"
done

# 如果没有新发现的 workflow
if [ -z "$REGISTERS" ]; then
    yellow "没有发现需要注册的新 workflow"
    rm -f "$TEMP_FILE"
    exit 0
fi

# 添加导入
for package_name in $PACKAGES; do
    echo "\t${package_name} \"ahs/internal/workflow/eino_imp/${package_name}\"" >> "$TEMP_FILE"
done

cat >> "$TEMP_FILE" << 'EOF'
)

func (m *Manager) RegisterWorkflows() {
	// Register workflow under here:
EOF

# 添加现有注册代码（如果存在）
grep "m.Register" "$REGISTER_FILE" | grep -v "// Register workflow" >> "$TEMP_FILE" || true

# 添加新注册代码
for register_info in $REGISTERS; do
    package_name=$(echo "$register_info" | cut -d'|' -f1)
    type_name=$(echo "$register_info" | cut -d'|' -f2)
    echo "\tm.Register(\"${package_name}\", &${package_name}.${type_name}{})" >> "$TEMP_FILE"
done

echo "}" >> "$TEMP_FILE"

# 备份原文件
cp "$REGISTER_FILE" "$REGISTER_FILE.bak"

# 替换文件
mv "$TEMP_FILE" "$REGISTER_FILE"

green "✅ 已更新注册文件: $REGISTER_FILE"
green "备份文件: $REGISTER_FILE.bak"

# 格式化代码
cd "$PROJECT_ROOT" && go fmt internal/workflow/register.go

# 显示更新内容
yellow "新增注册:"
echo "$REGISTERS" | tr ' ' '\n' | while IFS='|' read -r package_name type_name; do
    echo "  - $package_name ($type_name)"
done
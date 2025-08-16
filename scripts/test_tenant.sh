#!/bin/sh

# test_tenant.sh
# 测试租户头部验证功能
# 兼容 sh (POSIX shell)

set -e

API_BASE="http://localhost:8081"

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

blue() {
    printf '\033[34m%s\033[0m\n' "$1"
}

# 测试函数
test_case() {
    local name="$1"
    local expected_code="$2"
    shift 2
    
    yellow "=== $name ==="
    
    # 执行curl并捕获HTTP状态码
    actual_code=$(curl -s -o /dev/null -w "%{http_code}" "$@")
    
    if [ "$actual_code" = "$expected_code" ]; then
        green "✓ 通过 (HTTP $actual_code)"
    else
        red "✗ 失败 - 期望: $expected_code, 实际: $actual_code"
    fi
    echo
}

blue "开始测试租户头部验证功能..."
echo

# 测试 1: 缺少租户头部 (应该返回 400)
test_case "测试 1: 缺少租户头部" "400" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"workflow":"echo","input":"test without tenant"}' \
    "$API_BASE/api/execute"

# 测试 2: 只有 X-User-ID (应该返回 400)
test_case "测试 2: 只有 X-User-ID" "400" \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-User-ID: test_user" \
    -d '{"workflow":"echo","input":"test with user only"}' \
    "$API_BASE/api/execute"

# 测试 3: 只有 X-Archive-ID (应该返回 400)
test_case "测试 3: 只有 X-Archive-ID" "400" \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-Archive-ID: test_archive" \
    -d '{"workflow":"echo","input":"test with archive only"}' \
    "$API_BASE/api/execute"

# 测试 4: 完整租户头部 (应该返回 200)
test_case "测试 4: 完整租户头部" "200" \
    -X POST \
    -H "Content-Type: application/json" \
    -H "X-User-ID: test_user" \
    -H "X-Archive-ID: test_archive" \
    -d '{"workflow":"echo","input":"test with complete tenant"}' \
    "$API_BASE/api/execute"

# 测试 5: 健康检查 (排除路径，不需要租户)
test_case "测试 5: 健康检查 (排除路径)" "200" \
    "$API_BASE/health"

# 测试 6: 工作流列表 (排除路径，不需要租户)
test_case "测试 6: 工作流列表 (排除路径)" "200" \
    "$API_BASE/api/workflows"

green "测试完成！"
yellow "注意：确保服务器正在 $API_BASE 运行"
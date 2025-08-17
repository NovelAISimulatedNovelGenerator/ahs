w#!/bin/sh

# test_cors.sh
# 测试CORS配置功能
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

# 测试CORS预检请求
test_cors_preflight() {
    local name="$1"
    local origin="$2"
    local expected_allow="$3"  # "yes" 或 "no"
    
    yellow "=== $name ==="
    
    # 发送OPTIONS预检请求
    response=$(curl -s -i \
        -X OPTIONS \
        -H "Origin: $origin" \
        -H "Access-Control-Request-Method: POST" \
        -H "Access-Control-Request-Headers: Content-Type, X-User-ID, X-Archive-ID" \
        "$API_BASE/api/execute")
    
    # 检查HTTP状态码
    status_code=$(echo "$response" | head -1 | cut -d' ' -f2)
    
    # 检查Access-Control-Allow-Origin头部
    allow_origin=$(echo "$response" | grep -i "access-control-allow-origin" | cut -d' ' -f2- | tr -d '\r')
    
    if [ "$expected_allow" = "yes" ]; then
        if [ "$status_code" = "200" ] && [ "$allow_origin" = "$origin" ]; then
            green "✓ 通过 - Origin被允许 (HTTP $status_code, Origin: $allow_origin)"
        else
            red "✗ 失败 - 期望允许但被拒绝 (HTTP $status_code, Origin: $allow_origin)"
        fi
    else
        if [ "$status_code" = "403" ] || [ -z "$allow_origin" ]; then
            green "✓ 通过 - Origin被正确拒绝 (HTTP $status_code)"
        else
            red "✗ 失败 - 期望拒绝但被允许 (HTTP $status_code, Origin: $allow_origin)"
        fi
    fi
    echo
}

# 测试实际请求的CORS头部
test_cors_request() {
    local name="$1"
    local origin="$2"
    
    yellow "=== $name ==="
    
    # 发送实际请求并检查CORS头部
    response=$(curl -s -i \
        -X GET \
        -H "Origin: $origin" \
        "$API_BASE/health")
    
    # 检查Access-Control-Allow-Origin头部
    allow_origin=$(echo "$response" | grep -i "access-control-allow-origin" | cut -d' ' -f2- | tr -d '\r')
    
    if [ -n "$allow_origin" ]; then
        green "✓ CORS头部存在: $allow_origin"
    else
        yellow "! 无CORS头部 (可能因为Origin不被允许)"
    fi
    echo
}

blue "开始测试CORS配置功能..."
echo

# 测试 1: 允许的Origin (默认配置中的localhost:5173)
test_cors_preflight "测试 1: 允许的Origin (localhost:5173)" "http://localhost:5173" "yes"

# 测试 2: 不允许的Origin
test_cors_preflight "测试 2: 不允许的Origin (evil.com)" "https://evil.com" "no"

# 测试 3: 另一个不允许的Origin
test_cors_preflight "测试 3: 不允许的Origin (random.domain)" "https://random.domain" "no"

# 测试 4: 无Origin头部的请求
test_cors_request "测试 4: 无Origin的普通请求" ""

# 测试 5: 允许的Origin的实际请求
test_cors_request "测试 5: 允许Origin的实际请求" "http://localhost:5173"

# 测试 6: 不允许的Origin的实际请求
test_cors_request "测试 6: 不允许Origin的实际请求" "https://evil.com"

green "CORS测试完成！"
yellow "注意：确保服务器正在 $API_BASE 运行"
yellow "默认配置只允许 http://localhost:5173 访问"
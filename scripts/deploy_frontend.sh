#!/bin/sh

# deploy_frontend.sh
# 无脑一键部署前端 - npm install, build & 可选运行
# 兼容 sh (POSIX shell)

set -e

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
FRONTEND_DIR="$PROJECT_ROOT/frontend"

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

# 检查 Node.js 和 npm 是否安装
check_prerequisites() {
    blue "🔍 检查前置条件..."
    
    if ! command -v node > /dev/null 2>&1; then
        red "❌ 错误: 未找到 Node.js"
        red "请先安装 Node.js: https://nodejs.org/"
        exit 1
    fi
    
    if ! command -v npm > /dev/null 2>&1; then
        red "❌ 错误: 未找到 npm"
        red "请确保 npm 已正确安装"
        exit 1
    fi
    
    NODE_VERSION=$(node --version)
    NPM_VERSION=$(npm --version)
    green "✅ Node.js: $NODE_VERSION"
    green "✅ npm: $NPM_VERSION"
}

# 检查前端目录
check_frontend_dir() {
    if [ ! -d "$FRONTEND_DIR" ]; then
        red "❌ 错误: 前端目录不存在: $FRONTEND_DIR"
        exit 1
    fi
    
    if [ ! -f "$FRONTEND_DIR/package.json" ]; then
        red "❌ 错误: 未找到 package.json: $FRONTEND_DIR/package.json"
        exit 1
    fi
    
    green "✅ 前端目录检查通过: $FRONTEND_DIR"
}

# 安装依赖
install_dependencies() {
    blue "📦 安装依赖..."
    cd "$FRONTEND_DIR"
    
    if [ -f "package-lock.json" ]; then
        npm ci
    else
        npm install
    fi
    
    green "✅ 依赖安装完成"
}

# 构建项目
build_project() {
    blue "🔨 构建前端项目..."
    cd "$FRONTEND_DIR"
    
    npm run build
    
    if [ ! -d "dist" ]; then
        red "❌ 构建失败: 未生成 dist 目录"
        exit 1
    fi
    
    green "✅ 前端构建完成"
    green "📁 构建输出: $FRONTEND_DIR/dist"
}

# 启动开发服务器
start_dev_server() {
    blue "🚀 启动开发服务器..."
    cd "$FRONTEND_DIR"
    
    yellow "📝 开发服务器将在 http://localhost:5173 启动"
    yellow "📝 请确保后端服务器在 http://localhost:8081 运行"
    yellow "📝 按 Ctrl+C 停止服务器"
    echo ""
    
    npm run dev
}

# 显示帮助信息
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  无参数    - 执行完整部署 (install + build)"
    echo "  -r, --run - 部署后启动开发服务器"
    echo "  -d, --dev - 仅启动开发服务器 (跳过构建)"
    echo "  -h, --help - 显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0                # 安装依赖 + 构建"
    echo "  $0 --run          # 安装依赖 + 构建 + 启动开发服务器"
    echo "  $0 --dev          # 仅启动开发服务器"
}

# 主函数
main() {
    MODE="build"
    
    # 解析参数
    while [ $# -gt 0 ]; do
        case $1 in
            -r|--run)
                MODE="build_and_run"
                shift
                ;;
            -d|--dev)
                MODE="dev_only"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                red "❌ 未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    green "🎯 前端部署脚本开始执行..."
    echo ""
    
    # 检查前置条件
    check_prerequisites
    check_frontend_dir
    echo ""
    
    case $MODE in
        "dev_only")
            yellow "📋 模式: 仅启动开发服务器"
            start_dev_server
            ;;
        "build")
            yellow "📋 模式: 安装依赖 + 构建"
            install_dependencies
            echo ""
            build_project
            echo ""
            green "🎉 前端部署完成!"
            blue "💡 提示: 使用 '$0 --dev' 启动开发服务器"
            ;;
        "build_and_run")
            yellow "📋 模式: 安装依赖 + 构建 + 启动开发服务器"
            install_dependencies
            echo ""
            build_project
            echo ""
            start_dev_server
            ;;
    esac
}

# 执行主函数
main "$@"
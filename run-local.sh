#!/bin/bash

# Sunset 本地运行脚本 (macOS)
# 适用于 macOS 系统的本地开发和测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# 打印标题
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  🌅 Sunset 本地运行脚本 (macOS)${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# 检查 Go 环境
check_go() {
    print_info "检查 Go 环境..."

    if ! command -v go &> /dev/null; then
        print_error "未找到 Go 环境！"
        echo ""
        echo "请先安装 Go 1.21 或更高版本："
        echo "  方式1: 使用 Homebrew"
        echo "    brew install go"
        echo ""
        echo "  方式2: 从官网下载"
        echo "    https://go.dev/dl/"
        echo ""
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go 环境已安装: $GO_VERSION"
}

# 检查配置文件
check_config() {
    print_info "检查配置..."

    # 如果存在 .env 文件，提示用户
    if [ -f ".env" ]; then
        print_success "检测到 .env 配置文件"

        # 加载 .env 文件
        export $(cat .env | grep -v '^#' | xargs)
    else
        print_warning "未找到 .env 配置文件，将使用默认配置"
        echo ""
        echo "如需自定义配置，请创建 .env 文件："
        echo "  cp .env.example .env"
        echo "  然后编辑 .env 文件修改配置"
        echo ""
    fi

    # 检查必要的环境变量
    if [ -z "$WEBHOOK_URL" ]; then
        print_error "未设置 WEBHOOK_URL 环境变量！"
        echo ""
        echo "请设置企业微信 Webhook URL："
        echo "  export WEBHOOK_URL=\"https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的密钥\""
        echo ""
        echo "或者创建 .env 文件并添加："
        echo "  WEBHOOK_URL=https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的密钥"
        echo ""
        exit 1
    fi

    print_success "配置检查通过"
}

# 编译程序
build_app() {
    print_info "编译程序..."

    if go build -o sunset .; then
        print_success "编译成功"
    else
        print_error "编译失败"
        exit 1
    fi
}

# 显示配置信息
show_config() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo -e "${GREEN}  📋 当前配置${NC}"
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
    echo "  城市: ${CITY:-上海市-上海}"
    echo "  纬度: ${LATITUDE:-31.2304}"
    echo "  经度: ${LONGITUDE:-121.4737}"
    echo "  时区: 北京时间 (UTC+8)"

    if [ "${USE_SUNSET_TIME:-false}" = "true" ]; then
        echo "  触发模式: 日落时间自动触发"
        echo "  提前时间: ${SUNSET_ADVANCE_MINUTES:-30} 分钟"
    else
        echo "  触发模式: 固定时间触发"
        echo "  触发时间: ${SCHEDULE_HOUR:-17}:${SCHEDULE_MINUTE:-30}"
    fi

    echo "  HTTP端口: ${PORT:-8080}"
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════${NC}"
    echo ""
}

# 启动服务
start_service() {
    print_info "启动服务..."
    echo ""

    # 显示配置
    show_config

    # 提示用户
    print_info "服务即将启动..."
    print_info "按 Ctrl+C 可以停止服务"
    echo ""

    # 设置默认环境变量（如果未设置）
    export CITY="${CITY:-上海市-上海}"
    export LATITUDE="${LATITUDE:-31.2304}"
    export LONGITUDE="${LONGITUDE:-121.4737}"
    export SCHEDULE_HOUR="${SCHEDULE_HOUR:-17}"
    export SCHEDULE_MINUTE="${SCHEDULE_MINUTE:-30}"
    export USE_SUNSET_TIME="${USE_SUNSET_TIME:-false}"
    export SUNSET_ADVANCE_MINUTES="${SUNSET_ADVANCE_MINUTES:-30}"
    export PORT="${PORT:-8080}"

    # 运行服务
    ./sunset
}

# 主函数
main() {
    print_header

    # 检查 Go 环境
    check_go

    # 检查配置
    check_config

    # 编译程序
    build_app

    # 启动服务
    start_service
}

# 运行主函数
main

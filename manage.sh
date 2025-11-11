#!/bin/bash

# Sunset 服务管理脚本
# 用法: ./manage.sh [command]
# 命令: start|stop|restart|logs|status|config|test|update

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker Compose 命令
get_compose_cmd() {
    if docker compose version &> /dev/null; then
        echo "docker compose"
    elif command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    else
        print_error "Docker Compose 未安装"
        exit 1
    fi
}

COMPOSE_CMD=$(get_compose_cmd)

# 启动服务
start_service() {
    print_info "启动服务..."
    $COMPOSE_CMD up -d
    print_success "服务已启动"
    sleep 2
    show_status
}

# 停止服务
stop_service() {
    print_info "停止服务..."
    $COMPOSE_CMD down
    print_success "服务已停止"
}

# 重启服务
restart_service() {
    print_info "重启服务..."
    $COMPOSE_CMD restart
    print_success "服务已重启"
    sleep 2
    show_status
}

# 查看日志
show_logs() {
    print_info "显示服务日志 (Ctrl+C 退出)..."
    echo ""
    $COMPOSE_CMD logs -f sunset
}

# 查看状态
show_status() {
    print_info "服务状态:"
    echo ""
    $COMPOSE_CMD ps
    echo ""

    # 尝试检查健康状态
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        health=$(curl -s http://localhost:8080/health)
        print_success "服务健康检查: OK"
        echo "$health" | python3 -m json.tool 2>/dev/null || echo "$health"
    else
        print_error "服务健康检查: FAILED (服务可能未启动)"
    fi
}

# 查看配置
show_config() {
    print_info "当前配置:"
    echo ""

    if curl -s http://localhost:8080/config > /dev/null 2>&1; then
        config=$(curl -s http://localhost:8080/config)
        echo "$config" | python3 -m json.tool 2>/dev/null || echo "$config"
    else
        print_error "无法获取配置 (服务可能未启动)"
    fi
}

# 测试推送
test_push() {
    print_info "触发测试推送..."
    echo ""

    response=$(curl -s -X POST http://localhost:8080/trigger-push)

    if echo "$response" | grep -q "success"; then
        print_success "推送测试成功！"
        echo "$response" | python3 -m json.tool 2>/dev/null || echo "$response"
    else
        print_error "推送测试失败"
        echo "$response"
    fi
}

# 更新服务
update_service() {
    print_info "更新服务 (拉取最新代码并重新构建)..."

    # 如果是 git 仓库，拉取最新代码
    if [ -d ".git" ]; then
        print_info "拉取最新代码..."
        git pull
    fi

    print_info "重新构建并启动..."
    $COMPOSE_CMD up -d --build
    print_success "服务已更新"
    sleep 2
    show_status
}

# 显示帮助信息
show_help() {
    echo "Sunset 服务管理脚本"
    echo ""
    echo "用法: ./manage.sh [command]"
    echo ""
    echo "可用命令:"
    echo "  start     - 启动服务"
    echo "  stop      - 停止服务"
    echo "  restart   - 重启服务"
    echo "  logs      - 查看实时日志"
    echo "  status    - 查看服务状态"
    echo "  config    - 查看当前配置"
    echo "  test      - 测试推送功能"
    echo "  update    - 更新服务（重新构建）"
    echo ""
    echo "示例:"
    echo "  ./manage.sh start"
    echo "  ./manage.sh logs"
    echo "  ./manage.sh test"
}

# 主函数
main() {
    case "${1:-help}" in
        start)
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            restart_service
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        config)
            show_config
            ;;
        test)
            test_push
            ;;
        update)
            update_service
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"

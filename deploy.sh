#!/bin/bash

# Sunset 一键部署脚本
# 用法: ./deploy.sh

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 打印标题
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  Sunset 火烧云推送服务 - 一键部署${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# 检查 Docker 是否安装
check_docker() {
    print_info "检查 Docker 环境..."
    if ! command -v docker &> /dev/null; then
        print_error "Docker 未安装，请先安装 Docker"
        echo "安装指南: https://docs.docker.com/get-docker/"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker 未运行，请启动 Docker"
        exit 1
    fi

    print_success "Docker 环境检查通过"
}

# 检查 Docker Compose 是否可用
check_docker_compose() {
    print_info "检查 Docker Compose..."
    if docker compose version &> /dev/null; then
        print_success "Docker Compose 可用 (内置版本)"
        return 0
    elif command -v docker-compose &> /dev/null; then
        print_success "Docker Compose 可用 (独立版本)"
        return 0
    else
        print_error "Docker Compose 未安装"
        exit 1
    fi
}

# 检查必需文件
check_files() {
    print_info "检查必需文件..."

    if [ ! -f "docker-compose.yml" ]; then
        print_error "docker-compose.yml 文件不存在"
        exit 1
    fi

    if [ ! -f "Dockerfile" ]; then
        print_error "Dockerfile 文件不存在"
        exit 1
    fi

    if [ ! -f "main.go" ]; then
        print_error "main.go 文件不存在"
        exit 1
    fi

    print_success "所有必需文件检查通过"
}

# 停止并删除旧容器
cleanup_old_containers() {
    print_info "清理旧容器..."

    if docker ps -a | grep -q sunset-app; then
        print_warning "发现旧容器，正在停止并删除..."
        docker compose down || docker-compose down 2>/dev/null || true
        print_success "旧容器已清理"
    else
        print_info "未发现旧容器"
    fi
}

# 构建并启动服务
build_and_start() {
    print_info "构建并启动服务..."
    echo ""

    # 使用新版或旧版 docker-compose
    if docker compose version &> /dev/null; then
        docker compose up -d --build
    else
        docker-compose up -d --build
    fi

    echo ""
    print_success "服务启动成功！"
}

# 等待服务启动
wait_for_service() {
    print_info "等待服务启动..."

    max_attempts=30
    attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_success "服务已就绪"
            return 0
        fi

        attempt=$((attempt + 1))
        echo -n "."
        sleep 1
    done

    echo ""
    print_warning "服务健康检查超时，但容器可能正在启动"
    return 1
}

# 显示服务状态
show_status() {
    echo ""
    print_info "服务状态:"
    echo ""

    if docker compose version &> /dev/null; then
        docker compose ps
    else
        docker-compose ps
    fi
}

# 显示服务信息
show_info() {
    echo ""
    print_success "🎉 部署完成！"
    echo ""
    echo -e "${BLUE}服务信息:${NC}"
    echo "  - 容器名称: sunset-app"
    echo "  - 访问端口: 8080"
    echo ""
    echo -e "${BLUE}可用的 API 接口:${NC}"
    echo "  - 健康检查:   http://localhost:8080/health"
    echo "  - 查询配置:   http://localhost:8080/config"
    echo "  - 主动触发:   http://localhost:8080/trigger-push"
    echo "  - 日落时间:   http://localhost:8080/sunset-time"
    echo ""
    echo -e "${BLUE}常用命令:${NC}"
    echo "  - 查看日志:   docker compose logs -f sunset"
    echo "  - 停止服务:   docker compose down"
    echo "  - 重启服务:   docker compose restart"
    echo "  - 查看状态:   docker compose ps"
    echo ""
    echo -e "${BLUE}测试服务:${NC}"
    echo "  curl http://localhost:8080/config"
    echo ""
}

# 显示配置信息
show_config() {
    print_info "正在获取当前配置..."
    echo ""

    if curl -s http://localhost:8080/config > /dev/null 2>&1; then
        config=$(curl -s http://localhost:8080/config)
        echo -e "${BLUE}当前配置:${NC}"
        echo "$config" | python3 -m json.tool 2>/dev/null || echo "$config"
    else
        print_warning "无法获取配置信息（服务可能还在启动中）"
        echo "稍后可以通过以下命令查看:"
        echo "  curl http://localhost:8080/config"
    fi
}

# 主函数
main() {
    print_header

    # 执行检查
    check_docker
    check_docker_compose
    check_files

    echo ""

    # 清理并部署
    cleanup_old_containers
    build_and_start

    # 等待服务就绪
    wait_for_service

    # 显示状态和信息
    show_status
    show_config
    show_info

    print_success "✅ 一键部署完成！"
    echo ""
}

# 捕获错误
trap 'print_error "部署过程中发生错误"; exit 1' ERR

# 运行主函数
main

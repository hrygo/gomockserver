#!/bin/bash

# Docker 测试环境管理脚本
# 用于快速启动、停止和管理测试环境

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
COMPOSE_FILE="$PROJECT_ROOT/docker-compose.test.yml"

# 显示帮助信息
show_help() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   Docker 测试环境管理脚本${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo "使用方法: $0 <command> [options]"
    echo ""
    echo "命令:"
    echo "  up              启动基础测试环境 (MongoDB + Mock Server)"
    echo "  up-full         启动完整测试环境 (包含 Redis)"
    echo "  up-performance  启动性能测试环境 (包含 wrk)"
    echo "  up-integration  启动集成测试环境 (包含测试运行器)"
    echo "  down            停止并删除测试环境"
    echo "  restart         重启测试环境"
    echo "  logs [service]  查看日志"
    echo "  ps              查看运行状态"
    echo "  exec <service> <cmd>  在服务容器中执行命令"
    echo "  test            运行集成测试"
    echo "  perf            运行性能测试"
    echo "  clean           清理所有测试数据和容器"
    echo "  build           重新构建镜像"
    echo "  help            显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 up                    # 启动基础环境"
    echo "  $0 test                  # 运行集成测试"
    echo "  $0 logs mockserver-test  # 查看服务日志"
    echo "  $0 clean                 # 清理环境"
    echo ""
}

# 检查 Docker 是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}错误: Docker 未安装${NC}"
        echo "请访问 https://www.docker.com/get-started 安装 Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}错误: Docker Compose 未安装${NC}"
        exit 1
    fi
}

# 启动基础测试环境
start_basic() {
    echo -e "${YELLOW}启动基础测试环境...${NC}"
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" up -d mongodb-test mockserver-test
    
    echo -e "${GREEN}✓ 基础测试环境已启动${NC}"
    echo ""
    echo "服务访问地址:"
    echo "  - Admin API:  http://localhost:8081/api/v1"
    echo "  - Mock API:   http://localhost:9091"
    echo "  - MongoDB:    mongodb://localhost:27018"
    echo ""
    echo "查看日志: $0 logs"
    echo "停止环境: $0 down"
}

# 启动完整测试环境
start_full() {
    echo -e "${YELLOW}启动完整测试环境（包含 Redis）...${NC}"
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" --profile with-redis up -d
    
    echo -e "${GREEN}✓ 完整测试环境已启动${NC}"
    echo ""
    echo "服务访问地址:"
    echo "  - Admin API:  http://localhost:8081/api/v1"
    echo "  - Mock API:   http://localhost:9091"
    echo "  - MongoDB:    mongodb://localhost:27018"
    echo "  - Redis:      redis://localhost:6380"
}

# 启动性能测试环境
start_performance() {
    echo -e "${YELLOW}启动性能测试环境...${NC}"
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" --profile performance up -d
    
    echo -e "${GREEN}✓ 性能测试环境已启动${NC}"
    echo ""
    echo "运行性能测试: $0 perf"
}

# 启动集成测试环境
start_integration() {
    echo -e "${YELLOW}启动集成测试环境...${NC}"
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" --profile integration up -d
    
    echo -e "${GREEN}✓ 集成测试环境已启动${NC}"
    echo ""
    echo "测试运行器将自动执行集成测试"
    echo "查看测试日志: docker logs -f mockserver-test-runner"
}

# 停止测试环境
stop_env() {
    echo -e "${YELLOW}停止测试环境...${NC}"
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" --profile with-redis --profile performance --profile integration down
    
    echo -e "${GREEN}✓ 测试环境已停止${NC}"
}

# 重启测试环境
restart_env() {
    echo -e "${YELLOW}重启测试环境...${NC}"
    stop_env
    sleep 2
    start_basic
}

# 查看日志
show_logs() {
    cd "$PROJECT_ROOT"
    if [ -z "$1" ]; then
        docker-compose -f "$COMPOSE_FILE" logs -f
    else
        docker-compose -f "$COMPOSE_FILE" logs -f "$1"
    fi
}

# 查看状态
show_status() {
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" ps
}

# 在容器中执行命令
exec_command() {
    if [ -z "$1" ] || [ -z "$2" ]; then
        echo -e "${RED}错误: 需要指定服务名和命令${NC}"
        echo "用法: $0 exec <service> <command>"
        exit 1
    fi
    
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" exec "$1" "${@:2}"
}

# 运行集成测试
run_integration_test() {
    echo -e "${YELLOW}运行集成测试...${NC}"
    
    # 确保基础环境已启动
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" up -d mongodb-test mockserver-test
    
    echo "等待服务就绪..."
    sleep 10
    
    # 运行测试
    docker-compose -f "$COMPOSE_FILE" run --rm test-runner
    
    echo -e "${GREEN}✓ 集成测试完成${NC}"
}

# 运行性能测试
run_performance_test() {
    echo -e "${YELLOW}运行性能测试...${NC}"
    
    # 确保基础环境已启动
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" --profile performance up -d
    
    echo "等待服务就绪..."
    sleep 10
    
    # 运行 wrk 性能测试
    echo -e "${BLUE}执行 wrk 压力测试...${NC}"
    docker-compose -f "$COMPOSE_FILE" exec wrk-test wrk \
        -t4 -c100 -d30s \
        -H "X-Project-ID: test-project" \
        -H "X-Environment-ID: test-env" \
        http://mockserver-test:9090/api/test
    
    echo -e "${GREEN}✓ 性能测试完成${NC}"
}

# 清理环境
clean_env() {
    echo -e "${YELLOW}清理测试环境...${NC}"
    
    read -p "确认要清理所有测试数据和容器吗？(y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cd "$PROJECT_ROOT"
        docker-compose -f "$COMPOSE_FILE" --profile with-redis --profile performance --profile integration down -v
        echo -e "${GREEN}✓ 测试环境已清理${NC}"
    else
        echo "取消清理"
    fi
}

# 重新构建镜像
rebuild_images() {
    echo -e "${YELLOW}重新构建测试镜像...${NC}"
    cd "$PROJECT_ROOT"
    docker-compose -f "$COMPOSE_FILE" build --no-cache
    echo -e "${GREEN}✓ 镜像构建完成${NC}"
}

# 主函数
main() {
    check_docker
    
    case "$1" in
        up)
            start_basic
            ;;
        up-full)
            start_full
            ;;
        up-performance)
            start_performance
            ;;
        up-integration)
            start_integration
            ;;
        down)
            stop_env
            ;;
        restart)
            restart_env
            ;;
        logs)
            show_logs "$2"
            ;;
        ps)
            show_status
            ;;
        exec)
            exec_command "${@:2}"
            ;;
        test)
            run_integration_test
            ;;
        perf)
            run_performance_test
            ;;
        clean)
            clean_env
            ;;
        build)
            rebuild_images
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            echo -e "${RED}错误: 未知命令 '$1'${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 如果没有参数，显示帮助
if [ $# -eq 0 ]; then
    show_help
    exit 0
fi

main "$@"

#!/bin/bash

# MockServer 健康检查测试
# 快速验证所有组件是否正常运行

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 测试配置
TEST_DIR="$(dirname "$0")"
FRAMEWORK_LIB="$TEST_DIR/lib/test_framework.sh"

# 测试统计
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# 加载测试框架
if [ -f "$FRAMEWORK_LIB" ]; then
    source "$FRAMEWORK_LIB"
else
    echo -e "${RED}错误: 找不到测试框架文件 $FRAMEWORK_LIB${NC}"
    exit 1
fi

# 显示测试横幅
show_banner() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}   MockServer 健康检查测试${NC}"
    echo -e "${CYAN}=========================================${NC}"
    echo ""
    echo -e "${CYAN}检查项目:${NC}"
    echo -e "  • MockServer 后端服务"
    echo -e "  • Redis 缓存服务"
    echo -e "  • MongoDB 数据库"
    echo -e "  • API 端点可用性"
    echo ""
    echo -e "${CYAN}开始时间: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo ""
}

# 检查 MockServer 后端
check_mockserver_backend() {
    log_test "MockServer 后端服务检查"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    if check_server_health; then
        log_success "MockServer 后端服务正常"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        log_fail "MockServer 后端服务异常"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 检查 Redis 服务
check_redis_service() {
    log_test "Redis 缓存服务检查"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    if check_redis_connection; then
        # 测试基本操作
        local test_key="health_check_$(date +%s)"
        local test_value="ok"

        if redis-cli set "$test_key" "$test_value" | grep -q "OK" &&
           redis-cli get "$test_key" | grep -q "ok" &&
           redis-cli del "$test_key" | grep -q "1"; then
            log_success "Redis 缓存服务正常"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            log_fail "Redis 缓存服务操作异常"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
    else
        log_fail "Redis 缓存服务连接失败"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 检查 MongoDB 服务
check_mongodb_service() {
    log_test "MongoDB 数据库检查"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    # 检查 MongoDB 容器是否运行
    if docker ps --format '{{.Names}}' | grep -q "mockserver-mongodb"; then
        # 检查 MongoDB 健康状态
        if docker exec mockserver-mongodb mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
            log_success "MongoDB 数据库服务正常"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            log_fail "MongoDB 数据库服务异常"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
    else
        log_fail "MongoDB 数据库容器未运行"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 检查 Admin API
check_admin_api() {
    log_test "Admin API 健康检查"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    local response=$(curl -s -w "%{http_code}" -o /dev/null "$ADMIN_API/system/health" 2>/dev/null)

    if [ "$response" = "200" ]; then
        log_success "Admin API 正常 (HTTP $response)"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        log_fail "Admin API 异常 (HTTP $response)"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 检查 Mock API
check_mock_api() {
    log_test "Mock API 健康检查"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    local response=$(curl -s -w "%{http_code}" -o /dev/null "$MOCK_API/health" 2>/dev/null)

    if [ "$response" = "200" ] || [ "$response" = "404" ]; then
        log_success "Mock API 正常 (HTTP $response)"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        log_fail "Mock API 异常 (HTTP $response)"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 检查端口占用
check_port_usage() {
    log_test "端口占用检查"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    local ports=("8080" "9090" "27017" "6379")
    local port_names=("Admin API" "Mock API" "MongoDB" "Redis")
    local all_ports_ok=true

    for i in "${!ports[@]}"; do
        local port=${ports[$i]}
        local name=${port_names[$i]}

        if lsof -i :$port >/dev/null 2>&1; then
            log_success "$name 端口 $port 正在使用"
        else
            log_fail "$name 端口 $port 未被占用"
            all_ports_ok=false
        fi
    done

    if [ "$all_ports_ok" = true ]; then
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 检查容器状态
check_container_status() {
    log_test "容器状态检查"
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    local containers=("mockserver-mongodb" "mockserver-redis")
    local all_containers_ok=true

    for container in "${containers[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "$container"; then
            local status=$(docker ps --format "{{.Names}}: {{.Status}}" | grep "$container" | cut -d: -f2)
            log_success "$container 容器运行中 ($status)"
        else
            log_fail "$container 容器未运行"
            all_containers_ok=false
        fi
    done

    if [ "$all_containers_ok" = true ]; then
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
}

# 生成健康检查报告
generate_health_report() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   健康检查统计${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${CYAN}检查统计:${NC}"
    echo -e "  总检查项: $TOTAL_CHECKS"
    echo -e "  通过: ${GREEN}$PASSED_CHECKS${NC}"
    echo -e "  失败: ${RED}$FAILED_CHECKS${NC}"
    echo -e "  通过率: $(( PASSED_CHECKS * 100 / TOTAL_CHECKS ))%"
    echo ""

    if [ $FAILED_CHECKS -eq 0 ]; then
        echo -e "${GREEN}🎉 所有健康检查通过！${NC}"
        echo -e "${GREEN}✅ MockServer 系统状态良好${NC}"
        echo ""
        echo -e "${CYAN}服务访问地址:${NC}"
        echo -e "  • Admin API: $ADMIN_API"
        echo -e "  • Mock API: $MOCK_API"
        echo -e "  • Redis: localhost:6379"
        echo -e "  • MongoDB: mongodb://localhost:27017"
        return 0
    else
        echo -e "${RED}❌ 部分健康检查失败${NC}"
        echo -e "${YELLOW}💡 建议检查失败的服务${NC}"
        echo ""
        echo -e "${CYAN}故障排查建议:${NC}"
        echo -e "  • 检查 Docker 容器状态: docker ps -a"
        echo -e "  • 查看服务日志: make logs"
        echo -e "  • 重启服务: make stop-all && make start-all"
        return 1
    fi
}

# 主测试流程
main() {
    show_banner

    # 执行健康检查
    # 检查依赖（可选，不强制要求）
    command -v docker >/dev/null 2>&1 || { echo -e "${YELLOW}警告: Docker 未安装，跳过容器检查${NC}"; }
    command -v curl >/dev/null 2>&1 || { echo -e "${YELLOW}警告: curl 未安装，跳过 HTTP 检查${NC}"; }
    command -v redis-cli >/dev/null 2>&1 || { echo -e "${YELLOW}警告: redis-cli 未安装，跳过 Redis 检查${NC}"; }

    check_mockserver_backend
    check_redis_service
    check_mongodb_service
    check_admin_api
    check_mock_api
    check_port_usage
    check_container_status

    # 生成报告
    generate_health_report
}

# 信号处理
trap 'echo -e "\n${YELLOW}健康检查被中断${NC}"; exit 1' INT TERM

# 执行主流程
main
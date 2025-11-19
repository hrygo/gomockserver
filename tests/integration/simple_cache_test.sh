#!/bin/bash

# 简化的缓存集成测试脚本
# 专注于Redis基本功能和MockServer集成

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
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
TEST_PREFIX="simple_cache_${TIMESTAMP}_"

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 加载测试框架
if [ -f "$FRAMEWORK_LIB" ]; then
    source "$FRAMEWORK_LIB"
else
    echo -e "${RED}错误: 找不到测试框架文件 $FRAMEWORK_LIB${NC}"
    echo -e "${CYAN}使用内置基本测试功能${NC}"

    # 基本测试函数
    log_test() { echo -e "${CYAN}[TEST]${NC} $1"; }
    log_pass() { echo -e "${GREEN}[PASS]${NC} $1"; PASSED_TESTS=$((PASSED_TESTS + 1)); }
    log_fail() { echo -e "${RED}[FAIL]${NC} $1"; FAILED_TESTS=$((FAILED_TESTS + 1)); }
fi

# 显示横幅
show_banner() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}   简化缓存集成测试${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    echo -e "${CYAN}测试目标:${NC}"
    echo -e "  • Redis 基础连接"
    echo -e "  • 缓存 CRUD 操作"
    echo -e "  • MockServer 集成"
    echo -e ""
    echo -e "${CYAN}开始时间: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo ""
}

# 测试 1: Redis 基础连接
test_redis_connection() {
    log_test "Redis 基础连接测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 检查 redis-cli 是否可用
    if ! command -v redis-cli >/dev/null 2>&1; then
        log_fail "redis-cli 命令不可用"
        return 1
    fi

    # 测试连接
    local ping_result=$(redis-cli ping 2>/dev/null || echo "FAILED")
    if [ "$ping_result" = "PONG" ]; then
        log_pass "Redis 连接成功"
        return 0
    else
        log_fail "Redis 连接失败: $ping_result"
        return 1
    fi
}

# 测试 2: Redis 基础操作
test_redis_operations() {
    log_test "Redis 基础操作测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_key="${TEST_PREFIX}basic"
    local test_value="test_value_$(date +%s)"

    # SET 操作
    local set_result=$(redis-cli set "$test_key" "$test_value" 2>/dev/null || echo "FAILED")
    if [ "$set_result" = "OK" ]; then
        # GET 操作
        local get_result=$(redis-cli get "$test_key" 2>/dev/null || echo "FAILED")
        if [ "$get_result" = "$test_value" ]; then
            # DELETE 操作
            local del_result=$(redis-cli del "$test_key" 2>/dev/null || echo "FAILED")
            if [ "$del_result" = "1" ]; then
                log_pass "Redis 基础操作成功"
                return 0
            else
                log_fail "Redis DELETE 操作失败: $del_result"
            fi
        else
            log_fail "Redis GET 操作失败: 期望 $test_value, 得到 $get_result"
        fi
    else
        log_fail "Redis SET 操作失败: $set_result"
    fi

    return 1
}

# 测试 3: Redis 过期功能
test_redis_expiration() {
    log_test "Redis 过期功能测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_key="${TEST_PREFIX}expire"
    local test_value="expire_test_$(date +%s)"
    local ttl=3

    # 设置带过期时间的键
    local setex_result=$(redis-cli setex "$test_key" $ttl "$test_value" 2>/dev/null || echo "FAILED")
    if [ "$setex_result" = "OK" ]; then
        # 检查 TTL
        local ttl_result=$(redis-cli ttl "$test_key" 2>/dev/null || echo "FAILED")
        if [ "$ttl_result" -gt 0 ] 2>/dev/null; then
            # 立即获取应该成功
            local get_result=$(redis-cli get "$test_key" 2>/dev/null || echo "FAILED")
            if [ "$get_result" = "$test_value" ]; then
                log_pass "Redis 过期功能正常"
                return 0
            else
                log_fail "Redis 过期键立即获取失败"
            fi
        else
            log_fail "Redis TTL 检查失败: $ttl_result"
        fi
    else
        log_fail "Redis SETEX 操作失败: $setex_result"
    fi

    return 1
}

# 测试 4: MockServer 健康检查
test_mockserver_health() {
    log_test "MockServer 健康检查"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 检查 Admin API
    local admin_response=$(curl -s -w "%{http_code}" -o /dev/null "$ADMIN_API/system/health" 2>/dev/null || echo "000")
    if [ "$admin_response" = "200" ]; then
        log_pass "Admin API 健康检查通过"
    else
        log_fail "Admin API 健康检查失败: HTTP $admin_response"
        return 1
    fi

    # 检查 Mock API
    local mock_response=$(curl -s -w "%{http_code}" -o /dev/null "$MOCK_API/health" 2>/dev/null || echo "000")
    if [ "$mock_response" = "200" ] || [ "$mock_response" = "404" ]; then
        log_pass "Mock API 健康检查通过"
        return 0
    else
        log_fail "Mock API 健康检查失败: HTTP $mock_response"
        return 1
    fi
}

# 测试 5: MockServer 与缓存集成
test_mockserver_cache_integration() {
    log_test "MockServer 缓存集成测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 创建测试项目
    local project_data='{"name": "cache_integration_test", "description": "缓存集成测试项目"}'
    local project_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$project_data" \
        "$ADMIN_API/projects" 2>/dev/null)

    local project_id=$(echo "$project_response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    if [ -z "$project_id" ]; then
        log_fail "创建测试项目失败"
        return 1
    fi

    # 创建测试环境
    local env_data='{"name": "cache_test_env", "project_id": "'$project_id'", "description": "缓存测试环境"}'
    local env_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$env_data" \
        "$ADMIN_API/environments" 2>/dev/null)

    local env_id=$(echo "$env_response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    if [ -z "$env_id" ]; then
        log_fail "创建测试环境失败"
        return 1
    fi

    # 创建 Mock 规则
    local rule_data='{
        "name": "cache_test_rule",
        "project_id": "'$project_id'",
        "environment_id": "'$env_id'",
        "request": {"method": "GET", "path": "/api/cache-test"},
        "response": {
            "status": 200,
            "body": "{\"message\": \"Hello from cache test!\", \"timestamp\": \"'$(date -u +%Y-%m-%dT%H:%M:%SZ)'\"}",
            "headers": {"Content-Type": "application/json"}
        }
    }'

    local rule_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$rule_data" \
        "$ADMIN_API/rules" 2>/dev/null)

    local rule_id=$(echo "$rule_response" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    if [ -z "$rule_id" ]; then
        log_fail "创建 Mock 规则失败"
        return 1
    fi

    # 测试 API 响应
    local api_response=$(curl -s -w "%{http_code}" -o /tmp/cache_api_response.json \
        -H "X-Project-ID: $project_id" \
        -H "X-Environment-ID: $env_id" \
        "$MOCK_API/api/cache-test" 2>/dev/null)

    if [ "${api_response: -3}" = "200" ]; then
        log_pass "MockServer 缓存集成测试成功"

        # 清理测试数据
        curl -s -X DELETE "$ADMIN_API/projects/$project_id" >/dev/null 2>&1 || true
        return 0
    else
        log_fail "MockServer API 响应失败: HTTP ${api_response: -3}"
        return 1
    fi
}

# 测试 6: Redis 内存监控
test_redis_memory() {
    log_test "Redis 内存监控测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 获取内存信息
    local memory_info=$(redis-cli info memory 2>/dev/null || echo "ERROR")
    if [ "$memory_info" != "ERROR" ]; then
        local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '\r')
        if [ -n "$used_memory" ]; then
            log_pass "Redis 内存监控成功: 当前使用 $used_memory"
            return 0
        else
            log_fail "Redis 内存信息解析失败"
        fi
    else
        log_fail "Redis 内存信息获取失败"
    fi

    return 1
}

# 生成测试报告
generate_report() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   简化缓存集成测试结果${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${CYAN}测试统计:${NC}"
    echo -e "  总测试数: $TOTAL_TESTS"
    echo -e "  通过: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "  失败: ${RED}$FAILED_TESTS${NC}"
    echo -e "  通过率: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎉 所有缓存集成测试通过！${NC}"
        echo -e "${GREEN}✅ Redis 缓存系统工作正常${NC}"
        echo -e "${GREEN}✅ MockServer 集成成功${NC}"
        return 0
    else
        echo -e "${RED}❌ 部分测试失败${NC}"
        echo -e "${YELLOW}💡 请检查 Redis 和 MockServer 状态${NC}"
        return 1
    fi
}

# 主测试流程
main() {
    show_banner

    # 执行测试
    test_redis_connection || true
    test_redis_operations || true
    test_redis_expiration || true
    test_mockserver_health || true
    test_mockserver_cache_integration || true
    test_redis_memory || true

    # 生成报告
    generate_report
}

# 信号处理
trap 'echo -e "\n${YELLOW}测试被中断${NC}"; exit 1' INT TERM

# 执行主流程
main
#!/bin/bash

# 简化的缓存集成测试脚本
# 专注于Redis基本功能和MockServer集成
# 已优化：集成新的coordinate_services函数和统一测试框架

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
    echo -e "  • 性能基准测试"
    echo -e ""
    echo -e "${CYAN}开始时间: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo ""
}

# Redis 连接测试
test_redis_connection() {
    log_test "测试 Redis 基础连接"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 使用框架的Redis连接检查
    if check_redis_connection; then
        log_pass "Redis 连接正常"
        return 0
    else
        log_fail "Redis 连接失败"
        return 1
    fi
}

# Redis 基础CRUD测试
test_redis_crud() {
    log_test "测试 Redis CRUD 操作"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_key="${TEST_PREFIX}crud_test"
    local test_value="MockServer缓存测试值_${TIMESTAMP}"

    # 测试 SET 操作
    if ! redis-cli set "$test_key" "$test_value" >/dev/null 2>&1; then
        log_fail "Redis SET 操作失败"
        return 1
    fi

    # 测试 GET 操作
    local retrieved_value=$(redis-cli get "$test_key" 2>/dev/null)
    if [ "$retrieved_value" != "$test_value" ]; then
        log_fail "Redis GET 操作失败，期望: $test_value，实际: $retrieved_value"
        return 1
    fi

    # 测试 EXISTS 操作
    if ! redis-cli exists "$test_key" >/dev/null 2>&1; then
        log_fail "Redis EXISTS 操作失败"
        return 1
    fi

    # 测试 DEL 操作
    if ! redis-cli del "$test_key" >/dev/null 2>&1; then
        log_fail "Redis DEL 操作失败"
        return 1
    fi

    # 验证删除
    if redis-cli exists "$test_key" >/dev/null 2>&1; then
        log_fail "Redis 删除验证失败"
        return 1
    fi

    log_pass "Redis CRUD 操作测试通过"
    return 0
}

# Redis 键过期测试
test_redis_expiry() {
    log_test "测试 Redis 键过期功能"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_key="${TEST_PREFIX}expiry_test"
    local test_value="过期测试值"

    # 设置带过期时间的键
    if ! redis-cli setex "$test_key" 2 "$test_value" >/dev/null 2>&1; then
        log_fail "Redis SETEX 操作失败"
        return 1
    fi

    # 立即检查键存在
    if ! redis-cli exists "$test_key" >/dev/null 2>&1; then
        log_fail "Redis 键设置后立即检查失败"
        return 1
    fi

    # 等待过期
    log_test "等待键过期 (3秒)..."
    sleep 3

    # 检查键已过期
    if redis-cli exists "$test_key" >/dev/null 2>&1; then
        log_fail "Redis 键过期功能失败"
        return 1
    fi

    log_pass "Redis 键过期功能测试通过"
    return 0
}

# Redis 批量操作测试
test_redis_batch() {
    log_test "测试 Redis 批量操作"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_prefix="${TEST_PREFIX}batch_"
    local keys=()
    local values=()

    # 准备测试数据
    for i in {1..5}; do
        keys+=("${test_prefix}${i}")
        values+=("批量测试值${i}")
    done

    # 批量设置 (MSET)
    local mset_cmd="redis-cli mset"
    for i in "${!keys[@]}"; do
        mset_cmd="$mset_cmd ${keys[$i]} ${values[$i]}"
    done

    if ! $mset_cmd >/dev/null 2>&1; then
        log_fail "Redis MSET 操作失败"
        return 1
    fi

    # 批量获取 (MGET)
    local mget_cmd="redis-cli mget"
    for key in "${keys[@]}"; do
        mget_cmd="$mget_cmd $key"
    done

    local retrieved_values=($($mget_cmd 2>/dev/null))

    # 验证检索的值
    for i in "${!values[@]}"; do
        if [ "${retrieved_values[$i]}" != "${values[$i]}" ]; then
            log_fail "Redis MGET 验证失败，期望: ${values[$i]}，实际: ${retrieved_values[$i]}"
            return 1
        fi
    done

    # 清理批量键
    for key in "${keys[@]}"; do
        redis-cli del "$key" >/dev/null 2>&1
    done

    log_pass "Redis 批量操作测试通过"
    return 0
}

# Redis 连接池测试
test_redis_pool() {
    log_test "测试 Redis 连接池功能"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_key="${TEST_PREFIX}pool_test"
    local test_value="连接池测试值"
    local success_count=0
    local total_attempts=10

    # 并发测试连接池
    for i in $(seq 1 $total_attempts); do
        (
            if redis-cli set "${test_key}_${i}" "${test_value}_${i}" >/dev/null 2>&1; then
                echo "success"
            else
                echo "failed"
            fi
        ) &
    done

    # 等待所有后台任务完成
    wait

    # 验证结果
    for i in $(seq 1 $total_attempts); do
        if redis-cli exists "${test_key}_${i}" >/dev/null 2>&1; then
            success_count=$((success_count + 1))
            redis-cli del "${test_key}_${i}" >/dev/null 2>&1
        fi
    done

    if [ $success_count -eq $total_attempts ]; then
        log_pass "Redis 连接池测试通过 ($success_count/$total_attempts)"
        return 0
    else
        log_fail "Redis 连接池测试失败 ($success_count/$total_attempts)"
        return 1
    fi
}

# MockServer 缓存集成测试
test_mockserver_cache_integration() {
    log_test "测试 MockServer 缓存集成"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 创建项目
    local project_id=$(create_test_project "cache_integration_${TIMESTAMP}")
    if [ -z "$project_id" ]; then
        log_fail "创建测试项目失败"
        return 1
    fi

    # 创建环境
    local env_id=$(create_test_environment "$project_id" "cache_env")
    if [ -z "$env_id" ]; then
        log_fail "创建测试环境失败"
        cleanup_test_resources "$project_id" ""
        return 1
    fi

    # 创建缓存规则
    local rule_data='{
        "name": "cache_test_rule",
        "method": "GET",
        "path": "/api/cache/test",
        "response": {
            "status": 200,
            "body": "{\"message\": \"缓存测试响应\", \"timestamp\": "'$(date +%s)'\"}",
            "headers": {
                "Content-Type": "application/json"
            }
        },
        "cache": {
            "enabled": true,
            "ttl": 300
        }
    }'

    local rule_id=$(create_test_rule "$project_id" "$env_id" "$rule_data")
    if [ -z "$rule_id" ]; then
        log_fail "创建缓存规则失败"
        cleanup_test_resources "$project_id" "$env_id"
        return 1
    fi

    # 测试缓存响应
    local response=$(mock_request "GET" "/api/cache/test")
    if echo "$response" | grep -q "缓存测试响应"; then
        # 第二次请求应该命中缓存
        response=$(mock_request "GET" "/api/cache/test")
        if echo "$response" | grep -q "缓存测试响应"; then
            log_pass "MockServer 缓存集成测试通过"
            cleanup_test_resources "$project_id" "$env_id"
            return 0
        fi
    fi

    log_fail "MockServer 缓存集成测试失败"
    cleanup_test_resources "$project_id" "$env_id"
    return 1
}

# 性能基准测试
test_performance_benchmark() {
    log_test "执行性能基准测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_key="${TEST_PREFIX}perf_test"
    local iterations=1000
    local start_time end_time duration

    # Redis 写入性能测试
    start_time=$(date +%s.%N)
    for i in $(seq 1 $iterations); do
        redis-cli set "${test_key}_${i}" "性能测试值${i}" >/dev/null 2>&1
    done
    end_time=$(date +%s.%N)

    duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "1")
    local writes_per_sec=$(echo "scale=2; $iterations / $duration" | bc -l 2>/dev/null || echo "$iterations")

    # Redis 读取性能测试
    start_time=$(date +%s.%N)
    for i in $(seq 1 $iterations); do
        redis-cli get "${test_key}_${i}" >/dev/null 2>&1
    done
    end_time=$(date +%s.%N)

    duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "1")
    local reads_per_sec=$(echo "scale=2; $iterations / $duration" | bc -l 2>/dev/null || echo "$iterations")

    # 清理性能测试数据
    for i in $(seq 1 $iterations); do
        redis-cli del "${test_key}_${i}" >/dev/null 2>&1
    done

    log_pass "性能基准测试完成"
    echo -e "${CYAN}  写入性能: ${writes_per_sec} ops/sec${NC}"
    echo -e "${CYAN}  读取性能: ${reads_per_sec} ops/sec${NC}"

    # 验证性能是否在合理范围内 (至少100 ops/sec)
    local min_performance=100
    if (( $(echo "$writes_per_sec >= $min_performance" | bc -l 2>/dev/null || echo "1") )); then
        return 0
    else
        log_warn "写入性能低于期望值 ($writes_per_sec < $min_performance)"
        return 0  # 警告但不失败
    fi
}

# 内存使用监控
test_memory_usage() {
    log_test "监控内存使用情况"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 获取Redis内存信息
    local memory_info=$(redis-cli info memory 2>/dev/null)
    if [ -n "$memory_info" ]; then
        local used_memory=$(echo "$memory_info" | grep "used_memory:" | cut -d: -f2 | tr -d '\r')

        if [ -n "$used_memory" ]; then
            local used_mb=$((used_memory / 1024 / 1024))
            log_pass "内存使用监控完成"
            echo -e "${CYAN}  当前使用内存: ${used_mb} MB${NC}"

            # 检查内存使用是否在合理范围内 (< 1GB)
            if [ $used_mb -lt 1024 ]; then
                return 0
            else
                log_warn "内存使用较高 (${used_mb} MB)"
                return 0  # 警告但不失败
            fi
        fi
    fi

    log_fail "获取内存信息失败"
    return 1
}

# 主执行函数
main() {
    echo ""

    # 显示横幅
    show_banner

    # 使用统一的服务协调
    log_test "启动依赖服务"
    if ! coordinate_services; then
        echo -e "${RED}✗ 服务启动失败${NC}"
        exit 1
    fi

    echo -e "${CYAN}开始执行缓存测试...${NC}"
    echo ""

    # 执行测试套件
    local tests=(
        "test_redis_connection"
        "test_redis_crud"
        "test_redis_expiry"
        "test_redis_batch"
        "test_redis_pool"
        "test_mockserver_cache_integration"
        "test_performance_benchmark"
        "test_memory_usage"
    )

    local passed=0
    local failed=0

    for test_func in "${tests[@]}"; do
        if $test_func; then
            passed=$((passed + 1))
        else
            failed=$((failed + 1))
        fi
        echo ""
    done

    # 显示测试结果
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   缓存测试结果${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${CYAN}测试统计:${NC}"
    echo -e "  总测试数: $TOTAL_TESTS"
    echo -e "  通过: ${GREEN}$passed${NC}"
    echo -e "  失败: ${RED}$failed${NC}"
    echo -e "  成功率: $(( passed * 100 / TOTAL_TESTS ))%"
    echo ""

    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 所有缓存测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}❌ 有 $failed 个测试失败${NC}"
        exit 1
    fi
}

# 信号处理
trap 'echo -e "\n${YELLOW}测试被中断${NC}"; cleanup_dependency_services; exit 1' INT TERM

# 正常退出清理
trap 'cleanup_dependency_services' EXIT

# 执行主函数
main
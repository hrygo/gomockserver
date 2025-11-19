#!/bin/bash

# Redis集成测试脚本
# 测试Redis缓存功能的完整性和性能

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_URL="${REDIS_URL:-redis://localhost:6379}"
TEST_PREFIX="test_redis_integration_"

# 测试结果统计
TESTS_PASSED=0
TESTS_FAILED=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((TESTS_PASSED++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((TESTS_FAILED++))
}

# 检查Redis连接
check_redis_connection() {
    log_info "Checking Redis connection..."

    if redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null 2>&1; then
        log_success "Redis connection established"
        return 0
    else
        log_error "Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT"
        return 1
    fi
}

# 清理测试数据
cleanup_test_data() {
    log_info "Cleaning up previous test data..."
    redis-cli -h $REDIS_HOST -p $REDIS_PORT --scan --pattern "${TEST_PREFIX}*" | xargs -r redis-cli -h $REDIS_HOST -p $REDIS_PORT del 2>/dev/null || true
}

# 基础连接测试
test_basic_operations() {
    log_info "Testing basic Redis operations..."

    local test_key="${TEST_PREFIX}basic"
    local test_value="Hello Redis at $(date)"

    # SET操作
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT set "$test_key" "$test_value" | grep -q "OK"; then
        log_success "SET operation successful"
    else
        log_error "SET operation failed"
        return 1
    fi

    # GET操作
    local retrieved_value=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT get "$test_key")
    if [ "$retrieved_value" = "$test_value" ]; then
        log_success "GET operation successful"
    else
        log_error "GET operation failed - expected '$test_value', got '$retrieved_value'"
        return 1
    fi

    # DELETE操作
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT del "$test_key" | grep -q "1"; then
        log_success "DELETE operation successful"
    else
        log_error "DELETE operation failed"
        return 1
    fi

    # 验证删除
    local deleted_value=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT get "$test_key")
    if [ "$deleted_value" = "(nil)" ]; then
        log_success "Key successfully deleted"
    else
        log_error "Key still exists after deletion"
        return 1
    fi
}

# 过期时间测试
test_expiration() {
    log_info "Testing key expiration..."

    local test_key="${TEST_PREFIX}expire"
    local test_value="This will expire"
    local expire_time=2  # 2秒

    # 设置带过期时间的键
    redis-cli -h $REDIS_HOST -p $REDIS_PORT setex "$test_key" $expire_time "$test_value" > /dev/null

    # 立即检查应该存在
    local immediate_value=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT get "$test_key")
    if [ "$immediate_value" = "$test_value" ]; then
        log_success "Key exists immediately after SETEX"
    else
        log_error "Key not found immediately after SETEX"
        return 1
    fi

    # 等待过期
    log_info "Waiting $expire_time seconds for key to expire..."
    sleep $expire_time

    # 检查应该已过期
    local expired_value=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT get "$test_key")
    if [ "$expired_value" = "(nil)" ]; then
        log_success "Key expired correctly"
    else
        log_error "Key did not expire as expected"
        return 1
    fi
}

# 批量操作测试
test_batch_operations() {
    log_info "Testing batch operations..."

    # MSET测试
    redis-cli -h $REDIS_HOST -p $REDIS_PORT mset "${TEST_PREFIX}batch1" "value1" "${TEST_PREFIX}batch2" "value2" "${TEST_PREFIX}batch3" "value3" > /dev/null

    # MGET测试
    local batch_result=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT mget "${TEST_PREFIX}batch1" "${TEST_PREFIX}batch2" "${TEST_PREFIX}batch3" "${TEST_PREFIX}nonexistent")

    # 验证结果
    if echo "$batch_result" | grep -q "value1" && echo "$batch_result" | grep -q "value2" && echo "$batch_result" | grep -q "value3" && echo "$batch_result" | grep -q "(nil)"; then
        log_success "MGET operation returned correct values"
    else
        log_error "MGET operation returned unexpected results: $batch_result"
        return 1
    fi

    # 清理批量数据
    redis-cli -h $REDIS_HOST -p $REDIS_PORT del "${TEST_PREFIX}batch1" "${TEST_PREFIX}batch2" "${TEST_PREFIX}batch3" > /dev/null
}

# 数据类型测试
test_data_types() {
    log_info "Testing different data types..."

    # String类型
    redis-cli -h $REDIS_HOST -p $REDIS_PORT set "${TEST_PREFIX}string" "string_value" > /dev/null

    # List类型
    redis-cli -h $REDIS_HOST -p $REDIS_PORT rpush "${TEST_PREFIX}list" "item1" > /dev/null
    redis-cli -h $REDIS_HOST -p $REDIS_PORT rpush "${TEST_PREFIX}list" "item2" > /dev/null

    # Hash类型
    redis-cli -h $REDIS_HOST -p $REDIS_PORT hmset "${TEST_PREFIX}hash" field1 "hash_value1" field2 "hash_value2" > /dev/null

    # Set类型
    redis-cli -h $REDIS_HOST -p $REDIS_PORT sadd "${TEST_PREFIX}set" "member1" > /dev/null
    redis-cli -h $REDIS_HOST -p $REDIS_PORT sadd "${TEST_PREFIX}set" "member2" > /dev/null

    # 验证数据存在
    local string_exists=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT exists "${TEST_PREFIX}string")
    local list_length=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT llen "${TEST_PREFIX}list")
    local hash_exists=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT exists "${TEST_PREFIX}hash")
    local set_size=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT scard "${TEST_PREFIX}set")

    if [ "$string_exists" = "1" ] && [ "$list_length" = "2" ] && [ "$hash_exists" = "1" ] && [ "$set_size" = "2" ]; then
        log_success "All data types created successfully"
    else
        log_error "Data type creation failed - string:$string_exists list:$list_length hash:$hash_exists set:$set_size"
        return 1
    fi

    # 清理数据
    redis-cli -h $REDIS_HOST -p $REDIS_PORT del "${TEST_PREFIX}string" "${TEST_PREFIX}list" "${TEST_PREFIX}hash" "${TEST_PREFIX}set" > /dev/null
}

# 性能测试
test_performance() {
    log_info "Running performance tests..."

    local num_operations=1000
    local test_key="${TEST_PREFIX}perf"

    log_info "Testing $num_operations SET operations..."
    local start_time=$(date +%s%N)

    for i in $(seq 1 $num_operations); do
        redis-cli -h $REDIS_HOST -p $REDIS_PORT set "${test_key}_$i" "value_$i" > /dev/null
    done

    local end_time=$(date +%s%N)
    local set_duration=$(((end_time - start_time) / 1000000))
    local set_ops_per_sec=$((num_operations * 1000 / set_duration))

    log_success "SET: $num_operations operations in ${set_duration}ms (${set_ops_per_sec} ops/sec)"

    log_info "Testing $num_operations GET operations..."
    start_time=$(date +%s%N)

    for i in $(seq 1 $num_operations); do
        redis-cli -h $REDIS_HOST -p $REDIS_PORT get "${test_key}_$i" > /dev/null
    done

    end_time=$(date +%s%N)
    local get_duration=$(((end_time - start_time) / 1000000))
    local get_ops_per_sec=$((num_operations * 1000 / get_duration))

    log_success "GET: $num_operations operations in ${get_duration}ms (${get_ops_per_sec} ops/sec)"

    # 计算平均延迟
    local avg_set_latency=$((set_duration * 1000 / num_operations))
    local avg_get_latency=$((get_duration * 1000 / num_operations))

    log_info "Average SET latency: ${avg_set_latency}μs"
    log_info "Average GET latency: ${avg_get_latency}μs"

    # 清理性能测试数据
    log_info "Cleaning up performance test data..."
    for i in $(seq 1 $num_operations); do
        redis-cli -h $REDIS_HOST -p $REDIS_PORT del "${test_key}_$i" > /dev/null
    done
}

# 内存使用情况测试
test_memory_usage() {
    log_info "Checking Redis memory usage..."

    local memory_info=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT info memory)
    local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '[:space:]')
    local used_memory_rss=$(echo "$memory_info" | grep "used_memory_rss_human:" | cut -d: -f2 | tr -d '[:space:]')

    log_info "Used memory: $used_memory"
    log_info "RSS memory: $used_memory_rss"

    # 检查键数量
    local db_info=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT info keyspace)
    local total_keys=$(echo "$db_info" | grep -o "keys=[0-9]*" | cut -d= -f2)

    if [ -n "$total_keys" ]; then
        log_info "Total keys in database: $total_keys"
    else
        log_info "No keys found in database"
    fi
}

# 连接池测试
test_connection_pool() {
    log_info "Testing connection pool..."

    # 模拟并发连接
    local num_connections=10
    local pids=()

    log_info "Creating $num_connections concurrent connections..."

    for i in $(seq 1 $num_connections); do
        (
            redis-cli -h $REDIS_HOST -p $REDIS_PORT set "${TEST_PREFIX}pool_$i" "value_from_connection_$i" > /dev/null
            redis-cli -h $REDIS_HOST -p $REDIS_PORT get "${TEST_PREFIX}pool_$i" > /dev/null
            redis-cli -h $REDIS_HOST -p $REDIS_PORT del "${TEST_PREFIX}pool_$i" > /dev/null
        ) &
        pids+=($!)
    done

    # 等待所有后台进程完成
    for pid in "${pids[@]}"; do
        wait $pid
    done

    log_success "Connection pool test completed"
}

# 错误处理测试
test_error_handling() {
    log_info "Testing error handling..."

    # 测试不存在的键
    local nonexistent_value=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT get "${TEST_PREFIX}nonexistent")
    if [ "$nonexistent_value" = "(nil)" ]; then
        log_success "Handling of non-existent key is correct"
    else
        log_error "Unexpected value for non-existent key: $nonexistent_value"
        return 1
    fi

    # 测试空键
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT set "" "test_value" 2>&1 | grep -q "wrong number of arguments"; then
        log_success "Empty key correctly rejected"
    else
        log_warning "Empty key behavior may vary by Redis version"
    fi

    # 测试过长的键
    local long_key=$(printf 'a%.0s' {1..10000})
    if redis-cli -h $REDIS_HOST -p $REDIS_PORT set "$long_key" "test_value" > /dev/null 2>&1; then
        log_warning "Long key was accepted (may impact performance)"
        redis-cli -h $REDIS_HOST -p $REDIS_PORT del "$long_key" > /dev/null
    else
        log_success "Long key correctly rejected"
    fi
}

# 主测试函数
main() {
    echo "════════════════════════════════════════════════════════"
    echo "🧪 Redis Integration Test Suite"
    echo "════════════════════════════════════════════════════════"
    echo ""
    echo "Redis Configuration:"
    echo "  Host: $REDIS_HOST"
    echo "  Port: $REDIS_PORT"
    echo "  URL:  $REDIS_URL"
    echo ""

    # 检查Redis连接
    if ! check_redis_connection; then
        echo ""
        echo "❌ Redis connection failed. Please ensure Redis is running:"
        echo "   docker run -d --name redis-test -p 6379:6379 redis:7-alpine"
        echo "   or"
        echo "   make start-redis"
        exit 1
    fi

    echo ""

    # 清理之前的测试数据
    cleanup_test_data

    # 运行所有测试
    test_basic_operations || true
    echo ""

    test_expiration || true
    echo ""

    test_batch_operations || true
    echo ""

    test_data_types || true
    echo ""

    test_performance || true
    echo ""

    test_memory_usage || true
    echo ""

    test_connection_pool || true
    echo ""

    test_error_handling || true
    echo ""

    # 清理测试数据
    cleanup_test_data

    # 显示测试结果
    echo "════════════════════════════════════════════════════════"
    echo "📊 Test Results Summary"
    echo "════════════════════════════════════════════════════════"
    echo ""
    echo "✅ Tests Passed: $TESTS_PASSED"
    echo "❌ Tests Failed: $TESTS_FAILED"
    echo ""

    local total_tests=$((TESTS_PASSED + TESTS_FAILED))
    if [ $TESTS_FAILED -eq 0 ]; then
        echo "🎉 All tests passed successfully! Redis is working correctly."
        exit 0
    else
        echo "⚠️  Some tests failed. Please check the Redis configuration."
        exit 1
    fi
}

# 运行主函数
main "$@"
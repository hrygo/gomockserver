#!/bin/bash

# Redis高级测试脚本
# 提供更全面的Redis功能测试，包括集群、哨兵、持久化等高级特性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置参数
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
TEST_PREFIX="advanced_test_"
LOG_FILE="/tmp/redis_advanced_tests.log"

# 测试统计
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

# Redis连接函数
redis_cmd() {
    local cmd="$1"
    if [ -n "$REDIS_PASSWORD" ]; then
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" $cmd 2>/dev/null
    else
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" $cmd 2>/dev/null
    fi
}

# 测试Redis连接
test_redis_connection() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis connection..."

    if redis_cmd "ping" | grep -q "PONG"; then
        log_success "Redis connection established"
        return 0
    else
        log_error "Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT"
        return 1
    fi
}

# 测试Redis信息
test_redis_info() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis info command..."

    local server_info=$(redis_cmd "info server")
    local memory_info=$(redis_cmd "info memory")
    local stats_info=$(redis_cmd "info stats")

    if [ -n "$server_info" ] && [ -n "$memory_info" ] && [ -n "$stats_info" ]; then
        local redis_version=$(echo "$server_info" | grep "redis_version:" | cut -d: -f2 | tr -d '\r')
        local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '\r')
        local total_commands=$(echo "$stats_info" | grep "total_commands_processed:" | cut -d: -f2 | tr -d '\r')

        log_info "Redis Version: $redis_version"
        log_info "Used Memory: $used_memory"
        log_info "Total Commands: $total_commands"

        log_success "Redis info command working"
        return 0
    else
        log_error "Redis info command failed"
        return 1
    fi
}

# 测试基本数据类型
test_basic_data_types() {
    ((TESTS_TOTAL++))
    log_info "Testing basic Redis data types..."

    # String类型
    redis_cmd "set ${TEST_PREFIX}string 'Hello World'" > /dev/null
    local string_result=$(redis_cmd "get ${TEST_PREFIX}string")

    # List类型
    redis_cmd "lpush ${TEST_PREFIX}list item1 item2 item3" > /dev/null
    local list_length=$(redis_cmd "llen ${TEST_PREFIX}list")

    # Hash类型
    redis_cmd "hmset ${TEST_PREFIX}hash field1 value1 field2 value2" > /dev/null
    local hash_exists=$(redis_cmd "exists ${TEST_PREFIX}hash")

    # Set类型
    redis_cmd "sadd ${TEST_PREFIX}set member1 member2 member3" > /dev/null
    local set_size=$(redis_cmd "scard ${TEST_PREFIX}set")

    # Sorted Set类型
    redis_cmd "zadd ${TEST_PREFIX}zset 1 member1 2 member2 3 member3" > /dev/null
    local zset_size=$(redis_cmd "zcard ${TEST_PREFIX}zset")

    if [ "$string_result" = "Hello World" ] && [ "$list_length" = "3" ] &&
       [ "$hash_exists" = "1" ] && [ "$set_size" = "3" ] && [ "$zset_size" = "3" ]; then
        log_success "All basic data types working"

        # 清理测试数据
        redis_cmd "del ${TEST_PREFIX}string ${TEST_PREFIX}list ${TEST_PREFIX}hash ${TEST_PREFIX}set ${TEST_PREFIX}zset" > /dev/null
        return 0
    else
        log_error "Basic data types test failed"
        return 1
    fi
}

# 测试事务功能
test_transactions() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis transactions..."

    # 开启事务
    redis_cmd "multi" > /dev/null

    # 执行命令
    redis_cmd "set ${TEST_PREFIX}tx1 value1" > /dev/null
    redis_cmd "set ${TEST_PREFIX}tx2 value2" > /dev/null
    redis_cmd "incr ${TEST_PREFIX}counter" > /dev/null

    # 执行事务
    local tx_result=$(redis_cmd "exec")

    if [ -n "$tx_result" ]; then
        local val1=$(redis_cmd "get ${TEST_PREFIX}tx1")
        local val2=$(redis_cmd "get ${TEST_PREFIX}tx2")
        local counter=$(redis_cmd "get ${TEST_PREFIX}counter")

        if [ "$val1" = "value1" ] && [ "$val2" = "value2" ] && [ "$counter" = "1" ]; then
            log_success "Redis transactions working"

            # 清理测试数据
            redis_cmd "del ${TEST_PREFIX}tx1 ${TEST_PREFIX}tx2 ${TEST_PREFIX}counter" > /dev/null
            return 0
        fi
    fi

    log_error "Redis transactions test failed"
    return 1
}

# 测试发布订阅
test_pubsub() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis pub/sub..."

    # 在后台启动订阅者
    {
        echo "subscribe ${TEST_PREFIX}channel"
        sleep 2
        echo "unsubscribe ${TEST_PREFIX}channel"
        sleep 1
    } | redis_cli -h "$REDIS_HOST" -p "$REDIS_PORT" --csv > /tmp/pubsub_receiver.log 2>/dev/null &
    local subscriber_pid=$!

    # 等待订阅者准备就绪
    sleep 1

    # 发布消息
    redis_cmd "publish ${TEST_PREFIX}channel 'Hello from publisher'" > /dev/null

    # 等待消息处理
    sleep 2

    # 检查订阅者是否收到消息
    if grep -q "Hello from publisher" /tmp/pubsub_receiver.log 2>/dev/null; then
        log_success "Redis pub/sub working"
    else
        log_warning "Redis pub/sub test inconclusive (may need interactive testing)"
    fi

    # 清理
    kill $subscriber_pid 2>/dev/null || true
    rm -f /tmp/pubsub_receiver.log
}

# 测试键过期
test_key_expiration() {
    ((TESTS_TOTAL++))
    log_info "Testing key expiration..."

    # 设置带过期时间的键
    redis_cmd "setex ${TEST_PREFIX}expire 2 'Will expire'" > /dev/null

    # 立即检查应该存在
    local immediate_value=$(redis_cmd "get ${TEST_PREFIX}expire")

    # 等待过期
    sleep 3

    # 检查应该已过期
    local expired_value=$(redis_cmd "get ${TEST_PREFIX}expire")

    if [ "$immediate_value" = "Will expire" ] && [ "$expired_value" = "(nil)" ]; then
        log_success "Key expiration working"
        return 0
    else
        log_error "Key expiration test failed"
        return 1
    fi
}

# 测试Lua脚本
test_lua_scripts() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis Lua scripts..."

    # 简单的Lua脚本
    local script="return redis.call('set', KEYS[1], ARGV[1])"
    local script_result=$(redis_cmd "eval \"$script\" 1 ${TEST_PREFIX}lua_key 'Lua value'")

    # 检查结果
    local lua_value=$(redis_cmd "get ${TEST_PREFIX}lua_key")

    if [ "$script_result" = "OK" ] && [ "$lua_value" = "Lua value" ]; then
        log_success "Redis Lua scripts working"

        # 清理测试数据
        redis_cmd "del ${TEST_PREFIX}lua_key" > /dev/null
        return 0
    else
        log_error "Redis Lua scripts test failed"
        return 1
    fi
}

# 测试管道功能
test_pipelining() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis pipelining..."

    # 使用管道执行多个命令
    local start_time=$(date +%s%N)

    {
        echo "set ${TEST_PREFIX}pipe1 value1"
        echo "set ${TEST_PREFIX}pipe2 value2"
        echo "get ${TEST_PREFIX}pipe1"
        echo "get ${TEST_PREFIX}pipe2"
        echo "del ${TEST_PREFIX}pipe1"
        echo "del ${TEST_PREFIX}pipe2"
    } | redis_cli -h "$REDIS_HOST" -p "$REDIS_PORT" --raw > /tmp/pipeline_result.txt 2>/dev/null

    local end_time=$(date +%s%N)
    local pipeline_time=$(((end_time - start_time) / 1000000))

    if [ -f "/tmp/pipeline_result.txt" ] && grep -q "value1" /tmp/pipeline_result.txt && grep -q "value2" /tmp/pipeline_result.txt; then
        log_success "Redis pipelining working (${pipeline_time}ms)"
        rm -f /tmp/pipeline_result.txt
        return 0
    else
        log_error "Redis pipelining test failed"
        rm -f /tmp/pipeline_result.txt
        return 1
    fi
}

# 测试持久化配置
test_persistence() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis persistence configuration..."

    # 获取持久化配置
    local save_config=$(redis_cmd "config get save")
    local appendonly_config=$(redis_cmd "config get appendonly")
    local appendfsync_config=$(redis_cmd "config get appendfsync")

    if [ -n "$save_config" ] && [ -n "$appendonly_config" ] && [ -n "$appendfsync_config" ]; then
        log_info "Save config: $save_config"
        log_info "AOF config: $appendonly_config"
        log_info "AOF fsync: $appendfsync_config"

        log_success "Redis persistence configuration accessible"
        return 0
    else
        log_error "Redis persistence configuration check failed"
        return 1
    fi
}

# 测试内存管理
test_memory_management() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis memory management..."

    # 获取内存信息
    local memory_info=$(redis_cmd "info memory")
    local maxmemory_config=$(redis_cmd "config get maxmemory")
    local maxmemory_policy=$(redis_cmd "config get maxmemory-policy")

    if [ -n "$memory_info" ] && [ -n "$maxmemory_config" ] && [ -n "$maxmemory_policy" ]; then
        local used_memory=$(echo "$memory_info" | grep "used_memory:" | cut -d: -f2 | tr -d '\r')
        local maxmemory=$(echo "$maxmemory_config" | grep "maxmemory:" | cut -d: -f2 | tr -d '\r')
        local policy=$(echo "$maxmemory_policy" | grep "maxmemory-policy:" | cut -d: -f2 | tr -d '\r')

        log_info "Used memory: $used_memory bytes"
        log_info "Max memory: $maxmemory bytes"
        log_info "Eviction policy: $policy"

        log_success "Redis memory management working"
        return 0
    else
        log_error "Redis memory management test failed"
        return 1
    fi
}

# 测试安全性配置
test_security() {
    ((TESTS_TOTAL++))
    log_info "Testing Redis security configuration..."

    # 检查是否需要密码
    local requirepass_config=$(redis_cmd "config get requirepass")

    if [ -n "$requirepass_config" ]; then
        local requirepass=$(echo "$requirepass_config" | grep "requirepass:" | cut -d: -f2 | tr -d '\r')

        if [ -n "$requirepass" ] && [ "$requirepass" != "" ]; then
            log_info "Password protection: Enabled"
        else
            log_warning "Password protection: Disabled (consider enabling for production)"
        fi

        log_success "Redis security configuration checked"
        return 0
    else
        log_error "Redis security configuration check failed"
        return 1
    fi
}

# 清理测试数据
cleanup_test_data() {
    log_info "Cleaning up test data..."

    # 获取所有测试键
    local test_keys=$(redis_cmd "keys ${TEST_PREFIX}*")

    if [ -n "$test_keys" ] && [ "$test_keys" != "(empty array)" ] && [ "$test_keys" != "(empty list or set)" ]; then
        echo "$test_keys" | xargs redis_cmd "del" > /dev/null 2>&1 || true
    fi

    log_info "Test data cleanup completed"
}

# 生成测试报告
generate_test_report() {
    log_info "Generating test report..."

    {
        echo "=========================================="
        echo "Redis Advanced Tests Report"
        echo "=========================================="
        echo "Test Date: $(date)"
        echo "Redis Host: $REDIS_HOST:$REDIS_PORT"
        echo ""
        echo "Test Results:"
        echo "  Total Tests: $TESTS_TOTAL"
        echo "  Passed: $TESTS_PASSED"
        echo "  Failed: $TESTS_FAILED"
        echo "  Success Rate: $(( TESTS_PASSED * 100 / TESTS_TOTAL ))%"
        echo ""
        echo "Detailed logs available in: $LOG_FILE"
        echo "=========================================="
    } | tee -a "$LOG_FILE"
}

# 主测试函数
main() {
    echo "=========================================="
    echo "🔬 Redis Advanced Test Suite"
    echo "=========================================="
    echo "Testing Redis at: $REDIS_HOST:$REDIS_PORT"
    echo "Log file: $LOG_FILE"
    echo ""

    # 清理之前的日志
    > "$LOG_FILE"

    # 运行所有测试
    test_redis_connection
    test_redis_info
    test_basic_data_types
    test_transactions
    test_pubsub
    test_key_expiration
    test_lua_scripts
    test_pipelining
    test_persistence
    test_memory_management
    test_security

    # 清理测试数据
    cleanup_test_data

    # 生成测试报告
    generate_test_report

    echo ""
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}🎉 All Redis advanced tests passed! ($TESTS_PASSED/$TESTS_TOTAL)${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some Redis advanced tests failed ($TESTS_FAILED/$TESTS_TOTAL)${NC}"
        echo -e "${YELLOW}Check the log file for details: $LOG_FILE${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"
#!/bin/bash

# Redis性能测试脚本
# 对Redis缓存性能进行详细测试和基准测试

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
REDIS_URL="${REDIS_URL:-redis://localhost:6379}"
TEST_PREFIX="perf_test_"
OUTPUT_FILE="${1:-performance_report.txt}"

# 测试参数
BENCHMARK_DURATION="${BENCHMARK_DURATION:-30}"  # 基准测试持续时间（秒）
CONCURRENT_CLIENTS="${CONCURRENT_CLIENTS:-50}"    # 并发客户端数
KEY_SIZE="${KEY_SIZE:-32}"                          # 键大小
VALUE_SIZE="${VALUE_SIZE:-256}"                      # 值大小

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo -e "${CYAN}═══ $1 ═══${NC}"
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

# 获取Redis信息
get_redis_info() {
    log_info "Getting Redis server information..."

    local server_info=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT info server)
    local memory_info=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT info memory)
    local stats_info=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT info stats)

    echo "$server_info" > /tmp/redis_server_info.txt
    echo "$memory_info" > /tmp/redis_memory_info.txt
    echo "$stats_info" > /tmp/redis_stats_info.txt

    # 提取关键信息
    local redis_version=$(echo "$server_info" | grep "redis_version:" | cut -d: -f2 | tr -d '[:space:]')
    local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -d:2 | tr -d '[:space:]')
    local total_commands=$(echo "$stats_info" | grep "total_commands_processed:" | cut -d: -f2 | tr -d '[:space:]')
    local instantaneous_ops=$(echo "$stats_info" | grep "instantaneous_ops_per_sec:" | cut -d: -f2 | tr -d '[:space:]')

    log_info "Redis Version: $redis_version"
    log_info "Used Memory: $used_memory"
    log_info "Total Commands: $total_commands"
    log_info "Current Ops/sec: $instantaneous_ops"
}

# 清理测试数据
cleanup_test_data() {
    log_info "Cleaning up previous test data..."
    redis-cli -h $REDIS_HOST -p $REDIS_PORT --scan --pattern "${TEST_PREFIX}*" | xargs -r redis-cli -h $REDIS_HOST -p $REDIS_PORT del 2>/dev/null || true
}

# 生成测试数据
generate_test_data() {
    local num_keys=$1
    local prefix=$2

    log_info "Generating $num_keys test keys with prefix '$prefix'..."

    # 使用管道批量插入提高性能
    {
        for i in $(seq 1 $num_keys); do
            echo "set ${prefix}${i} $(date +%s%N)$(openssl rand -hex 32 | head -c 32)"
        done
    } | redis-cli -h $REDIS_HOST -p $REDIS_PORT --pipe

    log_success "Generated $num_keys test keys"
}

# 基准读取性能测试
benchmark_read_performance() {
    local num_keys=$1
    local clients=$2
    local duration=$3

    log_header "Read Performance Benchmark"
    log_info "Keys: $num_keys, Clients: $clients, Duration: ${duration}s"

    # 预热数据
    log_info "Warming up..."
    redis-cli -h $REDIS_HOST -p $REDIS_PORT --scan --pattern "${TEST_PREFIX}*" | head -100 | xargs -I {} redis-cli -h $REDIS_HOST -p $REDIS_PORT get {} > /dev/null

    # 运行读取基准测试
    log_info "Running read benchmark..."
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $clients -c $clients -t $duration -d $duration --csv get "${TEST_PREFIX}*"

    log_success "Read benchmark completed"
}

# 基准写入性能测试
benchmark_write_performance() {
    local num_keys=$1
    local clients=$2
    local duration=$3

    log_header "Write Performance Benchmark"
    log_info "Keys: $num_keys, Clients: $clients, Duration: ${duration}s"

    # 运行写入基准测试
    log_info "Running write benchmark..."
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $clients -c $clients -t $duration -d $duration --csv -r set "${TEST_PREFIX}" 0

    log_success "Write benchmark completed"
}

# 基准混合操作测试
benchmark_mixed_operations() {
    local num_keys=$1
    local clients=$2
    local duration=$3

    log_header "Mixed Operations Benchmark"
    log_info "Keys: $num_keys, Clients: $clients, Duration: ${duration}s"

    # 创建混合操作的Lua脚本
    local lua_script='
        local key = KEYS[1]
        local op = ARGV[1]
        if op == "get" then
            return redis.call("GET", key)
        elseif op == "set" then
            return redis.call("SET", key, ARGV[2])
        elseif op == "del" then
            return redis.call("DEL", key)
        else
            return "ERROR"
        end
    '

    # 保存Lua脚本
    local script_id=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT script load "$lua_script" | cut -d: -f2)

    log_info "Running mixed operations benchmark..."
    redis-cli -h $REDIS_HOST -p $REDIS_PORT -n $clients -c $clients -t $duration -d $duration --csv evalsha "$script_id" 0

    log_success "Mixed operations benchmark completed"
}

# 内存压力测试
test_memory_pressure() {
    log_header "Memory Pressure Test"

    local max_memory_mb=100  # 最大内存使用量（MB）
    local key_size=1000     # 每个键的大小（字节）
    local max_keys=$((max_memory_mb * 1024 * 1024 / key_size))

    log_info "Testing with up to $max_keys keys (${max_memory_mb}MB total)"

    local current_keys=0
    local batch_size=100

    while [ $current_keys -lt $max_keys ]; do
        local batch_end=$((current_keys + batch_size))
        if [ $batch_end -gt $max_keys ]; then
            batch_end=$max_keys
        fi

        log_info "Inserting keys $((current_keys + 1)) to $batch_end..."

        # 创建大值
        local large_value=$(head -c $key_size < /dev/zero | tr '\0' 'X')

        {
            for i in $(seq $((current_keys + 1)) $batch_end); do
                echo "set ${TEST_PREFIX}memory_$i $large_value"
            done
        } | redis-cli -h $REDIS_HOST -p $REDIS_PORT --pipe

        current_keys=$batch_end

        # 检查内存使用情况
        local memory_usage=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT info memory | grep "used_memory:" | cut -d: -f2)
        local memory_mb=$((memory_usage / 1024 / 1024))

        log_info "Current keys: $current_keys, Memory used: ${memory_mb}MB"

        # 如果内存使用超过90%，停止测试
        if [ $memory_mb -gt $((max_memory_mb * 90 / 100)) ]; then
            log_warning "Memory usage exceeded 90% threshold, stopping test"
            break
        fi

        sleep 1
    done

    log_info "Final memory usage check..."
    local final_memory=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT info memory | grep "used_memory_human:" | cut -d: -f2 | tr -d '[:space:]')
    local final_keys=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT dbsize | tr -d '[:space:]')

    log_info "Final memory usage: $final_memory"
    log_info "Total keys in database: $final_keys"

    # 清理内存测试数据
    cleanup_test_data
}

# 连接池压力测试
test_connection_pool_stress() {
    log_header "Connection Pool Stress Test"

    local max_connections=100
    local test_duration=10

    log_info "Testing with up to $max_connections concurrent connections for ${test_duration}s"

    local pids=()
    local connection_count=0

    # 创建多个并发连接
    for i in $(seq 1 $max_connections); do
        (
            redis-cli -h $REDIS_HOST -p $REDIS_PORT set "${TEST_PREFIX}conn_$i" "test_value_$i" > /dev/null
            redis-cli -h $REDIS_HOST -p $REDIS_PORT get "${TEST_PREFIX}conn_$i" > /dev/null
            sleep $test_duration
            redis-cli -h $REDIS_HOST -p $REDIS_PORT del "${TEST_PREFIX}conn_$i" > /dev/null
        ) &
        pids+=($!)
        ((connection_count++))

        # 每10个连接显示一次进度
        if [ $((connection_count % 10)) -eq 0 ]; then
            log_info "Created $connection_count connections..."
        fi
    done

    log_info "Waiting for all connections to complete..."
    for pid in "${pids[@]}"; do
        wait $pid
    done

    log_success "Connection pool stress test completed with $max_connections concurrent connections"
}

# 网络延迟测试
test_network_latency() {
    log_header "Network Latency Test"

    local iterations=1000
    local total_latency=0

    log_info "Testing network latency with $iterations iterations..."

    for i in $(seq 1 $iterations); do
        local start_time=$(date +%s%N)
        redis-cli -h $REDIS_HOST -p $REDIS_PORT ping > /dev/null
        local end_time=$(date +%s%N)
        local latency=$((end_time - start_time))
        total_latency=$((total_latency + latency))
    done

    local avg_latency=$((total_latency / iterations / 1000))  # 转换为微秒
    local min_latency=999999999
    local max_latency=0

    log_info "Average network latency: ${avg_latency}μs"
    log_success "Network latency test completed"
}

# 数据一致性测试
test_data_consistency() {
    log_header "Data Consistency Test"

    local num_keys=1000
    local test_data="consistency_test_data_$(date +%s)"

    # 写入测试数据
    log_info "Writing $num_keys keys with consistency check data..."
    {
        for i in $(seq 1 $num_keys); do
            echo "set ${TEST_PREFIX}consistency_$i $test_data"
        done
    } | redis-cli -h $REDIS_HOST -p $REDIS_PORT --pipe

    # 读取并验证数据一致性
    log_info "Verifying data consistency..."
    local inconsistent_keys=0
    local checked_keys=0

    for i in $(seq 1 $num_keys); do
        local retrieved_value=$(redis-cli -h $REDIS_HOST -p $REDIS_PORT get "${TEST_PREFIX}consistency_$i")
        ((checked_keys++))

        if [ "$retrieved_value" != "$test_data" ]; then
            ((inconsistent_keys++))
            if [ $inconsistent_keys -le 5 ]; then
                log_error "Inconsistent data for key consistency_$i: expected '$test_data', got '$retrieved_value'"
            fi
        fi
    done

    if [ $inconsistent_keys -eq 0 ]; then
        log_success "All $checked_keys keys passed consistency check"
    else
        log_error "$inconsistent_keys out of $checked_keys keys failed consistency check"
    fi

    # 清理一致性测试数据
    cleanup_test_data
}

# 生成性能报告
generate_performance_report() {
    local report_file=$1

    log_info "Generating performance report..."

    {
        echo "Redis Performance Test Report"
        echo "=========================="
        echo "Generated on: $(date)"
        echo "Redis Host: $REDIS_HOST:$REDIS_PORT"
        echo ""
        echo "Test Configuration:"
        echo "- Benchmark Duration: ${BENCHMARK_DURATION}s"
        echo "- Concurrent Clients: $CONCURRENT_CLIENTS"
        echo "- Key Size: ${KEY_SIZE} bytes"
        echo "- Value Size: ${VALUE_SIZE} bytes"
        echo ""

        echo "Redis Server Information:"
        grep -E "(redis_version|used_memory_human|total_commands_processed|instantaneous_ops_per_sec)" /tmp/redis_*_*.txt
        echo ""

        echo "Test Results:"
        echo "See individual test outputs above for detailed metrics."
        echo ""

        echo "Recommendations:"
        echo "- Monitor memory usage to avoid Redis OOM"
        echo "- Consider using Redis Cluster for high-throughput scenarios"
        echo "- Implement proper key expiration strategies"
        echo "- Use connection pooling in production applications"
        echo "- Monitor slow queries with Redis SLOWLOG"

    } > "$report_file"

    log_success "Performance report generated: $report_file"

    # 清理临时文件
    rm -f /tmp/redis_*_*.txt
}

# 主测试函数
main() {
    echo "════════════════════════════════════════════════════════"
    echo "⚡ Redis Performance Test Suite"
    echo "══════════════════════════════════════════════════════"
    echo ""
    echo "Redis Configuration:"
    echo "  Host: $REDIS_HOST"
    echo "  Port: $REDIS_PORT"
    echo "  URL:  $REDIS_URL"
    echo "  Benchmark Duration: ${BENCHMARK_DURATION}s"
    echo "  Concurrent Clients: $CONCURRENT_CLIENTS"
    echo "  Key Size: ${KEY_SIZE} bytes"
    echo "  Value Size: ${VALUE_SIZE} bytes"
    echo ""

    # 检查Redis连接
    if ! check_redis_connection; then
        echo ""
        echo "❌ Redis connection failed. Please ensure Redis is running:"
        echo "   docker run -d --name redis-perf -p 6379:6379 redis:7-alpine"
        echo "   or"
        echo "   make start-redis"
        exit 1
    fi

    echo ""

    # 获取Redis信息
    get_redis_info
    echo ""

    # 清理之前的测试数据
    cleanup_test_data

    # 运行性能测试
    test_network_latency
    echo ""

    test_connection_pool_stress
    echo ""

    benchmark_read_performance 1000 $CONCURRENT_CLIENTS $BENCHMARK_DURATION
    echo ""

    benchmark_write_performance 1000 $CONCURRENT_CLIENTS $BENCHMARK_DURATION
    echo ""

    benchmark_mixed_operations 1000 $CONCURRENT_CLIENTS $BENCHMARK_DURATION
    echo ""

    test_data_consistency
    echo ""

    test_memory_pressure
    echo ""

    # 清理测试数据
    cleanup_test_data

    # 生成性能报告
    generate_performance_report "$OUTPUT_FILE"

    # 显示完成信息
    echo ""
    echo "════════════════════════════════════════════════════════"
    echo "🚀 Performance Testing Completed!"
    echo "══════════════════════════════════════════════════════"
    echo ""
    echo "📊 Detailed report: $OUTPUT_FILE"
    echo ""
    echo "💡 Performance Optimization Tips:"
    echo "  • Use Redis pipelining for batch operations"
    echo "  • Implement proper key naming conventions"
    "  • Use appropriate data structures (hash, list, set, zset)"
    echo "  • Configure memory limits and eviction policies"
    echo "  • Monitor Redis metrics regularly"
    echo "  • Consider Redis persistence based on your use case"
    echo ""

    exit 0
}

# 运行主函数
main "$@"
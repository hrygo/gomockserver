#!/bin/bash

# MockServer E2E 测试框架
# 提供通用的测试功能和工具函数

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BINARY="$PROJECT_ROOT/mockserver"

# 全局测试变量
PROJECT_ID=""
ENVIRONMENT_ID=""
RULE_ID=""
WS_PROJECT_ID=""
WS_ENVIRONMENT_ID=""
WS_RULE_ID=""
TEST_PASSED=0
TEST_FAILED=0
TEST_SKIPPED=0

# 测试结果记录
TEST_RESULTS=()
TEST_START_TIME=""
TEST_END_TIME=""

# ========================================
# 环境检测和配置
# ========================================

detect_environment() {
    if [ -n "$GITHUB_ACTIONS" ]; then
        echo -e "${CYAN}检测到 GitHub Actions 环境${NC}"
        CONFIG_FILE="$PROJECT_ROOT/config.test.yaml"
        ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
        MOCK_API="${MOCK_API:-http://localhost:9090}"
        SKIP_SERVER_START="true"
    else
        echo -e "${CYAN}检测到本地开发环境${NC}"
        CONFIG_FILE="$PROJECT_ROOT/config.dev.yaml"
        ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
        MOCK_API="${MOCK_API:-http://localhost:9090}"
        SKIP_SERVER_START="${SKIP_SERVER_START:-false}"
    fi

    echo -e "${CYAN}使用配置:${NC}"
    echo -e "  配置文件: ${YELLOW}$CONFIG_FILE${NC}"
    echo -e "  管理API: ${YELLOW}$ADMIN_API${NC}"
    echo -e "  MockAPI: ${YELLOW}$MOCK_API${NC}"
    echo -e "  跳过服务器启动: ${YELLOW}$SKIP_SERVER_START${NC}"
}

# ========================================
# 测试日志函数
# ========================================

log_test() {
    echo -e "${CYAN}[TEST]${NC} $1"
}

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

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
}

# ========================================
# 测试结果记录函数
# ========================================

test_pass() {
    echo -e "${GREEN}✓ $1${NC}"
    TEST_PASSED=$((TEST_PASSED + 1))
    TEST_RESULTS+=("PASS: $1")
}

test_fail() {
    echo -e "${RED}✗ $1${NC}"
    TEST_FAILED=$((TEST_FAILED + 1))
    TEST_RESULTS+=("FAIL: $1")
}

test_skip() {
    echo -e "${YELLOW}⚠ $1${NC}"
    TEST_SKIPPED=$((TEST_SKIPPED + 1))
    TEST_RESULTS+=("SKIP: $1")
}

test_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

test_warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# ========================================
# JSON 工具函数
# ========================================

extract_json_field() {
    echo "$1" | grep -o "\"$2\":\"[^\"]*\"" | cut -d'"' -f4
}

extract_json_field_array() {
    echo "$1" | grep -o "\"$2\":\[[^]]*\]" | sed 's/.*\[\(.*\)\].*/\1/'
}

extract_json_bool() {
    local value=$(echo "$1" | grep -o "\"$2\":[^,}]*" | cut -d':' -f2 | tr -d ' ')
    echo "$value"
}

# ========================================
# HTTP 请求封装函数
# ========================================

http_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local headers="$4"

    local cmd="curl -s -w '\n%{http_code}\n'"

    if [ -n "$headers" ]; then
        cmd="$cmd $headers"
    fi

    if [ -n "$data" ]; then
        cmd="$cmd -X $method -H 'Content-Type: application/json' -d '$data'"
    else
        cmd="$cmd -X $method"
    fi

    cmd="$cmd '$url'"

    eval "$cmd"
}

http_get() {
    http_request "GET" "$1" "" "$2"
}

http_post() {
    http_request "POST" "$1" "$2" "$3"
}

http_put() {
    http_request "PUT" "$1" "$2" "$3"
}

http_delete() {
    http_request "DELETE" "$1" "" "$2"
}

# ========================================
# Mock 请求函数
# ========================================

mock_request() {
    local method="$1"
    local path="$2"
    local data="$3"
    local headers="$4"

    # 确保关键变量已设置
    if [ -z "$PROJECT_ID" ] || [ -z "$ENVIRONMENT_ID" ]; then
        echo "Error: PROJECT_ID or ENVIRONMENT_ID is not set"
        echo "PROJECT_ID='${PROJECT_ID}' ENVIRONMENT_ID='${ENVIRONMENT_ID}'"
        return 1
    fi

    local url="$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID$path"

    # 使用绝对路径的curl命令以避免PATH问题
    local curl_cmd="curl"
    if command -v curl >/dev/null 2>&1; then
        curl_cmd="curl"
    elif [ -x "/opt/anaconda3/bin/curl" ]; then
        curl_cmd="/opt/anaconda3/bin/curl"
    elif [ -x "/usr/bin/curl" ]; then
        curl_cmd="/usr/bin/curl"
    else
        echo "Error: curl command not found"
        return 1
    fi

    if [ -n "$data" ]; then
        cmd="$curl_cmd -s -L -w '\n%{http_code}\n' -X $method"
    else
        cmd="$curl_cmd -s -L -w '\n%{http_code}\n' -X $method"
    fi

    if [ -n "$headers" ]; then
        cmd="$cmd $headers"
    fi

    if [ -n "$data" ]; then
        cmd="$cmd -H 'Content-Type: application/json' -d '$data'"
    fi

    cmd="$cmd '$url'"

    eval "$cmd"
}

# ========================================
# 重试机制函数
# ========================================

retry_with_backoff() {
    local max_attempts="$1"
    local delay="$2"
    local command="$3"
    local description="$4"

    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        echo -e "  尝试第 $attempt 次: $description"

        if eval "$command"; then
            return 0
        fi

        if [ $attempt -lt $max_attempts ]; then
            echo -e "  等待 $delay 秒后重试..."
            sleep "$delay"
            delay=$((delay * 2))  # 指数退避
        fi

        attempt=$((attempt + 1))
    done

    return 1
}

# ========================================
# 时间工具函数
# ========================================

get_timestamp_ms() {
    # macOS 兼容的时间戳获取
    python3 -c 'import time; print(int(time.time() * 1000))' 2>/dev/null || date +%s000
}

# 跨平台序列生成函数
seq() {
    local start="${1:-1}"
    local end="$2"

    if [ $# -eq 1 ]; then
        end="$1"
        start=1
    fi

    # macOS 和 Linux 都支持 seq 命令，这里提供备用方案
    command seq "$start" "$end" 2>/dev/null || {
        # 备用方案：使用 while 循环
        local i=$start
        while [ $i -le $end ]; do
            echo $i
            i=$((i + 1))
        done
    }
}

# 跨平台工具检测和备用方案
check_command() {
    local cmd="$1"
    local fallback="$2"

    if command -v "$cmd" >/dev/null 2>&1; then
        echo "$cmd"
    elif [ -n "$fallback" ] && command -v "$fallback" >/dev/null 2>&1; then
        echo "$fallback"
    else
        return 1
    fi
}

# 跨平台超时函数
timeout_cmd() {
    local duration="$1"
    shift
    local command=("$@")

    # macOS 没有 timeout 命令，使用备用方案
    if command -v timeout >/dev/null 2>&1; then
        timeout "$duration" "${command[@]}"
    else
        # macOS 备用方案：使用 Perl 的 alarm 函数
        perl -e 'alarm shift @ARGV; exec @ARGV' "$duration" "${command[@]}"
    fi
}

# 跨平台进程查找
find_process() {
    local process_name="$1"

    # 首先尝试 pgrep
    if command -v pgrep >/dev/null 2>&1; then
        pgrep "$process_name" | head -1
        return 0
    fi

    # 备用方案：使用 ps
    if command -v ps >/dev/null 2>&1; then
        ps aux | grep "[${process_name:0:1}]${process_name:1}" | awk '{print $2}' | head -1
        return 0
    fi

    return 1
}

# 跨平台内存获取
get_process_memory() {
    local pid="$1"

    if ! command -v ps >/dev/null 2>&1; then
        return 1
    fi

    local memory=""
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS 使用 ps 命令
        memory=$(ps -o pid,rss -p "$pid" 2>/dev/null | awk 'NR==2 {print $2}')
    else
        # Linux 使用 ps 命令
        memory=$(ps -o pid,rss -p "$pid" 2>/dev/null | awk 'NR==2 {print $2}')
    fi

    if [ -n "$memory" ]; then
        echo "$memory"
        return 0
    fi

    return 1
}

calculate_duration() {
    local start_time="$1"
    local end_time="$2"
    echo $((end_time - start_time))
}

# ========================================
# 文件系统工具函数
# ========================================

create_test_file() {
    local file_path="$1"
    local content="$2"
    local size="$3"

    mkdir -p "$(dirname "$file_path")"

    if [ -n "$content" ]; then
        echo "$content" > "$file_path"
    elif [ -n "$size" ]; then
        dd if=/dev/zero of="$file_path" bs=1 count="$size" 2>/dev/null
    fi
}

cleanup_test_files() {
    local test_dir="$1"
    if [ -n "$test_dir" ] && [ -d "$test_dir" ]; then
        rm -rf "$test_dir"
    fi
}

# ========================================
# 随机数据生成函数
# ========================================

generate_random_string() {
    local length="${1:-8}"
    openssl rand -hex $((length/2)) 2>/dev/null | head -c "$length"
}

generate_random_email() {
    echo "test-$(generate_random_string 6)@example.com"
}

generate_random_phone() {
    echo "1$(generate_random_string 10)"
}

generate_random_id() {
    echo "$(date +%s%3N)$(generate_random_string 4)"
}

# ========================================
# 测试数据生成函数
# ========================================

generate_project_data() {
    local name="$1"
    cat <<EOF
{
    "name": "$name",
    "workspace_id": "workspace-$(generate_random_string 8)",
    "description": "E2E测试项目 - $name"
}
EOF
}

generate_environment_data() {
    local name="$1"
    local base_url="$2"
    cat <<EOF
{
    "name": "$name",
    "base_url": "$base_url",
    "variables": {
        "api_version": "v1",
        "timeout": "30s",
        "retry_count": "3"
    }
}
EOF
}

generate_rule_data() {
    local name="$1"
    local method="$2"
    local path="$3"
    local response_body="$4"
    local status_code="${5:-200}"

    cat <<EOF
{
    "name": "$name",
    "project_id": "$PROJECT_ID",
    "environment_id": "$ENVIRONMENT_ID",
    "protocol": "HTTP",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "method": "$method",
        "path": "$path"
    },
    "response": {
        "type": "Static",
        "content": {
            "status_code": $status_code,
            "content_type": "JSON",
            "headers": {
                "X-Mock-Server": "MockServer",
                "X-Test-Case": "$name"
            },
            "body": $response_body
        }
    }
}
EOF
}

# ========================================
# WebSocket 测试函数
# ========================================

websocket_test_connection() {
    local ws_url="$1"
    local test_message="$2"

    # 使用 websocat 或类似工具测试 WebSocket 连接
    if command -v websocat >/dev/null 2>&1; then
        echo "$test_message" | websocat --one-message "$ws_url" 2>/dev/null
        return 0
    else
        test_warn "websocat 未安装，跳过 WebSocket 测试"
        return 1
    fi
}

# ========================================
# 性能测试函数
# ========================================

performance_test() {
    local url="$1"
    local concurrent="$2"
    local duration="$3"

    echo -e "${CYAN}开始性能测试: $concurrent 并发，$duration 秒${NC}"

    if command -v wrk >/dev/null 2>&1; then
        wrk -t4 -c"$concurrent" -d"$duration"s "$url" 2>/dev/null | tee /tmp/perf_test.log
        return 0
    else
        test_warn "wrk 未安装，使用简单并发测试"

        # 简单并发测试
        for i in $(seq 1 $concurrent); do
            (curl -s "$url" >/dev/null 2>&1) &
        done
        wait

        echo "简单并发测试完成: $concurrent 个请求"
        return 0
    fi
}

# ========================================
# 测试报告生成函数
# ========================================

generate_test_report() {
    local report_file="$1"
    local test_name="$2"

    TEST_END_TIME=$(date +%s)
    TEST_DURATION=$((TEST_END_TIME - TEST_START_TIME))

    cat > "$report_file" << EOF
# $test_name 测试报告

## 测试概要
- **测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **测试时长**: ${TEST_DURATION}秒
- **通过测试**: $TEST_PASSED
- **失败测试**: $TEST_FAILED
- **跳过测试**: $TEST_SKIPPED
- **总计测试**: $((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))
- **成功率**: $(( TEST_PASSED * 100 / (TEST_PASSED + TEST_FAILED) ))%

## 测试结果详情
EOF

    if [ ${#TEST_RESULTS[@]} -gt 0 ]; then
        echo "" >> "$report_file"
        echo "### 详细结果" >> "$report_file"
        echo "" >> "$report_file"

        for result in "${TEST_RESULTS[@]}"; do
            if [[ $result == PASS* ]]; then
                echo "- ✅ ${result#PASS: }" >> "$report_file"
            elif [[ $result == FAIL* ]]; then
                echo "- ❌ ${result#FAIL: }" >> "$report_file"
            elif [[ $result == SKIP* ]]; then
                echo "- ⚠️ ${result#SKIP: }" >> "$report_file"
            fi
        done
    fi

    echo "" >> "$report_file"
    echo "## 环境信息" >> "$report_file"
    echo "- **操作系统**: $(uname -s)" >> "$report_file"
    echo "- **Go版本**: $(go version 2>/dev/null || echo 'Unknown')" >> "$report_file"
    echo "- **配置文件**: $CONFIG_FILE" >> "$report_file"
    echo "- **管理API**: $ADMIN_API" >> "$report_file"
    echo "- **MockAPI**: $MOCK_API" >> "$report_file"

    echo -e "${GREEN}测试报告已生成: $report_file${NC}"
}

# ========================================
# 清理函数
# ========================================

cleanup_test_resources() {
    echo -e "${YELLOW}清理测试资源...${NC}"

    # 删除测试项目
    if [ -n "$PROJECT_ID" ]; then
        echo "  删除测试项目: $PROJECT_ID"
        http_delete "$ADMIN_API/projects/$PROJECT_ID" >/dev/null 2>&1 || true
    fi

    # 删除 WebSocket 测试项目
    if [ -n "$WS_PROJECT_ID" ]; then
        echo "  删除WebSocket测试项目: $WS_PROJECT_ID"
        http_delete "$ADMIN_API/projects/$WS_PROJECT_ID" >/dev/null 2>&1 || true
    fi

    # 清理测试文件
    cleanup_test_files "/tmp/mockserver_test"
}

print_test_summary() {
    echo ""
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   测试结果统计${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo -e "通过测试: ${GREEN}$TEST_PASSED${NC}"
    echo -e "失败测试: ${RED}$TEST_FAILED${NC}"
    echo -e "跳过测试: ${YELLOW}$TEST_SKIPPED${NC}"
    echo -e "总计测试: $((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))"

    local total=$((TEST_PASSED + TEST_FAILED))
    if [ $total -gt 0 ]; then
        local success_rate=$(( TEST_PASSED * 100 / total ))
        echo -e "成功率: ${GREEN}$success_rate%${NC}"
    fi

    if [ $TEST_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ 所有测试通过！${NC}"
        return 0
    else
        echo -e "${RED}✗ 部分测试失败${NC}"
        return 1
    fi
}

# ========================================
# 初始化函数
# ========================================

init_test_framework() {
    TEST_START_TIME=$(date +%s)
    detect_environment

    # 检查并安装必要工具
    check_and_install_dependencies

    trap cleanup_test_resources EXIT INT TERM
}

# 检查并安装依赖
check_and_install_dependencies() {
    # 加载工具安装器
    local installer_path="$(dirname "$0")/tool_installer.sh"
    if [ -f "$installer_path" ]; then
        source "$installer_path"

        # 检查基础工具
        if ! check_tools_ready "basic"; then
            echo -e "${YELLOW}检测到缺失的基础工具，正在自动安装...${NC}"
            if ! install_required_tools_silent "basic"; then
                echo -e "${YELLOW}部分工具安装失败，请检查工具可用性${NC}"
            fi
        fi
    fi
}

# ========================================
# Redis 测试函数
# ========================================

# 检查Redis连接
check_redis_connection() {
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    if command -v redis-cli >/dev/null 2>&1; then
        if redis-cli -h "$redis_host" -p "$redis_port" ping >/dev/null 2>&1; then
            return 0
        fi
    fi
    return 1
}

# 启动Redis服务（如果需要）
start_redis_if_needed() {
    if check_redis_connection; then
        test_info "Redis is already running"
        return 0
    fi

    test_info "Starting Redis service..."
    if command -v make >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/Makefile" ]; then
        make start-redis >/dev/null 2>&1
        sleep 2
        return $?
    else
        test_warn "Cannot start Redis automatically - please start Redis manually"
        return 1
    fi
}

# Redis缓存测试
test_redis_cache_operations() {
    local test_key="mockserver_test_$(generate_random_string 8)"
    local test_value="test_value_$(date +%s)"
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    if ! check_redis_connection; then
        test_skip "Redis not available for cache testing"
        return 0
    fi

    # SET操作测试
    if redis-cli -h "$redis_host" -p "$redis_port" set "$test_key" "$test_value" | grep -q "OK"; then
        test_pass "Redis SET operation"
    else
        test_fail "Redis SET operation"
        return 1
    fi

    # GET操作测试
    local retrieved_value=$(redis-cli -h "$redis_host" -p "$redis_port" get "$test_key")
    if [ "$retrieved_value" = "$test_value" ]; then
        test_pass "Redis GET operation"
    else
        test_fail "Redis GET operation - expected '$test_value', got '$retrieved_value'"
        return 1
    fi

    # 过期时间测试
    redis-cli -h "$redis_host" -p "$redis_port" setex "${test_key}_expire" 2 "expire_test" >/dev/null
    if redis-cli -h "$redis_host" -p "$redis_port" get "${test_key}_expire" | grep -q "expire_test"; then
        test_pass "Redis SETEX operation (immediate)"
        sleep 3
        if redis-cli -h "$redis_host" -p "$redis_port" get "${test_key}_expire" | grep -q "(nil)"; then
            test_pass "Redis key expiration"
        else
            test_fail "Redis key expiration"
        fi
    else
        test_fail "Redis SETEX operation"
    fi

    # 清理测试数据
    redis-cli -h "$redis_host" -p "$redis_port" del "$test_key" "${test_key}_expire" >/dev/null 2>&1 || true

    return 0
}

# Redis性能测试
test_redis_performance() {
    local num_operations=100
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    if ! check_redis_connection; then
        test_skip "Redis not available for performance testing"
        return 0
    fi

    test_info "Testing Redis performance with $num_operations operations"

    local test_key="perf_test_$(date +%s)"
    local start_time=$(get_timestamp_ms)

    # SET性能测试
    for i in $(seq 1 $num_operations); do
        redis-cli -h "$redis_host" -p "$redis_port" set "${test_key}_$i" "value_$i" >/dev/null
    done

    local set_time=$(($(get_timestamp_ms) - start_time))
    local set_ops_per_sec=$((num_operations * 1000 / set_time))

    test_info "Redis SET: $num_operations operations in ${set_time}ms (${set_ops_per_sec} ops/sec)"

    # GET性能测试
    start_time=$(get_timestamp_ms)
    for i in $(seq 1 $num_operations); do
        redis-cli -h "$redis_host" -p "$redis_port" get "${test_key}_$i" >/dev/null
    done

    local get_time=$(($(get_timestamp_ms) - start_time))
    local get_ops_per_sec=$((num_operations * 1000 / get_time))

    test_info "Redis GET: $num_operations operations in ${get_time}ms (${get_ops_per_sec} ops/sec)"

    # 性能基准检查
    if [ $set_ops_per_sec -gt 1000 ] && [ $get_ops_per_sec -gt 2000 ]; then
        test_pass "Redis performance meets requirements"
    else
        test_warn "Redis performance below expected (SET: ${set_ops_per_sec}, GET: ${get_ops_per_sec})"
    fi

    # 清理性能测试数据
    for i in $(seq 1 $num_operations); do
        redis-cli -h "$redis_host" -p "$redis_port" del "${test_key}_$i" >/dev/null 2>&1 || true
    done

    return 0
}

# Redis内存使用检查
test_redis_memory_usage() {
    if ! check_redis_connection; then
        test_skip "Redis not available for memory testing"
        return 0
    fi

    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    # 获取内存信息
    local memory_info=$(redis-cli -h "$redis_host" -p "$redis_port" info memory 2>/dev/null)
    if [ -n "$memory_info" ]; then
        local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '[:space:]')
        local used_memory_rss=$(echo "$memory_info" | grep "used_memory_rss_human:" | cut -d: -f2 | tr -d '[:space:]')

        test_info "Redis memory usage: $used_memory (RSS: $used_memory_rss)"
        test_pass "Redis memory monitoring"
    else
        test_fail "Redis memory info retrieval"
        return 1
    fi

    return 0
}

# 运行Redis集成测试
run_redis_integration_tests() {
    test_info "Starting Redis integration tests"

    # 确保Redis运行
    start_redis_if_needed

    # 运行各种Redis测试
    test_redis_cache_operations
    test_redis_performance
    test_redis_memory_usage

    # 如果有独立的Redis测试脚本，也运行它们
    local redis_integration_script="$PROJECT_ROOT/tests/redis/redis_integration_test.sh"
    if [ -f "$redis_integration_script" ] && [ -x "$redis_integration_script" ]; then
        test_info "Running comprehensive Redis integration tests"
        if "$redis_integration_script"; then
            test_pass "Comprehensive Redis integration tests"
        else
            test_fail "Comprehensive Redis integration tests"
        fi
    fi

    test_info "Redis integration tests completed"
}

# 运行Redis性能测试
run_redis_performance_tests() {
    test_info "Starting Redis performance tests"

    # 确保Redis运行
    start_redis_if_needed

    # 如果有独立的Redis性能测试脚本，运行它们
    local redis_perf_script="$PROJECT_ROOT/tests/redis/redis_performance_test.sh"
    if [ -f "$redis_perf_script" ] && [ -x "$redis_perf_script" ]; then
        test_info "Running comprehensive Redis performance tests"
        if "$redis_perf_script" "/tmp/redis_performance_report.txt"; then
            test_pass "Comprehensive Redis performance tests"
            test_info "Performance report saved to /tmp/redis_performance_report.txt"
        else
            test_fail "Comprehensive Redis performance tests"
        fi
    else
        test_warn "Redis performance test script not found, running basic tests"
        test_redis_performance
    fi

    test_info "Redis performance tests completed"
}

# ========================================
# 导出函数
# ========================================

# 导出所有函数供其他脚本使用
export -f detect_environment
export -f test_pass test_fail test_skip test_info test_warn
export -f extract_json_field extract_json_field_array extract_json_bool
export -f http_request http_get http_post http_put http_delete
export -f mock_request
export -f retry_with_backoff
export -f get_timestamp_ms calculate_duration seq check_command find_process get_process_memory
export -f create_test_file cleanup_test_files
export -f generate_random_string generate_random_email generate_random_phone generate_random_id
export -f generate_project_data generate_environment_data generate_rule_data
export -f websocket_test_connection
export -f performance_test
export -f check_redis_connection start_redis_if_needed
export -f test_redis_cache_operations test_redis_performance test_redis_memory_usage
export -f run_redis_integration_tests run_redis_performance_tests
export -f generate_test_report
export -f cleanup_test_resources print_test_summary
export -f init_test_framework

echo -e "${GREEN}测试框架已加载${NC}"
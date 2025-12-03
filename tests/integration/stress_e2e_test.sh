#!/bin/bash

# MockServer 压力测试和负载测试脚本
# 测试系统在高负载下的性能表现
# 已优化：集成新的coordinate_services函数和统一测试框架

set -e

# 加载测试框架
source "$(dirname "$0")/lib/test_framework.sh"

# 初始化测试框架
init_test_framework

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# 测试配置
TEST_DIR="$(dirname "$0")"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_DIR="/tmp/stress_test_results_${TIMESTAMP}"
REPORT_FILE="$RESULTS_DIR/stress_test_report_${TIMESTAMP}.md"

# 创建结果目录
mkdir -p "$RESULTS_DIR"

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 压力测试配置
STRESS_CONFIGS=(
    "10:10:轻量级负载"
    "50:20:中等负载"
    "100:30:高负载"
    "200:60:极高负载"
)

# 显示横幅
show_banner() {
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   MockServer 压力测试和负载测试${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""
    echo -e "${CYAN}测试目标:${NC}"
    echo -e "  • 负载测试 (多并发连接)"
    echo -e "  • 响应时间基准测试"
    echo -e "  • 吞吐量性能测试"
    echo -e "  • 长时间稳定性测试"
    echo -e "  • 资源使用监控"
    echo -e ""
    echo -e "${CYAN}开始时间: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo -e "${CYAN}结果目录: $RESULTS_DIR${NC}"
    echo ""
}

# 检查压力测试工具
check_stress_tools() {
    log_test "检查压力测试工具"

    # 检查 wrk
    if command -v wrk >/dev/null 2>&1; then
        log_pass "找到 wrk 压力测试工具"
        echo "wrk version: $(wrk --version 2>/dev/null || echo 'unknown')"
        return 0
    fi

    # 检查 ab (Apache Bench)
    if command -v ab >/dev/null 2>&1; then
        log_pass "找到 Apache Bench (ab) 压力测试工具"
        echo "ab version: $(ab -V 2>&1 | head -1 || echo 'unknown')"
        return 0
    fi

    # 检查 hey
    if command -v hey >/dev/null 2>&1; then
        log_pass "找到 hey 压力测试工具"
        return 0
    fi

    log_fail "未找到压力测试工具 (wrk/ab/hey)"
    log_info "请安装其中一个工具:"
    log_info "  brew install wrk  # macOS"
    log_info "  sudo apt-get install apache2-utils  # Ubuntu"
    log_info "  go install github.com/rakyll/hey@latest"
    return 1
}

# 基础性能测试 (使用 curl)
basic_performance_test() {
    log_test "执行基础性能测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local url="$MOCK_API/api/test/performance"
    local iterations=100
    local total_time=0
    local success_count=0

    log_info "执行 $iterations 次基础请求..."

    for i in $(seq 1 $iterations); do
        local start_time=$(date +%s.%N)
        local response=$(curl -s -w "%{http_code}" \
            -H "Content-Type: application/json" \
            -d '{"test": "performance"}' \
            "$url" 2>/dev/null || echo "000")
        local end_time=$(date +%s.%N)

        local duration=$(echo "$end_time - $start_time" | bc -l 2>/dev/null || echo "0")
        total_time=$(echo "$total_time + $duration" | bc -l 2>/dev/null || echo "$total_time")

        if [ "$response" = "200" ]; then
            success_count=$((success_count + 1))
        fi

        # 显示进度
        if [ $((i % 20)) -eq 0 ]; then
            echo -n "."
        fi
    done
    echo ""

    local avg_time=$(echo "scale=3; $total_time / $iterations" | bc -l 2>/dev/null || echo "0")
    local success_rate=$((success_count * 100 / iterations))

    echo "基础性能测试结果:"
    echo "  成功率: $success_rate% ($success_count/$iterations)"
    echo "  平均响应时间: ${avg_time}s"
    echo "  总执行时间: ${total_time}s"

    # 记录结果
    cat >> "$RESULTS_DIR/basic_performance.txt" << EOF
基础性能测试 - $(date)
成功请求: $success_count/$iterations ($success_rate%)
平均响应时间: ${avg_time}s
总执行时间: ${total_time}s
EOF

    if [ $success_rate -ge 95 ]; then
        log_pass "基础性能测试通过 (成功率: $success_rate%)"
        return 0
    else
        log_fail "基础性能测试失败 (成功率: $success_rate%)"
        return 1
    fi
}

# 使用 wrk 进行压力测试
run_wrk_stress_test() {
    local concurrency="$1"
    local duration="$2"
    local test_name="$3"

    if ! command -v wrk >/dev/null 2>&1; then
        log_skip "跳过 wrk 压力测试 (工具不可用)"
        return 0
    fi

    log_test "执行 wrk 压力测试: $test_name"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local result_file="$RESULTS_DIR/wrk_${test_name}_${TIMESTAMP}.txt"

    echo "执行 wrk 压力测试..."
    echo "  并发连接: $concurrency"
    echo "  测试时长: ${duration}s"
    echo "  目标URL: $MOCK_API/api/test/load"

    # 执行 wrk 测试
    wrk -t4 -c"$concurrency" -d"${duration}s" \
        --timeout 10s \
        --latency \
        -H "Content-Type: application/json" \
        --script <(echo 'wrk.method = "POST"
wrk.body = \'{"test": "load"}\'
wrk.headers["Content-Type"] = "application/json"') \
        "$MOCK_API/api/test/load" > "$result_file" 2>&1

    # 分析结果
    if [ -f "$result_file" ]; then
        local requests=$(grep "requests in" "$result_file" | awk '{print $1}' || echo "0")
        local latency_avg=$(grep "Latency" "$result_file" | awk '{print $2}' || echo "0")
        local rps=$(grep "requests/sec" "$result_file" | awk '{print $1}' || echo "0")

        echo "wrk 测试结果:"
        echo "  总请求数: $requests"
        echo "  平均延迟: $latency_avg"
        echo "  RPS: $rps"

        if [ "$requests" -gt 0 ]; then
            log_pass "wrk 压力测试完成: $test_name"
            return 0
        else
            log_fail "wrk 压力测试失败: $test_name"
            return 1
        fi
    else
        log_fail "wrk 压力测试结果文件未生成"
        return 1
    fi
}

# 使用 Apache Bench 进行压力测试
run_ab_stress_test() {
    local concurrency="$1"
    local requests="$2"
    local test_name="$3"

    if ! command -v ab >/dev/null 2>&1; then
        log_skip "跳过 Apache Bench 压力测试 (工具不可用)"
        return 0
    fi

    log_test "执行 Apache Bench 压力测试: $test_name"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local result_file="$RESULTS_DIR/ab_${test_name}_${TIMESTAMP}.txt"

    echo "执行 Apache Bench 压力测试..."
    echo "  并发连接: $concurrency"
    echo "  请求数量: $requests"

    # 执行 ab 测试
    ab -n "$requests" -c "$concurrency" \
        -T "application/json" \
        -p <(echo '{"test": "benchmark"}') \
        -k \
        "$MOCK_API/api/test/benchmark" > "$result_file" 2>&1

    # 分析结果
    if [ -f "$result_file" ]; then
        local rps=$(grep "Requests per second" "$result_file" | awk '{print $4}' || echo "0")
        local time_per_req=$(grep "Time per request" "$result_file" | head -1 | awk '{print $4}' || echo "0")
        local failed=$(grep "Failed requests" "$result_file" | awk '{print $3}' || echo "0")

        echo "Apache Bench 测试结果:"
        echo "  RPS: $rps"
        echo "  每请求时间: ${time_per_req}ms"
        echo "  失败请求: $failed"

        # 转换成功率
        local success_rate=$(( (requests - failed) * 100 / requests ))
        if [ $success_rate -ge 95 ]; then
            log_pass "Apache Bench 压力测试通过: $test_name (成功率: $success_rate%)"
            return 0
        else
            log_fail "Apache Bench 压力测试失败: $test_name (成功率: $success_rate%)"
            return 1
        fi
    else
        log_fail "Apache Bench 压力测试结果文件未生成"
        return 1
    fi
}

# 使用 hey 进行压力测试
run_hey_stress_test() {
    local concurrency="$1"
    local duration="$2"
    local test_name="$3"

    if ! command -v hey >/dev/null 2>&1; then
        log_skip "跳过 hey 压力测试 (工具不可用)"
        return 0
    fi

    log_test "执行 hey 压力测试: $test_name"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local result_file="$RESULTS_DIR/hey_${test_name}_${TIMESTAMP}.txt"

    echo "执行 hey 压力测试..."
    echo "  并发连接: $concurrency"
    echo "  测试时长: ${duration}s"

    # 执行 hey 测试
    hey -n 0 -z "${duration}s" \
        -c "$concurrency" \
        -H "Content-Type: application/json" \
        -d '{"test": "hey"}' \
        "$MOCK_API/api/test/hey" > "$result_file" 2>&1

    # 分析结果
    if [ -f "$result_file" ]; then
        local status_distribution=$(grep -A 5 "Status code distribution" "$result_file" || echo "")
        local requests=$(grep "requests" "$result_file" | grep "total" | awk '{print $1}' || echo "0")
        local rps=$(grep "Requests/sec" "$result_file" | awk '{print $2}' || echo "0")

        echo "hey 测试结果:"
        echo "  总请求数: $requests"
        echo "  RPS: $rps"
        echo "$status_distribution"

        if [ "$requests" -gt 0 ]; then
            log_pass "hey 压力测试完成: $test_name"
            return 0
        else
            log_fail "hey 压力测试失败: $test_name"
            return 1
        fi
    else
        log_fail "hey 压力测试结果文件未生成"
        return 1
    fi
}

# 长时间稳定性测试
run_stability_test() {
    log_test "执行长时间稳定性测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local stability_duration=60  # 60秒稳定性测试
    local check_interval=10     # 每10秒检查一次
    local max_response_time=5   # 最大可接受响应时间(秒)
    local success_count=0
    local total_checks=0
    local slow_responses=0

    echo "执行 $stability_duration 秒稳定性测试..."
    echo "检查间隔: ${check_interval}s"
    echo "最大可接受响应时间: ${max_response_time}s"

    local end_time=$(( $(date +%s) + stability_duration ))

    while [ $(date +%s) -lt $end_time ]; do
        total_checks=$((total_checks + 1))

        # 记录开始时间
        local start_time=$(date +%s)

        # 执行请求
        local response=$(curl -s -w "%{http_code}" \
            -H "Content-Type: application/json" \
            -d '{"test": "stability"}' \
            "$MOCK_API/api/test/stability" 2>/dev/null || echo "000")

        local end_time_req=$(date +%s)
        local response_time=$((end_time_req - start_time))

        if [ "$response" = "200" ]; then
            success_count=$((success_count + 1))
        fi

        if [ $response_time -gt $max_response_time ]; then
            slow_responses=$((slow_responses + 1))
            echo "  慢响应警告: ${response_time}s (阈值: ${max_response_time}s)"
        fi

        echo -n "."
        sleep $check_interval
    done
    echo ""

    local success_rate=$((success_count * 100 / total_checks))
    local stability_score=$((success_rate - (slow_responses * 10 / total_checks)))

    echo "稳定性测试结果:"
    echo "  测试时长: ${stability_duration}s"
    echo "  检查次数: $total_checks"
    echo "  成功请求: $success_count"
    echo "  成功率: $success_rate%"
    echo "  慢响应: $slow_responses"
    echo "  稳定性评分: $stability_score"

    # 记录结果
    cat >> "$RESULTS_DIR/stability_test.txt" << EOF
稳定性测试 - $(date)
测试时长: ${stability_duration}s
检查次数: $total_checks
成功请求: $success_count
成功率: $success_rate%
慢响应: $slow_responses
稳定性评分: $stability_score
EOF

    if [ $success_rate -ge 95 ] && [ $slow_responses -lt $((total_checks / 10)) ]; then
        log_pass "长时间稳定性测试通过"
        return 0
    else
        log_fail "长时间稳定性测试失败"
        return 1
    fi
}

# 内存使用监控
monitor_memory_usage() {
    log_test "监控内存使用情况"

    local memory_info_file="$RESULTS_DIR/memory_usage_${TIMESTAMP}.txt"
    local duration=30
    local interval=5

    echo "监控内存使用 ${duration}s (间隔: ${interval}s)..."

    for i in $(seq 1 $((duration / interval))); do
        echo "=== 内存监控 $(date) ===" >> "$memory_info_file"

        # 系统内存
        if command -v free >/dev/null 2>&1; then
            free -h >> "$memory_info_file" 2>/dev/null
        fi

        # MockServer 进程内存
        local mockserver_pid=$(pgrep -f "mockserver" | head -1)
        if [ -n "$mockserver_pid" ]; then
            echo "MockServer PID: $mockserver_pid" >> "$memory_info_file"
            ps -p "$mockserver_pid" -o pid,ppid,pcpu,pmem,rss,vsz,etime,cmd >> "$memory_info_file" 2>/dev/null
        fi

        # Redis 内存
        if command -v redis-cli >/dev/null 2>&1; then
            echo "Redis 内存信息:" >> "$memory_info_file"
            redis-cli info memory | grep used_memory: >> "$memory_info_file" 2>/dev/null
        fi

        echo "" >> "$memory_info_file"
        sleep $interval
    done

    log_pass "内存使用监控完成"
    return 0
}

# 生成压力测试报告
generate_stress_report() {
    log_test "生成压力测试综合报告"

    cat > "$REPORT_FILE" << EOF
# MockServer 压力测试报告

## 测试概要

- **测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **测试持续时间**: $(($(date +%s) - START_TIME)) 秒
- **测试环境**: $(uname -s) $(uname -r)
- **MockServer 端点**: $MOCK_API

## 测试结果统计

### 总体结果
- **总测试数**: $TOTAL_TESTS
- **通过测试**: $PASSED_TESTS
- **失败测试**: $FAILED_TESTS
- **总体通过率**: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%

### 测试覆盖
EOF

    # 添加各种测试结果
    if [ -f "$RESULTS_DIR/basic_performance.txt" ]; then
        cat >> "$REPORT_FILE" << EOF

#### 基础性能测试
\`\`\`
$(cat "$RESULTS_DIR/basic_performance.txt")
\`\`\`
EOF
    fi

    if [ -f "$RESULTS_DIR/stability_test.txt" ]; then
        cat >> "$REPORT_FILE" << EOF

#### 稳定性测试
\`\`\`
$(cat "$RESULTS_DIR/stability_test.txt")
\`\`\`
EOF
    fi

    # 添加压力测试结果摘要
    echo "" >> "$REPORT_FILE"
    echo "## 压力测试详情" >> "$REPORT_FILE"

    for result_file in "$RESULTS_DIR"/wrk_*_${TIMESTAMP}.txt "$RESULTS_DIR"/ab_*_${TIMESTAMP}.txt "$RESULTS_DIR"/hey_*_${TIMESTAMP}.txt; do
        if [ -f "$result_file" ]; then
            local test_name=$(basename "$result_file" | sed "s/_${TIMESTAMP}.txt//")
            echo "" >> "$REPORT_FILE"
            echo "### $test_name" >> "$REPORT_FILE"
            echo "\`\`\`" >> "$REPORT_FILE"
            cat "$result_file" >> "$REPORT_FILE"
            echo "\`\`\`" >> "$REPORT_FILE"
        fi
    done

    cat >> "$REPORT_FILE" << EOF

## 性能基准

### 响应时间基准
- **优秀**: < 100ms
- **良好**: 100-500ms
- **可接受**: 500ms-1s
- **需要优化**: > 1s

### 吞吐量基准
- **优秀**: > 1000 RPS
- **良好**: 500-1000 RPS
- **可接受**: 100-500 RPS
- **需要优化**: < 100 RPS

### 成功率基准
- **优秀**: > 99.5%
- **良好**: 95-99.5%
- **可接受**: 90-95%
- **需要优化**: < 90%

## 建议和改进

### 性能优化建议
1. **响应时间优化**: 如平均响应时间超过500ms，建议检查数据库查询效率
2. **并发处理**: 如RPS低于预期，建议检查连接池配置和并发处理能力
3. **内存使用**: 监控内存泄漏，确保长期运行稳定性
4. **错误处理**: 优化错误处理逻辑，减少失败率

### 压力测试工具对比
- **wrk**: 适合高并发HTTP负载测试
- **ab**: Apache Bench，简单易用的基准测试工具
- **hey**: Go语言编写的现代化负载测试工具

## 测试环境信息

- **操作系统**: $(uname -s) $(uname -r)
- **处理器**: $(uname -m)
- **Go版本**: $(go version 2>/dev/null || echo "Unknown")
- **测试时间**: $(date)
- **MockServer版本**: $(./mockserver --version 2>/dev/null || echo "Unknown")

---

*报告生成时间: $(date)*
*测试工具: MockServer E2E Stress Test Suite*
EOF

    log_pass "压力测试报告已生成: $REPORT_FILE"
    echo -e "${CYAN}报告路径: $REPORT_FILE${NC}"
}

# 主执行函数
main() {
    # 记录开始时间
    START_TIME=$(date +%s)

    # 显示横幅
    show_banner

    # 使用统一的服务协调
    log_test "启动依赖服务"
    if ! coordinate_services; then
        echo -e "${RED}✗ 服务启动失败${NC}"
        exit 1
    fi

    echo -e "${CYAN}开始执行压力测试...${NC}"
    echo ""

    # 检查工具
    if ! check_stress_tools; then
        echo -e "${RED}压力测试工具检查失败，但继续执行基础测试${NC}"
        echo ""
    fi

    # 执行测试套件
    local tests=(
        "basic_performance_test"
    )

    # 根据可用工具添加压力测试
    if command -v wrk >/dev/null 2>&1; then
        for config in "${STRESS_CONFIGS[@]}"; do
            IFS=':' read -r concurrency duration description <<< "$config"
            tests+=("run_wrk_stress_test $concurrency $duration $description")
        done
    fi

    if command -v ab >/dev/null 2>&1; then
        for config in "${STRESS_CONFIGS[@]}"; do
            IFS=':' read -r concurrency duration description <<< "$config"
            local requests=$((concurrency * duration / 2))
            tests+=("run_ab_stress_test $concurrency $requests $description")
        done
    fi

    if command -v hey >/dev/null 2>&1; then
        for config in "${STRESS_CONFIGS[@]}"; do
            IFS=':' read -r concurrency duration description <<< "$config"
            tests+=("run_hey_stress_test $concurrency $duration $description")
        done
    fi

    tests+=(
        "run_stability_test"
        "monitor_memory_usage"
    )

    local passed=0
    local failed=0

    for test_cmd in "${tests[@]}"; do
        if $test_cmd; then
            passed=$((passed + 1))
        else
            failed=$((failed + 1))
        fi
        echo ""
    done

    # 生成综合报告
    generate_stress_report

    # 显示测试结果
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   压力测试结果${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo ""
    echo -e "${CYAN}测试统计:${NC}"
    echo -e "  总测试数: $TOTAL_TESTS"
    echo -e "  通过: ${GREEN}$passed${NC}"
    echo -e "  失败: ${RED}$failed${NC}"
    echo -e "  成功率: $(( passed * 100 / TOTAL_TESTS ))%"
    echo ""
    echo -e "${CYAN}测试结果文件:${NC}"
    echo -e "  结果目录: $RESULTS_DIR"
    echo -e "  综合报告: $REPORT_FILE"
    echo ""

    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 所有压力测试通过！系统性能稳定。${NC}"
        exit 0
    else
        echo -e "${YELLOW}⚠️  有 $failed 个测试失败，建议进行性能优化${NC}"
        exit 1
    fi
}

# 信号处理
trap 'echo -e "\n${YELLOW}压力测试被中断，正在清理...${NC}"; exit 1' INT TERM

# 执行主函数
main
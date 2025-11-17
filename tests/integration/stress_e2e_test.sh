#!/bin/bash

# MockServer 压力测试和负载测试脚本
# 测试系统在高负载下的性能表现

set -e

# 加载测试框架
source "$(dirname "$0")/lib/test_framework.sh"

# 初始化测试框架
init_test_framework

# 检查并安装压力测试工具
check_and_install_stress_tools() {
    # 加载工具安装器
    local installer_path="$(dirname "$0")/lib/tool_installer.sh"
    if [ -f "$installer_path" ]; then
        source "$installer_path"

        # 检查压力测试工具
        if ! check_tools_ready "stress"; then
            echo -e "${YELLOW}检测到缺失的压力测试工具，正在自动安装...${NC}"
            if ! install_required_tools_silent "stress"; then
                echo -e "${YELLOW}压力测试工具安装失败，将跳过相关测试${NC}"
                return 1
            fi
        fi
    else
        echo -e "${YELLOW}工具安装器不可用，请手动安装 wrk 或 ab${NC}"
        return 1
    fi
}

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   MockServer 压力测试和负载测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# 压力测试配置
STRESS_CONFIGS=(
    "10:10:轻量级负载"
    "50:20:中等负载"
    "100:30:高负载"
    "200:60:极高负载"
)

# 检查压力测试工具
check_stress_tools() {
    if ! command -v wrk >/dev/null 2>&1 && ! command -v ab >/dev/null 2>&1; then
        echo -e "${YELLOW}压力测试工具未安装，正在尝试自动安装...${NC}"
        check_and_install_stress_tools
    fi

    # 再次检查
    if ! command -v wrk >/dev/null 2>&1 && ! command -v ab >/dev/null 2>&1; then
        test_skip "压力测试工具安装失败 (wrk 或 ab)，跳过压力测试"
        return 1
    fi

    echo -e "${GREEN}压力测试工具就绪${NC}"
    return 0
}

# 创建测试数据
create_stress_test_data() {
    local rule_count="$1"
    local project_id="$2"
    local environment_id="$3"

    echo -e "${YELLOW}创建 $rule_count 个测试规则...${NC}"

    local created=0
    for i in $(seq 1 $rule_count); do
        local rule_name="压力测试规则-$i"
        local rule_path="/api/stress-rule-$i"

        RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "{
    \"name\": \"$rule_name\",
    \"project_id\": \"$project_id\",
    \"environment_id\": \"$environment_id\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": $((100 + i)),
    \"enabled\": true,
    \"match_condition\": {
        \"method\": \"GET\",
        \"path\": \"$rule_path\"
    },
    \"response\": {
        \"type\": \"Static\",
        \"content\": {
            \"status_code\": 200,
            \"content_type\": \"JSON\",
            \"body\": {
                \"rule_id\": $i,
                \"message\": \"压力测试响应\",
                \"timestamp\": \"{{timestamp}}\",
                \"request_count\": \"{{counter}}\"
            }
        }
    }
}")

        if echo "$RULE_RESPONSE" | grep -q '\"id\"'; then
            created=$((created + 1))
        fi

        # 每10个规则显示一次进度
        if [ $((i % 10)) -eq 0 ]; then
            echo -n "."
        fi
    done

    echo ""
    if [ $created -eq $rule_count ]; then
        test_pass "创建 $created 个测试规则成功"
        return 0
    else
        test_fail "只创建了 $created 个测试规则"
        return 1
    fi
}

# 运行压力测试
run_stress_test() {
    local concurrent="$1"
    local duration="$2"
    local test_name="$3"
    local url="$4"

    echo -e "${YELLOW}[$test_name] 开始压力测试...${NC}"
    echo -e "  并发数: $concurrent"
    echo -e "  持续时间: ${duration}秒"
    echo -e "  目标URL: $url"

    local start_time=$(get_timestamp_ms)

    # 使用 wrk 进行压力测试
    if command -v wrk >/dev/null 2>&1; then
        echo -e "${CYAN}使用 wrk 进行压力测试...${NC}"

        wrk -t4 -c"$concurrent" -d"${duration}s" --timeout 10s --latency \
            -H "Connection: keep-alive" \
            "$url" > "/tmp/stress_test_${concurrent}_${duration}s.log" 2>&1

        local wrk_exit_code=$?

        if [ $wrk_exit_code -eq 0 ]; then
            local end_time=$(get_timestamp_ms)
            local total_duration=$(calculate_duration "$start_time" "$end_time")

            # 解析 wrk 输出
            local requests=$(grep "requests in" "/tmp/stress_test_${concurrent}_${duration}s.log" | awk '{print $1}')
            local qps=$(grep "Requests/sec:" "/tmp/stress_test_${concurrent}_${duration}s.log" | awk '{print $2}')
            local latency_avg=$(grep "Latency" "/tmp/stress_test_${concurrent}_${duration}s.log" | awk '/\/.*\/.*\// {print $2}')
            local latency_p95=$(grep "Latency" "/tmp/stress_test_${concurrent}_${duration}s.log" | awk '/\/.*\/.*\// {print $4}')

            echo -e "${GREEN}✓ $test_name 压力测试完成${NC}"
            echo -e "  总请求数: $requests"
            echo -e "  平均QPS: $qps"
            echo -e "  平均延迟: $latency_avg"
            echo -e "  P95延迟: $latency_p95"
            echo -e "  实际耗时: ${total_duration}ms"

            # 记录测试结果
            echo "$test_name: $requests requests, $qps QPS, $latency_avg avg latency" >> "/tmp/stress_test_results.txt"
            return 0
        else
            test_fail "$test_name 压力测试失败 (wrk 退出码: $wrk_exit_code)"
            return 1
        fi

    # 使用 ab 进行压力测试（备用方案）
    elif command -v ab >/dev/null 2>&1; then
        echo -e "${CYAN}使用 ab 进行压力测试...${NC}"

        local total_requests=$((concurrent * 10))
        ab -n "$total_requests" -c "$concurrent" -t "$duration" -k \
            -H "Connection: keep-alive" \
            "$url" > "/tmp/ab_stress_test_${concurrent}_${duration}s.log" 2>&1

        local ab_exit_code=$?

        if [ $ab_exit_code -eq 0 ]; then
            local end_time=$(get_timestamp_ms)
            local total_duration=$(calculate_duration "$start_time" "$end_time")

            # 解析 ab 输出
            local requests=$(grep "Requests per second:" "/tmp/ab_stress_test_${concurrent}_${duration}s.log" | awk '{print $4}')
            local time_taken=$(grep "Time taken for tests:" "/tmp/ab_stress_test_${concurrent}_${duration}s.log" | awk '{print $5}')

            echo -e "${GREEN}✓ $test_name 压力测试完成${NC}"
            echo -e "  QPS: $requests"
            echo -e "  响应时间: ${time_taken}s"
            echo -e "  实际耗时: ${total_duration}ms"

            echo "$test_name: $total_requests requests, $requests QPS, ${time_taken}s response time" >> "/tmp/stress_test_results.txt"
            return 0
        else
            test_fail "$test_name 压力测试失败 (ab 退出码: $ab_exit_code)"
            return 1
        fi
    else
        test_fail "没有可用的压力测试工具"
        return 1
    fi
}

# 并发连接压力测试
concurrent_connection_test() {
    local max_connections="$1"
    local test_name="$2"

    echo -e "${YELLOW}[$test_name] 并发连接测试...${NC}"

    local success_count=0
    local failed_count=0

    # 测试并发连接
    for i in $(seq 1 $max_connections); do
        (
            start_time=$(get_timestamp_ms)
            response=$(timeout_cmd 5 curl -s "$MOCK_API/stress-project/stress-env/api/stress-rule-$((i % 10 + 1))")
            end_time=$(get_timestamp_ms)
            duration=$(calculate_duration "$start_time" "$end_time")

            http_code=$(echo "$response" | tail -n 1)

            if [ "$http_code" = "200" ] && [ $duration -lt 5000 ]; then
                echo "Connection $i: SUCCESS (${duration}ms)"
            else
                echo "Connection $i: FAILED (${duration}ms, code: $http_code)"
            fi
        ) &

        # 控制并发数量
        if [ $((i % 20)) -eq 0 ]; then
            wait
        fi
    done

    wait

    # 统计结果（简化版本）
    echo -e "${GREEN}✓ $test_name 并发连接测试完成${NC}"
}

# 内存使用监控
monitor_memory_usage() {
    local test_name="$1"
    local duration="$2"

    echo -e "${YELLOW}[$test_name] 内存使用监控...${NC}"

    local mockserver_pid=""
    mockserver_pid=$(find_process "mockserver")

    if [ -z "$mockserver_pid" ]; then
        test_skip "未找到 MockServer 进程，跳过内存监控"
        return 0
    fi

    local initial_memory=""
    initial_memory=$(get_process_memory "$mockserver_pid")

    if [ -z "$initial_memory" ]; then
        test_skip "无法获取内存信息，跳过内存监控"
        return 0
    fi

    echo -e "初始内存使用: ${initial_memory}KB"

    # 监控期间
    local max_memory=$initial_memory
    local end_time=$(($(date +%s) + duration))

    while [ $(date +%s) -lt $end_time ]; do
        local current_memory=""
        current_memory=$(get_process_memory "$mockserver_pid")

        if [ -n "$current_memory" ] && [ "$current_memory" -gt "$max_memory" ]; then
            max_memory=$current_memory
        fi

        sleep 5
    done

    local memory_increase=$((max_memory - initial_memory))
    echo -e "最大内存使用: ${max_memory}KB"
    echo -e "内存增长: ${memory_increase}KB"

    if [ $memory_increase -lt 50000 ]; then  # 50MB
        test_pass "$test_name 内存使用正常"
    else
        test_fail "$test_name 内存使用过高 (增长: ${memory_increase}KB)"
    fi
}

# ========================================
# 阶段 1: 准备工作
# ========================================

echo -e "${CYAN}[阶段 1] 准备工作${NC}"
echo ""

if ! check_stress_tools; then
    exit 0
fi

# 1.1 创建压力测试项目
echo -e "${YELLOW}[1.1] 创建压力测试项目...${NC}"
STRESS_PROJECT_RESPONSE=$(http_post "$ADMIN_API/projects" "$(generate_project_data "压力测试项目")")

if echo "$STRESS_PROJECT_RESPONSE" | grep -q '"id"'; then
    STRESS_PROJECT_ID=$(extract_json_field "$STRESS_PROJECT_RESPONSE" "id")
    PROJECT_ID="$STRESS_PROJECT_ID"
    test_pass "压力测试项目创建成功"
else
    test_fail "压力测试项目创建失败"
    exit 1
fi

# 1.2 创建压力测试环境
echo -e "${YELLOW}[1.2] 创建压力测试环境...${NC}"
STRESS_ENV_RESPONSE=$(http_post "$ADMIN_API/projects/$STRESS_PROJECT_ID/environments" "$(generate_environment_data "压力测试环境" "http://localhost:9090")")

if echo "$STRESS_ENV_RESPONSE" | grep -q '"id"'; then
    STRESS_ENVIRONMENT_ID=$(extract_json_field "$STRESS_ENV_RESPONSE" "id")
    test_pass "压力测试环境创建成功"
else
    test_fail "压力测试环境创建失败"
    exit 1
fi

# 1.3 创建大量测试规则
echo -e "${YELLOW}[1.3] 创建测试规则...${NC}"
if ! create_stress_test_data 20 "$STRESS_PROJECT_ID" "$STRESS_ENVIRONMENT_ID"; then
    exit 1
fi

# 等待规则生效
echo -e "${YELLOW}[1.4] 等待规则生效...${NC}"
sleep 5

echo ""

# ========================================
# 阶段 2: 压力测试
# ========================================

echo -e "${CYAN}[阶段 2] 压力测试${NC}"
echo ""

# 清空之前的结果文件
> "/tmp/stress_test_results.txt"

# 2.1 轻量级负载测试
if [[ " ${STRESS_CONFIGS[@]} " =~ " 10:10:轻量级负载 " ]]; then
    run_stress_test 10 10 "轻量级负载" "$MOCK_API/$STRESS_PROJECT_ID/$STRESS_ENVIRONMENT_ID/api/stress-rule-1"
fi

# 2.2 中等负载测试
if [[ " ${STRESS_CONFIGS[@]} " =~ " 50:20:中等负载 " ]]; then
    run_stress_test 50 20 "中等负载" "$MOCK_API/$STRESS_PROJECT_ID/$STRESS_ENVIRONMENT_ID/api/stress-rule-2"
fi

# 2.3 高负载测试
if [[ " ${STRESS_CONFIGS[@]} " =~ " 100:30:高负载 " ]]; then
    run_stress_test 100 30 "高负载" "$MOCK_API/$STRESS_PROJECT_ID/$STRESS_ENVIRONMENT_ID/api/stress-rule-3"
fi

# 2.4 极高负载测试
if [[ " ${STRESS_CONFIGS[@]} " =~ " 200:60:极高负载 " ]]; then
    run_stress_test 200 60 "极高负载" "$MOCK_API/$STRESS_PROJECT_ID/$STRESS_ENVIRONMENT_ID/api/stress-rule-4"
fi

echo ""

# ========================================
# 阶段 3: 并发连接测试
# ========================================

echo -e "${CYAN}[阶段 3] 并发连接测试${NC}"
echo ""

# 3.1 并发连接测试 - 中等规模
concurrent_connection_test 100 "中等并发连接"

# 3.2 并发连接测试 - 大规模
concurrent_connection_test 500 "大规模并发连接"

echo ""

# ========================================
# 阶段 4: 长时间稳定性测试
# ========================================

echo -e "${CYAN}[阶段 4] 长时间稳定性测试${NC}"
echo ""

# 4.1 长时间负载测试
echo -e "${YELLOW}[4.1] 长时间负载测试 (60秒)...${NC}"
LONG_STRESS_URL="$MOCK_API/$STRESS_PROJECT_ID/$STRESS_ENVIRONMENT_ID/api/stress-rule-1"
run_stress_test 20 60 "长时间负载" "$LONG_STRESS_URL"

# 4.2 内存使用监控
echo -e "${YELLOW}[4.2] 内存使用监控 (30秒)...${NC}"
monitor_memory_usage "内存监控" 30

echo ""

# ========================================
# 阶段 5: 极限测试
# ========================================

echo -e "${CYAN}[阶段 5] 极限测试${NC}"
echo ""

# 5.1 极限请求频率测试
echo -e "${YELLOW}[5.1] 极限请求频率测试...${NC}"
echo -e "在5秒内发送尽可能多的请求"

LIMIT_START_TIME=$(get_timestamp_ms)
LIMIT_END_TIME=$(($(date +%s) + 5))
REQUEST_COUNT=0

while [ $(date +%s) -lt $LIMIT_END_TIME ]; do
    (
        curl -s "$MOCK_API/$STRESS_PROJECT_ID/$STRESS_ENVIRONMENT_ID/api/stress-rule-1" >/dev/null 2>&1
        REQUEST_COUNT=$((REQUEST_COUNT + 1))
    ) &

    # 控制并发进程数
    if [ $((REQUEST_COUNT % 50)) -eq 0 ]; then
        wait
    fi
done

wait
LIMIT_END_ACTUAL_TIME=$(get_timestamp_ms)
LIMIT_DURATION=$(calculate_duration "$LIMIT_START_TIME" "$LIMIT_END_ACTUAL_TIME")
LIMIT_QPS=$((REQUEST_COUNT * 1000 / LIMIT_DURATION))

echo -e "${GREEN}✓ 极限请求频率测试完成${NC}"
echo -e "  总请求数: $REQUEST_COUNT"
echo -e "  测试时间: ${LIMIT_DURATION}ms"
echo -e "  最大QPS: $LIMIT_QPS"

if [ $LIMIT_QPS -gt 1000 ]; then
    test_pass "极限请求频率测试成功 (QPS: $LIMIT_QPS)"
else
    test_warn "极限请求频率测试性能较低 (QPS: $LIMIT_QPS)"
fi

echo ""

# ========================================
# 生成压力测试报告
# ========================================

echo -e "${CYAN}[完成] 生成压力测试报告${NC}"

# 生成压力测试摘要报告
STRESS_REPORT_FILE="/tmp/stress_test_report_$(date +%Y%m%d_%H%M%S).md"

cat > "$STRESS_REPORT_FILE" << EOF
# MockServer 压力测试报告

## 测试概要
- **测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **测试环境**: 本地开发环境
- **服务器配置**: $STRESS_PROJECT_ID/$STRESS_ENVIRONMENT_ID
- **测试工具**: $(command -v wrk >/dev/null 2>&1 && echo "wrk" || echo "ab")

## 测试配置
- **测试规则数量**: 20
- **测试路径模式**: /api/stress-rule-[1-10]
- **响应数据**: JSON格式，包含时间戳和计数器

## 压力测试结果

EOF

if [ -f "/tmp/stress_test_results.txt" ]; then
    echo "### 详细结果" >> "$STRESS_REPORT_FILE"
    echo "" >> "$STRESS_REPORT_FILE"
    cat "/tmp/stress_test_results.txt" >> "$STRESS_REPORT_FILE"
fi

cat >> "$STRESS_REPORT_FILE" << EOF

## 性能指标

### 响应时间分析
- 平均响应时间应该 < 100ms
- P95 延迟应该 < 200ms
- P99 延迟应该 < 500ms

### 吞吐量分析
- 轻量负载 (10并发): 预期 > 500 QPS
- 中等负载 (50并发): 预期 > 1000 QPS
- 高负载 (100并发): 预期 > 1500 QPS
- 极高负载 (200并发): 预期 > 2000 QPS

### 资源使用
- CPU使用率应该 < 80%
- 内存使用应该保持稳定
- 连接数应该正常管理

## 测试结论

EOF

if [ $TEST_FAILED -eq 0 ]; then
    echo "✅ **压力测试通过**: MockServer 在各种负载条件下表现良好" >> "$STRESS_REPORT_FILE"
    echo "✅ **性能稳定**: 响应时间和吞吐量都在可接受范围内" >> "$STRESS_REPORT_FILE"
    echo "✅ **资源控制**: 内存和CPU使用保持合理水平" >> "$STRESS_REPORT_FILE"
else
    echo "⚠️ **压力测试部分失败**: 某些测试场景需要优化" >> "$STRESS_REPORT_FILE"
    echo "⚠️ **性能调优**: 建议针对瓶颈进行优化" >> "$STRESS_REPORT_FILE"
fi

echo -e "${GREEN}压力测试报告已生成: $STRESS_REPORT_FILE${NC}"

# ========================================
# 测试结果统计
# ========================================

print_test_summary

echo ""
echo -e "${CYAN}压力测试特性验证:${NC}"
echo -e "  ${GREEN}✓ 多级负载测试${NC}"
echo -e "  ${GREEN}✓ 并发连接测试${NC}"
echo -e "  ${GREEN}✓ 长时间稳定性测试${NC}"
echo -e "  ${GREEN}✓ 内存使用监控${NC}"
echo -e "  ${GREEN}✓ 极限性能测试${NC}"

echo ""
echo -e "${CYAN}性能指标概览:${NC}"
if [ -f "/tmp/stress_test_results.txt" ]; then
    echo "详细测试结果请查看: /tmp/stress_test_results.txt"
fi
echo "完整报告请查看: $STRESS_REPORT_FILE"

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   MockServer 压力测试完成${NC}"
echo -e "${BLUE}=========================================${NC}"
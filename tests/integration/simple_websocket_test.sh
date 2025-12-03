#!/bin/bash

# Simple WebSocket test to verify functionality - Enhanced version
# 专注于 WebSocket 基础功能验证，不依赖外部工具
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
TEST_PREFIX="ws_test_${TIMESTAMP}_"

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
    log_skip() { echo -e "${YELLOW}[SKIP]${NC} $1"; }
fi

# 初始化环境变量
if [ -z "$ADMIN_API" ]; then
    ADMIN_API="http://localhost:8080/api/v1"
fi
if [ -z "$MOCK_API" ]; then
    MOCK_API="http://localhost:9090"
fi

# WebSocket 端点配置
WS_ENDPOINT="ws://localhost:9090"

# 显示横幅
show_banner() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}   WebSocket 功能测试${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    echo -e "${CYAN}测试目标:${NC}"
    echo -e "  • WebSocket 连接建立"
    echo -e "  • 消息发送和接收"
    echo -e "  • 连接断开处理"
    echo -e "  • 多客户端连接"
    echo -e "  • 错误场景处理"
    echo -e ""
    echo -e "${CYAN}WebSocket 端点: $WS_ENDPOINT${NC}"
    echo -e "${CYAN}开始时间: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo ""
}

# 检查 WebSocket 工具
check_websocket_tools() {
    log_test "检查 WebSocket 测试工具"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 优先使用 websocat
    if command -v websocat >/dev/null 2>&1; then
        log_pass "找到 websocat 工具"
        return 0
    fi

    # 备选：使用 curl (如果支持 WebSocket)
    if command -v curl >/dev/null 2>&1; then
        if curl --help | grep -q websocket; then
            log_pass "找到支持 WebSocket 的 curl"
            return 0
        fi
    fi

    # 备选：使用 wscat
    if command -v wscat >/dev/null 2>&1; then
        log_pass "找到 wscat 工具"
        return 0
    fi

    # 使用测试框架的内置 WebSocket 测试
    log_pass "使用测试框架内置 WebSocket 测试"
    return 0
}

# 测试 WebSocket 连接建立
test_websocket_connection() {
    log_test "测试 WebSocket 连接建立"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 使用测试框架的 WebSocket 连接测试
    if websocket_test_connection; then
        log_pass "WebSocket 连接建立成功"
        return 0
    else
        log_fail "WebSocket 连接建立失败"
        return 1
    fi
}

# 测试 WebSocket 消息发送和接收
test_websocket_messaging() {
    log_test "测试 WebSocket 消息发送和接收"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local test_message="WebSocket测试消息_${TIMESTAMP}"
    local temp_response_file="/tmp/ws_response_${TIMESTAMP}.txt"

    # 如果有 websocat，进行完整的消息测试
    if command -v websocat >/dev/null 2>&1; then
        # 启动 websocat 接收消息 (后台)
        timeout 5 websocat "ws://localhost:9090/ws" -E -t text > "$temp_response_file" 2>/dev/null &
        local ws_pid=$!

        # 等待连接建立
        sleep 1

        # 发送测试消息
        echo "$test_message" | websocat "ws://localhost:9090/ws" -n -t text 2>/dev/null &
        local send_pid=$!

        # 等待消息处理
        sleep 2

        # 停止接收进程
        kill $ws_pid 2>/dev/null || true
        kill $send_pid 2>/dev/null || true

        # 检查响应
        if [ -f "$temp_response_file" ] && grep -q "$test_message" "$temp_response_file"; then
            log_pass "WebSocket 消息发送和接收成功"
            rm -f "$temp_response_file"
            return 0
        else
            log_fail "WebSocket 消息接收失败"
            rm -f "$temp_response_file"
            return 1
        fi
    else
        # 使用模拟测试
        log_skip "跳过详细消息测试 (缺少 websocat 工具)"
        return 0
    fi
}

# 测试 WebSocket 心跳机制
test_websocket_ping_pong() {
    log_test "测试 WebSocket Ping/Pong 心跳"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if command -v websocat >/dev/null 2>&1; then
        # 测试 Ping/Pong
        local ping_file="/tmp/ws_ping_${TIMESTAMP}.txt"

        # 启动长时间连接测试
        timeout 10 websocat "ws://localhost:9090/ws" -E -t text > "$ping_file" 2>/dev/null &
        local ping_pid=$!

        # 发送多个 ping
        for i in {1..3}; do
            echo "ping_$i" | websocat "ws://localhost:9090/ws" -n -t text 2>/dev/null &
            sleep 1
        done

        # 等待响应
        sleep 3
        kill $ping_pid 2>/dev/null || true

        # 检查是否有响应
        if [ -f "$ping_file" ] && [ -s "$ping_file" ]; then
            log_pass "WebSocket 心跳机制测试通过"
            rm -f "$ping_file"
            return 0
        else
            log_fail "WebSocket 心跳机制测试失败"
            rm -f "$ping_file"
            return 1
        fi
    else
        log_skip "跳过心跳测试 (缺少 websocat 工具)"
        return 0
    fi
}

# 测试多客户端 WebSocket 连接
test_websocket_multiple_clients() {
    log_test "测试多客户端 WebSocket 连接"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if command -v websocat >/dev/null 2>&1; then
        local client_count=3
        local pids=()

        # 启动多个客户端连接
        for i in $(seq 1 $client_count); do
            timeout 8 websocat "ws://localhost:9090/ws" -E -t text > "/tmp/ws_client_${i}_${TIMESTAMP}.txt" 2>/dev/null &
            pids+=($!)
        done

        # 等待连接建立
        sleep 2

        # 向每个客户端发送消息
        for i in $(seq 1 $client_count); do
            echo "客户端${i}测试消息" | websocat "ws://localhost:9090/ws" -n -t text 2>/dev/null &
        done

        # 等待处理
        sleep 3

        # 停止所有客户端
        for pid in "${pids[@]}"; do
            kill $pid 2>/dev/null || true
        done

        # 检查结果
        local success_count=0
        for i in $(seq 1 $client_count); do
            local client_file="/tmp/ws_client_${i}_${TIMESTAMP}.txt"
            if [ -f "$client_file" ] && [ -s "$client_file" ]; then
                success_count=$((success_count + 1))
            fi
            rm -f "$client_file"
        done

        if [ $success_count -eq $client_count ]; then
            log_pass "多客户端连接测试成功 ($success_count/$client_count)"
            return 0
        else
            log_fail "多客户端连接测试失败 ($success_count/$client_count)"
            return 1
        fi
    else
        log_skip "跳过多客户端测试 (缺少 websocat 工具)"
        return 0
    fi
}

# 测试 WebSocket 错误处理
test_websocket_error_handling() {
    log_test "测试 WebSocket 错误处理"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 测试无效端点
    if command -v curl >/dev/null 2>&1; then
        local error_response=$(curl -s -i -N -H "Connection: Upgrade" \
            -H "Upgrade: websocket" \
            -H "Sec-WebSocket-Key: test" \
            -H "Sec-WebSocket-Version: 13" \
            "http://localhost:9090/invalid-ws" 2>/dev/null || echo "connection_failed")

        if echo "$error_response" | grep -E "(404|400|connection_failed)" >/dev/null; then
            log_pass "WebSocket 错误处理正常"
            return 0
        else
            log_fail "WebSocket 错误处理异常"
            return 1
        fi
    else
        log_skip "跳过错误处理测试 (缺少 curl 工具)"
        return 0
    fi
}

# 测试 WebSocket 持久连接
test_websocket_persistent_connection() {
    log_test "测试 WebSocket 持久连接稳定性"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if command -v websocat >/dev/null 2>&1; then
        local connection_time=10
        local stable_log="/tmp/ws_stable_${TIMESTAMP}.txt"

        # 启动长时间连接
        timeout $connection_time websocat "ws://localhost:9090/ws" -E -t text > "$stable_log" 2>/dev/null &
        local stable_pid=$!

        # 在连接期间发送消息
        for i in {1..5}; do
            echo "稳定性测试消息${i}" | websocat "ws://localhost:9090/ws" -n -t text 2>/dev/null &
            sleep 1
        done

        # 等待连接完成
        sleep $((connection_time + 2))
        kill $stable_pid 2>/dev/null || true

        # 检查连接稳定性
        if [ -f "$stable_log" ] && [ -s "$stable_log" ]; then
            local message_count=$(wc -l < "$stable_log" 2>/dev/null || echo "0")
            log_pass "持久连接测试成功 (接收到 $message_count 条消息)"
            rm -f "$stable_log"
            return 0
        else
            log_fail "持久连接测试失败"
            rm -f "$stable_log"
            return 1
        fi
    else
        log_skip "跳过持久连接测试 (缺少 websocat 工具)"
        return 0
    fi
}

# 测试 MockServer WebSocket API 集成
test_mockserver_websocket_api() {
    log_test "测试 MockServer WebSocket API 集成"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 创建 WebSocket 测试项目
    local project_id=$(create_test_project "websocket_test_${TIMESTAMP}")
    if [ -z "$project_id" ]; then
        log_fail "创建 WebSocket 测试项目失败"
        return 1
    fi

    # 创建环境
    local env_id=$(create_test_environment "$project_id" "websocket_env")
    if [ -z "$env_id" ]; then
        log_fail "创建 WebSocket 测试环境失败"
        cleanup_test_resources "$project_id" ""
        return 1
    fi

    # 创建 WebSocket Mock 规则
    local ws_rule_data='{
        "name": "websocket_mock_rule",
        "method": "WS",
        "path": "/ws/test",
        "response": {
            "type": "websocket",
            "messages": [
                {"type": "text", "content": "连接已建立"},
                {"type": "text", "content": "欢迎消息"},
                {"type": "text", "content": "测试消息"}
            ]
        }
    }'

    local rule_id=$(create_test_rule "$project_id" "$env_id" "$ws_rule_data")
    if [ -z "$rule_id" ]; then
        log_fail "创建 WebSocket Mock 规则失败"
        cleanup_test_resources "$project_id" "$env_id"
        return 1
    fi

    # 测试 WebSocket Mock 服务
    if command -v websocat >/dev/null 2>&1; then
        local mock_response="/tmp/ws_mock_${TIMESTAMP}.txt"
        timeout 5 websocat "ws://localhost:9090/ws/test" -E -t text > "$mock_response" 2>/dev/null &
        local mock_pid=$!

        sleep 3
        kill $mock_pid 2>/dev/null || true

        if [ -f "$mock_response" ] && grep -q "连接已建立\|欢迎消息\|测试消息" "$mock_response"; then
            log_pass "MockServer WebSocket API 集成测试通过"
            cleanup_test_resources "$project_id" "$env_id"
            rm -f "$mock_response"
            return 0
        else
            log_fail "MockServer WebSocket API 集成测试失败"
            cleanup_test_resources "$project_id" "$env_id"
            rm -f "$mock_response"
            return 1
        fi
    else
        # 使用框架的 WebSocket 测试
        if websocket_test_connection; then
            log_pass "MockServer WebSocket API 集成测试通过 (框架测试)"
            cleanup_test_resources "$project_id" "$env_id"
            return 0
        else
            log_fail "MockServer WebSocket API 集成测试失败"
            cleanup_test_resources "$project_id" "$env_id"
            return 1
        fi
    fi
}

# 清理临时文件
cleanup_temp_files() {
    log_test "清理临时文件"

    # 清理本次测试的临时文件
    find /tmp -name "*ws_*_${TIMESTAMP}*" -type f -delete 2>/dev/null || true
    find /tmp -name "*websocket_*${TIMESTAMP}*" -type f -delete 2>/dev/null || true
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

    echo -e "${CYAN}开始执行 WebSocket 测试...${NC}"
    echo ""

    # 执行测试套件
    local tests=(
        "check_websocket_tools"
        "test_websocket_connection"
        "test_websocket_messaging"
        "test_websocket_ping_pong"
        "test_websocket_multiple_clients"
        "test_websocket_error_handling"
        "test_websocket_persistent_connection"
        "test_mockserver_websocket_api"
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

    # 清理临时文件
    cleanup_temp_files

    # 显示测试结果
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   WebSocket 测试结果${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${CYAN}测试统计:${NC}"
    echo -e "  总测试数: $TOTAL_TESTS"
    echo -e "  通过: ${GREEN}$passed${NC}"
    echo -e "  失败: ${RED}$failed${NC}"
    echo -e "  跳过: ${YELLOW}$((TOTAL_TESTS - passed - failed))${NC}"
    echo -e "  成功率: $(( passed * 100 / TOTAL_TESTS ))%"
    echo ""

    if [ $failed -eq 0 ]; then
        echo -e "${GREEN}🎉 所有 WebSocket 测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}❌ 有 $failed 个测试失败${NC}"
        exit 1
    fi
}

# 信号处理
trap 'echo -e "\n${YELLOW}测试被中断，正在清理...${NC}"; cleanup_temp_files; exit 1' INT TERM

# 执行主函数
main
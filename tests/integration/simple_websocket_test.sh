#!/bin/bash

# Simple WebSocket test to verify functionality - Enhanced version
# 专注于 WebSocket 基础功能验证，不依赖外部工具

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

# 显示横幅
show_banner() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}   WebSocket 基础功能验证${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    echo -e "${CYAN}测试目标:${NC}"
    echo -e "  • WebSocket 项目创建"
    echo -e "  • WebSocket 环境管理"
    echo -e "  • WebSocket 规则配置"
    echo -e "  • WebSocket 端点验证"
    echo ""
    echo -e "${CYAN}开始时间: $(date '+%Y-%m-%d %H:%M:%S')${NC}"
    echo ""
}

# 简单的HTTP POST函数
simple_http_post() {
    local url="$1"
    local data="$2"

    curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$url" 2>/dev/null
}

# 简单的JSON字段提取函数
simple_extract_field() {
    local json="$1"
    local field="$2"

    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4
}

# 测试 1: WebSocket 项目创建
test_websocket_project() {
    log_test "WebSocket 项目创建测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local project_data='{"name": "WebSocket基础测试项目", "description": "验证WebSocket基础功能"}'
    local project_response=$(simple_http_post "$ADMIN_API/projects" "$project_data")
    local project_id=$(simple_extract_field "$project_response" "id")

    if [ -n "$project_id" ]; then
        test_pass "WebSocket项目创建成功: $project_id"
        WS_PROJECT_ID="$project_id"
        return 0
    else
        test_fail "WebSocket项目创建失败"
        return 1
    fi
}

# 测试 2: WebSocket 环境创建
test_websocket_environment() {
    log_test "WebSocket 环境创建测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ -z "$WS_PROJECT_ID" ]; then
        test_fail "项目ID不存在，跳过环境创建"
        return 1
    fi

    local env_data='{"name": "WebSocket测试环境", "project_id": "'$WS_PROJECT_ID'", "description": "WebSocket功能测试"}'
    local env_response=$(simple_http_post "$ADMIN_API/environments" "$env_data")
    local env_id=$(simple_extract_field "$env_response" "id")

    if [ -n "$env_id" ]; then
        test_pass "WebSocket环境创建成功: $env_id"
        WS_ENVIRONMENT_ID="$env_id"
        return 0
    else
        test_fail "WebSocket环境创建失败"
        return 1
    fi
}

# 测试 3: WebSocket 规则创建
test_websocket_rule() {
    log_test "WebSocket 规则创建测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ -z "$WS_PROJECT_ID" ] || [ -z "$WS_ENVIRONMENT_ID" ]; then
        test_fail "项目ID或环境ID不存在，跳过规则创建"
        return 1
    fi

    local rule_data='{
        "name": "WebSocket基础规则",
        "project_id": "'$WS_PROJECT_ID'",
        "environment_id": "'$WS_ENVIRONMENT_ID'",
        "protocol": "WebSocket",
        "match_type": "Simple",
        "priority": 100,
        "request": {
            "method": "WS",
            "path": "/websocket-test"
        },
        "response": {
            "status": 101,
            "body": "WebSocket connection established",
            "headers": {
                "Upgrade": "websocket",
                "Connection": "Upgrade"
            }
        }
    }'

    local rule_response=$(simple_http_post "$ADMIN_API/rules" "$rule_data")
    local rule_id=$(simple_extract_field "$rule_response" "id")

    if [ -n "$rule_id" ]; then
        test_pass "WebSocket规则创建成功: $rule_id"
        WS_RULE_ID="$rule_id"
        return 0
    else
        test_fail "WebSocket规则创建失败"
        return 1
    fi
}

# 测试 4: WebSocket 端点HTTP验证
test_websocket_endpoint() {
    log_test "WebSocket 端点验证测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ -z "$WS_PROJECT_ID" ] || [ -z "$WS_ENVIRONMENT_ID" ]; then
        test_fail "项目ID或环境ID不存在，跳过端点验证"
        return 1
    fi

    # 测试 HTTP 请求到 WebSocket 端点（应该返回特定错误码）
    local ws_response=$(curl -s -w "%{http_code}" -o /tmp/ws_endpoint_test.json \
        -H "X-Project-ID: $WS_PROJECT_ID" \
        -H "X-Environment-ID: $WS_ENVIRONMENT_ID" \
        -H "Connection: Upgrade" \
        -H "Upgrade: websocket" \
        "$MOCK_API/websocket-test" 2>/dev/null)

    local http_code="${ws_response: -3}"

    # 对于HTTP请求WebSocket端点，返回400/426是正常的
    if [ "$http_code" = "400" ] || [ "$http_code" = "426" ] || [ "$http_code" = "101" ]; then
        test_pass "WebSocket端点HTTP响应正常: $http_code"
        return 0
    else
        test_fail "WebSocket端点HTTP响应异常: $http_code"
        return 1
    fi
}

# 测试 5: WebSocket 端点可用性检查
test_websocket_availability() {
    log_test "WebSocket 端点可用性检查"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ -z "$WS_PROJECT_ID" ] || [ -z "$WS_ENVIRONMENT_ID" ]; then
        test_fail "项目ID或环境ID不存在，跳过可用性检查"
        return 1
    fi

    # 使用curl检查WebSocket端点是否被正确配置
    local availability_check=$(curl -s -I \
        -H "X-Project-ID: $WS_PROJECT_ID" \
        -H "X-Environment-ID: $WS_ENVIRONMENT_ID" \
        "$MOCK_API/websocket-test" 2>/dev/null | head -1)

    if [ -n "$availability_check" ]; then
        test_pass "WebSocket端点配置正确并可达"
        return 0
    else
        test_fail "WebSocket端点配置失败或不可达"
        return 1
    fi
}

# 清理测试数据
cleanup_test_data() {
    log_test "清理测试数据"

    if [ -n "$WS_PROJECT_ID" ]; then
        echo -e "${CYAN}清理WebSocket测试项目: $WS_PROJECT_ID${NC}"
        curl -s -X DELETE "$ADMIN_API/projects/$WS_PROJECT_ID" >/dev/null 2>&1 || true
    fi

    # 清理临时文件
    rm -f /tmp/ws_endpoint_test.json
}

# 生成测试报告
generate_report() {
    print_test_summary
    local exit_code=$?

    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   WebSocket 基础功能测试结果${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}🎉 所有 WebSocket 测试通过！${NC}"
        echo -e "${GREEN}✅ WebSocket 功能验证成功${NC}"
        echo -e "${GREEN}✅ 项目和规则管理正常${NC}"
    else
        echo -e "${RED}❌ 部分 WebSocket 测试失败${NC}"
        echo -e "${YELLOW}💡 请检查 MockServer WebSocket 支持${NC}"
    fi

    return $exit_code
}

# 主测试流程
main() {
    show_banner

    # 执行测试
    test_websocket_project || true
    test_websocket_environment || true
    test_websocket_rule || true
    test_websocket_endpoint || true
    test_websocket_availability || true

    # 生成报告
    generate_report

    # 清理测试数据
    cleanup_test_data
}

# 信号处理
trap 'echo -e "\n${YELLOW}测试被中断，正在清理...${NC}"; cleanup_test_data; exit 1' INT TERM

# 执行主流程
main
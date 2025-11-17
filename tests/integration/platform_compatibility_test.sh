#!/bin/bash

# 跨平台兼容性测试脚本
# 验证所有脚本在不同平台上的兼容性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   跨平台兼容性测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# 加载测试框架
source "$(dirname "$0")/lib/test_framework.sh"

# 测试结果统计
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"

    TESTS_TOTAL=$((TESTS_TOTAL + 1))

    echo -e "${YELLOW}[测试] $test_name${NC}"

    if eval "$test_command" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓ 通过${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "  ${RED}✗ 失败${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# 显示平台信息
echo -e "${CYAN}平台信息:${NC}"
echo -e "  操作系统: $(uname -s) $(uname -r)"
echo -e "  架构: $(uname -m)"
echo -e "  Shell: $SHELL"
echo -e "  Bash版本: $BASH_VERSION"
echo ""

# 测试基础命令兼容性
echo -e "${CYAN}基础命令兼容性测试:${NC}"

run_test "seq命令" "seq 1 5"
run_test "date命令" "date +%s"
run_test "curl命令" "curl --version"
run_test "python3命令" "python3 --version"
run_test "ps命令" "ps aux"
run_test "awk命令" "echo 'test' | awk '{print \$1}'"
run_test "grep命令" "echo 'test' | grep test"
run_test "tail命令" "echo -e 'test\ntest' | tail -n 1"
run_test "head命令" "echo -e 'test\ntest' | head -n 1"

echo ""

# 测试测试框架函数
echo -e "${CYAN}测试框架函数兼容性测试:${NC}"

run_test "时间戳获取" "get_timestamp_ms"
run_test "持续时间计算" "calculate_duration 1000 1500"
run_test "序列生成" "seq 1 3"
run_test "JSON字段提取" "extract_json_field '{\"id\": \"123\"}' 'id'"
run_test "进程查找" "find_process 'bash'"

echo ""

# 测试压力测试工具
echo -e "${CYAN}压力测试工具兼容性测试:${NC}"

run_test "wrk工具" "command -v wrk"
run_test "ab工具" "command -v ab"
run_test "websocat工具" "command -v websocat"

echo ""

# 测试脚本语法
echo -e "${CYAN}脚本语法检查:${NC}"

SCRIPTS=(
    "lib/test_framework.sh"
    "e2e_test.sh"
    "advanced_e2e_test.sh"
    "websocket_e2e_test.sh"
    "edge_case_e2e_test.sh"
    "stress_e2e_test.sh"
    "run_all_e2e_tests.sh"
)

for script in "${SCRIPTS[@]}"; do
    if [ -f "$(dirname "$0")/$script" ]; then
        run_test "脚本语法: $script" "bash -n $(dirname "$0")/$script"
    else
        echo -e "${YELLOW}跳过: $script (文件不存在)${NC}"
    fi
done

echo ""

# 测试网络连接
echo -e "${CYAN}网络连接测试:${NC}"

run_test "Admin API连接" "curl -s --max-time 5 http://localhost:8080/api/v1/system/health"
run_test "Mock API连接" "curl -s --max-time 5 http://localhost:9090"

echo ""

# 测试端口占用
echo -e "${CYAN}端口占用检查:${NC}"

if command -v lsof >/dev/null 2>&1; then
    run_test "端口8080检查" "lsof -i :8080 >/dev/null 2>&1"
    run_test "端口9090检查" "lsof -i :9090 >/dev/null 2>&1"
elif command -v netstat >/dev/null 2>&1; then
    run_test "端口8080检查" "netstat -an | grep :8080 >/dev/null 2>&1"
    run_test "端口9090检查" "netstat -an | grep :9090 >/dev/null 2>&1"
else
    echo -e "${YELLOW}跳过: 端口检查工具不可用${NC}"
fi

echo ""

# 显示测试结果
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   兼容性测试结果${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

echo -e "${CYAN}测试统计:${NC}"
echo -e "  总测试数: $TESTS_TOTAL"
echo -e "  通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "  失败: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "  通过率: ${GREEN}100%${NC}"
    echo ""
    echo -e "${GREEN}🎉 所有兼容性测试通过！${NC}"
    echo -e "${GREEN}✅ 系统准备就绪，可以运行 E2E 测试${NC}"
    exit 0
else
    local success_rate=$((TESTS_PASSED * 100 / TESTS_TOTAL))
    echo -e "  通过率: ${YELLOW}$success_rate%${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  部分兼容性测试失败${NC}"
    echo -e "${YELLOW}💡 建议检查失败的测试项${NC}"
    exit 1
fi
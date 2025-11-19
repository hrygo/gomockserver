#!/bin/bash

# Redis集成测试脚本 - 集成到主要测试框架中
# 使用统一的测试框架进行Redis功能测试

set -e

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# 加载测试框架
source "$PROJECT_ROOT/tests/integration/lib/test_framework.sh"

# 测试配置
TEST_NAME="Redis Integration Tests"
TEST_RESULTS_FILE="/tmp/redis_integration_results.txt"

# 主测试函数
main() {
    echo -e "${CYAN}=====================================${NC}"
    echo -e "${CYAN}  $TEST_NAME${NC}"
    echo -e "${CYAN}=====================================${NC}"
    echo ""

    # 初始化测试框架
    init_test_framework

    # 显示Redis配置信息
    echo -e "${BLUE}Redis Configuration:${NC}"
    echo -e "  Redis Host: ${YELLOW}${REDIS_HOST:-localhost}${NC}"
    echo -e "  Redis Port: ${YELLOW}${REDIS_PORT:-6379}${NC}"
    echo -e "  Redis URL: ${YELLOW}${REDIS_URL:-redis://localhost:6379}${NC}"
    echo ""

    # 检查Redis连接
    echo -e "${BLUE}Checking Redis connection...${NC}"
    if check_redis_connection; then
        test_pass "Redis connection established"
    else
        test_fail "Redis connection failed"
        echo ""
        echo -e "${YELLOW}To start Redis, run one of the following:${NC}"
        echo -e "  ${YELLOW}• make start-redis${NC}"
        echo -e "  ${YELLOW}• docker run -d --name mockserver-redis -p 6379:6379 redis:7-alpine${NC}"
        echo -e "  ${YELLOW}• make start-all${NC} (includes Redis)"
        exit 1
    fi

    echo ""

    # 运行Redis集成测试
    run_redis_integration_tests

    # 显示测试结果摘要
    echo ""
    print_test_summary

    # 生成测试报告
    generate_test_report "$TEST_RESULTS_FILE" "$TEST_NAME"

    # 检查是否有失败测试
    if [ $TEST_FAILED -eq 0 ]; then
        echo ""
        echo -e "${GREEN}🎉 All Redis integration tests passed!${NC}"
        exit 0
    else
        echo ""
        echo -e "${RED}❌ Some Redis integration tests failed.${NC}"
        echo -e "${YELLOW}Please check Redis configuration and status.${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"
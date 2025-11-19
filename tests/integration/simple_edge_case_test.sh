#!/bin/bash

# 简化的边界条件测试脚本
# 专注于基础边界条件验证，避免复杂语法

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
TEST_PREFIX="edge_test_${TIMESTAMP}_"

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
    echo -e "${CYAN}   边界条件简化测试${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    echo -e "${CYAN}测试目标:${NC}"
    echo -e "  • 基础边界条件验证"
    echo -e "  • 大数据量处理"
    echo -e "  • 特殊字符处理"
    echo -e "  • 错误场景验证"
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

# 简单的HTTP请求函数
simple_http_request() {
    local method="$1"
    local url="$2"
    local headers="$3"

    curl -s -X "$method" \
        -H "Content-Type: application/json" \
        $headers \
        "$url" 2>/dev/null
}

# 测试 1: 长路径处理
test_long_path() {
    log_test "长路径边界测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 创建测试项目
    local project_data='{"name": "长路径边界测试", "description": "测试长URL路径处理"}'
    local project_response=$(simple_http_post "$ADMIN_API/projects" "$project_data")
    local project_id=$(simple_extract_field "$project_response" "id")

    if [ -n "$project_id" ]; then
        test_pass "测试项目创建成功"

        # 创建测试环境
        local env_data='{"name": "长路径测试环境", "project_id": "'$project_id'", "description": "边界条件环境"}'
        local env_response=$(simple_http_post "$ADMIN_API/environments" "$env_data")
        local env_id=$(simple_extract_field "$env_response" "id")

        if [ -n "$env_id" ]; then
            test_pass "测试环境创建成功"

            # 创建长路径规则
            local long_path="/test/$(head -c 200 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 100)"
            local rule_data='{
                "name": "长路径规则",
                "project_id": "'$project_id'",
                "environment_id": "'$env_id'",
                "request": {
                    "method": "GET",
                    "path": "'$long_path'"
                },
                "response": {
                    "status": 200,
                    "body": "长路径测试成功"
                }
            }'

            local rule_response=$(simple_http_post "$ADMIN_API/rules" "$rule_data")
            local rule_id=$(simple_extract_field "$rule_response" "id")

            if [ -n "$rule_id" ]; then
                test_pass "长路径规则创建成功 (路径长度: ${#long_path})"

                # 测试长路径请求
                local path_response=$(simple_http_request "GET" \
                    "$MOCK_API/$long_path" \
                    "-H \"X-Project-ID: $project_id\" -H \"X-Environment-ID: $env_id\"")

                if [ -n "$path_response" ]; then
                    test_pass "长路径请求处理成功"

                    # 清理测试数据
                    curl -s -X DELETE "$ADMIN_API/projects/$project_id" >/dev/null 2>&1 || true
                    return 0
                else
                    test_fail "长路径请求处理失败"
                fi
            else
                test_fail "长路径规则创建失败"
            fi
        else
            test_fail "测试环境创建失败"
        fi
    else
        test_fail "测试项目创建失败"
    fi

    # 清理测试数据
    curl -s -X DELETE "$ADMIN_API/projects/$project_id" >/dev/null 2>&1 || true
    return 1
}

# 测试 2: 大请求体处理
test_large_payload() {
    log_test "大请求体边界测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 创建测试项目
    local project_data='{"name": "大请求体边界测试", "description": "测试大请求体处理"}'
    local project_response=$(simple_http_post "$ADMIN_API/projects" "$project_data")
    local project_id=$(simple_extract_field "$project_response" "id")

    if [ -n "$project_id" ]; then
        test_pass "测试项目创建成功"

        # 创建测试环境
        local env_data='{"name": "大请求体测试环境", "project_id": "'$project_id'", "description": "边界条件环境"}'
        local env_response=$(simple_http_post "$ADMIN_API/environments" "$env_data")
        local env_id=$(simple_extract_field "$env_response" "id")

        if [ -n "$env_id" ]; then
            test_pass "测试环境创建成功"

            # 创建大请求体规则
            local large_payload=$(head -c 10000 /dev/urandom | base64)
            local rule_data='{
                "name": "大请求体规则",
                "project_id": "'$project_id'",
                "environment_id": "'$env_id'",
                "request": {
                    "method": "POST",
                    "path": "/test/large-payload",
                    "headers": {
                        "Content-Type": "application/json"
                    }
                },
                "response": {
                    "status": 200,
                    "body": "大请求体测试成功"
                }
            }'

            local rule_response=$(simple_http_post "$ADMIN_API/rules" "$rule_data")
            local rule_id=$(simple_extract_field "$rule_response" "id")

            if [ -n "$rule_id" ]; then
                test_pass "大请求体规则创建成功 (载荷大小: ${#large_payload} 字节)"

                # 测试大请求体请求
                local payload_response=$(simple_http_request "POST" \
                    "$MOCK_API/test/large-payload" \
                    "-H \"X-Project-ID: $project_id\" -H \"X-Environment-ID: $env_id\" -H \"Content-Type: application/json\" -d '$large_payload'")

                if [ -n "$payload_response" ]; then
                    test_pass "大请求体处理成功"

                    # 清理测试数据
                    curl -s -X DELETE "$ADMIN_API/projects/$project_id" >/dev/null 2>&1 || true
                    return 0
                else
                    test_fail "大请求体处理失败"
                fi
            else
                test_fail "大请求体规则创建失败"
            fi
        else
            test_fail "测试环境创建失败"
        fi
    else
        test_fail "测试项目创建失败"
    fi

    # 清理测试数据
    curl -s -X DELETE "$ADMIN_API/projects/$project_id" >/dev/null 2>&1 || true
    return 1
}

# 测试 3: 特殊字符处理
test_special_characters() {
    log_test "特殊字符处理测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 创建测试项目
    local project_data='{"name": "特殊字符测试", "description": "测试特殊字符处理"}'
    local project_response=$(simple_http_post "$ADMIN_API/projects" "$project_data")
    local project_id=$(simple_extract_field "$project_response" "id")

    if [ -n "$project_id" ]; then
        test_pass "测试项目创建成功"

        # 创建测试环境
        local env_data='{"name": "特殊字符测试环境", "project_id": "'$project_id'", "description": "边界条件环境"}'
        local env_response=$(simple_http_post "$ADMIN_API/environments" "$env_data")
        local env_id=$(simple_extract_field "$env_response" "id")

        if [ -n "$env_id" ]; then
            test_pass "测试环境创建成功"

            # 创建特殊字符规则
            local special_chars='!@#$%^&*()_+-=[]{}|;:,.<>?'
            local rule_data='{
                "name": "特殊字符规则",
                "project_id": "'$project_id'",
                "environment_id": "'$env_id'",
                "request": {
                    "method": "GET",
                    "path": "/test/special-chars",
                    "headers": {
                        "X-Special": "'$special_chars'"
                    }
                },
                "response": {
                    "status": 200,
                    "body": "特殊字符处理成功",
                    "headers": {
                        "X-Special-Response": "'$special_chars'"
                    }
                }
            }'

            local rule_response=$(simple_http_post "$ADMIN_API/rules" "$rule_data")
            local rule_id=$(simple_extract_field "$rule_response" "id")

            if [ -n "$rule_id" ]; then
                test_pass "特殊字符规则创建成功"

                # 测试特殊字符请求
                local chars_response=$(simple_http_request "GET" \
                    "$MOCK_API/test/special-chars" \
                    "-H \"X-Project-ID: $project_id\" -H \"X-Environment-ID: $env_id\" -H \"X-Special: $special_chars\"")

                if [ -n "$chars_response" ]; then
                    test_pass "特殊字符处理成功"

                    # 清理测试数据
                    curl -s -X DELETE "$ADMIN_API/projects/$project_id" >/dev/null 2>&1 || true
                    return 0
                else
                    test_fail "特殊字符处理失败"
                fi
            else
                test_fail "特殊字符规则创建失败"
            fi
        else
            test_fail "测试环境创建失败"
        fi
    else
        test_fail "测试项目创建失败"
    fi

    # 清理测试数据
    curl -s -X DELETE "$ADMIN_API/projects/$project_id" >/dev/null 2>&1 || true
    return 1
}

# 测试 4: 错误场景处理
test_error_scenarios() {
    log_test "错误场景处理测试"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # 测试不存在的端点
    local error_response=$(simple_http_request "GET" "$MOCK_API/nonexistent-endpoint" "")

    if [ -n "$error_response" ]; then
        test_pass "不存在的端点正确返回响应"
    else
        test_fail "不存在的端点处理异常"
    fi

    # 测试无效的JSON格式（通过直接curl验证服务器健壮性）
    local invalid_response=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST \
        -H "Content-Type: application/json" \
        -d '{"invalid": json}' \
        "$ADMIN_API/projects" 2>/dev/null)

    local http_code="${invalid_response: -3}"

    if [ "$http_code" = "400" ] || [ "$http_code" = "422" ]; then
        test_pass "无效JSON格式正确返回错误码: $http_code"
        return 0
    else
        test_fail "无效JSON格式处理异常: HTTP $http_code"
        return 1
    fi
}

# 生成测试报告
generate_report() {
    print_test_summary
    local exit_code=$?

    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   边界条件测试结果${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}🎉 所有边界条件测试通过！${NC}"
        echo -e "${GREEN}✅ 系统边界处理能力正常${NC}"
        echo -e "${GREEN}✅ 错误场景处理健壮${NC}"
    else
        echo -e "${RED}❌ 部分边界条件测试失败${NC}"
        echo -e "${YELLOW}💡 请检查系统边界处理能力${NC}"
    fi

    return $exit_code
}

# 主测试流程
main() {
    show_banner

    # 执行测试
    test_long_path || true
    test_large_payload || true
    test_special_characters || true
    test_error_scenarios || true

    # 生成报告
    generate_report
}

# 信号处理
trap 'echo -e "\n${YELLOW}测试被中断${NC}"; exit 1' INT TERM

# 执行主流程
main
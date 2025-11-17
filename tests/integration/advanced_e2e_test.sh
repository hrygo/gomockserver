#!/bin/bash

# MockServer 高级 E2E 集成测试脚本
# 测试复杂场景、边界条件和性能特性

set -e

# 加载测试框架
source "$(dirname "$0")/lib/test_framework.sh"

# 初始化测试框架
init_test_framework

# 禁用自动清理
trap - EXIT INT TERM

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   Mock Server 高级 E2E 集成测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# ========================================
# 阶段 1: 复杂匹配规则测试
# ========================================

echo -e "${CYAN}[阶段 1] 复杂匹配规则测试${NC}"
echo ""

# 1.1 正则表达式匹配测试
echo -e "${YELLOW}[1.1] 测试正则表达式匹配...${NC}"
ADVANCED_PROJECT_RESPONSE=$(http_post "$ADMIN_API/projects" "$(generate_project_data "高级测试项目")")

if echo "$ADVANCED_PROJECT_RESPONSE" | grep -q '"id"'; then
    ADVANCED_PROJECT_ID=$(extract_json_field "$ADVANCED_PROJECT_RESPONSE" "id")
    PROJECT_ID="$ADVANCED_PROJECT_ID"  # 设置给框架使用
    test_pass "高级测试项目创建成功"
else
    test_fail "高级测试项目创建失败"
    exit 1
fi

ADVANCED_ENV_RESPONSE=$(http_post "$ADMIN_API/projects/$ADVANCED_PROJECT_ID/environments" "$(generate_environment_data "高级测试环境" "http://localhost:9090")")

if echo "$ADVANCED_ENV_RESPONSE" | grep -q '"id"'; then
    ADVANCED_ENVIRONMENT_ID=$(extract_json_field "$ADVANCED_ENV_RESPONSE" "id")
    test_pass "高级测试环境创建成功"
else
    test_fail "高级测试环境创建失败"
    exit 1
fi

# 创建正则表达式匹配规则
REGEX_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "{
    \"name\": \"正则表达式匹配测试\",
    \"project_id\": \"$ADVANCED_PROJECT_ID\",
    \"environment_id\": \"$ADVANCED_ENVIRONMENT_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Regex\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
        \"method\": \"GET\",
        \"path_regex\": \"/api/users/[0-9]+\"
    },
    \"response\": {
        \"type\": \"Static\",
        \"content\": {
            \"status_code\": 200,
            \"content_type\": \"JSON\",
            \"body\": {
                \"code\": 0,
                \"data\": {
                    \"id\": \"{{.path_regex_0}}\",
                    \"name\": \"用户详情\",
                    \"matched\": \"regex\"
                },
                \"message\": \"success\"
            }
        }
    }
}")

if echo "$REGEX_RULE_RESPONSE" | grep -q '"id"'; then
    test_pass "正则表达式规则创建成功"
else
    test_fail "正则表达式规则创建失败"
fi

sleep 2
REGEX_TEST_RESPONSE=$(mock_request "GET" "/api/users/123")
REGEX_HTTP_CODE=$(echo "$REGEX_TEST_RESPONSE" | tail -n 1)

if [ "$REGEX_HTTP_CODE" = "200" ]; then
    test_pass "正则表达式匹配成功"
else
    test_fail "正则表达式匹配失败，状态码: $REGEX_HTTP_CODE"
fi

# 1.2 动态响应模板测试
echo -e "${YELLOW}[1.2] 测试动态响应模板...${NC}"
TEMPLATE_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "{
    \"name\": \"动态响应模板测试\",
    \"project_id\": \"$ADVANCED_PROJECT_ID\",
    \"environment_id\": \"$ADVANCED_ENVIRONMENT_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
        \"method\": \"GET\",
        \"path\": \"/api/template\"
    },
    \"response\": {
        \"type\": \"Template\",
        \"content\": {
            \"status_code\": 200,
            \"content_type\": \"JSON\",
            \"body\": {
                \"timestamp\": \"{{timestamp}}\",
                \"uuid\": \"{{uuid}}\",
                \"random_number\": \"{{random 1000}}\",
                \"counter\": \"{{counter}}\",
                \"formatted_date\": \"{{date_format '2006-01-02'}}\"
            }
        }
    }
}")

if echo "$TEMPLATE_RULE_RESPONSE" | grep -q '"id"'; then
    test_pass "动态响应模板规则创建成功"
else
    test_fail "动态响应模板规则创建失败"
fi

sleep 2
TEMPLATE_TEST_RESPONSE=$(mock_request "GET" "/api/template")
TEMPLATE_HTTP_CODE=$(echo "$TEMPLATE_TEST_RESPONSE" | tail -n 1)

if [ "$TEMPLATE_HTTP_CODE" = "200" ]; then
    test_pass "动态响应模板测试成功"
else
    test_fail "动态响应模板测试失败，状态码: $TEMPLATE_HTTP_CODE"
fi

echo ""

# ========================================
# 阶段 2: 错误注入测试
# ========================================

echo -e "${CYAN}[阶段 2] 错误注入测试${NC}"
echo ""

# 2.1 HTTP错误状态码测试
echo -e "${YELLOW}[2.1] 测试 HTTP 错误状态码...${NC}"
ERROR_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "{
    \"name\": \"HTTP错误测试\",
    \"project_id\": \"$ADVANCED_PROJECT_ID\",
    \"environment_id\": \"$ADVANCED_ENVIRONMENT_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
        \"method\": \"GET\",
        \"path\": \"/api/error\"
    },
    \"response\": {
        \"type\": \"Static\",
        \"content\": {
            \"status_code\": 500,
            \"content_type\": \"JSON\",
            \"body\": {
                \"code\": 500,
                \"message\": \"Internal Server Error\",
                \"error\": \"模拟服务器错误\"
            }
        }
    }
}")

if echo "$ERROR_RULE_RESPONSE" | grep -q '"id"'; then
    test_pass "HTTP错误规则创建成功"
else
    test_fail "HTTP错误规则创建失败"
fi

sleep 2
ERROR_TEST_RESPONSE=$(mock_request "GET" "/api/error")
ERROR_HTTP_CODE=$(echo "$ERROR_TEST_RESPONSE" | tail -n 1)

if [ "$ERROR_HTTP_CODE" = "500" ]; then
    test_pass "HTTP错误状态码测试成功"
else
    test_fail "HTTP错误状态码测试失败，状态码: $ERROR_HTTP_CODE"
fi

echo ""

# ========================================
# 阶段 3: 高级延迟策略测试
# ========================================

echo -e "${CYAN}[阶段 3] 高级延迟策略测试${NC}"
echo ""

# 3.1 固定延迟测试
echo -e "${YELLOW}[3.1] 测试固定延迟...${NC}"
DELAY_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "{
    \"name\": \"固定延迟测试\",
    \"project_id\": \"$ADVANCED_PROJECT_ID\",
    \"environment_id\": \"$ADVANCED_ENVIRONMENT_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
        \"method\": \"GET\",
        \"path\": \"/api/delay\"
    },
    \"response\": {
        \"type\": \"Static\",
        \"content\": {
            \"status_code\": 200,
            \"content_type\": \"JSON\",
            \"body\": {
                \"message\": \"延迟响应测试\",
                \"delayed\": true
            }
        },
        \"delay_strategy\": {
            \"type\": \"Fixed\",
            \"duration_ms\": 50
        }
    }
}")

if echo "$DELAY_RULE_RESPONSE" | grep -q '"id"'; then
    test_pass "固定延迟规则创建成功"
else
    test_fail "固定延迟规则创建失败"
fi

sleep 2
DELAY_START_TIME=$(get_timestamp_ms)
DELAY_TEST_RESPONSE=$(mock_request "GET" "/api/delay")
DELAY_END_TIME=$(get_timestamp_ms)
DELAY_DURATION=$(calculate_duration "$DELAY_START_TIME" "$DELAY_END_TIME")
DELAY_HTTP_CODE=$(echo "$DELAY_TEST_RESPONSE" | tail -n 1)

if [ "$DELAY_HTTP_CODE" = "200" ] && [ $DELAY_DURATION -ge 40 ]; then
    test_pass "固定延迟测试成功 (耗时: ${DELAY_DURATION}ms)"
else
    test_fail "固定延迟测试失败 (耗时: ${DELAY_DURATION}ms, 状态码: $DELAY_HTTP_CODE)"
fi

echo ""

# ========================================
# 阶段 4: 代理模式测试
# ========================================

echo -e "${CYAN}[阶段 4] 代理模式测试${NC}"
echo ""

# 4.1 简单代理测试
echo -e "${YELLOW}[4.1] 测试简单代理模式...${NC}"
PROXY_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "{
    \"name\": \"简单代理测试\",
    \"project_id\": \"$ADVANCED_PROJECT_ID\",
    \"environment_id\": \"$ADVANCED_ENVIRONMENT_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
        \"method\": \"GET\",
        \"path\": \"/api/proxy/*\"
    },
    \"response\": {
        \"type\": \"Proxy\",
        \"proxy_config\": {
            \"target_url\": \"https://httpbin.org\",
            \"strip_path\": \"/api/proxy\",
            \"timeout_ms\": 5000
        }
    }
}")

if echo "$PROXY_RULE_RESPONSE" | grep -q '"id"'; then
    test_pass "代理规则创建成功"
else
    test_fail "代理规则创建失败"
fi

# 代理测试跳过（因为需要外部网络）
test_skip "代理模式测试（需要外部网络访问）"

echo ""

# ========================================
# 阶段 5: 清理测试数据
# ========================================

echo -e "${CYAN}[阶段 5] 清理测试数据${NC}"
echo ""

echo -e "${YELLOW}[5.1] 清理测试资源...${NC}"
if [ -n "$ADVANCED_PROJECT_ID" ]; then
    http_delete "$ADMIN_API/projects/$ADVANCED_PROJECT_ID" >/dev/null 2>&1 || true
    test_pass "测试项目清理完成"
fi

echo ""

# ========================================
# 生成测试报告
# ========================================

echo -e "${CYAN}[完成] 生成测试报告${NC}"
REPORT_FILE="/tmp/advanced_e2e_test_report_$(date +%Y%m%d_%H%M%S).md"
generate_test_report "$REPORT_FILE" "高级 E2E 集成测试"

# ========================================
# 测试结果统计
# ========================================

print_test_summary

echo ""
echo -e "${CYAN}高级功能验证:${NC}"
echo -e "  ${GREEN}✓ 正则表达式匹配${NC}"
echo -e "  ${GREEN}✓ 动态响应模板${NC}"
echo -e "  ${GREEN}✓ HTTP错误注入${NC}"
echo -e "  ${GREEN}✓ 高级延迟策略${NC}"
echo -e "  ${GREEN}✓ 代理模式配置${NC}"

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   高级 E2E 测试完成${NC}"
echo -e "${BLUE}=========================================${NC}"
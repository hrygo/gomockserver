#!/bin/bash

# 规则管理功能测试用例（精简版）

run_rule_tests() {
    title "规则管理功能测试"
    
    if [ -z "$PROJECT_ID" ] || [ -z "$ENVIRONMENT_ID" ]; then
        error "缺少项目ID或环境ID，请先执行相关测试"
        return
    fi
    
    if ! ask_confirmation "是否执行规则管理测试" "Y"; then
        warning "已跳过规则管理测试套件"
        return
    fi
    
    # 创建基本规则
    subtitle "RULE-001 & RULE-002: 创建HTTP规则测试"
    info "正在创建测试规则..."
    
    local response_body='{"code":0,"message":"success","data":{"id":1,"name":"张三"}}'
    local create_response=$(api_create_rule "获取用户信息" "$PROJECT_ID" "$ENVIRONMENT_ID" "GET" "/api/users/1" 200 "$response_body" 100)
    
    if [ $? -eq 0 ]; then
        RULE_ID=$(extract_json_field "$create_response" "id")
        echo -e "${CYAN}响应内容:${NC}"
        echo "$create_response" | jq '.' 2>/dev/null || echo "$create_response"
        echo ""
        success "规则创建成功，ID: $RULE_ID"
        test_pass "RULE-001 & RULE-002: 创建HTTP规则测试"
    else
        error "规则创建失败"
        test_fail "RULE-001 & RULE-002: 创建HTTP规则测试"
        return
    fi
    
    # 创建带延迟的规则
    subtitle "RULE-009: 创建带延迟的规则测试"
    info "正在创建带延迟的规则..."
    
    local delay_response=$(api_create_rule_with_delay "慢速接口" "$PROJECT_ID" "$ENVIRONMENT_ID" "GET" "/api/slow" 1000 200 '{"message":"slow response"}')
    
    if [ $? -eq 0 ]; then
        DELAY_RULE_ID=$(extract_json_field "$delay_response" "id")
        echo -e "${CYAN}响应内容:${NC}"
        echo "$delay_response" | jq '.' 2>/dev/null || echo "$delay_response"
        echo ""
        success "延迟规则创建成功，ID: $DELAY_RULE_ID"
        test_pass "RULE-009: 创建带延迟的规则测试"
    else
        error "延迟规则创建失败"
        test_fail "RULE-009: 创建带延迟的规则测试"
    fi
    
    # 禁用规则
    subtitle "RULE-008: 禁用规则测试"
    info "正在禁用规则..."
    
    local disable_response=$(api_disable_rule "$RULE_ID")
    if [ $? -eq 0 ]; then
        echo -e "${CYAN}响应内容:${NC}"
        echo "$disable_response" | jq '.' 2>/dev/null || echo "$disable_response"
        echo ""
        test_pass "RULE-008: 禁用规则测试"
    else
        test_fail "RULE-008: 禁用规则测试"
    fi
    
    # 启用规则
    subtitle "RULE-007: 启用规则测试"
    info "正在启用规则..."
    
    local enable_response=$(api_enable_rule "$RULE_ID")
    if [ $? -eq 0 ]; then
        echo -e "${CYAN}响应内容:${NC}"
        echo "$enable_response" | jq '.' 2>/dev/null || echo "$enable_response"
        echo ""
        test_pass "RULE-007: 启用规则测试"
    else
        test_fail "RULE-007: 启用规则测试"
    fi
    
    # 导出规则ID供其他测试使用
    export RULE_ID
    export DELAY_RULE_ID
}

export -f run_rule_tests

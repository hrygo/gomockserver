#!/bin/bash

# Mock服务功能测试用例（精简版）

run_mock_tests() {
    title "Mock服务功能测试"
    
    if [ -z "$PROJECT_ID" ] || [ -z "$ENVIRONMENT_ID" ] || [ -z "$RULE_ID" ]; then
        error "缺少必要的ID，请先执行相关测试"
        return
    fi
    
    if ! ask_confirmation "是否执行Mock服务测试" "Y"; then
        warning "已跳过Mock服务测试套件"
        return
    fi
    
    # 测试基本Mock请求
    subtitle "MOCK-001: GET请求Mock响应测试"
    info "正在发送Mock请求..."
    
    local mock_response=$(mock_request "$PROJECT_ID" "$ENVIRONMENT_ID" "/api/users/1" "GET" "" 200)
    if [ $? -eq 0 ]; then
        echo -e "${CYAN}Mock响应内容:${NC}"
        echo "$mock_response" | jq '.' 2>/dev/null || echo "$mock_response"
        echo ""
        
        if echo "$mock_response" | grep -q "张三"; then
            success "Mock响应内容正确"
            test_pass "MOCK-001: GET请求Mock响应测试"
        else
            warning "Mock响应内容可能不正确"
            ask_test_result "MOCK-001: GET请求Mock响应测试"
        fi
    else
        error "Mock请求失败"
        test_fail "MOCK-001: GET请求Mock响应测试"
    fi
    
    # 测试延迟响应
    if [ -n "$DELAY_RULE_ID" ]; then
        subtitle "MOCK-006: 响应延迟功能测试"
        info "正在测试延迟响应（预期1秒延迟）..."
        
        local start_ms=$(date +%s%3N)
        local delay_response=$(mock_request "$PROJECT_ID" "$ENVIRONMENT_ID" "/api/slow" "GET" "" 200)
        local end_ms=$(date +%s%3N)
        local actual_delay=$((end_ms - start_ms))
        
        echo -e "${CYAN}响应时间: ${actual_delay}ms${NC}"
        echo -e "${CYAN}响应内容:${NC}"
        echo "$delay_response" | jq '.' 2>/dev/null || echo "$delay_response"
        echo ""
        
        if [ $actual_delay -ge 900 ] && [ $actual_delay -le 1500 ]; then
            success "延迟时间符合预期（1000ms ± 500ms）"
            test_pass "MOCK-006: 响应延迟功能测试"
        else
            warning "延迟时间可能不准确，实际: ${actual_delay}ms，预期: 1000ms"
            ask_test_result "MOCK-006: 响应延迟功能测试"
        fi
    fi
    
    # 测试不匹配的请求
    subtitle "MOCK-012: 未匹配规则返回404测试"
    info "正在测试不存在的路径..."
    
    local notfound_response=$(mock_request "$PROJECT_ID" "$ENVIRONMENT_ID" "/api/nonexistent" "GET" "" 404)
    if [ $? -eq 0 ]; then
        echo -e "${CYAN}响应内容:${NC}"
        echo "$notfound_response"
        echo ""
        success "未匹配规则正确返回404"
        test_pass "MOCK-012: 未匹配规则返回404测试"
    else
        warning "404响应可能不正确"
        ask_test_result "MOCK-012: 未匹配规则返回404测试"
    fi
}

export -f run_mock_tests

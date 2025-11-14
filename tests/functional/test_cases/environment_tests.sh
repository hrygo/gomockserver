#!/bin/bash

# 环境管理功能测试用例（精简版）

run_environment_tests() {
    title "环境管理功能测试"
    
    if [ -z "$PROJECT_ID" ]; then
        error "缺少项目ID，请先执行项目管理测试"
        return
    fi
    
    if ! ask_confirmation "是否执行环境管理测试" "Y"; then
        warning "已跳过环境管理测试套件"
        return
    fi
    
    # 创建环境
    subtitle "ENV-001 & ENV-002: 创建环境测试"
    info "正在创建测试环境..."
    
    local create_response=$(api_create_environment "功能测试环境" "$PROJECT_ID" "http://localhost:9090")
    if [ $? -eq 0 ]; then
        ENVIRONMENT_ID=$(extract_json_field "$create_response" "id")
        echo -e "${CYAN}响应内容:${NC}"
        echo "$create_response" | jq '.' 2>/dev/null || echo "$create_response"
        echo ""
        success "环境创建成功，ID: $ENVIRONMENT_ID"
        test_pass "ENV-001 & ENV-002: 创建环境测试"
    else
        error "环境创建失败"
        test_fail "ENV-001 & ENV-002: 创建环境测试"
        return
    fi
    
    # 查询环境
    subtitle "ENV-003: 查询环境详情测试"
    info "正在查询环境详情..."
    
    local get_response=$(api_get_environment "$ENVIRONMENT_ID")
    if [ $? -eq 0 ]; then
        echo -e "${CYAN}响应内容:${NC}"
        echo "$get_response" | jq '.' 2>/dev/null || echo "$get_response"
        echo ""
        test_pass "ENV-003: 查询环境详情测试"
    else
        test_fail "ENV-003: 查询环境详情测试"
    fi
    
    # 导出环境ID供其他测试使用
    export ENVIRONMENT_ID
}

export -f run_environment_tests

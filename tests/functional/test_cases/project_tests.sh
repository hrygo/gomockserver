#!/bin/bash

# 项目管理功能测试用例（精简版）

run_project_tests() {
    title "项目管理功能测试"
    
    info "本测试套件将测试项目的CRUD操作"
    echo ""
    
    if ! ask_confirmation "是否执行项目管理测试" "Y"; then
        warning "已跳过项目管理测试套件"
        return
    fi
    
    # 创建项目
    subtitle "PRJ-001 & PRJ-002: 创建项目测试"
    info "正在创建测试项目..."
    
    local create_response=$(api_create_project "功能测试项目" "functional-test" "自动化功能测试项目")
    if [ $? -eq 0 ]; then
        PROJECT_ID=$(extract_json_field "$create_response" "id")
        echo -e "${CYAN}响应内容:${NC}"
        echo "$create_response" | jq '.' 2>/dev/null || echo "$create_response"
        echo ""
        success "项目创建成功，ID: $PROJECT_ID"
        test_pass "PRJ-001 & PRJ-002: 创建项目测试"
    else
        error "项目创建失败"
        test_fail "PRJ-001 & PRJ-002: 创建项目测试"
        return
    fi
    
    # 查询项目
    subtitle "PRJ-003: 查询项目详情测试"
    info "正在查询项目详情..."
    
    local get_response=$(api_get_project "$PROJECT_ID")
    if [ $? -eq 0 ]; then
        echo -e "${CYAN}响应内容:${NC}"
        echo "$get_response" | jq '.' 2>/dev/null || echo "$get_response"
        echo ""
        test_pass "PRJ-003: 查询项目详情测试"
    else
        test_fail "PRJ-003: 查询项目详情测试"
    fi
    
    # 更新项目
    subtitle "PRJ-004: 更新项目信息测试"
    info "正在更新项目信息..."
    
    local update_response=$(api_update_project "$PROJECT_ID" "功能测试项目(已更新)" "更新后的描述")
    if [ $? -eq 0 ]; then
        echo -e "${CYAN}响应内容:${NC}"
        echo "$update_response" | jq '.' 2>/dev/null || echo "$update_response"
        echo ""
        test_pass "PRJ-004: 更新项目信息测试"
    else
        test_fail "PRJ-004: 更新项目信息测试"
    fi
    
    # 导出项目ID供其他测试使用
    export PROJECT_ID
}

export -f run_project_tests

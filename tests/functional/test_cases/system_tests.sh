#!/bin/bash

# 系统管理功能测试用例

# ==================== 测试用例 ====================

# SYS-001 & SYS-002: 健康检查
test_sys_health_check() {
    subtitle "SYS-001 & SYS-002: 健康检查测试"
    
    info "测试目的: 验证健康检查接口正常响应且返回正确状态"
    info "测试步骤:"
    echo "  1. 调用健康检查API"
    echo "  2. 验证HTTP状态码为200"
    echo "  3. 验证响应包含'ok'或'healthy'状态"
    info "预期结果: 返回200状态码，响应体包含健康状态信息"
    echo ""
    
    if ask_confirmation "是否执行此测试"; then
        echo ""
        info "正在执行测试..."
        
        local response=$(api_health_check)
        local result=$?
        
        echo ""
        echo -e "${CYAN}响应内容:${NC}"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        echo ""
        
        if [ $result -eq 0 ]; then
            if echo "$response" | grep -qi "ok\|healthy\|status"; then
                info "✓ 健康检查接口返回正常"
                ask_test_result "SYS-001 & SYS-002: 健康检查测试"
            else
                error "响应内容不包含健康状态信息"
                test_fail "SYS-001 & SYS-002: 健康检查测试"
            fi
        else
            error "健康检查接口调用失败"
            test_fail "SYS-001 & SYS-002: 健康检查测试"
        fi
    else
        test_skip "SYS-001 & SYS-002: 健康检查测试"
    fi
}

# SYS-003 & SYS-004: 版本信息
test_sys_version_info() {
    subtitle "SYS-003 & SYS-004: 版本信息测试"
    
    info "测试目的: 验证版本信息接口正常响应且包含完整信息"
    info "测试步骤:"
    echo "  1. 调用版本信息API"
    echo "  2. 验证HTTP状态码为200"
    echo "  3. 验证响应包含version字段"
    echo "  4. 检查是否包含构建时间、Git提交号等信息"
    info "预期结果: 返回版本号、构建时间等完整信息"
    echo ""
    
    if ask_confirmation "是否执行此测试"; then
        echo ""
        info "正在执行测试..."
        
        local response=$(api_get_version)
        local result=$?
        
        echo ""
        echo -e "${CYAN}响应内容:${NC}"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        echo ""
        
        if [ $result -eq 0 ]; then
            if echo "$response" | grep -q "version"; then
                info "✓ 版本信息接口返回正常"
                
                # 检查详细字段
                local has_version=$(echo "$response" | grep -c "version")
                local has_buildtime=$(echo "$response" | grep -c "build_time\|buildTime")
                local has_gitcommit=$(echo "$response" | grep -c "git_commit\|gitCommit\|commit")
                
                info "检测到的字段："
                [ $has_version -gt 0 ] && echo "  ✓ version" || echo "  ✗ version"
                [ $has_buildtime -gt 0 ] && echo "  ✓ build_time" || echo "  ⊙ build_time (可选)"
                [ $has_gitcommit -gt 0 ] && echo "  ✓ git_commit" || echo "  ⊙ git_commit (可选)"
                
                ask_test_result "SYS-003 & SYS-004: 版本信息测试"
            else
                error "响应内容不包含version字段"
                test_fail "SYS-003 & SYS-004: 版本信息测试"
            fi
        else
            error "版本信息接口调用失败"
            test_fail "SYS-003 & SYS-004: 版本信息测试"
        fi
    else
        test_skip "SYS-003 & SYS-004: 版本信息测试"
    fi
}

# SYS-005: 服务启动时间
test_sys_startup_time() {
    subtitle "SYS-005: 服务启动时间测试"
    
    info "测试目的: 验证服务启动时间在可接受范围内（< 10秒）"
    info "测试说明: 此项需要手工测试，记录服务启动时间"
    info "测试步骤:"
    echo "  1. 停止Mock Server服务"
    echo "  2. 记录当前时间"
    echo "  3. 启动Mock Server服务"
    echo "  4. 等待健康检查接口返回200"
    echo "  5. 计算启动耗时"
    info "预期结果: 启动时间 < 10秒"
    echo ""
    
    warning "这是一个手工测试项，需要手动执行"
    echo ""
    
    if ask_confirmation "您是否已完成此测试"; then
        echo ""
        read -p "请输入实际启动时间（秒）: " startup_time
        
        if [ -n "$startup_time" ] && [ "$startup_time" -lt 10 ]; then
            success "启动时间 ${startup_time}秒，符合要求（< 10秒）"
            test_pass "SYS-005: 服务启动时间测试"
        elif [ -n "$startup_time" ]; then
            warning "启动时间 ${startup_time}秒，超过预期（< 10秒）"
            ask_test_result "SYS-005: 服务启动时间测试"
        else
            test_skip "SYS-005: 服务启动时间测试"
        fi
    else
        test_skip "SYS-005: 服务启动时间测试"
    fi
}

# SYS-006: 服务异常重启后数据完整性
test_sys_data_integrity_after_restart() {
    subtitle "SYS-006: 服务异常重启后数据完整性测试"
    
    info "测试目的: 验证服务重启后数据不丢失"
    info "测试说明: 此项需要手工测试，验证数据持久化"
    info "测试步骤:"
    echo "  1. 创建测试数据（项目、环境、规则）"
    echo "  2. 记录创建的数据ID"
    echo "  3. 重启Mock Server服务"
    echo "  4. 查询之前创建的数据"
    echo "  5. 验证数据完整性"
    info "预期结果: 所有数据保持完整，无丢失"
    echo ""
    
    warning "这是一个手工测试项，需要手动执行"
    echo ""
    
    if ask_confirmation "您是否已完成此测试"; then
        ask_test_result "SYS-006: 服务异常重启后数据完整性测试"
    else
        test_skip "SYS-006: 服务异常重启后数据完整性测试"
    fi
}

# ==================== 测试套件执行 ====================

run_system_tests() {
    title "系统管理功能测试"
    
    info "本测试套件包含以下测试用例："
    echo "  SYS-001 & SYS-002: 健康检查测试"
    echo "  SYS-003 & SYS-004: 版本信息测试"
    echo "  SYS-005: 服务启动时间测试（手工）"
    echo "  SYS-006: 服务异常重启后数据完整性测试（手工）"
    echo ""
    
    if ask_confirmation "是否执行所有系统管理测试" "Y"; then
        test_sys_health_check
        test_sys_version_info
        test_sys_startup_time
        test_sys_data_integrity_after_restart
    else
        warning "已跳过系统管理测试套件"
    fi
}

# ==================== 导出函数 ====================

export -f test_sys_health_check test_sys_version_info test_sys_startup_time
export -f test_sys_data_integrity_after_restart run_system_tests

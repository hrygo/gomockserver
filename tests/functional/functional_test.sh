#!/bin/bash

# Mock Server 交互式功能测试主脚本
# 版本: 1.0
# 说明: 提供人机交互的功能测试执行环境

set -e

# ==================== 脚本目录 ====================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# ==================== 加载公共库 ====================
source "$SCRIPT_DIR/lib/common.sh"
source "$SCRIPT_DIR/lib/api_client.sh"
source "$SCRIPT_DIR/lib/report_generator.sh"

# ==================== 加载测试用例 ====================
source "$SCRIPT_DIR/test_cases/system_tests.sh"
source "$SCRIPT_DIR/test_cases/project_tests.sh"
source "$SCRIPT_DIR/test_cases/environment_tests.sh"
source "$SCRIPT_DIR/test_cases/rule_tests.sh"
source "$SCRIPT_DIR/test_cases/mock_tests.sh"

# ==================== 全局变量 ====================
PROJECT_ID=""
ENVIRONMENT_ID=""
RULE_ID=""
DELAY_RULE_ID=""

# ==================== 清理函数 ====================
cleanup() {
    echo ""
    title "测试清理"
    
    if ask_confirmation "是否清理测试数据"; then
        info "正在清理测试数据..."
        
        # 删除创建的规则
        if [ -n "$RULE_ID" ]; then
            api_delete_rule "$RULE_ID" 2>/dev/null && success "规则 $RULE_ID 已删除" || warning "规则删除失败"
        fi
        if [ -n "$DELAY_RULE_ID" ]; then
            api_delete_rule "$DELAY_RULE_ID" 2>/dev/null && success "规则 $DELAY_RULE_ID 已删除" || warning "规则删除失败"
        fi
        
        # 删除创建的环境
        if [ -n "$ENVIRONMENT_ID" ]; then
            api_delete_environment "$ENVIRONMENT_ID" 2>/dev/null && success "环境 $ENVIRONMENT_ID 已删除" || warning "环境删除失败"
        fi
        
        # 删除创建的项目
        if [ -n "$PROJECT_ID" ]; then
            api_delete_project "$PROJECT_ID" 2>/dev/null && success "项目 $PROJECT_ID 已删除" || warning "项目删除失败"
        fi
    else
        info "保留测试数据，可用于问题排查"
        echo "  项目ID: $PROJECT_ID"
        echo "  环境ID: $ENVIRONMENT_ID"
        echo "  规则ID: $RULE_ID"
    fi
    
    echo ""
    show_test_summary
    
    # 生成测试报告
    if ask_confirmation "是否生成测试报告"; then
        local report_path=$(generate_test_report "$SCRIPT_DIR/reports/functional_test_report_$(date +%Y%m%d_%H%M%S).md" "功能测试")
        info "测试报告已生成: $report_path"
        
        # 同时生成HTML报告
        if ask_confirmation "是否生成HTML报告"; then
            local html_report=$(generate_html_report "$SCRIPT_DIR/reports/functional_test_report_$(date +%Y%m%d_%H%M%S).html")
            info "HTML报告已生成: $html_report"
        fi
    fi
    
    echo ""
    success "测试完成！"
}

trap cleanup EXIT

# ==================== 环境检查 ====================
check_environment() {
    title "环境检查"
    
    # 检查必要命令
    info "检查必要命令..."
    local missing_cmds=()
    
    for cmd in curl jq; do
        if check_command "$cmd"; then
            success "$cmd 已安装"
        else
            error "$cmd 未安装"
            missing_cmds+=("$cmd")
        fi
    done
    
    if [ ${#missing_cmds[@]} -gt 0 ]; then
        error "缺少必要命令: ${missing_cmds[*]}"
        error "请安装后重试"
        exit 1
    fi
    
    echo ""
    
    # 检查服务可用性
    info "检查Mock Server服务..."
    if check_url "$ADMIN_API/system/health" 5; then
        success "管理API服务可访问: $ADMIN_API"
    else
        error "管理API服务不可访问: $ADMIN_API"
        error "请确保Mock Server正在运行"
        exit 1
    fi
    
    echo ""
    info "环境检查通过！"
    echo ""
}

# ==================== 测试菜单 ====================
show_menu() {
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}    Mock Server 功能测试菜单${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    echo "  1. 系统管理功能测试"
    echo "  2. 项目管理功能测试"
    echo "  3. 环境管理功能测试"
    echo "  4. 规则管理功能测试"
    echo "  5. Mock服务功能测试"
    echo ""
    echo "  0. 执行完整测试流程"
    echo ""
    echo "  q. 退出"
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo ""
}

# ==================== 主函数 ====================
main() {
    # 初始化
    init_test_env
    
    # 欢迎信息
    title "Mock Server 交互式功能测试"
    info "版本: 1.0"
    info "日志文件: $TEST_LOG_FILE"
    echo ""
    
    # 环境检查
    check_environment
    
    # 主循环
    while true; do
        show_menu
        read -p "请选择测试项 [0-5, q]: " choice
        
        case "$choice" in
            1)
                run_system_tests
                ;;
            2)
                run_project_tests
                ;;
            3)
                run_environment_tests
                ;;
            4)
                run_rule_tests
                ;;
            5)
                run_mock_tests
                ;;
            0)
                info "开始执行完整测试流程..."
                echo ""
                run_system_tests
                run_project_tests
                run_environment_tests
                run_rule_tests
                run_mock_tests
                
                info "完整测试流程执行完毕"
                break
                ;;
            [Qq]*)
                warning "用户退出测试"
                break
                ;;
            *)
                error "无效选择，请重新输入"
                ;;
        esac
    done
}

# ==================== 执行主函数 ====================
main "$@"

#!/bin/bash

# 功能测试通用函数库
# 提供颜色输出、日志记录、结果统计等通用功能

# ==================== 颜色定义 ====================
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# ==================== 全局变量 ====================
TEST_PASSED=0
TEST_FAILED=0
TEST_SKIPPED=0
TEST_START_TIME=""
TEST_LOG_FILE="/tmp/functional_test_$(date +%Y%m%d_%H%M%S).log"

# ==================== 日志函数 ====================

# 记录日志到文件
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$TEST_LOG_FILE"
}

# 信息输出
info() {
    echo -e "${BLUE}ℹ  $1${NC}"
    log "INFO: $1"
}

# 成功输出
success() {
    echo -e "${GREEN}✓ $1${NC}"
    log "SUCCESS: $1"
}

# 警告输出
warning() {
    echo -e "${YELLOW}⚠  $1${NC}"
    log "WARNING: $1"
}

# 错误输出
error() {
    echo -e "${RED}✗ $1${NC}"
    log "ERROR: $1"
}

# 标题输出
title() {
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
    log "TITLE: $1"
}

# 子标题输出
subtitle() {
    echo ""
    echo -e "${MAGENTA}[  $1  ]${NC}"
    echo ""
    log "SUBTITLE: $1"
}

# ==================== 测试结果记录 ====================

# 记录测试通过
test_pass() {
    success "$1"
    TEST_PASSED=$((TEST_PASSED + 1))
    log "TEST PASSED: $1"
}

# 记录测试失败
test_fail() {
    error "$1"
    TEST_FAILED=$((TEST_FAILED + 1))
    log "TEST FAILED: $1"
}

# 记录测试跳过
test_skip() {
    warning "$1"
    TEST_SKIPPED=$((TEST_SKIPPED + 1))
    log "TEST SKIPPED: $1"
}

# ==================== JSON 处理函数 ====================

# 从JSON中提取字段值
extract_json_field() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | cut -d'"' -f4
}

# 从JSON中提取数字字段
extract_json_number() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\":[0-9]*" | cut -d':' -f2
}

# 验证JSON字段值
verify_json_field() {
    local json="$1"
    local field="$2"
    local expected="$3"
    local actual=$(extract_json_field "$json" "$field")
    
    if [ "$actual" = "$expected" ]; then
        return 0
    else
        return 1
    fi
}

# 检查JSON是否包含字段
check_json_field_exists() {
    local json="$1"
    local field="$2"
    if echo "$json" | grep -q "\"$field\""; then
        return 0
    else
        return 1
    fi
}

# ==================== 输入验证 ====================

# 询问用户确认
ask_confirmation() {
    local prompt="$1"
    local default="${2:-N}"
    
    if [ "$default" = "Y" ]; then
        read -p "$(echo -e ${YELLOW}$prompt [Y/n]: ${NC})" response
        response=${response:-Y}
    else
        read -p "$(echo -e ${YELLOW}$prompt [y/N]: ${NC})" response
        response=${response:-N}
    fi
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        return 0
    else
        return 1
    fi
}

# 询问测试结果
ask_test_result() {
    local test_name="$1"
    echo ""
    echo -e "${YELLOW}请确认测试结果：$test_name${NC}"
    echo -e "  ${GREEN}Y${NC} - 通过"
    echo -e "  ${RED}N${NC} - 失败"
    echo -e "  ${YELLOW}S${NC} - 跳过"
    read -p "请输入 [Y/N/S]: " result
    
    case "$result" in
        [Yy]* )
            test_pass "$test_name"
            return 0
            ;;
        [Nn]* )
            test_fail "$test_name"
            return 1
            ;;
        [Ss]* )
            test_skip "$test_name"
            return 2
            ;;
        * )
            warning "无效输入，默认为跳过"
            test_skip "$test_name"
            return 2
            ;;
    esac
}

# ==================== 环境检查 ====================

# 检查命令是否存在
check_command() {
    local cmd="$1"
    if command -v "$cmd" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# 检查端口是否可用
check_port() {
    local port="$1"
    if nc -z localhost "$port" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# 检查URL是否可访问
check_url() {
    local url="$1"
    local timeout="${2:-5}"
    if curl -s --max-time "$timeout" "$url" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# ==================== 统计报告 ====================

# 显示测试统计
show_test_summary() {
    local total=$((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))
    local pass_rate=0
    
    if [ $total -gt 0 ]; then
        pass_rate=$(awk "BEGIN {printf \"%.2f\", ($TEST_PASSED/$total)*100}")
    fi
    
    echo ""
    title "测试结果统计"
    echo -e "${CYAN}总计测试数:${NC} $total"
    echo -e "${GREEN}通过测试数:${NC} $TEST_PASSED"
    echo -e "${RED}失败测试数:${NC} $TEST_FAILED"
    echo -e "${YELLOW}跳过测试数:${NC} $TEST_SKIPPED"
    echo -e "${CYAN}通过率:${NC} $pass_rate%"
    echo ""
    
    log "TEST SUMMARY: Total=$total, Passed=$TEST_PASSED, Failed=$TEST_FAILED, Skipped=$TEST_SKIPPED, Rate=$pass_rate%"
}

# 获取测试统计数据
get_test_stats() {
    local total=$((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))
    local pass_rate=0
    
    if [ $total -gt 0 ]; then
        pass_rate=$(awk "BEGIN {printf \"%.2f\", ($TEST_PASSED/$total)*100}")
    fi
    
    echo "$total $TEST_PASSED $TEST_FAILED $TEST_SKIPPED $pass_rate"
}

# ==================== 时间处理 ====================

# 记录开始时间
start_timer() {
    TEST_START_TIME=$(date +%s)
}

# 计算耗时
get_duration() {
    local end_time=$(date +%s)
    local duration=$((end_time - TEST_START_TIME))
    echo "$duration"
}

# 格式化时间
format_duration() {
    local duration="$1"
    local hours=$((duration / 3600))
    local minutes=$(((duration % 3600) / 60))
    local seconds=$((duration % 60))
    
    if [ $hours -gt 0 ]; then
        echo "${hours}h ${minutes}m ${seconds}s"
    elif [ $minutes -gt 0 ]; then
        echo "${minutes}m ${seconds}s"
    else
        echo "${seconds}s"
    fi
}

# ==================== 清理函数 ====================

# 清理临时文件
cleanup_temp_files() {
    # 预留清理临时文件的逻辑
    log "Cleanup temporary files"
}

# ==================== 初始化 ====================

# 初始化测试环境
init_test_env() {
    start_timer
    log "Test environment initialized"
    info "测试日志文件: $TEST_LOG_FILE"
}

# ==================== 导出函数 ====================

# 导出所有函数供其他脚本使用
export -f log info success warning error title subtitle
export -f test_pass test_fail test_skip
export -f extract_json_field extract_json_number verify_json_field check_json_field_exists
export -f ask_confirmation ask_test_result
export -f check_command check_port check_url
export -f show_test_summary get_test_stats
export -f start_timer get_duration format_duration
export -f cleanup_temp_files init_test_env

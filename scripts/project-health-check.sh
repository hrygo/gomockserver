#!/bin/bash

# MockServer 项目质量检查脚本
# Author: MockServer Team
# Created: 2025-11-19
# Description: 检查项目文件管理规范，防止项目腐化

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检查结果统计
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# 检查函数
check_item() {
    local description="$1"
    local check_command="$2"
    local expected_result="$3"
    local severity="${4:-error}" # error, warning, info

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    echo -n "检查: $description ... "

    if eval "$check_command" $expected_result; then
        echo -e "${GREEN}✅ 通过${NC}"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        if [[ "$severity" == "error" ]]; then
            echo -e "${RED}❌ 失败${NC}"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        elif [[ "$severity" == "warning" ]]; then
            echo -e "${YELLOW}⚠️ 警告${NC}"
            WARNINGS=$((WARNINGS + 1))
        else
            echo -e "${CYAN}ℹ️ 信息${NC}"
        fi
        return 1
    fi
}

# 主检查函数
main() {
    echo "🔍 MockServer 项目质量检查"
    echo "=================================="
    echo ""

    # 切换到项目根目录
    cd "$(dirname "$0")/.."

    # 检查临时文件
    log_info "检查临时文件和构建产物..."
    check_item "根目录无二进制文件" "! test -f" "mockserver" "error"
    check_item "根目录无日志文件" "! test -f" "mockserver.log" "error"

    # 检查目录结构
    log_info "检查目录结构完整性..."
    local required_dirs=("cmd" "internal" "pkg" "web" "tests" "docs" "scripts")
    for dir in "${required_dirs[@]}"; do
        check_item "必需目录存在: $dir" "test -d" "$dir" "error"
    done

    # 检查必需文件
    log_info "检查必需文件..."
    local required_files=("go.mod" "go.sum" "Makefile" ".gitignore" "README.md" "LICENSE")
    for file in "${required_files[@]}"; do
        check_item "必需文件存在: $file" "test -f" "$file" "error"
    done

    # 生成报告
    echo ""
    echo "=================================="
    echo "📊 项目质量检查报告"
    echo "=================================="
    echo "总检查项目: $TOTAL_CHECKS"
    echo -e "通过检查: ${GREEN}$PASSED_CHECKS${NC}"
    echo -e "失败检查: ${RED}$FAILED_CHECKS${NC}"
    echo -e "警告项目: ${YELLOW}$WARNINGS${NC}"

    local pass_rate=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
    echo "通过率: $pass_rate%"

    echo ""
    if [[ $pass_rate -ge 90 ]]; then
        echo -e "评级: ${GREEN}🟢 优秀 (A级)${NC}"
    elif [[ $pass_rate -ge 80 ]]; then
        echo -e "评级: ${GREEN}🟢 良好 (B级)${NC}"
    else
        echo -e "评级: ${YELLOW}🟡 需要改进${NC}"
    fi

    # 返回适当的退出码
    if [[ $FAILED_CHECKS -gt 0 ]]; then
        exit 1
    elif [[ $WARNINGS -gt 0 ]]; then
        exit 2
    else
        exit 0
    fi
}

# 如果直接运行此脚本
if [[ "${BASH_SOURCE[0]}" = "${0}" ]]; then
    main "$@"
fi
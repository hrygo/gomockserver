#!/bin/bash

# MockServer 项目结构质量检查脚本
# Author: MockServer Team
# Created: 2025-11-19
# Description: 检查项目目录结构规范性和完整性

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# 检查结果统计
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# 检查函数
check_item() {
    local description="$1"
    local check_command="$2"
    local expected_result="$3"

    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

    echo -n "检查: $description ... "

    if eval "$check_command" $expected_result; then
        echo -e "${GREEN}✅ 通过${NC}"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
        return 0
    else
        echo -e "${RED}❌ 失败${NC}"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
        return 1
    fi
}

# 必需的目录结构
REQUIRED_DIRS=(
    "cmd/mockserver"
    "internal"
    "internal/adapter"
    "internal/api"
    "internal/config"
    "internal/engine"
    "internal/service"
    "internal/repository"
    "pkg"
    "web/frontend"
    "web/frontend/src"
    "web/frontend/src/api"
    "web/frontend/src/components"
    "web/frontend/src/pages"
    "tests"
    "tests/integration"
    "docs"
    "scripts"
    "scripts/quality"
)

# 禁止存在的文件/目录
FORBIDDEN_ITEMS=(
    "bin/mockserver"  # 根目录不应该有编译产物
    "temp"
    "tmp"
    ".DS_Store"
    "*.log"
)

# 检查必需目录
check_required_directories() {
    log_info "检查必需的目录结构..."

    for dir in "${REQUIRED_DIRS[@]}"; do
        check_item "目录存在: $dir" "test -d" "$dir"
    done
}

# 检查禁止的项目
check_forbidden_items() {
    log_info "检查禁止的文件和目录..."

    for item in "${FORBIDDEN_ITEMS[@]}"; do
        if [[ $item == *"*"* ]]; then
            # 通配符检查
            check_item "不存在: $item" "find . -maxdepth 1 -name" "$item" && return 1 || return 0
        else
            check_item "不存在: $item" "test -e" "$item" && return 1 || return 0
        fi
    done
}

# 检查目录命名规范
check_naming_conventions() {
    log_info "检查目录命名规范..."

    # 检查是否有大写字母的目录名
    local uppercase_dirs
    uppercase_dirs=$(find . -type d -name "*[A-Z]*" ! -path "./.git/*" ! -path "./node_modules/*" 2>/dev/null || true)

    if [[ -n "$uppercase_dirs" ]]; then
        log_error "发现大写目录名:"
        echo "$uppercase_dirs" | sed 's/^/  - /'
        check_item "目录命名规范 (无大写字母)" "false" "" && return 1
    else
        check_item "目录命名规范 (无大写字母)" "true" "" && return 0
    fi
}

# 检查目录深度
check_directory_depth() {
    log_info "检查目录深度..."

    # 查找深度超过4层的目录
    local deep_dirs
    deep_dirs=$(find . -type d -path "./.git" -prune -o -path "./node_modules" -prune -o -type d -printf '%d\t%p\n' | awk -F'\t' '$1 > 4 {print $2}' || true)

    if [[ -n "$deep_dirs" ]]; then
        log_warn "发现深度超过4层的目录:"
        echo "$deep_dirs" | sed 's/^/  - /'

        # 检查是否是已知的深度问题目录
        if echo "$deep_dirs" | grep -q "internal/graphql"; then
            log_warn "已知问题: internal/graphql 目录需要重构以减少深度"
        fi

        # 这不是严重错误，只是警告
        return 0
    else
        check_item "目录深度 (≤4层)" "true" "" && return 0
    fi
}

# 检查关键配置文件
check_configuration_files() {
    log_info "检查关键配置文件..."

    local config_files=(
        "go.mod"
        "go.sum"
        "Makefile"
        "config.yaml"
        "config.dev.yaml"
        "config.test.yaml"
        "docker-compose.yml"
        "Dockerfile"
        "README.md"
        ".gitignore"
        ".golangci.yml"
    )

    for file in "${config_files[@]}"; do
        check_item "配置文件存在: $file" "test -f" "$file"
    done
}

# 检查前端项目结构
check_frontend_structure() {
    log_info "检查前端项目结构..."

    local frontend_files=(
        "web/frontend/package.json"
        "web/frontend/tsconfig.json"
        "web/frontend/vite.config.ts"
        "web/frontend/src/index.html"
    )

    for file in "${frontend_files[@]}"; do
        check_item "前端文件存在: $file" "test -f" "$file"
    done
}

# 检查Go模块结构
check_go_module_structure() {
    log_info "检查Go模块结构..."

    # 检查主程序入口
    check_item "主程序入口存在" "test -f" "cmd/mockserver/main.go"

    # 检查是否有go.mod文件
    check_item "Go模块文件存在" "test -f" "go.mod"

    # 检查是否有循环依赖（简单检查）
    if command -v go mod graph >/dev/null 2>&1; then
        local has_cycles
        has_cycles=$(go mod graph | grep -c "self" || true)
        if [[ "$has_cycles" -gt 0 ]]; then
            log_warn "检测到可能的循环依赖"
        fi
    fi
}

# 检查文档完整性
check_documentation() {
    log_info "检查文档完整性..."

    local doc_files=(
        "docs"
        "README.md"
        "CHANGELOG.md"
    )

    for file in "${doc_files[@]}"; do
        check_item "文档存在: $file" "test -e" "$file"
    done
}

# 检查测试结构
check_test_structure() {
    log_info "检查测试结构..."

    # 检查是否有基本的测试目录
    check_item "集成测试目录存在" "test -d" "tests/integration"

    # 检查覆盖率目录
    check_item "测试覆盖率目录存在" "test -d" "tests/coverage"

    # 检查是否有基本的测试文件
    local test_files
    test_files=$(find . -name "*_test.go" | wc -l)
    if [[ "$test_files" -gt 0 ]]; then
        check_item "存在Go测试文件" "true" "" && return 0
    else
        check_item "存在Go测试文件" "false" "" && return 1
    fi
}

# 检查安全性
check_security() {
    log_info "检查安全性配置..."

    # 检查.gitignore是否包含敏感文件
    if [[ -f ".gitignore" ]]; then
        local gitignore_content
        gitignore_content=$(cat .gitignore)

        local should_ignore=(
            "*.log"
            "*.env"
            "*.pem"
            "*.key"
            "config.prod.yaml"
            "secrets/"
        )

        for pattern in "${should_ignore[@]}"; do
            if echo "$gitignore_content" | grep -q "$pattern"; then
                check_item ".gitignore包含: $pattern" "true" "" && return 0
            else
                check_item ".gitignore包含: $pattern" "false" "" && return 1
            fi
        done
    else
        log_error ".gitignore文件不存在"
        return 1
    fi
}

# 生成检查报告
generate_report() {
    echo ""
    echo "=================================="
    echo "📊 项目结构质量检查报告"
    echo "=================================="
    echo "总检查项目: $TOTAL_CHECKS"
    echo -e "通过检查: ${GREEN}$PASSED_CHECKS${NC}"
    echo -e "失败检查: ${RED}$FAILED_CHECKS${NC}"

    local pass_rate
    pass_rate=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
    echo "通过率: $pass_rate%"

    if [[ $pass_rate -ge 90 ]]; then
        echo -e "\n🎉 ${GREEN}项目结构质量优秀！${NC}"
        return 0
    elif [[ $pass_rate -ge 75 ]]; then
        echo -e "\n✅ ${GREEN}项目结构质量良好${NC}"
        return 0
    elif [[ $pass_rate -ge 60 ]]; then
        echo -e "\n⚠️  ${YELLOW}项目结构质量一般，建议改进${NC}"
        return 1
    else
        echo -e "\n❌ ${RED}项目结构质量较差，需要立即改进${NC}"
        return 1
    fi
}

# 主函数
main() {
    echo "🔍 MockServer 项目结构质量检查"
    echo "=================================="

    # 切换到项目根目录
    cd "$(dirname "$0")/../.."

    # 执行各项检查
    check_required_directories
    check_forbidden_items
    check_naming_conventions
    check_directory_depth
    check_configuration_files
    check_frontend_structure
    check_go_module_structure
    check_documentation
    check_test_structure
    check_security

    # 生成报告
    generate_report
}

# 如果直接运行此脚本
if [[ "${BASH_SOURCE[0]}" = "${0}" ]]; then
    main "$@"
fi
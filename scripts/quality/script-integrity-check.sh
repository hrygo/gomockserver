#!/bin/bash

# MockServer 脚本完整性检查工具
# Author: MockServer Team
# Created: 2025-11-19
# Description: 检查项目脚本完整性，防止脚本腐化
# Usage: ./scripts/quality/script-integrity-check.sh [options]
# Dependencies: find, grep, wc, shellcheck (optional)

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
VERBOSE=false
FIX_MODE=false
EXIT_ON_ERROR=false

# 统计变量
TOTAL_ISSUES=0
FIXED_ISSUES=0

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
    if [[ "$VERBOSE" == true ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $1"
    fi
}

# 显示帮助信息
show_help() {
    cat << EOF
MockServer 脚本完整性检查工具

用法: $0 [选项]

选项:
    -v, --verbose      详细输出模式
    -f, --fix          自动修复可修复的问题
    -e, --exit-error   发现问题时立即退出
    -h, --help         显示此帮助信息

检查项目:
    1. 脚本执行权限完整性
    2. 重复脚本检测
    3. 孤立脚本检测
    4. 脚本引用关系验证
    5. 脚本质量检查（需要shellcheck）

示例:
    $0                 # 基础检查
    $0 -v              # 详细检查
    $0 -f              # 检查并自动修复
    $0 -v -f           # 详细检查并修复

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -f|--fix)
                FIX_MODE=true
                shift
                ;;
            -e|--exit-error)
                EXIT_ON_ERROR=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 1. 检查脚本执行权限
check_script_permissions() {
    log_info "检查脚本执行权限..."

    local no_exec_scripts
    no_exec_scripts=$(find "$SCRIPT_DIR" -name "*.sh" -type f ! -perm +111 2>/dev/null)

    if [[ -z "$no_exec_scripts" ]]; then
        log_info "✅ 所有脚本都有执行权限"
        return 0
    else
        local count
        count=$(echo "$no_exec_scripts" | wc -l)
        log_error "❌ 发现 $count 个脚本没有执行权限:"
        echo "$no_exec_scripts" | while IFS= read -r script; do
            echo "  - $script"
        done

        if [[ "$FIX_MODE" == true ]]; then
            log_info "🔧 修复执行权限..."
            echo "$no_exec_scripts" | while IFS= read -r script; do
                if chmod +x "$script" 2>/dev/null; then
                    log_info "✅ 已修复: $script"
                    FIXED_ISSUES=$((FIXED_ISSUES + 1))
                else
                    log_error "❌ 修复失败: $script"
                fi
            done
        fi

        TOTAL_ISSUES=$((TOTAL_ISSUES + count))
        return 1
    fi
}

# 2. 检查重复脚本
check_duplicate_scripts() {
    log_info "检查重复脚本..."

    local duplicates
    duplicates=$(find "$SCRIPT_DIR" -name "*.sh" -type f -exec basename {} \; | sort | uniq -d)

    if [[ -z "$duplicates" ]]; then
        log_info "✅ 未发现重复脚本"
        return 0
    else
        log_error "❌ 发现重复的脚本名称:"
        echo "$duplicates" | while IFS= read -r dup; do
            echo "  - $dup"
            find "$SCRIPT_DIR" -name "$dup" -type f | while IFS= read -r script; do
                echo "    $script"
            done
        done

        local count
        count=$(echo "$duplicates" | wc -l)
        TOTAL_ISSUES=$((TOTAL_ISSUES + count))
        return 1
    fi
}

# 3. 检查孤立脚本
check_orphaned_scripts() {
    log_info "检查孤立脚本..."

    # 获取所有脚本文件
    local all_scripts
    all_scripts=$(find "$SCRIPT_DIR" -name "*.sh" -type f | sort)

    # 获取被引用的脚本
    local referenced_scripts
    referenced_scripts=$(
        # 从Makefile中查找
        if [[ -f "$SCRIPT_DIR/Makefile" ]]; then
            grep -oE '\./[^[:space:]]+\.sh' "$SCRIPT_DIR/Makefile" 2>/dev/null | sed 's|^\./||' || true
        fi

        # 从GitHub Actions中查找
        if [[ -d "$SCRIPT_DIR/.github/workflows" ]]; then
            grep -oE '\./[^[:space:]]+\.sh' "$SCRIPT_DIR/.github/workflows"/*.yml 2>/dev/null | sed 's|^\./||' || true
        fi

        # 从其他脚本中查找（排除注释）
        find "$SCRIPT_DIR" -name "*.sh" -type f -exec grep -Hv '^[[:space:]]*#' {} \; 2>/dev/null | \
        grep -oE '\./[^[:space:]]+\.sh' | sed 's|^\./||' | sort -u || true
    )

    log_debug "被引用的脚本:"
    if [[ "$VERBOSE" == true ]]; then
        echo "$referenced_scripts" | while IFS= read -r script; do
            echo "  $script"
        done
    fi

    # 查找孤立脚本
    local orphaned_count=0
    echo "$all_scripts" | while IFS= read -r script; do
        local script_name
        script_name=$(basename "$script")
        local relative_path
        relative_path=${script#$SCRIPT_DIR/}

        # 检查是否被引用
        if ! echo "$referenced_scripts" | grep -q "$script_name" && \
           ! echo "$referenced_scripts" | grep -q "$relative_path"; then
            # 排除一些特殊情况
            if [[ "$script_name" =~ ^(script-integrity-check\.sh|test_.*\.sh)$ ]]; then
                log_debug "跳过检查脚本或测试脚本: $script_name"
                continue
            fi

            echo "  - $script"
            orphaned_count=$((orphaned_count + 1))
        fi
    done

    if [[ $orphaned_count -eq 0 ]]; then
        log_info "✅ 未发现孤立脚本"
        return 0
    else
        log_error "❌ 发现 $orphaned_count 个可能的孤立脚本"
        TOTAL_ISSUES=$((TOTAL_ISSUES + orphaned_count))
        return 1
    fi
}

# 4. 检查脚本引用关系
check_script_references() {
    log_info "检查脚本引用关系..."

    local reference_issues=0

    # 检查Makefile中的引用
    if [[ -f "$SCRIPT_DIR/Makefile" ]]; then
        local makefile_scripts
        makefile_scripts=$(grep -oE '\./[^[:space:]]+\.sh' "$SCRIPT_DIR/Makefile" 2>/dev/null || true)

        if [[ -n "$makefile_scripts" ]]; then
            echo "$makefile_scripts" | while IFS= read -r script_ref; do
                local script_path
                script_path="${script_ref#\./}"
                if [[ ! -f "$SCRIPT_DIR/$script_path" ]]; then
                    log_error "❌ Makefile引用的脚本不存在: $script_ref"
                    reference_issues=$((reference_issues + 1))
                fi
            done
        fi
    fi

    # 检查GitHub Actions中的引用
    if [[ -d "$SCRIPT_DIR/.github/workflows" ]]; then
        local workflow_files=("$SCRIPT_DIR/.github/workflows"/*.yml)
        local workflow_scripts
        workflow_scripts=$(grep -oE '\./[^[:space:]]+\.sh' "${workflow_files[@]}" 2>/dev/null || true)

        if [[ -n "$workflow_scripts" ]]; then
            echo "$workflow_scripts" | while IFS= read -r script_ref; do
                local script_path
                script_path="${script_ref#\./}"
                if [[ ! -f "$SCRIPT_DIR/$script_path" ]]; then
                    # 提供详细的错误信息，包含文件位置
                    local workflow_file
                    workflow_file=$(grep -l "$script_ref" "${workflow_files[@]}" 2>/dev/null | head -1 || echo "unknown")
                    log_error "❌ ${workflow_file##*/}引用的脚本不存在: $script_ref"
                    reference_issues=$((reference_issues + 1))
                fi
            done
        fi
    fi

    if [[ $reference_issues -eq 0 ]]; then
        log_info "✅ 所有脚本引用都有效"
        return 0
    else
        TOTAL_ISSUES=$((TOTAL_ISSUES + reference_issues))
        return 1
    fi
}

# 5. 检查脚本质量（使用shellcheck）
check_script_quality() {
    log_info "检查脚本质量..."

    if ! command -v shellcheck >/dev/null 2>&1; then
        log_warn "⚠️ shellcheck 未安装，跳过质量检查"
        log_info "  安装方法: brew install shellcheck (macOS) 或 apt-get install shellcheck (Ubuntu)"
        return 0
    fi

    local quality_issues=0
    local script_count=0

    find "$SCRIPT_DIR" -name "*.sh" -type f | while IFS= read -r script; do
        script_count=$((script_count + 1))

        if shellcheck "$script" >/dev/null 2>&1; then
            log_debug "✅ $script: 质量检查通过"
        else
            log_error "❌ $script: 存在质量问题"
            if [[ "$VERBOSE" == true ]]; then
                shellcheck "$script" 2>&1 | head -5 | sed 's/^/    /'
            fi
            quality_issues=$((quality_issues + 1))
        fi
    done

    if [[ $quality_issues -eq 0 ]]; then
        log_info "✅ 所有脚本质量检查通过 (共 $script_count 个脚本)"
        return 0
    else
        log_error "❌ $quality_issues 个脚本存在质量问题"
        TOTAL_ISSUES=$((TOTAL_ISSUES + quality_issues))
        return 1
    fi
}

# 生成统计报告
generate_report() {
    echo ""
    echo "=================================="
    echo "📊 脚本完整性检查报告"
    echo "=================================="
    echo "发现问题总数: $TOTAL_ISSUES"

    if [[ "$FIX_MODE" == true && $FIXED_ISSUES -gt 0 ]]; then
        echo -e "已修复问题数: ${GREEN}$FIXED_ISSUES${NC}"
        echo -e "剩余问题数: ${RED}$((TOTAL_ISSUES - FIXED_ISSUES))${NC}"
    fi

    echo ""
    echo "🎯 改进建议:"

    if [[ $TOTAL_ISSUES -eq 0 ]]; then
        echo -e "${GREEN}🎉 所有检查都通过！脚本管理状态优秀。${NC}"
    else
        echo "1. 🚨 立即修复发现的问题"

        if [[ $FIXED_ISSUES -lt $((TOTAL_ISSUES)) ]]; then
            echo "2. 🔧 手动修复自动化工具无法解决的问题"
            echo "3. 📚 参考脚本管理最佳实践文档"
        fi

        echo "4. 🔄 定期运行此检查脚本"
        echo "5. 📈 将检查集成到CI/CD流程"
    fi

    echo ""
    echo "📋 快速修复命令:"
    echo "  修复权限: find . -name '*.sh' -type f ! -perm +111 -exec chmod +x {} \\;"
    echo "  查找重复: find . -name '*.sh' -type f -exec basename {} \\; | sort | uniq -d"
    echo "  质量检查: shellcheck **/*.sh"
}

# 主函数
main() {
    echo "🔍 MockServer 脚本完整性检查"
    echo "=================================="
    echo ""

    # 解析参数
    parse_args "$@"

    # 切换到项目根目录
    cd "$SCRIPT_DIR"

    log_debug "项目目录: $SCRIPT_DIR"
    log_debug "详细模式: $VERBOSE"
    log_debug "修复模式: $FIX_MODE"

    # 执行检查
    local check_start_time
    check_start_time=$(date +%s)

    local failed_checks=0

    check_script_permissions || failed_checks=$((failed_checks + 1))
    if [[ "$EXIT_ON_ERROR" == true && $failed_checks -gt 0 ]]; then
        exit 1
    fi

    check_duplicate_scripts || failed_checks=$((failed_checks + 1))
    if [[ "$EXIT_ON_ERROR" == true && $failed_checks -gt 0 ]]; then
        exit 1
    fi

    check_orphaned_scripts || failed_checks=$((failed_checks + 1))
    if [[ "$EXIT_ON_ERROR" == true && $failed_checks -gt 0 ]]; then
        exit 1
    fi

    check_script_references || failed_checks=$((failed_checks + 1))
    if [[ "$EXIT_ON_ERROR" == true && $failed_checks -gt 0 ]]; then
        exit 1
    fi

    check_script_quality || failed_checks=$((failed_checks + 1))

    local check_end_time
    check_end_time=$(date +%s)
    local duration=$((check_end_time - check_start_time))

    echo ""
    log_info "检查完成，耗时: ${duration}s"

    # 生成报告
    generate_report

    # 返回适当的退出码
    if [[ $TOTAL_ISSUES -gt 0 ]]; then
        exit 1
    else
        exit 0
    fi
}

# 如果直接运行此脚本
if [[ "${BASH_SOURCE[0]}" = "${0}" ]]; then
    main "$@"
fi
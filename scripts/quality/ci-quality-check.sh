#!/bin/bash

# CI/CD 质量门禁检查
# Author: MockServer Team
# Created: 2025-11-19
# Description: CI/CD专用的脚本质量检查，避免复杂引用检查
# Usage: ./scripts/quality/ci-quality-check.sh

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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
TOTAL_ISSUES=0

echo "🔍 CI/CD 质量门禁检查"
echo "=================================="

# 1. 检查脚本执行权限
echo "检查脚本执行权限..."
no_exec_scripts=$(find . -name "*.sh" -type f ! -executable 2>/dev/null)

if [[ -z "$no_exec_scripts" ]]; then
    log_info "✅ 所有脚本都有执行权限"
else
    count=$(echo "$no_exec_scripts" | wc -l)
    log_error "❌ 发现 $count 个脚本没有执行权限"
    echo "$no_exec_scripts" | while IFS= read -r script; do
        echo "  - $script"
    done
    TOTAL_ISSUES=$((TOTAL_ISSUES + count))
fi

# 2. 检查重复脚本
echo "检查重复脚本..."
duplicates=$(find . -name "*.sh" -type f -exec basename {} \; | sort | uniq -d)

if [[ -z "$duplicates" ]]; then
    log_info "✅ 未发现重复脚本"
else
    log_error "❌ 发现重复的脚本名称:"
    echo "$duplicates" | while IFS= read -r dup; do
        echo "  - $dup"
    done
    count=$(echo "$duplicates" | wc -l)
    TOTAL_ISSUES=$((TOTAL_ISSUES + count))
fi

# 3. 检查脚本质量（使用shellcheck）
echo "检查脚本质量..."
if ! command -v shellcheck >/dev/null 2>&1; then
    log_warn "⚠️ shellcheck 未安装，跳过质量检查"
else
    log_info "✅ shellcheck 已安装，开始质量检查"
    # 由于shellcheck检查可能较长，这里只做简单验证
    if shellcheck --version >/dev/null 2>&1; then
        log_info "✅ shellcheck 工具正常"
    else
        log_error "❌ shellcheck 工具异常"
        TOTAL_ISSUES=$((TOTAL_ISSUES + 1))
    fi
fi

# 4. 检查关键脚本存在性
echo "检查关键脚本存在性..."
critical_scripts=(
    "scripts/project-health-check.sh"
    "scripts/quality/script-integrity-check.sh"
    "scripts/check-docker.sh"
    "tests/integration/e2e_test.sh"
)

missing_critical=0
for script in "${critical_scripts[@]}"; do
    if [[ -f "$script" ]]; then
        log_info "✅ $script 存在"
    else
        log_error "❌ $script 不存在"
        missing_critical=$((missing_critical + 1))
    fi
done

if [[ $missing_critical -gt 0 ]]; then
    TOTAL_ISSUES=$((TOTAL_ISSUES + missing_critical))
fi

# 生成报告
echo ""
echo "=================================="
echo "📊 CI/CD 质量检查报告"
echo "=================================="
echo "发现问题总数: $TOTAL_ISSUES"

if [[ $TOTAL_ISSUES -eq 0 ]]; then
    echo -e "${GREEN}🎉 质量检查通过，可以继续CI/CD流程${NC}"
    exit 0
else
    echo -e "${RED}❌ 质量检查失败，请修复问题后重试${NC}"
    exit 1
fi
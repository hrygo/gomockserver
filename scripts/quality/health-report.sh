#!/bin/bash

# MockServer 项目健康报告生成脚本
# Author: MockServer Team
# Created: 2025-11-19
# Description: 生成项目健康状态报告

set -euo pipefail

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# 配置
REPORT_FILE="docs/reports/health-report-$(date +%Y%m%d).md"
TEMP_DIR="/tmp/mockserver-health-$$"

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

# 初始化临时目录
init_temp_dir() {
    mkdir -p "$TEMP_DIR"
    mkdir -p "$(dirname "$REPORT_FILE")"
}

# 清理临时目录
cleanup() {
    rm -rf "$TEMP_DIR"
}

# 设置退出时清理
trap cleanup EXIT

# 获取项目基本信息
get_project_info() {
    local project_name="MockServer"
    local version=$(grep -r "Version.*=" internal/service/health.go | sed 's/.*Version = "\(.*\)".*/\1/' || echo "unknown")
    local go_version=$(go version | awk '{print $3}' | sed 's/go//')
    local node_version=$(cd web/frontend 2>/dev/null && node --version 2>/dev/null || echo "N/A")

    cat > "$TEMP_DIR/project-info.md" << EOF
## 📋 项目基本信息

- **项目名称**: $project_name
- **当前版本**: v$version
- **Go版本**: $go_version
- **Node.js版本**: $node_version
- **报告生成时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **Git分支**: $(git branch --show-current 2>/dev/null || echo "N/A")
- **Git提交**: $(git rev-parse --short HEAD 2>/dev/null || echo "N/A")

EOF
}

# 分析代码质量
analyze_code_quality() {
    log_info "分析代码质量..."

    # Go代码统计
    local go_files=$(find . -name "*.go" -not -path "./.git/*" -not -path "./vendor/*" | wc -l)
    local go_lines=$(find . -name "*.go" -not -path "./.git/*" -not -path "./vendor/*" | xargs wc -l | tail -1 | awk '{print $1}')
    local go_packages=$(go list ./... | wc -l)

    # 前端代码统计
    local ts_files=0
    local ts_lines=0
    if [[ -d "web/frontend" ]]; then
        ts_files=$(find web/frontend -name "*.ts" -o -name "*.tsx" | wc -l)
        ts_lines=$(find web/frontend -name "*.ts" -o -name "*.tsx" | xargs wc -l | tail -1 | awk '{print $1}' 2>/dev/null || echo "0")
    fi

    cat > "$TEMP_DIR/code-quality.md" << EOF
## 💻 代码质量分析

### Go 后端
- **源文件数量**: $go_files 个 .go 文件
- **代码行数**: $go_lines 行
- **包数量**: $go_packages 个包
- **平均每文件行数**: $((go_lines / go_files)) 行/文件

### 前端 (TypeScript)
- **TypeScript文件**: $ts_files 个 .ts/.tsx 文件
- **代码行数**: $ts_lines 行
- **平均每文件行数**: $((ts_lines > 0 ? ts_lines / ts_files : 0)) 行/文件

### 代码复杂度
EOF

    # 运行 golangci-lint 获取复杂度统计
    if command -v golangci-lint >/dev/null 2>&1; then
        local complexity_stats
        complexity_stats=$(golangci-lint run --disable-all -E cyclop --enable-only cyclop 2>/dev/null || echo "无法获取复杂度统计")
        echo "- **圈复杂度统计**:" >> "$TEMP_DIR/code-quality.md"
        echo "$complexity_stats" | head -5 | sed 's/^/  /' >> "$TEMP_DIR/code-quality.md" 2>/dev/null || true
    fi

    echo "" >> "$TEMP_DIR/code-quality.md"
}

# 分析测试覆盖率
analyze_test_coverage() {
    log_info "分析测试覆盖率..."

    cat > "$TEMP_DIR/test-coverage.md" << EOF
## 🧪 测试覆盖率分析

### Go 单元测试
EOF

    # 运行测试并获取覆盖率
    if go test -cover ./... >/dev/null 2>&1; then
        local coverage_output
        coverage_output=$(go test -cover ./... 2>/dev/null | tail -1)
        local coverage_percentage
        coverage_percentage=$(echo "$coverage_output" | grep -o '[0-9.]*%' | head -1)

        echo "- **总体覆盖率**: $coverage_percentage" >> "$TEMP_DIR/test-coverage.md"

        # 按包统计覆盖率
        echo "- **各包覆盖率详情**:" >> "$TEMP_DIR/test-coverage.md"
        go test -cover ./... 2>/dev/null | grep "ok" | grep -v "\[no test files\]" | while read line; do
            local pkg_name=$(echo "$line" | awk '{print $1}' | sed 's|.*/||')
            local pkg_coverage=$(echo "$line" | grep -o '[0-9.]*%')
            echo "  - $pkg_name: $pkg_coverage" >> "$TEMP_DIR/test-coverage.md"
        done
    else
        echo "- ❌ 测试运行失败，无法获取覆盖率信息" >> "$TEMP_DIR/test-coverage.md"
    fi

    # 前端测试覆盖率
    if [[ -d "web/frontend" ]] && cd web/frontend; then
        echo "" >> "$TEMP_DIR/test-coverage.md"
        echo "### 前端测试" >> "$TEMP_DIR/test-coverage.md"

        if npm test -- --watchAll=false --coverage --coverageReport=text >/dev/null 2>&1; then
            local frontend_coverage
            frontend_coverage=$(npm test -- --watchAll=false --coverage --coverageReport=text 2>/dev/null | grep "All files" | grep -o '[0-9.]*%' || echo "N/A")
            echo "- **前端覆盖率**: $frontend_coverage" >> "$TEMP_DIR/test-coverage.md"
        else
            echo "- ❌ 前端测试运行失败" >> "$TEMP_DIR/test-coverage.md"
        fi
        cd - >/dev/null
    fi

    echo "" >> "$TEMP_DIR/test-coverage.md"
}

# 分析依赖状态
analyze_dependencies() {
    log_info "分析依赖状态..."

    cat > "$TEMP_DIR/dependencies.md" << EOF
## 📦 依赖状态分析

### Go 模块依赖
EOF

    if [[ -f "go.mod" ]]; then
        local go_modules_count
        go_modules_count=$(go list -m all | wc -l)
        echo "- **Go模块数量**: $go_modules_count" >> "$TEMP_DIR/dependencies.md"

        # 检查Go模块安全漏洞
        if command -v govulncheck >/dev/null 2>&1; then
            local vuln_count
            vuln_count=$(govulncheck ./... 2>/dev/null | grep -c "Vulnerability" || echo "0")
            if [[ "$vuln_count" -gt 0 ]]; then
                echo "- ⚠️ 发现 $vuln_count 个安全漏洞" >> "$TEMP_DIR/dependencies.md"
            else
                echo "- ✅ 未发现已知安全漏洞" >> "$TEMP_DIR/dependencies.md"
            fi
        fi

        # 分析直接依赖
        echo "- **直接依赖数量**: $(go list -m direct | wc -l)" >> "$TEMP_DIR/dependencies.md"
    fi

    # Node.js 依赖
    if [[ -f "web/frontend/package.json" ]]; then
        echo "" >> "$TEMP_DIR/dependencies.md"
        echo "### Node.js 依赖" >> "$TEMP_DIR/dependencies.md"

        cd web/frontend
        local npm_deps_count
        npm_deps_count=$(npm list --depth=0 --prod 2>/dev/null | grep -c "├\|└" || echo "0")
        local npm_dev_deps_count
        npm_dev_deps_count=$(npm list --depth=0 --dev 2>/dev/null | grep -c "├\|└" || echo "0")

        echo "- **生产依赖**: $npm_deps_count 个" >> "$TEMP_DIR/dependencies.md"
        echo "- **开发依赖**: $npm_dev_deps_count 个" >> "$TEMP_DIR/dependencies.md"

        # 检查npm安全漏洞
        if npm audit --audit-level high >/dev/null 2>&1; then
            local npm_vuln_count
            npm_vuln_count=$(npm audit --json 2>/dev/null | jq -r '.vulnerabilities | length' 2>/dev/null || echo "unknown")
            if [[ "$npm_vuln_count" -gt 0 ]]; then
                echo "- ⚠️ 发现 $npm_vuln_count 个安全漏洞" >> "$TEMP_DIR/dependencies.md"
            else
                echo "- ✅ 未发现高危安全漏洞" >> "$TEMP_DIR/dependencies.md"
            fi
        fi
        cd - >/dev/null
    fi

    echo "" >> "$TEMP_DIR/dependencies.md"
}

# 分析文档状态
analyze_documentation() {
    log_info "分析文档状态..."

    cat > "$TEMP_DIR/documentation.md" << EOF
## 📚 文档状态分析

### 文档统计
EOF

    # 统计各类文档
    local api_docs=$(find docs -name "*.md" -path "*/api/*" 2>/dev/null | wc -l)
    local arch_docs=$(find docs -name "*.md" -path "*/architecture/*" 2>/dev/null | wc -l)
    local dev_docs=$(find docs -name "*.md" -path "*/development/*" 2>/dev/null | wc -l)
    local total_docs=$(find docs -name "*.md" 2>/dev/null | wc -l)

    echo "- **API文档**: $api_docs 个" >> "$TEMP_DIR/documentation.md"
    echo "- **架构文档**: $arch_docs 个" >> "$TEMP_DIR/documentation.md"
    echo "- **开发文档**: $dev_docs 个" >> "$TEMP_DIR/documentation.md"
    echo "- **文档总数**: $total_docs 个" >> "$TEMP_DIR/documentation.md"

    # 检查关键文档
    echo "" >> "$TEMP_DIR/documentation.md"
    echo "### 关键文档检查" >> "$TEMP_DIR/documentation.md"

    local key_docs=("README.md" "CHANGELOG.md" "docs/ARCHITECTURE.md")
    for doc in "${key_docs[@]}"; do
        if [[ -f "$doc" ]]; then
            local doc_size=$(wc -l < "$doc")
            local last_modified=$(stat -f "%Sm" -t "%Y-%m-%d" "$doc" 2>/dev/null || stat -c "%y" "$doc" 2>/dev/null | cut -d' ' -f1)
            echo "- ✅ $doc ($doc_size 行, 更新于 $last_modified)" >> "$TEMP_DIR/documentation.md"
        else
            echo "- ❌ $doc (缺失)" >> "$TEMP_DIR/documentation.md"
        fi
    done

    echo "" >> "$TEMP_DIR/documentation.md"
}

# 分析性能指标
analyze_performance() {
    log_info "分析性能指标..."

    cat > "$TEMP_DIR/performance.md" << EOF
## ⚡ 性能指标分析

### 构建性能
EOF

    # Go构建性能
    local go_build_start=$(date +%s)
    if go build -o /tmp/mockserver-test ./cmd/mockserver >/dev/null 2>&1; then
        local go_build_end=$(date +%s)
        local go_build_time=$((go_build_end - go_build_start))
        echo "- **Go构建时间**: ${go_build_time}秒" >> "$TEMP_DIR/performance.md"
        rm -f /tmp/mockserver-test
    else
        echo "- ❌ Go构建失败" >> "$TEMP_DIR/performance.md"
    fi

    # 前端构建性能
    if [[ -d "web/frontend" ]]; then
        cd web/frontend
        local npm_build_start=$(date +%s)
        if npm run build >/dev/null 2>&1; then
            local npm_build_end=$(date +%s)
            local npm_build_time=$((npm_build_end - npm_build_start))
            echo "- **前端构建时间**: ${npm_build_time}秒" >> "$TEMP_DIR/performance.md"

            # 构建产物大小
            if [[ -d "dist" ]]; then
                local dist_size=$(du -sh dist | cut -f1)
                echo "- **构建产物大小**: $dist_size" >> "$TEMP_DIR/performance.md"
            fi
        else
            echo "- ❌ 前端构建失败" >> "$TEMP_DIR/performance.md"
        fi
        cd - >/dev/null
    fi

    echo "" >> "$TEMP_DIR/performance.md"
}

# 分析Git历史
analyze_git_history() {
    log_info "分析Git历史..."

    cat > "$TEMP_DIR/git-history.md" << EOF
## 📈 Git 活动分析

### 代码提交统计
EOF

    # 总提交数
    local total_commits
    total_commits=$(git rev-list --count HEAD 2>/dev/null || echo "N/A")
    echo "- **总提交数**: $total_commits" >> "$TEMP_DIR/git-history.md"

    # 最近30天活动
    local recent_commits
    recent_commits=$(git rev-list --count --since="30 days ago" HEAD 2>/dev/null || echo "N/A")
    echo "- **最近30天提交**: $recent_commits" >> "$TEMP_DIR/git-history.md"

    # 活跃贡献者
    local contributors
    contributors=$(git shortlog -sn --since="30 days ago" 2>/dev/null | wc -l || echo "N/A")
    echo "- **活跃贡献者**: $contributors 人" >> "$TEMP_DIR/git-history.md"

    # 最大文件变更
    echo "" >> "$TEMP_DIR/git-history.md"
    echo "### 最近变更" >> "$TEMP_DIR/git-history.md"
    git log --oneline -5 2>/dev/null | sed 's/^/- /' >> "$TEMP_DIR/git-history.md" 2>/dev/null || echo "- 无法获取Git日志" >> "$TEMP_DIR/git-history.md"

    echo "" >> "$TEMP_DIR/git-history.md"
}

# 生成健康评分
calculate_health_score() {
    log_info "计算健康评分..."

    local score=0
    local max_score=100

    # 代码质量 (25分)
    local go_files=$(find . -name "*.go" -not -path "./.git/*" | wc -l)
    if [[ $go_files -gt 10 ]]; then ((score += 5)); fi
    if [[ $go_files -gt 50 ]]; then ((score += 5)); fi

    # 测试覆盖率 (25分)
    if [[ -f "go.mod" ]] && go test -cover ./... >/dev/null 2>&1; then
        local coverage=$(go test -cover ./... 2>/dev/null | tail -1 | grep -o '[0-9.]*' | head -1 | cut -d'.' -f1)
        if [[ ${coverage:-0} -ge 50 ]]; then ((score += 10)); fi
        if [[ ${coverage:-0} -ge 70 ]]; then ((score += 10)); fi
        if [[ ${coverage:-0} -ge 80 ]]; then ((score += 5)); fi
    fi

    # 文档完整性 (20分)
    local doc_files=$(find docs -name "*.md" 2>/dev/null | wc -l)
    if [[ $doc_files -ge 5 ]]; then ((score += 5)); fi
    if [[ $doc_files -ge 10 ]]; then ((score += 10)); fi
    if [[ -f "README.md" ]] && [[ -f "CHANGELOG.md" ]]; then ((score += 5)); fi

    # 项目结构 (15分)
    local required_dirs=("cmd" "internal" "pkg" "docs" "tests")
    local existing_dirs=0
    for dir in "${required_dirs[@]}"; do
        if [[ -d "$dir" ]]; then ((existing_dirs++)); fi
    done
    local structure_score=$((existing_dirs * 3))
    ((score += structure_score))

    # 依赖管理 (15分)
    if [[ -f "go.mod" ]] && [[ -f "go.sum" ]]; then ((score += 5)); fi
    if [[ -f "web/frontend/package.json" ]]; then ((score += 5)); fi
    if command -v govulncheck >/dev/null 2>&1 && ! govulncheck ./... 2>/dev/null | grep -q "Vulnerability"; then
        ((score += 5))
    fi

    echo "$score"
}

# 生成最终报告
generate_final_report() {
    log_info "生成最终健康报告..."

    local health_score
    health_score=$(calculate_health_score)

    cat > "$REPORT_FILE" << EOF
# MockServer 项目健康报告

> 📅 生成时间: $(date '+%Y-%m-%d %H:%M:%S')
> 🎯 健康评分: $health_score/100
> 📊 评分等级: $(get_score_grade $health_score)

---

$(cat "$TEMP_DIR/project-info.md")
$(cat "$TEMP_DIR/code-quality.md")
$(cat "$TEMP_DIR/test-coverage.md")
$(cat "$TEMP_DIR/dependencies.md")
$(cat "$TEMP_DIR/documentation.md")
$(cat "$TEMP_DIR/performance.md")
$(cat "$TEMP_DIR/git-history.md")

## 🎯 健康评分详情

### 总分: $health_score/100

$(get_score_breakdown $health_score)

### 改进建议

$(get_improvement_suggestions $health_score)

---

## 📋 下次检查时间

- **日常检查**: $(date -v+7d '+%Y-%m-%d') (每周)
- **详细评估**: $(date -v+30d '+%Y-%m-%d') (每月)
- **架构审查**: $(date -v+90d '+%Y-%m-%d') (每季度)

---

*此报告由 MockServer 自动化工具生成*
EOF

    log_info "健康报告已生成: $REPORT_FILE"
}

# 获取评分等级
get_score_grade() {
    local score=$1
    if [[ $score -ge 90 ]]; then echo "🟢 优秀"; fi
    if [[ $score -ge 75 ]] && [[ $score -lt 90 ]]; then echo "🟡 良好"; fi
    if [[ $score -ge 60 ]] && [[ $score -lt 75 ]]; then echo "🟠 一般"; fi
    if [[ $score -lt 60 ]]; then echo "🔴 需要改进"; fi
}

# 获取评分明细
get_score_breakdown() {
    local score=$1

    echo "| 评估项目 | 得分 | 权重 | 说明 |"
    echo "|---------|------|------|------|"

    # 代码质量
    local code_score=0
    if [[ $score -ge 10 ]]; then code_score=$((score > 20 ? 20 : 10)); fi
    echo "| 代码质量 | $code_score/25 | 25% | 代码规范、结构、复杂度 |"

    # 测试覆盖率
    local test_score=0
    if [[ $score -ge 15 ]]; then test_score=$((score > 40 ? 25 : (score - 15) * 25 / 25)); fi
    echo "| 测试覆盖率 | $test_score/25 | 25% | 单元测试、集成测试覆盖率 |"

    # 文档完整性
    local doc_score=0
    if [[ $score -ge 10 ]]; then doc_score=$((score > 60 ? 20 : (score - 40) * 20 / 20)); fi
    echo "| 文档完整性 | $doc_score/20 | 20% | README、API文档、架构文档 |"

    # 项目结构
    local struct_score=0
    if [[ $score -ge 5 ]]; then struct_score=$((score > 75 ? 15 : (score - 60) * 15 / 15)); fi
    echo "| 项目结构 | $struct_score/15 | 15% | 目录组织、命名规范 |"

    # 依赖管理
    local dep_score=0
    if [[ $score -ge 5 ]]; then dep_score=$((score > 90 ? 15 : (score - 75) * 15 / 15)); fi
    echo "| 依赖管理 | $dep_score/15 | 15% | 版本管理、安全漏洞检查 |"
}

# 获取改进建议
get_improvement_suggestions() {
    local score=$1

    if [[ $score -lt 60 ]]; then
        echo "#### 🔴 紧急改进项"
        echo "- **增加测试覆盖率**: 当前测试覆盖率可能不足，建议增加单元测试和集成测试"
        echo "- **完善文档**: 检查并补充缺失的API文档和架构文档"
        echo "- **规范代码结构**: 检查目录命名和文件组织是否符合最佳实践"
        echo "- **安全漏洞修复**: 检查依赖库的安全漏洞并及时修复"
        echo ""
    fi

    if [[ $score -lt 75 ]]; then
        echo "#### 🟡 重要改进项"
        echo "- **提升代码质量**: 优化代码复杂度，增加代码注释"
        echo "- **完善测试体系**: 补充边界测试和异常测试用例"
        echo "- **增加性能监控**: 添加性能指标监控和基准测试"
        echo ""
    fi

    if [[ $score -lt 90 ]]; then
        echo "#### 🟢 可选改进项"
        echo "- **优化构建速度**: 使用缓存和并行编译提升构建速度"
        echo "- **增强文档交互性**: 添加代码示例和交互式文档"
        echo "- **完善CI/CD流程**: 增加自动化测试和部署流水线"
        echo ""
    fi
}

# 主函数
main() {
    echo "🏥 生成 MockServer 项目健康报告..."
    echo "=================================="

    # 初始化
    init_temp_dir

    # 收集数据
    get_project_info
    analyze_code_quality
    analyze_test_coverage
    analyze_dependencies
    analyze_documentation
    analyze_performance
    analyze_git_history

    # 生成报告
    generate_final_report

    # 显示健康评分
    local health_score
    health_score=$(calculate_health_score)
    echo ""
    echo "=================================="
    echo "🎯 项目健康评分: $health_score/100"
    echo -e "评分等级: $(get_score_grade $health_score)"
    echo "=================================="

    if [[ $health_score -ge 75 ]]; then
        echo -e "${GREEN}✅ 项目健康状况良好！${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠️  项目需要改进，请查看详细报告${NC}"
        return 1
    fi
}

# 如果直接运行此脚本
if [[ "${BASH_SOURCE[0]}" = "${0}" ]]; then
    main "$@"
fi
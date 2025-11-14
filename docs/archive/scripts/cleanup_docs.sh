#!/bin/bash

# Mock Server 文档目录清理脚本
# 功能：清理冗余和过时的测试相关文档

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs"

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   Mock Server 文档清理${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# 确认操作
echo -e "${YELLOW}此脚本将清理以下类型的文件：${NC}"
echo "  1. coverage 目录下的历史覆盖率文件"
echo "  2. reports 目录下的重复和过期报告"
echo "  3. 临时测试输出文件"
echo ""
read -p "确认继续？(y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "取消操作"
    exit 0
fi
echo ""

# 统计清理前的文件数
BEFORE_COUNT=$(find "$DOCS_DIR" -type f | wc -l | tr -d ' ')
echo -e "${BLUE}清理前文件数: $BEFORE_COUNT${NC}"
echo ""

# 1. 清理 coverage 目录
echo -e "${YELLOW}[1/3] 清理 coverage 目录...${NC}"
COVERAGE_DIR="$DOCS_DIR/testing/coverage"

if [ -d "$COVERAGE_DIR" ]; then
    echo "  保留文件："
    echo "    - unit-coverage-*.out (最新的单元测试覆盖率数据)"
    echo "    - unit-coverage-*.html (最新的单元测试覆盖率报告)"
    echo ""
    echo "  删除文件："
    
    # 删除历史的 coverage-all 文件
    if [ -n "$(ls -A "$COVERAGE_DIR"/coverage-all*.out 2>/dev/null)" ]; then
        rm -f "$COVERAGE_DIR"/coverage-all*.out
        echo "    ✓ coverage-all*.out"
    fi
    if [ -n "$(ls -A "$COVERAGE_DIR"/coverage-all*.html 2>/dev/null)" ]; then
        rm -f "$COVERAGE_DIR"/coverage-all*.html
        echo "    ✓ coverage-all*.html"
    fi
    
    # 删除 phase 2 的覆盖率文件
    if [ -n "$(ls -A "$COVERAGE_DIR"/coverage-phase*.* 2>/dev/null)" ]; then
        rm -f "$COVERAGE_DIR"/coverage-phase*.*
        echo "    ✓ coverage-phase*.*"
    fi
    
    # 删除旧的单模块覆盖率文件（不是 unit-coverage- 前缀的）
    for file in adapter-coverage.out engine-coverage.out engine-coverage.html executor-coverage.out repository-coverage.out; do
        if [ -f "$COVERAGE_DIR/$file" ]; then
            rm -f "$COVERAGE_DIR/$file"
            echo "    ✓ $file"
        fi
    done
    
    # 删除集成测试的覆盖率文件
    if [ -n "$(ls -A "$COVERAGE_DIR"/integration*.* 2>/dev/null)" ]; then
        rm -f "$COVERAGE_DIR"/integration*.*
        echo "    ✓ integration*.*"
    fi
    
    echo -e "${GREEN}  ✓ coverage 目录清理完成${NC}"
else
    echo "  coverage 目录不存在，跳过"
fi
echo ""

# 2. 清理 reports 目录
echo -e "${YELLOW}[2/3] 清理 reports 目录...${NC}"
REPORTS_DIR="$DOCS_DIR/testing/reports"

if [ -d "$REPORTS_DIR" ]; then
    echo "  保留文件："
    echo "    - COVERAGE_ANALYSIS_AND_IMPROVEMENT.md (最新的覆盖率分析)"
    echo "    - UNIT_TEST_EXECUTION_SUMMARY.md (最新的执行总结)"
    echo "    - 最新的一组带时间戳报告（unit_test_summary, unit_test_output, coverage_analysis）"
    echo ""
    echo "  删除文件："
    
    # 删除重复的报告文件（保留最核心的几个）
    declare -a redundant_reports=(
        "TESTING.md"
        "TEST_EXECUTION_SUMMARY.md"
        "INTEGRATION_TEST_SUMMARY.md"
        "FINAL_TEST_REPORT.md"
        "PHASE_1_SUMMARY.md"
        "REPOSITORY_TEST_SUMMARY.md"
        "UNIT_TEST_FINAL_SUMMARY.md"
        "INTEGRATION_TEST_FINAL_SUMMARY.md"
        "REAL_DB_TEST_EXECUTION_REPORT.md"
    )
    
    for file in "${redundant_reports[@]}"; do
        if [ -f "$REPORTS_DIR/$file" ]; then
            rm -f "$REPORTS_DIR/$file"
            echo "    ✓ $file (已归档到上层目录的总结文档)"
        fi
    done
    
    # 清理重复的带时间戳报告文件（保留最新的一组）
    # 1. 清理旧的 test-report 文件（昨天的）
    if [ -n "$(ls -A "$REPORTS_DIR"/test-report-*.md 2>/dev/null)" ]; then
        rm -f "$REPORTS_DIR"/test-report-*.md
        echo "    ✓ test-report-*.md (已整合到 UNIT_TEST_REPORT.md)"
    fi
    
    # 2. 只保留最新的一组测试报告（按时间戳排序，删除旧的）
    # 获取所有 unit_test_summary 文件，按时间排序，删除除最新外的所有文件
    SUMMARY_FILES=($(ls -t "$REPORTS_DIR"/unit_test_summary_*.md 2>/dev/null))
    if [ ${#SUMMARY_FILES[@]} -gt 1 ]; then
        for ((i=1; i<${#SUMMARY_FILES[@]}; i++)); do
            rm -f "${SUMMARY_FILES[$i]}"
        done
        echo "    ✓ 旧的 unit_test_summary 文件 (保留最新的)"
    fi
    
    # 获取所有 unit_test_output 文件
    OUTPUT_FILES=($(ls -t "$REPORTS_DIR"/unit_test_output_*.txt 2>/dev/null))
    if [ ${#OUTPUT_FILES[@]} -gt 1 ]; then
        for ((i=1; i<${#OUTPUT_FILES[@]}; i++)); do
            rm -f "${OUTPUT_FILES[$i]}"
        done
        echo "    ✓ 旧的 unit_test_output 文件 (保留最新的)"
    fi
    
    # 获取所有 coverage_analysis 文件
    COVERAGE_FILES=($(ls -t "$REPORTS_DIR"/coverage_analysis_*.txt 2>/dev/null))
    if [ ${#COVERAGE_FILES[@]} -gt 1 ]; then
        for ((i=1; i<${#COVERAGE_FILES[@]}; i++)); do
            rm -f "${COVERAGE_FILES[$i]}"
        done
        echo "    ✓ 旧的 coverage_analysis 文件 (保留最新的)"
    fi
    
    echo -e "${GREEN}  ✓ reports 目录清理完成${NC}"
else
    echo "  reports 目录不存在，跳过"
fi
echo ""

# 3. 清理 testing 根目录的临时文件
echo -e "${YELLOW}[3/3] 清理 testing 根目录...${NC}"
TESTING_DIR="$DOCS_DIR/testing"

if [ -d "$TESTING_DIR" ]; then
    echo "  保留文件："
    echo "    - README.md (测试文档索引)"
    echo "    - COVERAGE_ANALYSIS_AND_IMPROVEMENT.md (覆盖率分析)"
    echo "    - UNIT_TEST_EXECUTION_SUMMARY.md (执行总结)"
    echo "    - UNIT_TEST_REPORT.md (单元测试报告)"
    echo ""
    echo "  删除文件："
    
    # 删除冗余的文档
    declare -a redundant_docs=(
        "ARCHIVE_SUMMARY.md"
        "INTEGRATION_TEST_GUIDE.md"
        "unit-test-output.txt"
    )
    
    for file in "${redundant_docs[@]}"; do
        if [ -f "$TESTING_DIR/$file" ]; then
            rm -f "$TESTING_DIR/$file"
            echo "    ✓ $file (内容已整合到其他文档)"
        fi
    done
    
    echo -e "${GREEN}  ✓ testing 根目录清理完成${NC}"
else
    echo "  testing 目录不存在，跳过"
fi
echo ""

# 统计清理后的文件数
AFTER_COUNT=$(find "$DOCS_DIR" -type f | wc -l | tr -d ' ')
DELETED_COUNT=$((BEFORE_COUNT - AFTER_COUNT))

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   清理完成${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo -e "${GREEN}清理前文件数: $BEFORE_COUNT${NC}"
echo -e "${GREEN}清理后文件数: $AFTER_COUNT${NC}"
echo -e "${GREEN}删除文件数: $DELETED_COUNT${NC}"
echo ""

# 显示清理后的目录结构
echo "清理后保留的核心文档："
echo ""
echo "docs/testing/"
echo "├── README.md                              # 测试文档总索引"
echo "├── COVERAGE_ANALYSIS_AND_IMPROVEMENT.md   # 覆盖率分析和改进方案"
echo "├── UNIT_TEST_EXECUTION_SUMMARY.md         # 单元测试执行总结"
echo "├── UNIT_TEST_REPORT.md                    # 单元测试报告"
echo "├── coverage/                              # 覆盖率数据"
echo "│   ├── unit-coverage-all.out              # 总体覆盖率数据"
echo "│   ├── unit-coverage-all.html             # 总体覆盖率报告"
echo "│   ├── unit-coverage-adapter.html         # Adapter 模块报告"
echo "│   ├── unit-coverage-api.html             # API 模块报告"
echo "│   ├── unit-coverage-engine.html          # Engine 模块报告"
echo "│   ├── unit-coverage-executor.html        # Executor 模块报告"
echo "│   ├── unit-coverage-service.html         # Service 模块报告"
echo "│   └── unit-coverage-repository.html      # Repository 模块报告"
echo "├── reports/                               # 测试报告（最近的）"
echo "│   ├── unit_test_summary_*.md             # 最近的测试总结"
echo "│   ├── coverage_analysis_*.txt            # 最近的覆盖率分析"
echo "│   └── unit_test_output_*.txt             # 最近的测试输出"
echo "└── plans/                                 # 测试计划"
echo "    └── perfect-mvp-testing-plan.md        # MVP测试计划"
echo ""
echo -e "${GREEN}✓ 所有清理任务完成！${NC}"
echo ""

#!/bin/bash

# Mock Server 完整单元测试执行脚本
# 功能：运行所有单元测试、生成覆盖率报告、分析测试结果

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
COVERAGE_DIR="$PROJECT_ROOT/docs/testing/coverage"
REPORTS_DIR="$PROJECT_ROOT/docs/testing/reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# 确保目录存在
mkdir -p "$COVERAGE_DIR"
mkdir -p "$REPORTS_DIR"

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   Mock Server 完整单元测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# 1. 清理旧的测试输出
echo -e "${YELLOW}[1/6] 清理旧的测试输出...${NC}"

# 清理 coverage 目录下的临时文件
if [ -d "$COVERAGE_DIR" ]; then
    echo "  清理 coverage 目录..."
    # 保留最新的 unit-coverage-*.out 和 unit-coverage-*.html
    # 删除其他历史文件
    rm -f "$COVERAGE_DIR"/coverage-all*.out
    rm -f "$COVERAGE_DIR"/coverage-all*.html
    rm -f "$COVERAGE_DIR"/coverage-phase*.out
    rm -f "$COVERAGE_DIR"/coverage-phase*.html
    rm -f "$COVERAGE_DIR"/coverage-phase*.txt
    rm -f "$COVERAGE_DIR"/adapter-coverage.out
    rm -f "$COVERAGE_DIR"/engine-coverage.out
    rm -f "$COVERAGE_DIR"/engine-coverage.html
    rm -f "$COVERAGE_DIR"/executor-coverage.out
    rm -f "$COVERAGE_DIR"/repository-coverage.out
    rm -f "$COVERAGE_DIR"/integration-coverage.out
    rm -f "$COVERAGE_DIR"/integration-real-coverage.out
    rm -f "$COVERAGE_DIR"/integration-real-coverage.html
    echo "  ✓ 已清理历史覆盖率文件"
fi

# 清理 reports 目录下的带时间戳的临时文件（只保留最新的一组）
if [ -d "$REPORTS_DIR" ]; then
    echo "  清理 reports 目录..."
    
    # 清理旧的 test-report 文件
    if [ -n "$(ls -A "$REPORTS_DIR"/test-report-*.md 2>/dev/null)" ]; then
        rm -f "$REPORTS_DIR"/test-report-*.md
    fi
    
    # 只保留最新的一组测试报告（除了即将生成的新报告）
    # 删除除最新外的 unit_test_summary 文件
    SUMMARY_FILES=($(ls -t "$REPORTS_DIR"/unit_test_summary_*.md 2>/dev/null))
    if [ ${#SUMMARY_FILES[@]} -gt 1 ]; then
        for ((i=1; i<${#SUMMARY_FILES[@]}; i++)); do
            rm -f "${SUMMARY_FILES[$i]}"
        done
    fi
    
    # 删除除最新外的 unit_test_output 文件
    OUTPUT_FILES=($(ls -t "$REPORTS_DIR"/unit_test_output_*.txt 2>/dev/null))
    if [ ${#OUTPUT_FILES[@]} -gt 1 ]; then
        for ((i=1; i<${#OUTPUT_FILES[@]}; i++)); do
            rm -f "${OUTPUT_FILES[$i]}"
        done
    fi
    
    # 删除除最新外的 coverage_analysis 文件
    COVERAGE_FILES=($(ls -t "$REPORTS_DIR"/coverage_analysis_*.txt 2>/dev/null))
    if [ ${#COVERAGE_FILES[@]} -gt 1 ]; then
        for ((i=1; i<${#COVERAGE_FILES[@]}; i++)); do
            rm -f "${COVERAGE_FILES[$i]}"
        done
    fi
    
    echo "  ✓ 已清理过期的报告文件"
fi

echo -e "${GREEN}✓ 清理完成${NC}"
echo ""

# 2. 检查测试文件
echo -e "${YELLOW}[2/6] 检查测试文件...${NC}"
TEST_FILES=$(find internal -name "*_test.go" | wc -l | tr -d ' ')
SOURCE_FILES=$(find internal -name "*.go" -not -name "*_test.go" | wc -l | tr -d ' ')
echo "  - 源文件数: $SOURCE_FILES"
echo "  - 测试文件数: $TEST_FILES"
echo ""
echo "  各模块测试文件分布："
for dir in internal/*/; do
    module=$(basename "$dir")
    src_count=$(find "$dir" -name "*.go" -not -name "*_test.go" | wc -l | tr -d ' ')
    test_count=$(find "$dir" -name "*_test.go" | wc -l | tr -d ' ')
    if [ "$test_count" -gt 0 ]; then
        echo "    ✓ $module: $src_count 源文件, $test_count 测试文件"
    else
        echo "    ✗ $module: $src_count 源文件, 无测试文件"
    fi
done
echo -e "${GREEN}✓ 检查完成${NC}"
echo ""

# 3. 运行所有单元测试
echo -e "${YELLOW}[3/6] 运行所有单元测试...${NC}"
TEST_OUTPUT="$REPORTS_DIR/unit_test_output_$TIMESTAMP.txt"

if go test ./internal/... -v -coverprofile="$COVERAGE_DIR/unit-coverage-all.out" 2>&1 | tee "$TEST_OUTPUT"; then
    echo -e "${GREEN}✓ 所有测试通过${NC}"
    TEST_RESULT="PASS"
else
    echo -e "${RED}✗ 测试失败${NC}"
    TEST_RESULT="FAIL"
    exit 1
fi
echo ""

# 4. 生成覆盖率报告
echo -e "${YELLOW}[4/6] 生成覆盖率报告...${NC}"

# 生成总体覆盖率 HTML 报告
go tool cover -html="$COVERAGE_DIR/unit-coverage-all.out" -o "$COVERAGE_DIR/unit-coverage-all.html"
echo "  ✓ 总体覆盖率报告: $COVERAGE_DIR/unit-coverage-all.html"

# 生成各模块覆盖率报告
for module in adapter api engine executor repository service; do
    if go test "./internal/$module" -coverprofile="$COVERAGE_DIR/unit-coverage-$module.out" >/dev/null 2>&1; then
        go tool cover -html="$COVERAGE_DIR/unit-coverage-$module.out" -o "$COVERAGE_DIR/unit-coverage-$module.html" 2>/dev/null
        echo "  ✓ $module 模块覆盖率报告: $COVERAGE_DIR/unit-coverage-$module.html"
    fi
done
echo -e "${GREEN}✓ 覆盖率报告生成完成${NC}"
echo ""

# 5. 分析覆盖率
echo -e "${YELLOW}[5/6] 分析覆盖率...${NC}"
COVERAGE_REPORT="$REPORTS_DIR/coverage_analysis_$TIMESTAMP.txt"

{
    echo "========================================="
    echo "Mock Server 单元测试覆盖率分析"
    echo "生成时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "========================================="
    echo ""
    
    echo "总体覆盖率:"
    go tool cover -func="$COVERAGE_DIR/unit-coverage-all.out" | tail -1
    echo ""
    
    echo "各模块覆盖率:"
    for module in adapter api engine executor repository service; do
        if [ -f "$COVERAGE_DIR/unit-coverage-$module.out" ]; then
            coverage=$(go tool cover -func="$COVERAGE_DIR/unit-coverage-$module.out" | tail -1 | awk '{print $3}')
            printf "  %-15s %s\n" "$module:" "$coverage"
        fi
    done
    echo ""
    
    echo "详细文件覆盖率 (低于80%的文件):"
    go tool cover -func="$COVERAGE_DIR/unit-coverage-all.out" | grep -v "100.0%" | grep -v "total:" | awk '{if($3+0 < 80) print "  " $1 " " $3}' | head -20
    echo ""
    
    echo "测试统计:"
    total_tests=$(grep -c "^=== RUN" "$TEST_OUTPUT" || echo "0")
    passed_tests=$(grep -c "^--- PASS" "$TEST_OUTPUT" || echo "0")
    echo "  总测试数: $total_tests"
    echo "  通过测试: $passed_tests"
    echo ""
    
} | tee "$COVERAGE_REPORT"

echo -e "${GREEN}✓ 覆盖率分析完成${NC}"
echo ""

# 6. 生成测试总结报告
echo -e "${YELLOW}[6/6] 生成测试总结报告...${NC}"
SUMMARY_REPORT="$REPORTS_DIR/unit_test_summary_$TIMESTAMP.md"

{
    echo "# Mock Server 单元测试总结报告"
    echo ""
    echo "**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')  "
    echo "**测试结果**: $TEST_RESULT"
    echo ""
    
    echo "## 📊 测试统计"
    echo ""
    total_tests=$(grep -c "^=== RUN" "$TEST_OUTPUT" || echo "0")
    passed_tests=$(grep -c "^--- PASS" "$TEST_OUTPUT" || echo "0")
    echo "| 指标 | 数值 |"
    echo "|------|------|"
    echo "| 总测试数 | $total_tests |"
    echo "| 通过测试 | $passed_tests |"
    echo "| 源文件数 | $SOURCE_FILES |"
    echo "| 测试文件数 | $TEST_FILES |"
    echo ""
    
    echo "## 📈 覆盖率详情"
    echo ""
    echo "### 总体覆盖率"
    echo "\`\`\`"
    go tool cover -func="$COVERAGE_DIR/unit-coverage-all.out" | tail -1
    echo "\`\`\`"
    echo ""
    
    echo "### 各模块覆盖率"
    echo ""
    echo "| 模块 | 覆盖率 | 测试文件 |"
    echo "|------|--------|---------|"
    for module in adapter api engine executor repository service; do
        test_count=$(find "internal/$module" -name "*_test.go" | wc -l | tr -d ' ')
        if [ -f "$COVERAGE_DIR/unit-coverage-$module.out" ]; then
            coverage=$(go tool cover -func="$COVERAGE_DIR/unit-coverage-$module.out" | tail -1 | awk '{print $3}')
            echo "| $module | $coverage | $test_count |"
        else
            echo "| $module | N/A | $test_count |"
        fi
    done
    echo ""
    
    echo "## 🎯 测试覆盖模块"
    echo ""
    for dir in internal/*/; do
        module=$(basename "$dir")
        echo "### $module"
        test_files=$(find "$dir" -name "*_test.go")
        if [ -n "$test_files" ]; then
            echo ""
            while IFS= read -r file; do
                test_count=$(grep -c "^func Test" "$file" || echo "0")
                echo "- $(basename "$file"): $test_count 个测试函数"
            done <<< "$test_files"
        else
            echo ""
            echo "- 无测试文件"
        fi
        echo ""
    done
    
    echo "## 📁 生成文件"
    echo ""
    echo "- 覆盖率数据: \`$COVERAGE_DIR/unit-coverage-all.out\`"
    echo "- HTML 报告: \`$COVERAGE_DIR/unit-coverage-all.html\`"
    echo "- 测试输出: \`$TEST_OUTPUT\`"
    echo "- 覆盖率分析: \`$COVERAGE_REPORT\`"
    echo ""
    
    echo "## 🔍 低覆盖率文件（< 80%）"
    echo ""
    echo "\`\`\`"
    go tool cover -func="$COVERAGE_DIR/unit-coverage-all.out" | grep -v "100.0%" | grep -v "total:" | awk '{if($3+0 < 80) print $1 " " $3}' | head -20
    echo "\`\`\`"
    echo ""
    
} > "$SUMMARY_REPORT"

echo -e "${GREEN}✓ 测试总结报告生成完成${NC}"
echo ""

# 最终总结
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   测试完成${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo -e "测试结果: ${GREEN}$TEST_RESULT${NC}"
echo ""
echo "生成的报告文件:"
echo "  1. 覆盖率 HTML: $COVERAGE_DIR/unit-coverage-all.html"
echo "  2. 测试输出: $TEST_OUTPUT"
echo "  3. 覆盖率分析: $COVERAGE_REPORT"
echo "  4. 总结报告: $SUMMARY_REPORT"
echo ""

# 显示总体覆盖率
echo "总体覆盖率:"
go tool cover -func="$COVERAGE_DIR/unit-coverage-all.out" | tail -1

echo ""
echo -e "${GREEN}✓ 所有任务完成！${NC}"
echo ""

# 提示如何查看报告
echo "查看报告："
echo "  HTML 覆盖率: open $COVERAGE_DIR/unit-coverage-all.html"
echo "  Markdown 总结: cat $SUMMARY_REPORT"
echo ""

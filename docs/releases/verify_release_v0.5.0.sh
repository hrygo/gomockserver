#!/bin/bash

# MockServer v0.5.0 版本发布验证脚本
# 用于验证版本号一致性和文档完整性

set -e

echo "═══════════════════════════════════════════════════════"
echo "  MockServer v0.5.0 版本发布验证"
echo "═══════════════════════════════════════════════════════"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 错误计数
errors=0

# 验证函数
check_version() {
    local file=$1
    local expected=$2
    local pattern=$3
    
    if grep -q "$pattern" "$file" 2>/dev/null; then
        echo -e "${GREEN}✅${NC} $file - 版本号正确"
    else
        echo -e "${RED}❌${NC} $file - 版本号错误或文件不存在"
        ((errors++))
    fi
}

check_file_exists() {
    local file=$1
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅${NC} $file - 文件存在"
    else
        echo -e "${RED}❌${NC} $file - 文件缺失"
        ((errors++))
    fi
}

echo "1️⃣  检查版本号..."
echo "─────────────────────────────────────────────────────"

# 检查后端版本号
check_version "internal/service/health.go" "0.5.0" 'Version = "0.5.0"'

# 检查 CHANGELOG
check_version "CHANGELOG.md" "0.5.0" "## \[0.5.0\] - 2025-01-17"

# 检查 README
check_version "README.md" "0.5.0" "### 当前版本（v0.5.0）"

# 检查 PROJECT_SUMMARY
check_version "PROJECT_SUMMARY.md" "0.5.0" "已成功完成 v0.5.0 版本"

echo ""
echo "2️⃣  检查必需文件..."
echo "─────────────────────────────────────────────────────"

# 检查核心文件
check_file_exists "CHANGELOG.md"
check_file_exists "README.md"
check_file_exists "RELEASE_NOTES_v0.5.0.md"
check_file_exists "PROJECT_SUMMARY.md"
check_file_exists "RELEASE_v0.5.0_SUMMARY.md"

echo ""
echo "3️⃣  检查 CHANGELOG 内容..."
echo "─────────────────────────────────────────────────────"

# 检查 CHANGELOG 关键内容
if grep -q "请求日志系统" CHANGELOG.md; then
    echo -e "${GREEN}✅${NC} CHANGELOG 包含请求日志系统说明"
else
    echo -e "${RED}❌${NC} CHANGELOG 缺少请求日志系统说明"
    ((errors++))
fi

if grep -q "Prometheus" CHANGELOG.md; then
    echo -e "${GREEN}✅${NC} CHANGELOG 包含 Prometheus 监控说明"
else
    echo -e "${RED}❌${NC} CHANGELOG 缺少 Prometheus 监控说明"
    ((errors++))
fi

if grep -q "80.7%" CHANGELOG.md; then
    echo -e "${GREEN}✅${NC} CHANGELOG 包含测试覆盖率数据"
else
    echo -e "${RED}❌${NC} CHANGELOG 缺少测试覆盖率数据"
    ((errors++))
fi

echo ""
echo "4️⃣  检查 README 新增 API 文档..."
echo "─────────────────────────────────────────────────────"

if grep -q "请求日志 API" README.md; then
    echo -e "${GREEN}✅${NC} README 包含请求日志 API 文档"
else
    echo -e "${RED}❌${NC} README 缺少请求日志 API 文档"
    ((errors++))
fi

if grep -q "/api/v1/request-logs" README.md; then
    echo -e "${GREEN}✅${NC} README 包含请求日志 API 端点"
else
    echo -e "${RED}❌${NC} README 缺少请求日志 API 端点"
    ((errors++))
fi

if grep -q "/api/v1/health/metrics" README.md; then
    echo -e "${GREEN}✅${NC} README 包含 Prometheus 指标端点"
else
    echo -e "${RED}❌${NC} README 缺少 Prometheus 指标端点"
    ((errors++))
fi

echo ""
echo "5️⃣  运行测试验证..."
echo "─────────────────────────────────────────────────────"

# 运行测试
echo "正在运行单元测试..."
if go test -v ./internal/... > /tmp/test_output.txt 2>&1; then
    echo -e "${GREEN}✅${NC} 所有单元测试通过"
else
    echo -e "${RED}❌${NC} 部分单元测试失败"
    echo "查看详细输出: /tmp/test_output.txt"
    ((errors++))
fi

echo ""
echo "6️⃣  检查测试覆盖率..."
echo "─────────────────────────────────────────────────────"

# 生成覆盖率报告
go test -coverprofile=coverage.out ./internal/... > /dev/null 2>&1

# 检查核心模块覆盖率
check_coverage() {
    local module=$1
    local threshold=$2
    
    coverage=$(go tool cover -func=coverage.out | grep "$module" | awk '{print $NF}' | sed 's/%//' | head -1)
    
    if [ -n "$coverage" ]; then
        if (( $(echo "$coverage >= $threshold" | bc -l) )); then
            echo -e "${GREEN}✅${NC} $module 覆盖率: ${coverage}% (>= ${threshold}%)"
        else
            echo -e "${RED}❌${NC} $module 覆盖率: ${coverage}% (< ${threshold}%)"
            ((errors++))
        fi
    fi
}

check_coverage "executor" 80
check_coverage "service" 80
check_coverage "engine" 80

# 总体覆盖率
total_coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $NF}')
echo ""
echo -e "${YELLOW}📊 总体覆盖率: ${total_coverage}${NC}"

echo ""
echo "7️⃣  验证构建..."
echo "─────────────────────────────────────────────────────"

# 尝试构建
if go build -o /tmp/mockserver ./cmd/mockserver > /dev/null 2>&1; then
    echo -e "${GREEN}✅${NC} 项目构建成功"
    rm -f /tmp/mockserver
else
    echo -e "${RED}❌${NC} 项目构建失败"
    ((errors++))
fi

echo ""
echo "═══════════════════════════════════════════════════════"
echo "  验证结果"
echo "═══════════════════════════════════════════════════════"
echo ""

if [ $errors -eq 0 ]; then
    echo -e "${GREEN}🎉 所有验证通过！v0.5.0 版本准备就绪${NC}"
    echo ""
    echo "下一步操作："
    echo "  1. 运行 'git add .' 添加所有更改"
    echo "  2. 运行 'git commit -m \"Release v0.5.0\"' 提交变更"
    echo "  3. 运行 'git tag -a v0.5.0 -m \"Release v0.5.0\"' 创建标签"
    echo "  4. 运行 'git push origin main --tags' 推送到远程仓库"
    echo ""
    exit 0
else
    echo -e "${RED}❌ 发现 $errors 个错误，请修复后重试${NC}"
    echo ""
    exit 1
fi

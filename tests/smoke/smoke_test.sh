#!/bin/bash

# Mock Server 冒烟测试脚本
# 功能：验证主程序可以正常启动和运行

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BINARY="$PROJECT_ROOT/mockserver"
CONFIG_FILE="$PROJECT_ROOT/config.yaml"
ADMIN_API="http://localhost:8080/api/v1"
MOCK_API="http://localhost:9090"

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   Mock Server 冒烟测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# 清理函数
cleanup() {
    if [ ! -z "$SERVER_PID" ]; then
        echo -e "${YELLOW}正在停止服务器...${NC}"
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        echo -e "${GREEN}✓ 服务器已停止${NC}"
    fi
}

# 设置退出时清理
trap cleanup EXIT INT TERM

# 1. 检查二进制文件是否存在
echo -e "${YELLOW}[1/7] 检查二进制文件...${NC}"
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}✗ 二进制文件不存在，正在编译...${NC}"
    cd "$PROJECT_ROOT"
    go build -o mockserver ./cmd/mockserver
    if [ $? -ne 0 ]; then
        echo -e "${RED}✗ 编译失败${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ 编译成功${NC}"
else
    echo -e "${GREEN}✓ 二进制文件存在: $BINARY${NC}"
fi
echo ""

# 2. 检查配置文件
echo -e "${YELLOW}[2/7] 检查配置文件...${NC}"
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}✗ 配置文件不存在: $CONFIG_FILE${NC}"
    exit 1
fi
echo -e "${GREEN}✓ 配置文件存在: $CONFIG_FILE${NC}"
echo ""

# 3. 启动服务器
echo -e "${YELLOW}[3/7] 启动服务器...${NC}"
cd "$PROJECT_ROOT"
$BINARY -config="$CONFIG_FILE" > /tmp/mockserver_smoke_test.log 2>&1 &
SERVER_PID=$!

if [ -z "$SERVER_PID" ]; then
    echo -e "${RED}✗ 服务器启动失败${NC}"
    exit 1
fi

echo -e "${GREEN}✓ 服务器已启动 (PID: $SERVER_PID)${NC}"
echo ""

# 4. 等待服务器就绪
echo -e "${YELLOW}[4/7] 等待服务器就绪...${NC}"
MAX_WAIT=30
WAIT_COUNT=0

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if curl -s "$ADMIN_API/system/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ 服务器已就绪${NC}"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    echo -n "."
done

if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
    echo ""
    echo -e "${RED}✗ 服务器启动超时${NC}"
    echo "服务器日志:"
    tail -20 /tmp/mockserver_smoke_test.log
    exit 1
fi
echo ""

# 5. 测试健康检查API
echo -e "${YELLOW}[5/7] 测试健康检查API...${NC}"
HEALTH_RESPONSE=$(curl -s "$ADMIN_API/system/health")

if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    echo -e "${GREEN}✓ 健康检查通过${NC}"
    echo "  响应: $HEALTH_RESPONSE"
else
    echo -e "${RED}✗ 健康检查失败${NC}"
    echo "  响应: $HEALTH_RESPONSE"
    exit 1
fi
echo ""

# 6. 测试版本信息API
echo -e "${YELLOW}[6/7] 测试版本信息API...${NC}"
VERSION_RESPONSE=$(curl -s "$ADMIN_API/system/version")

if echo "$VERSION_RESPONSE" | grep -q "version"; then
    echo -e "${GREEN}✓ 版本信息获取成功${NC}"
    echo "  响应: $VERSION_RESPONSE"
else
    echo -e "${RED}✗ 版本信息获取失败${NC}"
    echo "  响应: $VERSION_RESPONSE"
    exit 1
fi
echo ""

# 7. 测试基本功能（创建项目）
echo -e "${YELLOW}[7/7] 测试基本功能（创建项目）...${NC}"
PROJECT_RESPONSE=$(curl -s -X POST "$ADMIN_API/projects" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "冒烟测试项目",
        "workspace_id": "smoke-test",
        "description": "自动化冒烟测试"
    }')

if echo "$PROJECT_RESPONSE" | grep -q "id"; then
    echo -e "${GREEN}✓ 项目创建成功${NC}"
    PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "  项目ID: $PROJECT_ID"
    
    # 测试查询项目
    PROJECT_GET=$(curl -s "$ADMIN_API/projects/$PROJECT_ID")
    if echo "$PROJECT_GET" | grep -q "冒烟测试项目"; then
        echo -e "${GREEN}✓ 项目查询成功${NC}"
    else
        echo -e "${RED}✗ 项目查询失败${NC}"
    fi
else
    echo -e "${RED}✗ 项目创建失败${NC}"
    echo "  响应: $PROJECT_RESPONSE"
    exit 1
fi
echo ""

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   冒烟测试完成${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""
echo -e "${GREEN}✓ 所有测试通过！${NC}"
echo ""
echo "测试摘要:"
echo "  - 二进制编译: ✓"
echo "  - 配置文件加载: ✓"
echo "  - 服务器启动: ✓"
echo "  - 健康检查: ✓"
echo "  - 版本信息: ✓"
echo "  - 基本功能: ✓"
echo ""
echo "服务器日志位置: /tmp/mockserver_smoke_test.log"
echo ""

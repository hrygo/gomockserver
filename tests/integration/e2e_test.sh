#!/bin/bash

# Mock Server 端到端集成测试脚本
# 功能：测试完整业务流程 - 项目创建→环境创建→规则创建→Mock请求→规则更新→验证

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BINARY="$PROJECT_ROOT/mockserver"

# 检测运行环境
if [ -n "$GITHUB_ACTIONS" ]; then
    # GitHub Actions 环境
    echo -e "${CYAN}检测到 GitHub Actions 环境${NC}"
    CONFIG_FILE="$PROJECT_ROOT/config.test.yaml"
    ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
    MOCK_API="${MOCK_API:-http://localhost:9090}"
    SKIP_SERVER_START="true"  # 在 GitHub Actions 中使用已运行的服务
else
    # 本地开发环境
    echo -e "${CYAN}检测到本地开发环境${NC}"
    CONFIG_FILE="$PROJECT_ROOT/config.dev.yaml"
    ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
    MOCK_API="${MOCK_API:-http://localhost:9090}"
    SKIP_SERVER_START="${SKIP_SERVER_START:-false}"
fi

# 测试数据变量
PROJECT_ID=""
ENVIRONMENT_ID=""
RULE_ID=""
TEST_PASSED=0
TEST_FAILED=0

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   Mock Server 端到端集成测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

echo -e "${CYAN}使用配置:${NC}"
echo -e "  配置文件: ${YELLOW}$CONFIG_FILE${NC}"
echo -e "  管理API: ${YELLOW}$ADMIN_API${NC}"
echo -e "  MockAPI: ${YELLOW}$MOCK_API${NC}"
echo -e "  跳过服务器启动: ${YELLOW}$SKIP_SERVER_START${NC}"
echo ""

# 清理函数
cleanup() {
    if [ ! -z "$SERVER_PID" ] && [ "$SKIP_SERVER_START" != "true" ]; then
        echo -e "${YELLOW}正在停止服务器...${NC}"
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        echo -e "${GREEN}✓ 服务器已停止${NC}"
    fi

    # 清理依赖服务
    if [ "$START_MONGODB_BY_SCRIPT" = "true" ]; then
        echo -e "${YELLOW}停止 MongoDB 服务 (由脚本启动)...${NC}"
        if command -v make >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/Makefile" ]; then
            make stop-mongo >/dev/null 2>&1 || true
        else
            docker stop mongodb >/dev/null 2>&1 || true
        fi
        echo -e "${GREEN}✓ MongoDB 已停止${NC}"
    fi

    if [ "$START_REDIS_BY_SCRIPT" = "true" ]; then
        echo -e "${YELLOW}停止 Redis 服务 (由脚本启动)...${NC}"
        if command -v make >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/Makefile" ]; then
            make stop-redis >/dev/null 2>&1 || true
        else
            docker stop mockserver-redis >/dev/null 2>&1 || true
        fi
        echo -e "${GREEN}✓ Redis 已停止${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   测试结果统计${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo -e "通过测试: ${GREEN}$TEST_PASSED${NC}"
    echo -e "失败测试: ${RED}$TEST_FAILED${NC}"
    echo -e "总计测试: $((TEST_PASSED + TEST_FAILED))"
    
    if [ $TEST_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ 所有测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}✗ 部分测试失败${NC}"
        exit 1
    fi
}

# 设置退出时清理
trap cleanup EXIT INT TERM

# 测试结果记录函数
test_pass() {
    echo -e "${GREEN}✓ $1${NC}"
    TEST_PASSED=$((TEST_PASSED + 1))
}

test_fail() {
    echo -e "${RED}✗ $1${NC}"
    TEST_FAILED=$((TEST_FAILED + 1))
}

# JSON 提取函数
extract_json_field() {
    echo "$1" | grep -o "\"$2\":\"[^\"]*\"" | cut -d'"' -f4
}

# 服务状态跟踪
START_MONGODB_BY_SCRIPT=false
START_REDIS_BY_SCRIPT=false

# ========================================
# 阶段 0: 准备工作
# ========================================

echo -e "${CYAN}[阶段 0] 准备工作${NC}"
echo ""

# 检查是否需要启动服务器
if [ "$SKIP_SERVER_START" != "true" ]; then
    # 0.0 检查并启动依赖服务
    echo -e "${YELLOW}[0.0] 检查依赖服务...${NC}"
    
    # 检查 MongoDB
    if ! docker ps --format '{{.Names}}' | grep -q '^mongodb$'; then
        echo "启动 MongoDB..."
        if command -v make >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/Makefile" ]; then
            make start-mongo >/dev/null 2>&1
        else
            docker run -d --name mongodb -p 27017:27017 -v mongodb_data:/data/db m.daocloud.io/docker.io/mongo:6.0 >/dev/null 2>&1
        fi
        START_MONGODB_BY_SCRIPT=true
        sleep 3
    else
        echo "MongoDB 已运行"
    fi

    # 检查 Redis
    if ! docker ps --format '{{.Names}}' | grep -q '^mockserver-redis$'; then
        echo "启动 Redis..."
        if command -v make >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/Makefile" ]; then
            make start-redis >/dev/null 2>&1
        else
            docker run -d --name mockserver-redis -p 6379:6379 -v redis_data:/data m.daocloud.io/docker.io/redis:7.2-alpine redis-server --appendonly yes >/dev/null 2>&1
        fi
        START_REDIS_BY_SCRIPT=true
        sleep 2
    else
        echo "Redis 已运行"
    fi
    echo ""

    # 0.1 检查并编译二进制文件
    echo -e "${YELLOW}[0.1] 检查二进制文件...${NC}"
    if [ ! -f "$BINARY" ]; then
        echo "二进制文件不存在，正在编译..."
        cd "$PROJECT_ROOT"
        go build -o mockserver ./cmd/mockserver
        if [ $? -ne 0 ]; then
            test_fail "编译失败"
            exit 1
        fi
        test_pass "编译成功"
    else
        test_pass "二进制文件存在"
    fi
    echo ""

    # 0.2 启动服务器
    echo -e "${YELLOW}[0.2] 启动服务器...${NC}"
    cd "$PROJECT_ROOT"
    echo "使用配置文件: $CONFIG_FILE"
    $BINARY -config="$CONFIG_FILE" > /tmp/mockserver_e2e_test.log 2>&1 &
    SERVER_PID=$!

    if [ -z "$SERVER_PID" ]; then
        test_fail "服务器启动失败"
        # 显示启动日志
        if [ -f "/tmp/mockserver_e2e_test.log" ]; then
            echo "服务器启动日志:"
            tail -20 /tmp/mockserver_e2e_test.log
        fi
        exit 1
    fi
    test_pass "服务器已启动 (PID: $SERVER_PID)"
    echo ""
else
    echo -e "${YELLOW}[0.1] 跳过服务器启动（使用已运行的服务）${NC}"
    test_pass "使用已运行的服务器"
    echo ""
fi

# 0.3 等待服务器就绪
echo -e "${YELLOW}[0.3] 等待服务器就绪...${NC}"
MAX_WAIT=30
WAIT_COUNT=0

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if curl -s "$ADMIN_API/system/health" > /dev/null 2>&1; then
        test_pass "服务器已就绪"
        break
    fi
    sleep 1
    WAIT_COUNT=$((WAIT_COUNT + 1))
    echo -n "."
done

if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
    echo ""
    test_fail "服务器启动超时"
    if [ -f "/tmp/mockserver_e2e_test.log" ]; then
        tail -20 /tmp/mockserver_e2e_test.log
    fi
    exit 1
fi
echo ""

# ========================================
# 阶段 1: 项目管理
# ========================================

echo -e "${CYAN}[阶段 1] 项目管理测试${NC}"
echo ""

# 1.1 创建项目
echo -e "${YELLOW}[1.1] 创建项目...${NC}"
PROJECT_RESPONSE=$(curl -s -X POST "$ADMIN_API/projects" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "E2E测试项目",
        "workspace_id": "e2e-test-workspace",
        "description": "端到端集成测试项目"
    }')

PROJECT_ID=$(extract_json_field "$PROJECT_RESPONSE" "id")
if [ -z "$PROJECT_ID" ]; then
    test_fail "项目创建失败"
    echo "响应: $PROJECT_RESPONSE"
    exit 1
else
    test_pass "项目创建成功 (ID: $PROJECT_ID)"
fi
echo ""

# 1.2 查询项目
echo -e "${YELLOW}[1.2] 查询项目详情...${NC}"
PROJECT_GET=$(curl -s "$ADMIN_API/projects/$PROJECT_ID")
if echo "$PROJECT_GET" | grep -q "E2E测试项目"; then
    test_pass "项目查询成功"
else
    test_fail "项目查询失败"
    echo "响应: $PROJECT_GET"
fi
echo ""

# 1.3 更新项目
echo -e "${YELLOW}[1.3] 更新项目信息...${NC}"
PROJECT_UPDATE=$(curl -s -X PUT "$ADMIN_API/projects/$PROJECT_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "E2E测试项目(已更新)",
        "description": "更新后的描述"
    }')

if echo "$PROJECT_UPDATE" | grep -q "已更新"; then
    test_pass "项目更新成功"
else
    test_fail "项目更新失败"
fi
echo ""

# 1.4 列出所有项目
echo -e "${YELLOW}[1.4] 列出所有项目...${NC}"
PROJECTS_LIST=$(curl -s "$ADMIN_API/projects")
if echo "$PROJECTS_LIST" | grep -q "$PROJECT_ID"; then
    test_pass "项目列表查询成功"
else
    test_fail "项目列表查询失败"
fi
echo ""

# ========================================
# 阶段 2: 环境管理
# ========================================

echo -e "${CYAN}[阶段 2] 环境管理测试${NC}"
echo ""

# 2.1 创建环境
echo -e "${YELLOW}[2.1] 创建环境...${NC}"
ENV_RESPONSE=$(curl -s -X POST "$ADMIN_API/projects/$PROJECT_ID/environments" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "开发环境",
        "base_url": "http://dev.example.com",
        "variables": {
            "api_version": "v1",
            "timeout": "30s"
        }
    }')

ENVIRONMENT_ID=$(extract_json_field "$ENV_RESPONSE" "id")
if [ -z "$ENVIRONMENT_ID" ]; then
    test_fail "环境创建失败"
    echo "响应: $ENV_RESPONSE"
    exit 1
else
    test_pass "环境创建成功 (ID: $ENVIRONMENT_ID)"
fi
echo ""

# 2.2 查询环境
echo -e "${YELLOW}[2.2] 查询环境详情...${NC}"
ENV_GET=$(curl -s "$ADMIN_API/projects/$PROJECT_ID/environments/$ENVIRONMENT_ID")
if echo "$ENV_GET" | grep -q "开发环境"; then
    test_pass "环境查询成功"
else
    test_fail "环境查询失败"
fi
echo ""

# 2.3 更新环境
echo -e "${YELLOW}[2.3] 更新环境信息...${NC}"
ENV_UPDATE=$(curl -s -X PUT "$ADMIN_API/projects/$PROJECT_ID/environments/$ENVIRONMENT_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "开发环境(已更新)",
        "base_url": "http://dev-updated.example.com"
    }')

if echo "$ENV_UPDATE" | grep -q "已更新"; then
    test_pass "环境更新成功"
else
    test_fail "环境更新失败"
fi
echo ""

# 2.4 列出项目的所有环境
echo -e "${YELLOW}[2.4] 列出项目所有环境...${NC}"
ENVS_LIST=$(curl -s "$ADMIN_API/projects/$PROJECT_ID/environments")
if echo "$ENVS_LIST" | grep -q "$ENVIRONMENT_ID"; then
    test_pass "环境列表查询成功"
else
    test_fail "环境列表查询失败"
fi
echo ""

# ========================================
# 阶段 3: 规则管理
# ========================================

echo -e "${CYAN}[阶段 3] 规则管理测试${NC}"
echo ""

# 3.1 创建 HTTP Mock 规则
echo -e "${YELLOW}[3.1] 创建 HTTP Mock 规则...${NC}"
RULE_RESPONSE=$(curl -s -X POST "$ADMIN_API/rules" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "获取用户列表API",
        "project_id": "'"$PROJECT_ID"'",
        "environment_id": "'"$ENVIRONMENT_ID"'",
        "protocol": "HTTP",
        "match_type": "Simple",
        "priority": 100,
        "enabled": true,
        "match_condition": {
            "method": "GET",
            "path": "/api/users"
        },
        "response": {
            "type": "Static",
            "content": {
                "status_code": 200,
                "content_type": "JSON",
                "headers": {
                    "X-Custom-Header": "test-value"
                },
                "body": {
                    "code": 0,
                    "message": "success",
                    "data": [
                        {"id": 1, "name": "张三"},
                        {"id": 2, "name": "李四"}
                    ]
                }
            }
        }
    }')

RULE_ID=$(extract_json_field "$RULE_RESPONSE" "id")
if [ -z "$RULE_ID" ]; then
    test_fail "规则创建失败"
    echo "响应: $RULE_RESPONSE"
    exit 1
else
    test_pass "规则创建成功 (ID: $RULE_ID)"
fi
echo ""

# 3.2 查询规则
echo -e "${YELLOW}[3.2] 查询规则详情...${NC}"
RULE_GET=$(curl -s "$ADMIN_API/rules/$RULE_ID")
if echo "$RULE_GET" | grep -q "获取用户列表API"; then
    test_pass "规则查询成功"
else
    test_fail "规则查询失败"
fi
echo ""

# 3.3 更新规则
echo -e "${YELLOW}[3.3] 更新规则...${NC}"
RULE_UPDATE=$(curl -s -X PUT "$ADMIN_API/rules/$RULE_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "获取用户列表API(v2)",
        "priority": 200,
        "enabled": true,
        "response": {
            "type": "Static",
            "content": {
                "status_code": 200,
                "content_type": "JSON",
                "body": {
                    "code": 0,
                    "message": "success",
                    "data": [
                        {"id": 1, "name": "张三", "age": 25},
                        {"id": 2, "name": "李四", "age": 30},
                        {"id": 3, "name": "王五", "age": 28}
                    ]
                }
            }
        }
    }')

if echo "$RULE_UPDATE" | grep -q "v2"; then
    test_pass "规则更新成功"
else
    test_fail "规则更新失败"
fi
echo ""

# 3.4 创建带延迟的规则
echo -e "${YELLOW}[3.4] 创建带延迟的规则...${NC}"
DELAY_RULE_RESPONSE=$(curl -s -X POST "$ADMIN_API/rules" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "延迟响应测试",
        "project_id": "'"$PROJECT_ID"'",
        "environment_id": "'"$ENVIRONMENT_ID"'",
        "protocol": "HTTP",
        "match_type": "Simple",
        "priority": 50,
        "enabled": true,
        "match_condition": {
            "method": "GET",
            "path": "/api/slow"
        },
        "response": {
            "type": "Static",
            "delay": {
                "type": "fixed",
                "fixed": 100
            },
            "content": {
                "status_code": 200,
                "content_type": "JSON",
                "body": {
                    "message": "delayed response"
                }
            }
        }
    }')

DELAY_RULE_ID=$(extract_json_field "$DELAY_RULE_RESPONSE" "id")
if [ -z "$DELAY_RULE_ID" ]; then
    test_fail "延迟规则创建失败"
else
    test_pass "延迟规则创建成功 (ID: $DELAY_RULE_ID)"
fi
echo ""

# 3.5 列出所有规则
echo -e "${YELLOW}[3.5] 列出所有规则...${NC}"
RULES_LIST=$(curl -s "$ADMIN_API/rules?project_id=$PROJECT_ID&environment_id=$ENVIRONMENT_ID")
if echo "$RULES_LIST" | grep -q "$RULE_ID"; then
    test_pass "规则列表查询成功"
else
    test_fail "规则列表查询失败"
fi

# 调试：立即验证第一个规则是否工作
echo -e "${YELLOW}[3.6] 立即验证第一个规则...${NC}"
echo "  等待 2 秒让规则生效..."
sleep 2

DEBUG_RESPONSE=$(curl -s -w "\n%{http_code}" \
    "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/users")

DEBUG_CODE=$(echo "$DEBUG_RESPONSE" | tail -n 1)
DEBUG_BODY=$(echo "$DEBUG_RESPONSE" | sed '$d')

echo "  调试请求状态码: $DEBUG_CODE"
echo "  调试请求响应体: $DEBUG_BODY"

if [ "$DEBUG_CODE" = "200" ]; then
    echo "  ✓ 第一个规则正常工作"
else
    echo "  ✗ 第一个规则有问题，需要进一步调试"
    echo "  规则ID: $RULE_ID"
fi
echo ""

# ========================================
# 阶段 4: Mock 请求测试
# ========================================

echo -e "${CYAN}[阶段 4] Mock 请求测试${NC}"
echo ""

# 确保使用更新后的规则（避免 headers 字段可能的问题）
echo -e "${YELLOW}[4.0] 确保规则配置正确...${NC}"
echo "  重新更新规则以确保配置正确..."

RULE_FIX=$(curl -s -X PUT "$ADMIN_API/rules/$RULE_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "获取用户列表API(测试版)",
        "priority": 300,
        "enabled": true,
        "response": {
            "type": "Static",
            "content": {
                "status_code": 200,
                "content_type": "JSON",
                "headers": {
                    "X-Custom-Header": "test-value"
                },
                "body": {
                    "code": 0,
                    "message": "success",
                    "data": [
                        {"id": 1, "name": "张三"},
                        {"id": 2, "name": "李四"}
                    ]
                }
            }
        }
    }')

echo "  规则修复响应: $RULE_FIX"

# 给服务器更多时间处理规则
echo "  等待规则生效..."
sleep 3

# 4.1 测试基本 Mock 请求
echo -e "${YELLOW}[4.1] 测试基本 Mock 请求...${NC}"

# 添加重试机制，最多尝试3次
MAX_RETRIES=3
RETRY_COUNT=0
TEST_SUCCESS=false

while [ $RETRY_COUNT -lt $MAX_RETRIES ] && [ "$TEST_SUCCESS" = false ]; do
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "  尝试第 $RETRY_COUNT 次..."

    MOCK_RESPONSE=$(curl -s -L -w "\n%{http_code}" \
        "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/users")

    HTTP_CODE=$(echo "$MOCK_RESPONSE" | tail -n 1)
    RESPONSE_BODY=$(echo "$MOCK_RESPONSE" | sed '$d')

    echo "  状态码: $HTTP_CODE"
    echo "  响应体: $RESPONSE_BODY"

    if [ "$HTTP_CODE" = "200" ]; then
        if echo "$RESPONSE_BODY" | grep -q "张三"; then
            test_pass "Mock 请求成功，返回正确数据"
            TEST_SUCCESS=true
        else
            echo "  响应数据不正确，准备重试..."
        fi
    else
        echo "  HTTP状态码错误: $HTTP_CODE，准备重试..."
    fi

    if [ "$TEST_SUCCESS" = false ] && [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
        echo "  等待 2 秒后重试..."
        sleep 2
    fi
done

if [ "$TEST_SUCCESS" = false ]; then
    test_fail "Mock 请求失败，重试 $MAX_RETRIES 次后仍失败"
    echo "  最终状态码: $HTTP_CODE"
    echo "  最终响应体: $RESPONSE_BODY"
fi
echo ""

# 4.2 测试自定义 Header
echo -e "${YELLOW}[4.2] 测试自定义 Header...${NC}"

# 添加重试机制
HEADER_TEST_SUCCESS=false
HEADER_RETRY_COUNT=0

while [ $HEADER_RETRY_COUNT -lt $MAX_RETRIES ] && [ "$HEADER_TEST_SUCCESS" = false ]; do
    HEADER_RETRY_COUNT=$((HEADER_RETRY_COUNT + 1))
    echo "  Header 测试尝试第 $HEADER_RETRY_COUNT 次..."

    HEADER_RESPONSE=$(curl -s -i -L \
        "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/users")

    echo "  完整响应头:"
    echo "$HEADER_RESPONSE" | head -20

    if echo "$HEADER_RESPONSE" | grep -q "X-Custom-Header: test-value"; then
        test_pass "自定义 Header 正确返回"
        HEADER_TEST_SUCCESS=true
    else
        echo "  自定义 Header 未找到，准备重试..."
        if [ $HEADER_RETRY_COUNT -lt $MAX_RETRIES ]; then
            echo "  等待 2 秒后重试..."
            sleep 2
        fi
    fi
done

if [ "$HEADER_TEST_SUCCESS" = false ]; then
    test_fail "自定义 Header 测试失败，重试 $MAX_RETRIES 次后仍失败"
    echo "  响应内容:"
    echo "$HEADER_RESPONSE"
fi
echo ""

# 4.3 测试延迟响应
echo -e "${YELLOW}[4.3] 测试延迟响应...${NC}"
# macOS compatible way to get milliseconds
START_TIME=$(python3 -c 'import time; print(int(time.time() * 1000))' 2>/dev/null || date +%s000)
DELAY_RESPONSE=$(curl -s -L \
    "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/slow")
# macOS compatible way to get milliseconds
END_TIME=$(python3 -c 'import time; print(int(time.time() * 1000))' 2>/dev/null || date +%s000)
DURATION=$((END_TIME - START_TIME))

if [ $DURATION -ge 100 ]; then
    test_pass "延迟响应正确 (耗时: ${DURATION}ms)"
else
    test_fail "延迟时间不足 (耗时: ${DURATION}ms)"
fi
echo ""

# 4.4 测试不匹配的请求（应返回404）
echo -e "${YELLOW}[4.4] 测试不匹配的请求...${NC}"
NOT_FOUND_RESPONSE=$(curl -s -L -w "\n%{http_code}" \
    "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/not-exists")

NOT_FOUND_CODE=$(echo "$NOT_FOUND_RESPONSE" | tail -n 1)
if [ "$NOT_FOUND_CODE" = "404" ]; then
    test_pass "不匹配请求正确返回404"
else
    test_fail "不匹配请求状态码错误: $NOT_FOUND_CODE"
fi
echo ""

# 4.5 测试POST请求
echo -e "${YELLOW}[4.5] 创建并测试 POST 请求规则...${NC}"

# 先创建POST规则
POST_RULE_RESPONSE=$(curl -s -X POST "$ADMIN_API/rules" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "创建用户API",
        "project_id": "'"$PROJECT_ID"'",
        "environment_id": "'"$ENVIRONMENT_ID"'",
        "protocol": "HTTP",
        "match_type": "Simple",
        "priority": 100,
        "enabled": true,
        "match_condition": {
            "method": "POST",
            "path": "/api/users"
        },
        "response": {
            "type": "Static",
            "content": {
                "status_code": 201,
                "content_type": "JSON",
                "body": {
                    "code": 0,
                    "message": "用户创建成功",
                    "data": {
                        "id": 123,
                        "name": "新用户"
                    }
                }
            }
        }
    }')

sleep 1  # 等待规则生效

# 测试POST请求
POST_MOCK_RESPONSE=$(curl -s -L -w "\n%{http_code}" -X POST \
    -H "Content-Type: application/json" \
    -d '{"name": "测试用户"}' \
    "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/users")

POST_CODE=$(echo "$POST_MOCK_RESPONSE" | tail -n 1)
POST_BODY=$(echo "$POST_MOCK_RESPONSE" | sed '$d')

if [ "$POST_CODE" = "201" ]; then
    if echo "$POST_BODY" | grep -q "用户创建成功"; then
        test_pass "POST 请求 Mock 成功"
    else
        test_fail "POST 请求返回数据不正确"
    fi
else
    test_fail "POST 请求失败，状态码: $POST_CODE"
fi
echo ""

# ========================================
# 阶段 5: 规则状态管理
# ========================================

echo -e "${CYAN}[阶段 5] 规则状态管理测试${NC}"
echo ""

# 5.1 禁用规则
echo -e "${YELLOW}[5.1] 禁用规则...${NC}"
DISABLE_RULE=$(curl -s -X PUT "$ADMIN_API/rules/$RULE_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "enabled": false
    }')

if echo "$DISABLE_RULE" | grep -q "false"; then
    test_pass "规则禁用成功"
else
    test_fail "规则禁用失败"
fi

sleep 1

# 验证禁用后请求返回404
DISABLED_RESPONSE=$(curl -s -L -w "\n%{http_code}" \
    "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/users")

DISABLED_CODE=$(echo "$DISABLED_RESPONSE" | tail -n 1)
if [ "$DISABLED_CODE" = "404" ]; then
    test_pass "禁用规则后正确返回404"
else
    test_fail "禁用规则后状态码错误: $DISABLED_CODE"
fi
echo ""

# 5.2 重新启用规则
echo -e "${YELLOW}[5.2] 重新启用规则...${NC}"
ENABLE_RULE=$(curl -s -X PUT "$ADMIN_API/rules/$RULE_ID" \
    -H "Content-Type: application/json" \
    -d '{
        "enabled": true
    }')

if echo "$ENABLE_RULE" | grep -q "true"; then
    test_pass "规则启用成功"
else
    test_fail "规则启用失败"
fi

sleep 1

# 验证启用后请求正常
ENABLED_RESPONSE=$(curl -s -L -w "\n%{http_code}" \
    "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/api/users")

ENABLED_CODE=$(echo "$ENABLED_RESPONSE" | tail -n 1)
if [ "$ENABLED_CODE" = "200" ]; then
    test_pass "启用规则后请求正常"
else
    test_fail "启用规则后请求失败: $ENABLED_CODE"
fi
echo ""

# ========================================
# 阶段 6: 清理测试数据
# ========================================

echo -e "${CYAN}[阶段 6] 清理测试数据${NC}"
echo ""

# 6.1 删除规则
echo -e "${YELLOW}[6.1] 删除规则...${NC}"
DELETE_RULE=$(curl -s -X DELETE "$ADMIN_API/rules/$RULE_ID")
if [ $? -eq 0 ]; then
    test_pass "规则删除成功"
else
    test_fail "规则删除失败"
fi
echo ""

# 6.2 删除环境
echo -e "${YELLOW}[6.2] 删除环境...${NC}"
DELETE_ENV=$(curl -s -X DELETE "$ADMIN_API/projects/$PROJECT_ID/environments/$ENVIRONMENT_ID")
if [ $? -eq 0 ]; then
    test_pass "环境删除成功"
else
    test_fail "环境删除失败"
fi
echo ""

# 6.3 删除项目
echo -e "${YELLOW}[6.3] 删除项目...${NC}"
DELETE_PROJECT=$(curl -s -X DELETE "$ADMIN_API/projects/$PROJECT_ID")
if [ $? -eq 0 ]; then
    test_pass "项目删除成功"
else
    test_fail "项目删除失败"
fi
echo ""

# cleanup 函数会在脚本退出时自动调用，打印测试统计
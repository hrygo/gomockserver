#!/bin/bash

# Simple WebSocket test to verify functionality

# 加载测试框架
source "$(dirname "$0")/lib/test_framework.sh"

# 初始化测试框架
init_test_framework

echo "Testing WebSocket functionality..."

# 创建测试项目
echo "Creating test project..."
PROJECT_RESPONSE=$(http_post "$ADMIN_API/projects" "$(generate_project_data "WebSocket简单测试")")
PROJECT_ID=$(extract_json_field "$PROJECT_RESPONSE" "id")

if [ -n "$PROJECT_ID" ]; then
    test_pass "项目创建成功 (ID: $PROJECT_ID)"
else
    test_fail "项目创建失败"
    exit 1
fi

# 创建测试环境
echo "Creating test environment..."
ENV_RESPONSE=$(http_post "$ADMIN_API/projects/$PROJECT_ID/environments" "$(generate_environment_data "WebSocket环境" "ws://localhost:9090")")
ENVIRONMENT_ID=$(extract_json_field "$ENV_RESPONSE" "id")

if [ -n "$ENVIRONMENT_ID" ]; then
    test_pass "环境创建成功 (ID: $ENVIRONMENT_ID)"
else
    test_fail "环境创建失败"
    exit 1
fi

# 创建 WebSocket 规则
echo "Creating WebSocket rule..."
RULE_RESPONSE=$(http_post "$ADMIN_API/rules" '{
    "name": "WebSocket测试规则",
    "project_id": "'$PROJECT_ID'",
    "environment_id": "'$ENVIRONMENT_ID'",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/test"
    },
    "response": {
        "type": "Static",
        "content": {
            "message": "WebSocket测试成功",
            "type": "test_response",
            "timestamp": "{{timestamp}}"
        }
    }
}')

RULE_ID=$(extract_json_field "$RULE_RESPONSE" "id")

if [ -n "$RULE_ID" ]; then
    test_pass "WebSocket规则创建成功 (ID: $RULE_ID)"
else
    test_fail "WebSocket规则创建失败"
    exit 1
fi

# 等待规则生效
echo "等待规则生效..."
sleep 3

# 测试 WebSocket 连接
echo "Testing WebSocket connection..."
WS_URL="ws://localhost:9090/$PROJECT_ID/$ENVIRONMENT_ID/ws/test"

if command -v websocat >/dev/null 2>&1; then
    echo "使用 websocat 测试 WebSocket 连接..."
    WS_RESULT=$(echo '{"test": "hello"}' | timeout_cmd 5 websocat --one-message "$WS_URL" 2>/dev/null || echo "TIMEOUT")

    if [ "$WS_RESULT" != "TIMEOUT" ]; then
        if echo "$WS_RESULT" | grep -q "WebSocket测试成功"; then
            test_pass "WebSocket连接测试成功"
            echo "响应内容: $WS_RESULT"
        else
            test_fail "WebSocket响应内容不正确: $WS_RESULT"
        fi
    else
        test_fail "WebSocket连接超时"
    fi
else
    test_skip "websocat 未安装，跳过 WebSocket 连接测试"
fi

print_test_summary
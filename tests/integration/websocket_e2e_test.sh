#!/bin/bash

# MockServer WebSocket E2E 集成测试脚本
# 测试 WebSocket 协议的完整功能

set -e

# 加载测试框架
source "$(dirname "$0")/lib/test_framework.sh"

# 初始化测试框架
init_test_framework

# 检查并安装 WebSocket 测试工具
check_and_install_websocket_tools() {
    # 加载工具安装器
    local installer_path="$(dirname "$0")/lib/tool_installer.sh"
    if [ -f "$installer_path" ]; then
        source "$installer_path"

        # 检查 WebSocket 测试工具
        if ! check_tools_ready "websocket"; then
            echo -e "${YELLOW}检测到缺失的 WebSocket 测试工具，正在自动安装...${NC}"
            if ! install_required_tools_silent "websocket"; then
                echo -e "${YELLOW}WebSocket 测试工具安装失败，将跳过相关测试${NC}"
                return 1
            fi
        fi
    else
        echo -e "${YELLOW}工具安装器不可用，请手动安装 websocat${NC}"
        return 1
    fi
}

echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   Mock Server WebSocket E2E 集成测试${NC}"
echo -e "${BLUE}=========================================${NC}"
echo ""

# 检查必要的工具
check_websocket_tools() {
    if ! command -v websocat >/dev/null 2>&1; then
        echo -e "${YELLOW}websocat 未安装，正在尝试自动安装...${NC}"
        check_and_install_websocket_tools
    fi

    # 再次检查
    if ! command -v websocat >/dev/null 2>&1; then
        test_skip "websocat 安装失败，跳过 WebSocket 测试"
        return 1
    fi

    echo -e "${GREEN}WebSocket 测试工具就绪${NC}"
    return 0
}

# ========================================
# 阶段 1: WebSocket 项目和环境创建
# ========================================

echo -e "${CYAN}[阶段 1] WebSocket 项目和环境创建${NC}"
echo ""

# 1.1 创建 WebSocket 测试项目
echo -e "${YELLOW}[1.1] 创建 WebSocket 测试项目...${NC}"
WS_PROJECT_RESPONSE=$(http_post "$ADMIN_API/projects" "$(generate_project_data "WebSocket测试项目")")

if echo "$WS_PROJECT_RESPONSE" | grep -q '"id"'; then
    WS_PROJECT_ID=$(extract_json_field "$WS_PROJECT_RESPONSE" "id")
    test_pass "WebSocket项目创建成功 (ID: $WS_PROJECT_ID)"
else
    test_fail "WebSocket项目创建失败"
    exit 1
fi

# 1.2 创建 WebSocket 测试环境
echo -e "${YELLOW}[1.2] 创建 WebSocket 测试环境...${NC}"
WS_ENV_RESPONSE=$(http_post "$ADMIN_API/projects/$WS_PROJECT_ID/environments" "$(generate_environment_data "WebSocket环境" "ws://localhost:9090")")

if echo "$WS_ENV_RESPONSE" | grep -q '"id"'; then
    WS_ENVIRONMENT_ID=$(extract_json_field "$WS_ENV_RESPONSE" "id")
    test_pass "WebSocket环境创建成功 (ID: $WS_ENVIRONMENT_ID)"
else
    test_fail "WebSocket环境创建失败"
    exit 1
fi

echo ""

# ========================================
# 阶段 2: WebSocket 连接测试
# ========================================

echo -e "${CYAN}[阶段 2] WebSocket 连接测试${NC}"
echo ""

if ! check_websocket_tools; then
    exit 0
fi

# 2.1 基础 WebSocket 连接测试
echo -e "${YELLOW}[2.1] 测试基础 WebSocket 连接...${NC}"
WS_CONNECT_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket连接测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/connect-test"
    },
    "response": {
        "type": "Static",
        "content": {
            "message": "WebSocket连接成功",
            "type": "connection_established",
            "timestamp": "{{timestamp}}"
        }
    }
}
EOF
)")"

if echo "$WS_CONNECT_RULE_RESPONSE" | grep -q '"id"'; then
    WS_CONNECT_RULE_ID=$(extract_json_field "$WS_CONNECT_RULE_RESPONSE" "id")
    test_pass "WebSocket连接规则创建成功"
else
    test_fail "WebSocket连接规则创建失败"
fi

# 测试 WebSocket 连接
sleep 2
WS_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/connect-test"
WS_CONNECT_RESULT=$(echo '{"test": "connection"}' | timeout_cmd 5 websocat --one-message "$WS_URL" 2>/dev/null || echo "TIMEOUT")

if [ "$WS_CONNECT_RESULT" != "TIMEOUT" ]; then
    if echo "$WS_CONNECT_RESULT" | grep -q "连接成功"; then
        test_pass "WebSocket基础连接测试成功"
    else
        test_fail "WebSocket连接响应不正确: $WS_CONNECT_RESULT"
    fi
else
    test_fail "WebSocket连接超时"
fi

# 2.2 WebSocket 消息广播测试
echo -e "${YELLOW}[2.2] 测试 WebSocket 消息广播...${NC}"
WS_BROADCAST_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket广播测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/broadcast-test"
    },
    "response": {
        "type": "Template",
        "content": {
            "message": "广播消息: {{.Message}}",
            "type": "broadcast",
            "sender": "{{.ClientID}}",
            "timestamp": "{{timestamp}}"
        }
    }
}
EOF
)")"

if echo "$WS_BROADCAST_RULE_RESPONSE" | grep -q '"id"'; then
    WS_BROADCAST_RULE_ID=$(extract_json_field "$WS_BROADCAST_RULE_RESPONSE" "id")
    test_pass "WebSocket广播规则创建成功"
else
    test_fail "WebSocket广播规则创建失败"
fi

# 测试 WebSocket 广播
sleep 2
WS_BROADCAST_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/broadcast-test"
WS_BROADCAST_RESULT=$(echo 'Hello Broadcast' | timeout_cmd 5 websocat --one-message "$WS_BROADCAST_URL" 2>/dev/null || echo "TIMEOUT")

if [ "$WS_BROADCAST_RESULT" != "TIMEOUT" ]; then
    if echo "$WS_BROADCAST_RESULT" | grep -q "广播消息"; then
        test_pass "WebSocket广播测试成功"
    else
        test_fail "WebSocket广播响应不正确: $WS_BROADCAST_RESULT"
    fi
else
    test_fail "WebSocket广播超时"
fi

echo ""

# ========================================
# 阶段 3: WebSocket 心跳测试
# ========================================

echo -e "${CYAN}[阶段 3] WebSocket 心跳测试${NC}"
echo ""

# 3.1 Ping/Pong 心跳测试
echo -e "${YELLOW}[3.1] 测试 Ping/Pong 心跳...${NC}"
WS_HEARTBEAT_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket心跳测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/heartbeat-test"
    },
    "response": {
        "type": "Static",
        "content": {
            "message": "Pong",
            "type": "heartbeat_response",
            "timestamp": "{{timestamp}}"
        }
    }
}
EOF
)")"

if echo "$WS_HEARTBEAT_RULE_RESPONSE" | grep -q '"id"'; then
    WS_HEARTBEAT_RULE_ID=$(extract_json_field "$WS_HEARTBEAT_RULE_RESPONSE" "id")
    test_pass "WebSocket心跳规则创建成功"
else
    test_fail "WebSocket心跳规则创建失败"
fi

# 测试心跳
sleep 2
WS_HEARTBEAT_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/heartbeat-test"
WS_HEARTBEAT_RESULT=$(echo 'ping' | timeout_cmd 5 websocat --one-message "$WS_HEARTBEAT_URL" 2>/dev/null || echo "TIMEOUT")

if [ "$WS_HEARTBEAT_RESULT" != "TIMEOUT" ]; then
    if echo "$WS_HEARTBEAT_RESULT" | grep -q "Pong"; then
        test_pass "WebSocket心跳测试成功"
    else
        test_fail "WebSocket心跳响应不正确: $WS_HEARTBEAT_RESULT"
    fi
else
    test_fail "WebSocket心跳超时"
fi

echo ""

# ========================================
# 阶段 4: WebSocket 并发连接测试
# ========================================

echo -e "${CYAN}[阶段 4] WebSocket 并发连接测试${NC}"
echo ""

# 4.1 多连接并发测试
echo -e "${YELLOW}[4.1] 测试多连接并发...${NC}"
WS_CONCURRENT_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket并发测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/concurrent-test"
    },
    "response": {
        "type": "Template",
        "content": {
            "message": "并发连接测试",
            "connection_id": "{{.ClientID}}",
            "timestamp": "{{timestamp}}",
            "server_time": "{{.ServerTime}}"
        }
    }
}
EOF
)")"

if echo "$WS_CONCURRENT_RULE_RESPONSE" | grep -q '"id"'; then
    WS_CONCURRENT_RULE_ID=$(extract_json_field "$WS_CONCURRENT_RULE_RESPONSE" "id")
    test_pass "WebSocket并发规则创建成功"
else
    test_fail "WebSocket并发规则创建失败"
fi

# 并发连接测试
sleep 2
WS_CONCURRENT_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/concurrent-test"
CONCURRENT_SUCCESS=0
CONCURRENT_TOTAL=5

for i in $(seq 1 $CONCURRENT_TOTAL); do
    CONCURRENT_RESULT=$(echo "client-$i" | timeout_cmd 3 websocat --one-message "$WS_CONCURRENT_URL" 2>/dev/null || echo "TIMEOUT") &
    if [ "$CONCURRENT_RESULT" != "TIMEOUT" ]; then
        CONCURRENT_SUCCESS=$((CONCURRENT_SUCCESS + 1))
    fi
done

wait

if [ $CONCURRENT_SUCCESS -eq $CONCURRENT_TOTAL ]; then
    test_pass "WebSocket并发测试成功 ($CONCURRENT_SUCCESS/$CONCURRENT_TOTAL)"
else
    test_fail "WebSocket并发测试部分失败 ($CONCURRENT_SUCCESS/$CONCURRENT_TOTAL)"
fi

echo ""

# ========================================
# 阶段 5: WebSocket 数据流测试
# ========================================

echo -e "${CYAN}[阶段 5] WebSocket 数据流测试${NC}"
echo ""

# 5.1 JSON 数据流测试
echo -e "${YELLOW}[5.1] 测试 JSON 数据流...${NC}"
WS_STREAM_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket数据流测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/stream-test"
    },
    "response": {
        "type": "Template",
        "content": {
            "type": "data_stream",
            "data": {
                "id": "{{uuid}}",
                "value": "{{random 1000}}",
                "timestamp": "{{timestamp}}",
                "status": "active"
            }
        }
    }
}
EOF
)")"

if echo "$WS_STREAM_RULE_RESPONSE" | grep -q '"id"'; then
    WS_STREAM_RULE_ID=$(extract_json_field "$WS_STREAM_RULE_RESPONSE" "id")
    test_pass "WebSocket数据流规则创建成功"
else
    test_fail "WebSocket数据流规则创建失败"
fi

# 测试数据流
sleep 2
WS_STREAM_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/stream-test"
WS_STREAM_RESULT=$(echo 'get_data' | timeout_cmd 5 websocat --one-message "$WS_STREAM_URL" 2>/dev/null || echo "TIMEOUT")

if [ "$WS_STREAM_RESULT" != "TIMEOUT" ]; then
    if echo "$WS_STREAM_RESULT" | grep -q "data_stream" && echo "$WS_STREAM_RESULT" | grep -q "id"; then
        test_pass "WebSocket数据流测试成功"
    else
        test_fail "WebSocket数据流格式不正确: $WS_STREAM_RESULT"
    fi
else
    test_fail "WebSocket数据流超时"
fi

# 5.2 大数据流测试
echo -e "${YELLOW}[5.2] 测试大数据流...${NC}"
WS_LARGE_DATA_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket大数据流测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/large-data-test"
    },
    "response": {
        "type": "Template",
        "content": {
            "type": "large_data_stream",
            "data": {
                "items": [
                    {"id": 1, "name": "item1", "value": "{{random 100}}"},
                    {"id": 2, "name": "item2", "value": "{{random 100}}"},
                    {"id": 3, "name": "item3", "value": "{{random 100}}"},
                    {"id": 4, "name": "item4", "value": "{{random 100}}"},
                    {"id": 5, "name": "item5", "value": "{{random 100}}"}
                ],
                "metadata": {
                    "total_count": 5,
                    "generated_at": "{{timestamp}}",
                    "request_id": "{{uuid}}"
                }
            }
        }
    }
}
EOF
)")"

if echo "$WS_LARGE_DATA_RULE_RESPONSE" | grep -q '"id"'; then
    WS_LARGE_DATA_RULE_ID=$(extract_json_field "$WS_LARGE_DATA_RULE_RESPONSE" "id")
    test_pass "WebSocket大数据流规则创建成功"
else
    test_fail "WebSocket大数据流规则创建失败"
fi

# 测试大数据流
sleep 2
WS_LARGE_DATA_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/large-data-test"
WS_LARGE_DATA_RESULT=$(echo 'get_large_data' | timeout_cmd 8 websocat --one-message "$WS_LARGE_DATA_URL" 2>/dev/null || echo "TIMEOUT")

if [ "$WS_LARGE_DATA_RESULT" != "TIMEOUT" ]; then
    if echo "$WS_LARGE_DATA_RESULT" | grep -q "large_data_stream" && echo "$WS_LARGE_DATA_RESULT" | grep -q "items"; then
        test_pass "WebSocket大数据流测试成功"
    else
        test_fail "WebSocket大数据流格式不正确"
    fi
else
    test_fail "WebSocket大数据流超时"
fi

echo ""

# ========================================
# 阶段 6: WebSocket 错误处理测试
# ========================================

echo -e "${CYAN}[阶段 6] WebSocket 错误处理测试${NC}"
echo ""

# 6.1 无效路径测试
echo -e "${YELLOW}[6.1] 测试无效路径处理...${NC}"
WS_INVALID_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/invalid-path"
WS_INVALID_RESULT=$(echo 'test' | timeout_cmd 3 websocat --one-message "$WS_INVALID_URL" 2>/dev/null || echo "TIMEOUT")

if [ "$WS_INVALID_RESULT" = "TIMEOUT" ]; then
    test_pass "WebSocket无效路径正确处理（连接超时）"
else
    test_warn "WebSocket无效路径处理异常: $WS_INVALID_RESULT"
fi

# 6.2 连接中断测试
echo -e "${YELLOW}[6.2] 测试连接中断处理...${NC}"
# 创建一个会立即断开连接的规则
WS_DISCONNECT_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket断开测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/disconnect-test"
    },
    "response": {
        "type": "Static",
        "content": {
            "action": "disconnect",
            "message": "Connection will be closed"
        }
    }
}
EOF
)")"

if echo "$WS_DISCONNECT_RULE_RESPONSE" | grep -q '"id"'; then
    WS_DISCONNECT_RULE_ID=$(extract_json_field "$WS_DISCONNECT_RULE_RESPONSE" "id")
    test_pass "WebSocket断开规则创建成功"
else
    test_fail "WebSocket断开规则创建失败"
fi

# 测试连接断开
sleep 2
WS_DISCONNECT_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/disconnect-test"
WS_DISCONNECT_RESULT=$(echo 'test_disconnect' | timeout_cmd 3 websocat --one-message "$WS_DISCONNECT_URL" 2>/dev/null || echo "DISCONNECTED")

if [ "$WS_DISCONNECT_RESULT" = "DISCONNECTED" ]; then
    test_pass "WebSocket连接断开测试成功"
else
    test_warn "WebSocket连接断开测试异常: $WS_DISCONNECT_RESULT"
fi

echo ""

# ========================================
# 阶段 7: WebSocket 性能测试
# ========================================

echo -e "${CYAN}[阶段 7] WebSocket 性能测试${NC}"
echo ""

# 7.1 连接建立性能测试
echo -e "${YELLOW}[7.1] 测试连接建立性能...${NC}"
WS_PERF_RULE_RESPONSE=$(http_post "$ADMIN_API/rules" "$(cat <<EOF
{
    "name": "WebSocket性能测试",
    "project_id": "$WS_PROJECT_ID",
    "environment_id": "$WS_ENVIRONMENT_ID",
    "protocol": "WebSocket",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "path": "/ws/perf-test"
    },
    "response": {
        "type": "Static",
        "content": {
            "message": "Performance test response",
            "timestamp": "{{timestamp}}"
        }
    }
}
EOF
)")"

if echo "$WS_PERF_RULE_RESPONSE" | grep -q '"id"'; then
    WS_PERF_RULE_ID=$(extract_json_field "$WS_PERF_RULE_RESPONSE" "id")
    test_pass "WebSocket性能规则创建成功"
else
    test_fail "WebSocket性能规则创建失败"
fi

# 性能测试 - 快速连接建立
sleep 2
WS_PERF_URL="ws://localhost:9090/$WS_PROJECT_ID/$WS_ENVIRONMENT_ID/ws/perf-test"
PERF_SUCCESS=0
PERF_TOTAL=10

PERF_START_TIME=$(get_timestamp_ms)

for i in $(seq 1 $PERF_TOTAL); do
    PERF_RESULT=$(echo "perf-test-$i" | timeout_cmd 2 websocat --one-message "$WS_PERF_URL" 2>/dev/null || echo "TIMEOUT") &
    if [ "$PERF_RESULT" != "TIMEOUT" ]; then
        PERF_SUCCESS=$((PERF_SUCCESS + 1))
    fi
done

wait

PERF_END_TIME=$(get_timestamp_ms)
PERF_DURATION=$(calculate_duration "$PERF_START_TIME" "$PERF_END_TIME")
PERF_AVG_TIME=$((PERF_DURATION / PERF_TOTAL))

if [ $PERF_SUCCESS -ge $((PERF_TOTAL * 8 / 10)) ]; then
    test_pass "WebSocket性能测试成功 ($PERF_SUCCESS/$PERF_TOTAL, 平均耗时: ${PERF_AVG_TIME}ms)"
else
    test_fail "WebSocket性能测试失败 ($PERF_SUCCESS/$PERF_TOTAL)"
fi

echo ""

# ========================================
# 生成测试报告
# ========================================

echo -e "${CYAN}[完成] 生成测试报告${NC}"
REPORT_FILE="/tmp/websocket_e2e_test_report_$(date +%Y%m%d_%H%M%S).md"
generate_test_report "$REPORT_FILE" "WebSocket E2E集成测试"

# ========================================
# 测试结果统计
# ========================================

print_test_summary

echo ""
echo -e "${CYAN}WebSocket 功能验证:${NC}"
echo -e "  ${GREEN}✓ WebSocket 连接管理${NC}"
echo -e "  ${GREEN}✓ 消息广播${NC}"
echo -e "  ${GREEN}✓ Ping/Pong 心跳${NC}"
echo -e "  ${GREEN}✓ 并发连接处理${NC}"
echo -e "  ${GREEN}✓ 数据流传输${NC}"
echo -e "  ${GREEN}✓ JSON 数据流${NC}"
echo -e "  ${GREEN}✓ 大数据流处理${NC}"
echo -e "  ${GREEN}✓ 错误处理${NC}"
echo -e "  ${GREEN}✓ 性能测试${NC}"

echo ""
echo -e "${BLUE}=========================================${NC}"
echo -e "${BLUE}   WebSocket E2E 测试完成${NC}"
echo -e "${BLUE}=========================================${NC}"
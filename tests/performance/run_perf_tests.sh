#!/bin/bash

# Mock Server 性能测试脚本
# 使用 wrk 进行压力测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
MOCK_API="${MOCK_API:-http://localhost:9090}"
PROJECT_ID=""
ENVIRONMENT_ID=""
RULE_ID=""
RESULTS_DIR="$(pwd)/tests/performance/results"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Mock Server 性能测试${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 创建结果目录
mkdir -p "$RESULTS_DIR"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULT_FILE="$RESULTS_DIR/perf_test_$TIMESTAMP.txt"

# 检查wrk是否安装
check_wrk() {
    if ! command -v wrk &> /dev/null; then
        echo -e "${RED}错误: wrk 未安装${NC}"
        echo "安装方法:"
        echo "  macOS: brew install wrk"
        echo "  Ubuntu: sudo apt-get install wrk"
        exit 1
    fi
    echo -e "${GREEN}✓ wrk 已安装${NC}"
}

# 准备测试数据
prepare_test_data() {
    echo -e "${YELLOW}准备测试数据...${NC}"
    
    # 创建项目
    PROJECT_RESPONSE=$(curl -s -X POST "$ADMIN_API/projects" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "性能测试项目",
            "workspace_id": "perf-test"
        }')
    
    PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "  项目ID: $PROJECT_ID"
    
    # 创建环境
    ENV_RESPONSE=$(curl -s -X POST "$ADMIN_API/projects/$PROJECT_ID/environments" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "性能测试环境"
        }')
    
    ENVIRONMENT_ID=$(echo "$ENV_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "  环境ID: $ENVIRONMENT_ID"
    
    # 创建简单规则
    RULE_RESPONSE=$(curl -s -X POST "$ADMIN_API/rules" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "性能测试规则",
            "project_id": "'"$PROJECT_ID"'",
            "environment_id": "'"$ENVIRONMENT_ID"'",
            "protocol": "HTTP",
            "match_type": "Simple",
            "priority": 100,
            "enabled": true,
            "match_condition": {
                "method": "GET",
                "path": "/perf/test"
            },
            "response": {
                "type": "Static",
                "content": {
                    "status_code": 200,
                    "content_type": "JSON",
                    "body": {
                        "message": "performance test",
                        "timestamp": 1234567890
                    }
                }
            }
        }')
    
    RULE_ID=$(echo "$RULE_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "  规则ID: $RULE_ID"
    
    # 等待规则生效
    sleep 2
    echo -e "${GREEN}✓ 测试数据准备完成${NC}"
}

# 清理测试数据
cleanup() {
    if [ ! -z "$PROJECT_ID" ]; then
        echo -e "${YELLOW}清理测试数据...${NC}"
        curl -s -X DELETE "$ADMIN_API/projects/$PROJECT_ID" > /dev/null
        echo -e "${GREEN}✓ 清理完成${NC}"
    fi
}

trap cleanup EXIT

# 测试1: 轻负载测试
test_light_load() {
    echo -e "${YELLOW}[测试1] 轻负载测试 (10并发, 30秒)${NC}"
    wrk -t4 -c10 -d30s \
        -H "X-Project-ID: $PROJECT_ID" \
        -H "X-Environment-ID: $ENVIRONMENT_ID" \
        "$MOCK_API/perf/test" \
        | tee -a "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"
}

# 测试2: 中负载测试
test_medium_load() {
    echo -e "${YELLOW}[测试2] 中负载测试 (100并发, 60秒)${NC}"
    wrk -t8 -c100 -d60s \
        -H "X-Project-ID: $PROJECT_ID" \
        -H "X-Environment-ID: $ENVIRONMENT_ID" \
        "$MOCK_API/perf/test" \
        | tee -a "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"
}

# 测试3: 高负载测试
test_high_load() {
    echo -e "${YELLOW}[测试3] 高负载测试 (500并发, 120秒)${NC}"
    wrk -t12 -c500 -d120s \
        -H "X-Project-ID: $PROJECT_ID" \
        -H "X-Environment-ID: $ENVIRONMENT_ID" \
        "$MOCK_API/perf/test" \
        | tee -a "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"
}

# 测试4: 压力测试
test_stress() {
    echo -e "${YELLOW}[测试4] 压力测试 (1000并发, 60秒)${NC}"
    wrk -t16 -c1000 -d60s \
        -H "X-Project-ID: $PROJECT_ID" \
        -H "X-Environment-ID: $ENVIRONMENT_ID" \
        "$MOCK_API/perf/test" \
        | tee -a "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"
}

# 主函数
main() {
    check_wrk
    
    echo "测试配置:"
    echo "  Admin API: $ADMIN_API"
    echo "  Mock API: $MOCK_API"
    echo "  结果文件: $RESULT_FILE"
    echo ""
    
    prepare_test_data
    
    echo -e "${BLUE}开始性能测试...${NC}"
    echo "" >> "$RESULT_FILE"
    echo "========================================" >> "$RESULT_FILE"
    echo "Mock Server 性能测试报告" >> "$RESULT_FILE"
    echo "测试时间: $(date)" >> "$RESULT_FILE"
    echo "========================================" >> "$RESULT_FILE"
    echo "" >> "$RESULT_FILE"
    
    test_light_load
    sleep 5
    
    test_medium_load
    sleep 5
    
    test_high_load
    sleep 5
    
    test_stress
    
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}   性能测试完成${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo "结果已保存到: $RESULT_FILE"
}

main "$@"

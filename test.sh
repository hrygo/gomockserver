#!/bin/bash

# Mock Server 快速测试脚本

API_BASE="http://localhost:8080/api/v1"
MOCK_BASE="http://localhost:9090"

echo "========================================="
echo "Mock Server 快速测试"
echo "========================================="
echo ""

# 1. 检查服务健康状态
echo "1. 检查服务健康状态..."
curl -s $API_BASE/system/health | jq .
echo ""

# 2. 创建项目
echo "2. 创建测试项目..."
PROJECT_RESPONSE=$(curl -s -X POST $API_BASE/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "示例项目",
    "workspace_id": "default",
    "description": "这是一个示例项目"
  }')
echo $PROJECT_RESPONSE | jq .
PROJECT_ID=$(echo $PROJECT_RESPONSE | jq -r '.id')
echo "项目ID: $PROJECT_ID"
echo ""

# 3. 创建环境
echo "3. 创建开发环境..."
ENV_RESPONSE=$(curl -s -X POST $API_BASE/environments \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"开发环境\",
    \"project_id\": \"$PROJECT_ID\",
    \"base_url\": \"http://localhost:9090\"
  }")
echo $ENV_RESPONSE | jq .
ENV_ID=$(echo $ENV_RESPONSE | jq -r '.id')
echo "环境ID: $ENV_ID"
echo ""

# 4. 创建 Mock 规则 - 用户列表
echo "4. 创建Mock规则 - 用户列表..."
RULE_RESPONSE=$(curl -s -X POST $API_BASE/rules \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"获取用户列表\",
    \"project_id\": \"$PROJECT_ID\",
    \"environment_id\": \"$ENV_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
      \"method\": \"GET\",
      \"path\": \"/api/users\"
    },
    \"response\": {
      \"type\": \"Static\",
      \"content\": {
        \"status_code\": 200,
        \"content_type\": \"JSON\",
        \"headers\": {
          \"Content-Type\": \"application/json\"
        },
        \"body\": {
          \"code\": 0,
          \"message\": \"success\",
          \"data\": [
            {
              \"id\": 1,
              \"name\": \"张三\",
              \"email\": \"zhangsan@example.com\",
              \"age\": 25
            },
            {
              \"id\": 2,
              \"name\": \"李四\",
              \"email\": \"lisi@example.com\",
              \"age\": 30
            }
          ]
        }
      }
    }
  }")
echo $RULE_RESPONSE | jq .
RULE_ID=$(echo $RULE_RESPONSE | jq -r '.id')
echo "规则ID: $RULE_ID"
echo ""

# 5. 创建 Mock 规则 - 获取单个用户
echo "5. 创建Mock规则 - 获取单个用户..."
curl -s -X POST $API_BASE/rules \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"获取用户详情\",
    \"project_id\": \"$PROJECT_ID\",
    \"environment_id\": \"$ENV_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
      \"method\": \"GET\",
      \"path\": \"/api/users/:id\"
    },
    \"response\": {
      \"type\": \"Static\",
      \"content\": {
        \"status_code\": 200,
        \"content_type\": \"JSON\",
        \"headers\": {
          \"Content-Type\": \"application/json\"
        },
        \"body\": {
          \"code\": 0,
          \"message\": \"success\",
          \"data\": {
            \"id\": 1,
            \"name\": \"张三\",
            \"email\": \"zhangsan@example.com\",
            \"age\": 25,
            \"phone\": \"13800138000\"
          }
        }
      }
    }
  }" | jq .
echo ""

# 6. 创建 Mock 规则 - POST 请求
echo "6. 创建Mock规则 - 创建用户..."
curl -s -X POST $API_BASE/rules \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"创建用户\",
    \"project_id\": \"$PROJECT_ID\",
    \"environment_id\": \"$ENV_ID\",
    \"protocol\": \"HTTP\",
    \"match_type\": \"Simple\",
    \"priority\": 100,
    \"enabled\": true,
    \"match_condition\": {
      \"method\": \"POST\",
      \"path\": \"/api/users\"
    },
    \"response\": {
      \"type\": \"Static\",
      \"delay\": {
        \"type\": \"fixed\",
        \"fixed\": 100
      },
      \"content\": {
        \"status_code\": 201,
        \"content_type\": \"JSON\",
        \"headers\": {
          \"Content-Type\": \"application/json\"
        },
        \"body\": {
          \"code\": 0,
          \"message\": \"创建成功\",
          \"data\": {
            \"id\": 3,
            \"name\": \"新用户\",
            \"email\": \"newuser@example.com\"
          }
        }
      }
    }
  }" | jq .
echo ""

# 7. 测试 Mock 接口
echo "7. 测试Mock接口..."
echo ""

echo "7.1 测试获取用户列表:"
curl -s "$MOCK_BASE/$PROJECT_ID/$ENV_ID/api/users" | jq .
echo ""

echo "7.2 测试获取用户详情:"
curl -s "$MOCK_BASE/$PROJECT_ID/$ENV_ID/api/users/1" | jq .
echo ""

echo "7.3 测试创建用户:"
curl -s -X POST "$MOCK_BASE/$PROJECT_ID/$ENV_ID/api/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"测试用户","email":"test@example.com"}' | jq .
echo ""

# 8. 查询规则列表
echo "8. 查询规则列表..."
curl -s "$API_BASE/rules?project_id=$PROJECT_ID&environment_id=$ENV_ID" | jq .
echo ""

echo "========================================="
echo "测试完成！"
echo "========================================="
echo ""
echo "项目ID: $PROJECT_ID"
echo "环境ID: $ENV_ID"
echo ""
echo "你可以使用以下命令继续测试："
echo "  查看规则: curl $API_BASE/rules?project_id=$PROJECT_ID"
echo "  禁用规则: curl -X POST $API_BASE/rules/$RULE_ID/disable"
echo "  启用规则: curl -X POST $API_BASE/rules/$RULE_ID/enable"
echo ""

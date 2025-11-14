#!/bin/bash

# API 客户端函数库
# 提供统一的 API 调用接口

# ==================== 配置 ====================
ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
MOCK_API="${MOCK_API:-http://localhost:9090}"
API_TIMEOUT="${API_TIMEOUT:-10}"

# ==================== HTTP 请求函数 ====================

# GET 请求
api_get() {
    local url="$1"
    local expected_status="${2:-200}"
    
    log "API GET: $url"
    
    local response=$(curl -s -w "\n%{http_code}" --max-time "$API_TIMEOUT" "$url")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    log "Response Code: $http_code"
    log "Response Body: $body"
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "$body"
        return 0
    else
        error "期望状态码 $expected_status，实际 $http_code"
        echo "$body"
        return 1
    fi
}

# POST 请求
api_post() {
    local url="$1"
    local data="$2"
    local expected_status="${3:-200}"
    
    log "API POST: $url"
    log "Request Data: $data"
    
    local response=$(curl -s -w "\n%{http_code}" --max-time "$API_TIMEOUT" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$url")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    log "Response Code: $http_code"
    log "Response Body: $body"
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "$body"
        return 0
    else
        error "期望状态码 $expected_status，实际 $http_code"
        echo "$body"
        return 1
    fi
}

# PUT 请求
api_put() {
    local url="$1"
    local data="$2"
    local expected_status="${3:-200}"
    
    log "API PUT: $url"
    log "Request Data: $data"
    
    local response=$(curl -s -w "\n%{http_code}" --max-time "$API_TIMEOUT" \
        -X PUT \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$url")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    log "Response Code: $http_code"
    log "Response Body: $body"
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "$body"
        return 0
    else
        error "期望状态码 $expected_status，实际 $http_code"
        echo "$body"
        return 1
    fi
}

# DELETE 请求
api_delete() {
    local url="$1"
    local expected_status="${2:-200}"
    
    log "API DELETE: $url"
    
    local response=$(curl -s -w "\n%{http_code}" --max-time "$API_TIMEOUT" \
        -X DELETE \
        "$url")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    log "Response Code: $http_code"
    log "Response Body: $body"
    
    if [ "$http_code" = "$expected_status" ]; then
        echo "$body"
        return 0
    else
        error "期望状态码 $expected_status，实际 $http_code"
        echo "$body"
        return 1
    fi
}

# ==================== 系统管理 API ====================

# 健康检查
api_health_check() {
    api_get "$ADMIN_API/system/health" 200
}

# 获取版本信息
api_get_version() {
    api_get "$ADMIN_API/system/version" 200
}

# ==================== 项目管理 API ====================

# 创建项目
api_create_project() {
    local name="$1"
    local workspace_id="$2"
    local description="${3:-}"
    
    local data="{\"name\":\"$name\",\"workspace_id\":\"$workspace_id\""
    if [ -n "$description" ]; then
        data="$data,\"description\":\"$description\""
    fi
    data="$data}"
    
    api_post "$ADMIN_API/projects" "$data" 200
}

# 获取项目详情
api_get_project() {
    local project_id="$1"
    api_get "$ADMIN_API/projects/$project_id" 200
}

# 更新项目
api_update_project() {
    local project_id="$1"
    local name="$2"
    local description="${3:-}"
    
    local data="{\"name\":\"$name\""
    if [ -n "$description" ]; then
        data="$data,\"description\":\"$description\""
    fi
    data="$data}"
    
    api_put "$ADMIN_API/projects/$project_id" "$data" 200
}

# 删除项目
api_delete_project() {
    local project_id="$1"
    api_delete "$ADMIN_API/projects/$project_id" 200
}

# 列出所有项目
api_list_projects() {
    api_get "$ADMIN_API/projects" 200
}

# ==================== 环境管理 API ====================

# 创建环境
api_create_environment() {
    local name="$1"
    local project_id="$2"
    local base_url="${3:-}"
    
    local data="{\"name\":\"$name\",\"project_id\":\"$project_id\""
    if [ -n "$base_url" ]; then
        data="$data,\"base_url\":\"$base_url\""
    fi
    data="$data}"
    
    api_post "$ADMIN_API/environments" "$data" 200
}

# 获取环境详情
api_get_environment() {
    local env_id="$1"
    api_get "$ADMIN_API/environments/$env_id" 200
}

# 更新环境
api_update_environment() {
    local env_id="$1"
    local name="$2"
    local base_url="${3:-}"
    
    local data="{\"name\":\"$name\""
    if [ -n "$base_url" ]; then
        data="$data,\"base_url\":\"$base_url\""
    fi
    data="$data}"
    
    api_put "$ADMIN_API/environments/$env_id" "$data" 200
}

# 删除环境
api_delete_environment() {
    local env_id="$1"
    api_delete "$ADMIN_API/environments/$env_id" 200
}

# 列出项目的所有环境
api_list_environments() {
    local project_id="$1"
    api_get "$ADMIN_API/environments?project_id=$project_id" 200
}

# ==================== 规则管理 API ====================

# 创建规则（简化版，支持基本HTTP规则）
api_create_rule() {
    local name="$1"
    local project_id="$2"
    local env_id="$3"
    local method="$4"
    local path="$5"
    local status_code="${6:-200}"
    local response_body="${7:-{}}"
    local priority="${8:-100}"
    
    local data="{
        \"name\":\"$name\",
        \"project_id\":\"$project_id\",
        \"environment_id\":\"$env_id\",
        \"protocol\":\"HTTP\",
        \"match_type\":\"Simple\",
        \"priority\":$priority,
        \"enabled\":true,
        \"match_condition\":{
            \"method\":\"$method\",
            \"path\":\"$path\"
        },
        \"response\":{
            \"type\":\"Static\",
            \"content\":{
                \"status_code\":$status_code,
                \"content_type\":\"JSON\",
                \"body\":$response_body
            }
        }
    }"
    
    api_post "$ADMIN_API/rules" "$data" 200
}

# 创建带延迟的规则
api_create_rule_with_delay() {
    local name="$1"
    local project_id="$2"
    local env_id="$3"
    local method="$4"
    local path="$5"
    local delay_ms="$6"
    local status_code="${7:-200}"
    local response_body="${8:-{}}"
    
    local data="{
        \"name\":\"$name\",
        \"project_id\":\"$project_id\",
        \"environment_id\":\"$env_id\",
        \"protocol\":\"HTTP\",
        \"match_type\":\"Simple\",
        \"priority\":100,
        \"enabled\":true,
        \"match_condition\":{
            \"method\":\"$method\",
            \"path\":\"$path\"
        },
        \"response\":{
            \"type\":\"Static\",
            \"delay\":{
                \"type\":\"Fixed\",
                \"value\":$delay_ms
            },
            \"content\":{
                \"status_code\":$status_code,
                \"content_type\":\"JSON\",
                \"body\":$response_body
            }
        }
    }"
    
    api_post "$ADMIN_API/rules" "$data" 200
}

# 获取规则详情
api_get_rule() {
    local rule_id="$1"
    api_get "$ADMIN_API/rules/$rule_id" 200
}

# 更新规则
api_update_rule() {
    local rule_id="$1"
    local name="$2"
    local enabled="${3:-true}"
    
    local data="{\"name\":\"$name\",\"enabled\":$enabled}"
    
    api_put "$ADMIN_API/rules/$rule_id" "$data" 200
}

# 删除规则
api_delete_rule() {
    local rule_id="$1"
    api_delete "$ADMIN_API/rules/$rule_id" 200
}

# 列出规则
api_list_rules() {
    local project_id="$1"
    local env_id="$2"
    api_get "$ADMIN_API/rules?project_id=$project_id&environment_id=$env_id" 200
}

# 启用规则
api_enable_rule() {
    local rule_id="$1"
    api_post "$ADMIN_API/rules/$rule_id/enable" "{}" 200
}

# 禁用规则
api_disable_rule() {
    local rule_id="$1"
    api_post "$ADMIN_API/rules/$rule_id/disable" "{}" 200
}

# ==================== Mock 服务 API ====================

# 发送 Mock 请求
mock_request() {
    local project_id="$1"
    local env_id="$2"
    local path="$3"
    local method="${4:-GET}"
    local data="${5:-}"
    local expected_status="${6:-200}"
    
    local url="$MOCK_API/$project_id/$env_id$path"
    
    log "MOCK REQUEST: $method $url"
    
    if [ "$method" = "GET" ]; then
        api_get "$url" "$expected_status"
    elif [ "$method" = "POST" ]; then
        api_post "$url" "$data" "$expected_status"
    elif [ "$method" = "PUT" ]; then
        api_put "$url" "$data" "$expected_status"
    elif [ "$method" = "DELETE" ]; then
        api_delete "$url" "$expected_status"
    else
        error "不支持的HTTP方法: $method"
        return 1
    fi
}

# 测量Mock响应时间
mock_request_with_timing() {
    local project_id="$1"
    local env_id="$2"
    local path="$3"
    
    local url="$MOCK_API/$project_id/$env_id$path"
    local start_time=$(date +%s%3N)
    
    curl -s "$url" > /dev/null
    
    local end_time=$(date +%s%3N)
    local duration=$((end_time - start_time))
    
    echo "$duration"
}

# ==================== 导出函数 ====================

export -f api_get api_post api_put api_delete
export -f api_health_check api_get_version
export -f api_create_project api_get_project api_update_project api_delete_project api_list_projects
export -f api_create_environment api_get_environment api_update_environment api_delete_environment api_list_environments
export -f api_create_rule api_create_rule_with_delay api_get_rule api_update_rule api_delete_rule api_list_rules
export -f api_enable_rule api_disable_rule
export -f mock_request mock_request_with_timing

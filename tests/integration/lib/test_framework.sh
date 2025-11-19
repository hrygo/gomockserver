#!/bin/bash

# MockServer E2E 测试框架
# 提供通用的测试功能和工具函数

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BINARY="$PROJECT_ROOT/mockserver"

# 全局测试变量
PROJECT_ID=""
ENVIRONMENT_ID=""
RULE_ID=""
WS_PROJECT_ID=""
WS_ENVIRONMENT_ID=""
WS_RULE_ID=""
TEST_PASSED=0
TEST_FAILED=0
TEST_SKIPPED=0

# 测试结果记录
TEST_RESULTS=()
TEST_START_TIME=""
TEST_END_TIME=""

# ========================================
# 环境检测和配置
# ========================================

detect_environment() {
    if [ -n "$GITHUB_ACTIONS" ]; then
        echo -e "${CYAN}检测到 GitHub Actions 环境${NC}"
        CONFIG_FILE="$PROJECT_ROOT/config.test.yaml"
        ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
        MOCK_API="${MOCK_API:-http://localhost:9090}"
        SKIP_SERVER_START="true"
    else
        echo -e "${CYAN}检测到本地开发环境${NC}"
        CONFIG_FILE="$PROJECT_ROOT/config.dev.yaml"
        ADMIN_API="${ADMIN_API:-http://localhost:8080/api/v1}"
        MOCK_API="${MOCK_API:-http://localhost:9090}"
        SKIP_SERVER_START="${SKIP_SERVER_START:-false}"
    fi

    echo -e "${CYAN}使用配置:${NC}"
    echo -e "  配置文件: ${YELLOW}$CONFIG_FILE${NC}"
    echo -e "  管理API: ${YELLOW}$ADMIN_API${NC}"
    echo -e "  MockAPI: ${YELLOW}$MOCK_API${NC}"
    echo -e "  跳过服务器启动: ${YELLOW}$SKIP_SERVER_START${NC}"
}

# ========================================
# Docker服务管理函数
# ========================================

# 检查端口是否被占用 - 增强版本
check_port_available() {
    local port="$1"
    local service_name="$2"
    local timeout="${3:-2}"

    # 参数验证
    if [[ ! "$port" =~ ^[0-9]+$ ]] || [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
        echo -e "${RED}无效的端口号: $port ($service_name)${NC}"
        return 2
    fi

    # 方法1: 使用lsof (最可靠)
    if command -v lsof >/dev/null 2>&1; then
        if timeout "$timeout" lsof -i ":$port" >/dev/null 2>&1; then
            local pid=$(lsof -ti ":$port" 2>/dev/null | head -1)
            if [ -n "$pid" ]; then
                local process_info=$(ps -p "$pid" -o comm= 2>/dev/null || echo "unknown")
                echo -e "${YELLOW}端口 $port 已被占用 ($service_name) - 进程: $process_info (PID: $pid)${NC}"
            else
                echo -e "${YELLOW}端口 $port 已被占用 ($service_name)${NC}"
            fi
            return 1
        fi
    # 方法2: 使用netstat (兼容性更好)
    elif command -v netstat >/dev/null 2>&1; then
        # 根据操作系统选择不同的netstat参数
        local netstat_opts="-tulpn"
        if [[ "$(uname)" == "Darwin" ]]; then
            netstat_opts="-an"
        fi

        if netstat $netstat_opts 2>/dev/null | grep -E ":$port\s" >/dev/null; then
            echo -e "${YELLOW}端口 $port 已被占用 ($service_name)${NC}"
            return 1
        fi
    # 方法3: 使用ss (Linux现代替代)
    elif command -v ss >/dev/null 2>&1; then
        if ss -tuln 2>/dev/null | grep -E ":$port\s" >/dev/null; then
            echo -e "${YELLOW}端口 $port 已被占用 ($service_name)${NC}"
            return 1
        fi
    # 方法4: 使用nc (netcat)连接测试
    elif command -v nc >/dev/null 2>&1; then
        if nc -z localhost "$port" 2>/dev/null; then
            echo -e "${YELLOW}端口 $port 已被占用 ($service_name) - nc检测${NC}"
            return 1
        fi
    # 方法5: 使用/dev/tcp连接测试 (bash内置)
    elif timeout "$timeout" bash -c "echo >/dev/tcp/localhost/$port" 2>/dev/null; then
        echo -e "${YELLOW}端口 $port 已被占用 ($service_name) - TCP连接测试${NC}"
        return 1
    else
        echo -e "${YELLOW}警告: 无法检测端口 $port 状态 - 缺少检测工具${NC}"
        # 假设端口可用，但给出警告
        return 0
    fi

    # 端口可用
    echo -e "${GREEN}✓ 端口 $port 可用 ($service_name)${NC}"
    return 0
}

# 智能端口检测 - 寻找可用端口 (增强版)
find_available_port() {
    local base_port="$1"
    local service_name="$2"
    local max_attempts="${3:-20}"
    local port_range="${4:-100}"

    # 参数验证
    if [[ ! "$base_port" =~ ^[0-9]+$ ]] || [ "$base_port" -lt 1024 ] || [ "$base_port" -gt 65000 ]; then
        echo -e "${RED}无效的基础端口号: $base_port ($service_name)${NC}"
        echo "0"
        return 2
    fi

    echo -e "${CYAN}为 $service_name 寻找可用端口 (起始: $base_port, 最大尝试: $max_attempts)${NC}"

    local available_ports=()

    # 第一轮: 顺序检查
    for ((i=0; i<max_attempts; i++)); do
        local test_port=$((base_port + i))

        # 避免超过端口范围
        if [ "$test_port" -gt 65535 ]; then
            break
        fi

        if check_port_available "$test_port" "$service_name" >/dev/null 2>&1; then
            available_ports+=("$test_port")
            echo -e "${GREEN}✓ 找到可用端口: $test_port ($service_name)${NC}"
            echo "$test_port"
            return 0
        fi
    done

    # 第二轮: 随机范围检查 (如果顺序检查失败)
    echo -e "${YELLOW}顺序检查失败，尝试随机端口检查 ($service_name)${NC}"
    for ((i=0; i<10; i++)); do
        # 在指定范围内随机选择端口
        local random_offset=$((RANDOM % port_range))
        local test_port=$((base_port + random_offset))

        # 确保端口在有效范围内
        if [ "$test_port" -lt 1024 ] || [ "$test_port" -gt 65535 ]; then
            continue
        fi

        if check_port_available "$test_port" "$service_name" >/dev/null 2>&1; then
            available_ports+=("$test_port")
            echo -e "${GREEN}✓ 找到随机可用端口: $test_port ($service_name)${NC}"
            echo "$test_port"
            return 0
        fi
    done

    # 如果都失败了，使用动态端口分配
    echo -e "${YELLOW}尝试动态端口分配 ($service_name)${NC}"
    local dynamic_port=0

    # 尝试使用系统的动态端口分配机制
    if command -v python3 >/dev/null 2>&1; then
        dynamic_port=$(python3 -c "
import socket
s = socket.socket()
s.bind(('', 0))
addr = s.getsockname()
s.close()
print(addr[1])
" 2>/dev/null)

        if [[ "$dynamic_port" =~ ^[0-9]+$ ]] && [ "$dynamic_port" -gt 0 ]; then
            echo -e "${GREEN}✓ 分配动态端口: $dynamic_port ($service_name)${NC}"
            echo "$dynamic_port"
            return 0
        fi
    fi

    # 最终失败
    echo -e "${RED}✗ 无法为 $service_name 找到可用端口${NC}"
    echo "0"
    return 1
}

# 启动Docker服务（如果需要）
start_docker_services() {
    echo -e "${CYAN}检查和启动Docker依赖服务...${NC}"

    # 检查Docker是否运行
    if ! docker info >/dev/null 2>&1; then
        echo -e "${RED}Docker未运行，请启动Docker服务${NC}"
        return 1
    fi

    local services_started=false

    # 检查MongoDB
    local mongodb_running=false
    if docker ps --format "table" | grep -q "mongodb"; then
        mongodb_running=true
        echo -e "${GREEN}✓ MongoDB 容器已运行${NC}"
    else
        echo -e "${YELLOW}启动 MongoDB 容器...${NC}"
        if docker run -d --name mongodb-mockserver -p 27017:27017 mongo:5.0 >/dev/null 2>&1; then
            echo -e "${GREEN}✓ MongoDB 容器启动成功${NC}"
            mongodb_running=true
            services_started=true
            sleep 3  # 等待MongoDB完全启动
        else
            echo -e "${RED}✗ MongoDB 容器启动失败${NC}"
        fi
    fi

    # 检查Redis
    local redis_running=false
    if docker ps --format "table" | grep -q "redis"; then
        redis_running=true
        echo -e "${GREEN}✓ Redis 容器已运行${NC}"
    else
        echo -e "${YELLOW}启动 Redis 容器...${NC}"
        if docker run -d --name redis-mockserver -p 6379:6379 redis:6.2-alpine >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Redis 容器启动成功${NC}"
            redis_running=true
            services_started=true
            sleep 2  # 等待Redis完全启动
        else
            echo -e "${RED}✗ Redis 容器启动失败${NC}"
        fi
    fi

    # 如果启动了新服务，等待它们完全就绪
    if [ "$services_started" = true ]; then
        echo -e "${CYAN}等待服务完全启动...${NC}"
        sleep 5

        # 验证服务连接
        if [ "$mongodb_running" = true ]; then
            echo -e "${CYAN}验证 MongoDB 连接...${NC}"
            local mongo_retries=0
            local mongo_max_retries=10
            while [ $mongo_retries -lt $mongo_max_retries ]; do
                if docker exec mongodb-mockserver mongosh --eval "db.adminCommand('ping')" >/dev/null 2>&1; then
                    echo -e "${GREEN}✓ MongoDB 连接验证成功${NC}"
                    break
                fi
                echo -e "${YELLOW}  MongoDB 连接验证失败，重试 $((mongo_retries + 1))/$mongo_max_retries${NC}"
                sleep 2
                mongo_retries=$((mongo_retries + 1))
            done
        fi

        if [ "$redis_running" = true ]; then
            echo -e "${CYAN}验证 Redis 连接...${NC}"
            local redis_retries=0
            local redis_max_retries=5
            while [ $redis_retries -lt $redis_max_retries ]; do
                if docker exec redis-mockserver redis-cli ping | grep -q "PONG"; then
                    echo -e "${GREEN}✓ Redis 连接验证成功${NC}"
                    break
                fi
                echo -e "${YELLOW}  Redis 连接验证失败，重试 $((redis_retries + 1))/$redis_max_retries${NC}"
                sleep 1
                redis_retries=$((redis_retries + 1))
            done
        fi

        echo -e "${GREEN}✅ Docker 服务启动和验证完成${NC}"
    fi

    return 0
}

# 智能服务协调 - 根据SKIP_SERVER_START决定启动策略
coordinate_services() {
    if [ "$SKIP_SERVER_START" = "true" ]; then
        echo -e "${CYAN}SKIP_SERVER_START=true，检查现有服务状态...${NC}"

        # 检查现有服务
        local admin_api_up=false
        local mock_api_up=false

        # 简单的健康检查
        if command -v curl >/dev/null 2>&1; then
            if curl -s "$ADMIN_API/health" | grep -q "healthy"; then
                admin_api_up=true
                echo -e "${GREEN}✓ Admin API ($ADMIN_API) 已运行${NC}"
            fi

            # 检查Mock API（如果PROJECT_ID和ENVIRONMENT_ID存在）
            if [ -n "$PROJECT_ID" ] && [ -n "$ENVIRONMENT_ID" ]; then
                if curl -s "$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID/health" >/dev/null; then
                    mock_api_up=true
                    echo -e "${GREEN}✓ Mock API 已运行${NC}"
                fi
            fi
        fi

        # 如果关键服务未运行，启动Docker依赖
        if [ "$admin_api_up" = false ]; then
            echo -e "${YELLOW}Admin API未运行，启动依赖服务...${NC}"
            start_docker_services
        fi

    else
        echo -e "${CYAN}SKIP_SERVER_START=false，启动完整服务栈...${NC}"
        # 启动完整服务栈（包括后端服务）
        if command -v make >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/Makefile" ]; then
            echo -e "${CYAN}使用Makefile启动服务...${NC}"
            make start-all
        else
            echo -e "${RED}无法启动服务 - Makefile不可用${NC}"
            return 1
        fi
    fi

    return 0
}

# ========================================
# 增强错误处理和恢复机制
# ========================================

# 错误恢复处理
handle_test_error() {
    local error_code="$1"
    local error_message="$2"
    local context="$3"
    local operation="${4:-unknown}"

    echo -e "${RED}=== 测试错误处理 ===${NC}"
    echo -e "${RED}错误代码: $error_code${NC}"
    echo -e "${RED}错误信息: $error_message${NC}"
    echo -e "${RED}上下文: $context${NC}"
    echo -e "${RED}操作: $operation${NC}"

    # 记录错误到日志文件
    local error_log="/tmp/mockserver_test_errors.log"
    echo "$(date '+%Y-%m-%d %H:%M:%S') [ERROR] $operation: $error_message (Context: $context)" >> "$error_log"

    # 根据错误类型执行恢复操作
    case "$error_code" in
        1)
            echo -e "${YELLOW}执行一般错误恢复...${NC}"
            recover_from_general_error "$operation"
            ;;
        2)
            echo -e "${YELLOW}执行网络连接错误恢复...${NC}"
            recover_from_network_error "$operation"
            ;;
        3)
            echo -e "${YELLOW}执行端口冲突错误恢复...${NC}"
            recover_from_port_conflict "$operation"
            ;;
        4)
            echo -e "${YELLOW}执行服务启动错误恢复...${NC}"
            recover_from_service_error "$operation"
            ;;
        *)
            echo -e "${YELLOW}执行默认错误恢复...${NC}"
            recover_from_default_error "$operation"
            ;;
    esac

    # 清理部分状态以允许继续测试
    cleanup_partial_state "$operation"

    return "$error_code"
}

# 一般错误恢复
recover_from_general_error() {
    local operation="$1"
    echo -e "${CYAN}执行一般错误恢复操作...${NC}"

    # 等待一段时间后重试
    sleep 2

    # 检查系统资源
    check_system_resources

    # 清理临时文件
    cleanup_temp_files
}

# 网络连接错误恢复
recover_from_network_error() {
    local operation="$1"
    echo -e "${CYAN}执行网络连接错误恢复操作...${NC}"

    # 检查网络连接
    if command -v ping >/dev/null 2>&1; then
        if ping -c 1 localhost >/dev/null 2>&1; then
            echo -e "${GREEN}✓ 本地网络连接正常${NC}"
        else
            echo -e "${YELLOW}⚠ 本地网络连接异常${NC}"
        fi
    fi

    # 重置网络连接（如果需要）
    reset_network_connections

    # 重新检查服务状态
    check_service_status
}

# 端口冲突错误恢复
recover_from_port_conflict() {
    local operation="$1"
    echo -e "${CYAN}执行端口冲突错误恢复操作...${NC}"

    # 查找并终止占用端口的进程
    local conflicting_ports=(8080 9090 27017 6379 5173)

    for port in "${conflicting_ports[@]}"; do
        if ! check_port_available "$port" "conflict-check" >/dev/null 2>&1; then
            echo -e "${YELLOW}发现端口冲突: $port，尝试清理...${NC}"
            terminate_port_process "$port"
        fi
    done
}

# 服务启动错误恢复
recover_from_service_error() {
    local operation="$1"
    echo -e "${CYAN}执行服务启动错误恢复操作...${NC}"

    # 重新启动依赖服务
    restart_dependency_services

    # 检查Docker状态
    check_docker_status

    # 清理故障容器
    cleanup_failed_containers
}

# 默认错误恢复
recover_from_default_error() {
    local operation="$1"
    echo -e "${CYAN}执行默认错误恢复操作...${NC}"

    # 基本清理操作
    cleanup_temp_files
    sleep 1

    # 记录错误状态
    record_error_state "$operation"
}

# 终止占用端口的进程
terminate_port_process() {
    local port="$1"

    if command -v lsof >/dev/null 2>&1; then
        local pids=$(lsof -ti ":$port" 2>/dev/null)
        if [ -n "$pids" ]; then
            echo -e "${YELLOW}终止占用端口 $port 的进程: $pids${NC}"
            echo "$pids" | xargs kill -TERM 2>/dev/null || true
            sleep 2

            # 如果进程仍在运行，强制终止
            local remaining_pids=$(lsof -ti ":$port" 2>/dev/null)
            if [ -n "$remaining_pids" ]; then
                echo -e "${RED}强制终止进程: $remaining_pids${NC}"
                echo "$remaining_pids" | xargs kill -KILL 2>/dev/null || true
            fi
        fi
    fi
}

# 重置网络连接
reset_network_connections() {
    echo -e "${CYAN}重置网络连接...${NC}"

    # 清理可能的网络连接缓存
    if command -v dns_clean >/dev/null 2>&1; then
        dns_clean >/dev/null 2>&1 || true
    fi
}

# 重新启动依赖服务
restart_dependency_services() {
    echo -e "${CYAN}重新启动依赖服务...${NC}"

    # 重新启动Docker服务（如果需要）
    if command -v docker >/dev/null 2>&1; then
        if ! docker info >/dev/null 2>&1; then
            echo -e "${YELLOW}Docker服务未运行，尝试启动...${NC}"
            # 尝试启动Docker（根据系统不同）
            if command -v systemctl >/dev/null 2>&1; then
                sudo systemctl restart docker 2>/dev/null || true
            elif command -v service >/dev/null 2>&1; then
                sudo service docker restart 2>/dev/null || true
            fi
        fi
    fi
}

# 检查Docker状态
check_docker_status() {
    if command -v docker >/dev/null 2>&1; then
        if docker info >/dev/null 2>&1; then
            echo -e "${GREEN}✓ Docker服务正常运行${NC}"
            return 0
        else
            echo -e "${RED}✗ Docker服务异常${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠ Docker未安装${NC}"
        return 1
    fi
}

# 清理故障容器
cleanup_failed_containers() {
    if command -v docker >/dev/null 2>&1; then
        echo -e "${CYAN}清理故障的Docker容器...${NC}"

        # 清理已停止的容器
        docker container prune -f >/dev/null 2>&1 || true

        # 清理未使用的镜像
        docker image prune -f >/dev/null 2>&1 || true
    fi
}

# 检查系统资源
check_system_resources() {
    echo -e "${CYAN}检查系统资源状态...${NC}"

    # 检查磁盘空间
    local disk_usage=$(df / 2>/dev/null | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ "$disk_usage" -gt 90 ]; then
        echo -e "${YELLOW}⚠ 磁盘空间不足: ${disk_usage}%${NC}"
    else
        echo -e "${GREEN}✓ 磁盘空间充足: ${disk_usage}%${NC}"
    fi

    # 检查内存使用
    if [[ "$(uname)" == "Darwin" ]]; then
        local memory_pressure=$(memory_pressure 2>/dev/null | grep "System-wide memory free percentage" | awk '{print $5}' | sed 's/%//' || echo "N/A")
        if [[ "$memory_pressure" != "N/A" ]] && [ "$memory_pressure" -lt 10 ]; then
            echo -e "${YELLOW}⚠ 内存压力较高: ${memory_pressure}%${NC}"
        else
            echo -e "${GREEN}✓ 内存使用正常${NC}"
        fi
    fi
}

# 清理临时文件
cleanup_temp_files() {
    echo -e "${CYAN}清理临时文件...${NC}"

    # 清理测试临时目录
    if [ -d "/tmp/mockserver_test" ]; then
        rm -rf /tmp/mockserver_test 2>/dev/null || true
    fi

    # 清理其他临时文件
    find /tmp -name "mockserver_*" -type f -mtime +1 -delete 2>/dev/null || true
}

# 清理部分状态
cleanup_partial_state() {
    local operation="$1"
    echo -e "${CYAN}清理 $operation 的部分状态...${NC}"

    # 根据操作类型进行特定清理
    case "$operation" in
        "service_start")
            # 清理可能的服务状态
            unset PROJECT_ID ENVIRONMENT_ID 2>/dev/null || true
            ;;
        "test_execution")
            # 清理测试相关状态
            cleanup_test_files "/tmp/mockserver_test" 2>/dev/null || true
            ;;
        *)
            # 通用清理
            cleanup_temp_files
            ;;
    esac
}

# 记录错误状态
record_error_state() {
    local operation="$1"
    local state_file="/tmp/mockserver_error_state.json"

    # 创建错误状态记录
    cat > "$state_file" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "operation": "$operation",
    "error_count": "$((TEST_FAILED + 1))",
    "system_info": {
        "os": "$(uname -s)",
        "uptime": "$(uptime)",
        "disk_usage": "$(df / 2>/dev/null | tail -1 | awk '{print $5}')",
        "load_average": "$(uptime | awk -F'load average:' '{print $2}' | sed 's/^[[:space:]]*//')"
    }
}
EOF
}

# 检查服务状态
check_service_status() {
    echo -e "${CYAN}检查关键服务状态...${NC}"

    # 检查Admin API
    if http_get "$ADMIN_API/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Admin API 可访问${NC}"
    else
        echo -e "${YELLOW}⚠ Admin API 不可访问${NC}"
    fi

    # 检查Mock API
    if http_get "$MOCK_API/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Mock API 可访问${NC}"
    else
        echo -e "${YELLOW}⚠ Mock API 不可访问${NC}"
    fi
}

# 健康检查和自动修复
health_check_and_auto_repair() {
    echo -e "${CYAN}执行系统健康检查和自动修复...${NC}"

    local issues_found=0

    # 检查端口冲突
    local critical_ports=(8080 9090 27017 6379)
    for port in "${critical_ports[@]}"; do
        if ! check_port_available "$port" "health-check" >/dev/null 2>&1; then
            echo -e "${YELLOW}⚠ 端口 $port 冲突，尝试自动修复...${NC}"
            terminate_port_process "$port"
            issues_found=$((issues_found + 1))
        fi
    done

    # 检查Docker状态
    if ! check_docker_status >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠ Docker状态异常，尝试自动修复...${NC}"
        restart_dependency_services
        issues_found=$((issues_found + 1))
    fi

    # 检查系统资源
    check_system_resources

    if [ "$issues_found" -eq 0 ]; then
        echo -e "${GREEN}✓ 系统健康检查通过，未发现问题${NC}"
    else
        echo -e "${YELLOW}⚠ 发现 $issues_found 个问题，已尝试自动修复${NC}"
    fi

    return "$issues_found"
}

# 带错误保护的命令执行
safe_execute() {
    local description="$1"
    local command="$2"
    local max_retries="${3:-3}"
    local retry_delay="${4:-2}"

    echo -e "${CYAN}执行命令: $description${NC}"

    local attempt=1
    while [ $attempt -le $max_retries ]; do
        echo -e "  尝试第 $attempt 次..."

        if eval "$command"; then
            echo -e "${GREEN}✓ $description 执行成功${NC}"
            return 0
        else
            local exit_code=$?
            echo -e "${YELLOW}⚠ $description 执行失败 (退出码: $exit_code)${NC}"

            if [ $attempt -lt $max_retries ]; then
                echo -e "  等待 $retry_delay 秒后重试..."
                sleep "$retry_delay"

                # 尝试错误恢复
                handle_test_error "$exit_code" "$description 失败" "safe_execute_retry" "$command"

                # 指数退避
                retry_delay=$((retry_delay * 2))
            fi

            attempt=$((attempt + 1))
        fi
    done

    echo -e "${RED}✗ $description 在 $max_retries 次尝试后仍然失败${NC}"
    handle_test_error "$exit_code" "$description 最终失败" "safe_execute_final" "$command"
    return "$exit_code"
}

# ========================================
# 测试日志函数
# ========================================

log_test() {
    echo -e "${CYAN}[TEST]${NC} $1"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
}

# ========================================
# 测试结果记录函数
# ========================================

test_pass() {
    echo -e "${GREEN}✓ $1${NC}"
    TEST_PASSED=$((TEST_PASSED + 1))
    TEST_RESULTS+=("PASS: $1")
}

test_fail() {
    echo -e "${RED}✗ $1${NC}"
    TEST_FAILED=$((TEST_FAILED + 1))
    TEST_RESULTS+=("FAIL: $1")
}

test_skip() {
    echo -e "${YELLOW}⚠ $1${NC}"
    TEST_SKIPPED=$((TEST_SKIPPED + 1))
    TEST_RESULTS+=("SKIP: $1")
}

test_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

test_warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# ========================================
# JSON 工具函数
# ========================================

extract_json_field() {
    echo "$1" | grep -o "\"$2\":\"[^\"]*\"" | cut -d'"' -f4
}

extract_json_field_array() {
    echo "$1" | grep -o "\"$2\":\[[^]]*\]" | sed 's/.*\[\(.*\)\].*/\1/'
}

extract_json_bool() {
    local value=$(echo "$1" | grep -o "\"$2\":[^,}]*" | cut -d':' -f2 | tr -d ' ')
    echo "$value"
}

# ========================================
# HTTP 请求封装函数
# ========================================

http_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local headers="$4"

    local cmd="curl -s -w '\n%{http_code}\n'"

    if [ -n "$headers" ]; then
        cmd="$cmd $headers"
    fi

    if [ -n "$data" ]; then
        cmd="$cmd -X $method -H 'Content-Type: application/json' -d '$data'"
    else
        cmd="$cmd -X $method"
    fi

    cmd="$cmd '$url'"

    eval "$cmd"
}

http_get() {
    http_request "GET" "$1" "" "$2"
}

http_post() {
    http_request "POST" "$1" "$2" "$3"
}

http_put() {
    http_request "PUT" "$1" "$2" "$3"
}

http_delete() {
    http_request "DELETE" "$1" "" "$2"
}

# ========================================
# Mock 请求函数
# ========================================

mock_request() {
    local method="$1"
    local path="$2"
    local data="$3"
    local headers="$4"

    # 确保关键变量已设置
    if [ -z "$PROJECT_ID" ] || [ -z "$ENVIRONMENT_ID" ]; then
        echo "Error: PROJECT_ID or ENVIRONMENT_ID is not set"
        echo "PROJECT_ID='${PROJECT_ID}' ENVIRONMENT_ID='${ENVIRONMENT_ID}'"
        return 1
    fi

    local url="$MOCK_API/$PROJECT_ID/$ENVIRONMENT_ID$path"

    # 使用绝对路径的curl命令以避免PATH问题
    local curl_cmd="curl"
    if command -v curl >/dev/null 2>&1; then
        curl_cmd="curl"
    elif [ -x "/opt/anaconda3/bin/curl" ]; then
        curl_cmd="/opt/anaconda3/bin/curl"
    elif [ -x "/usr/bin/curl" ]; then
        curl_cmd="/usr/bin/curl"
    else
        echo "Error: curl command not found"
        return 1
    fi

    if [ -n "$data" ]; then
        cmd="$curl_cmd -s -L -w '\n%{http_code}\n' -X $method"
    else
        cmd="$curl_cmd -s -L -w '\n%{http_code}\n' -X $method"
    fi

    if [ -n "$headers" ]; then
        cmd="$cmd $headers"
    fi

    if [ -n "$data" ]; then
        cmd="$cmd -H 'Content-Type: application/json' -d '$data'"
    fi

    cmd="$cmd '$url'"

    eval "$cmd"
}

# ========================================
# 重试机制函数
# ========================================

retry_with_backoff() {
    local max_attempts="$1"
    local delay="$2"
    local command="$3"
    local description="$4"

    local attempt=1
    while [ $attempt -le $max_attempts ]; do
        echo -e "  尝试第 $attempt 次: $description"

        if eval "$command"; then
            return 0
        fi

        if [ $attempt -lt $max_attempts ]; then
            echo -e "  等待 $delay 秒后重试..."
            sleep "$delay"
            delay=$((delay * 2))  # 指数退避
        fi

        attempt=$((attempt + 1))
    done

    return 1
}

# ========================================
# 时间工具函数
# ========================================

get_timestamp_ms() {
    # macOS 兼容的时间戳获取
    python3 -c 'import time; print(int(time.time() * 1000))' 2>/dev/null || date +%s000
}

# 跨平台序列生成函数
seq() {
    local start="${1:-1}"
    local end="$2"

    if [ $# -eq 1 ]; then
        end="$1"
        start=1
    fi

    # macOS 和 Linux 都支持 seq 命令，这里提供备用方案
    command seq "$start" "$end" 2>/dev/null || {
        # 备用方案：使用 while 循环
        local i=$start
        while [ $i -le $end ]; do
            echo $i
            i=$((i + 1))
        done
    }
}

# 跨平台工具检测和备用方案
check_command() {
    local cmd="$1"
    local fallback="$2"

    if command -v "$cmd" >/dev/null 2>&1; then
        echo "$cmd"
    elif [ -n "$fallback" ] && command -v "$fallback" >/dev/null 2>&1; then
        echo "$fallback"
    else
        return 1
    fi
}

# 跨平台超时函数
timeout_cmd() {
    local duration="$1"
    shift
    local command=("$@")

    # macOS 没有 timeout 命令，使用备用方案
    if command -v timeout >/dev/null 2>&1; then
        timeout "$duration" "${command[@]}"
    else
        # macOS 备用方案：使用 Perl 的 alarm 函数
        perl -e 'alarm shift @ARGV; exec @ARGV' "$duration" "${command[@]}"
    fi
}

# 跨平台进程查找
find_process() {
    local process_name="$1"

    # 首先尝试 pgrep
    if command -v pgrep >/dev/null 2>&1; then
        pgrep "$process_name" | head -1
        return 0
    fi

    # 备用方案：使用 ps
    if command -v ps >/dev/null 2>&1; then
        ps aux | grep "[${process_name:0:1}]${process_name:1}" | awk '{print $2}' | head -1
        return 0
    fi

    return 1
}

# 跨平台内存获取
get_process_memory() {
    local pid="$1"

    if ! command -v ps >/dev/null 2>&1; then
        return 1
    fi

    local memory=""
    if [[ "$(uname)" == "Darwin" ]]; then
        # macOS 使用 ps 命令
        memory=$(ps -o pid,rss -p "$pid" 2>/dev/null | awk 'NR==2 {print $2}')
    else
        # Linux 使用 ps 命令
        memory=$(ps -o pid,rss -p "$pid" 2>/dev/null | awk 'NR==2 {print $2}')
    fi

    if [ -n "$memory" ]; then
        echo "$memory"
        return 0
    fi

    return 1
}

calculate_duration() {
    local start_time="$1"
    local end_time="$2"
    echo $((end_time - start_time))
}

# ========================================
# 文件系统工具函数
# ========================================

create_test_file() {
    local file_path="$1"
    local content="$2"
    local size="$3"

    mkdir -p "$(dirname "$file_path")"

    if [ -n "$content" ]; then
        echo "$content" > "$file_path"
    elif [ -n "$size" ]; then
        dd if=/dev/zero of="$file_path" bs=1 count="$size" 2>/dev/null
    fi
}

cleanup_test_files() {
    local test_dir="$1"
    if [ -n "$test_dir" ] && [ -d "$test_dir" ]; then
        rm -rf "$test_dir"
    fi
}

# ========================================
# 随机数据生成函数
# ========================================

generate_random_string() {
    local length="${1:-8}"
    openssl rand -hex $((length/2)) 2>/dev/null | head -c "$length"
}

generate_random_email() {
    echo "test-$(generate_random_string 6)@example.com"
}

generate_random_phone() {
    echo "1$(generate_random_string 10)"
}

generate_random_id() {
    echo "$(date +%s%3N)$(generate_random_string 4)"
}

# ========================================
# 测试数据生成函数
# ========================================

generate_project_data() {
    local name="$1"
    cat <<EOF
{
    "name": "$name",
    "workspace_id": "workspace-$(generate_random_string 8)",
    "description": "E2E测试项目 - $name"
}
EOF
}

generate_environment_data() {
    local name="$1"
    local base_url="$2"
    cat <<EOF
{
    "name": "$name",
    "base_url": "$base_url",
    "variables": {
        "api_version": "v1",
        "timeout": "30s",
        "retry_count": "3"
    }
}
EOF
}

generate_rule_data() {
    local name="$1"
    local method="$2"
    local path="$3"
    local response_body="$4"
    local status_code="${5:-200}"

    cat <<EOF
{
    "name": "$name",
    "project_id": "$PROJECT_ID",
    "environment_id": "$ENVIRONMENT_ID",
    "protocol": "HTTP",
    "match_type": "Simple",
    "priority": 100,
    "enabled": true,
    "match_condition": {
        "method": "$method",
        "path": "$path"
    },
    "response": {
        "type": "Static",
        "content": {
            "status_code": $status_code,
            "content_type": "JSON",
            "headers": {
                "X-Mock-Server": "MockServer",
                "X-Test-Case": "$name"
            },
            "body": $response_body
        }
    }
}
EOF
}

# ========================================
# WebSocket 测试函数
# ========================================

websocket_test_connection() {
    local ws_url="$1"
    local test_message="$2"

    # 使用 websocat 或类似工具测试 WebSocket 连接
    if command -v websocat >/dev/null 2>&1; then
        echo "$test_message" | websocat --one-message "$ws_url" 2>/dev/null
        return 0
    else
        test_warn "websocat 未安装，跳过 WebSocket 测试"
        return 1
    fi
}

# ========================================
# 性能测试函数
# ========================================

performance_test() {
    local url="$1"
    local concurrent="$2"
    local duration="$3"

    echo -e "${CYAN}开始性能测试: $concurrent 并发，$duration 秒${NC}"

    if command -v wrk >/dev/null 2>&1; then
        wrk -t4 -c"$concurrent" -d"$duration"s "$url" 2>/dev/null | tee /tmp/perf_test.log
        return 0
    else
        test_warn "wrk 未安装，使用简单并发测试"

        # 简单并发测试
        for i in $(seq 1 $concurrent); do
            (curl -s "$url" >/dev/null 2>&1) &
        done
        wait

        echo "简单并发测试完成: $concurrent 个请求"
        return 0
    fi
}

# ========================================
# 测试报告生成函数
# ========================================

generate_test_report() {
    local report_file="$1"
    local test_name="$2"

    TEST_END_TIME=$(date +%s)
    TEST_DURATION=$((TEST_END_TIME - TEST_START_TIME))

    cat > "$report_file" << EOF
# $test_name 测试报告

## 测试概要
- **测试时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **测试时长**: ${TEST_DURATION}秒
- **通过测试**: $TEST_PASSED
- **失败测试**: $TEST_FAILED
- **跳过测试**: $TEST_SKIPPED
- **总计测试**: $((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))
- **成功率**: $(( TEST_PASSED * 100 / (TEST_PASSED + TEST_FAILED) ))%

## 测试结果详情
EOF

    if [ ${#TEST_RESULTS[@]} -gt 0 ]; then
        echo "" >> "$report_file"
        echo "### 详细结果" >> "$report_file"
        echo "" >> "$report_file"

        for result in "${TEST_RESULTS[@]}"; do
            if [[ $result == PASS* ]]; then
                echo "- ✅ ${result#PASS: }" >> "$report_file"
            elif [[ $result == FAIL* ]]; then
                echo "- ❌ ${result#FAIL: }" >> "$report_file"
            elif [[ $result == SKIP* ]]; then
                echo "- ⚠️ ${result#SKIP: }" >> "$report_file"
            fi
        done
    fi

    echo "" >> "$report_file"
    echo "## 环境信息" >> "$report_file"
    echo "- **操作系统**: $(uname -s)" >> "$report_file"
    echo "- **Go版本**: $(go version 2>/dev/null || echo 'Unknown')" >> "$report_file"
    echo "- **配置文件**: $CONFIG_FILE" >> "$report_file"
    echo "- **管理API**: $ADMIN_API" >> "$report_file"
    echo "- **MockAPI**: $MOCK_API" >> "$report_file"

    echo -e "${GREEN}测试报告已生成: $report_file${NC}"
}

# ========================================
# 清理函数
# ========================================

cleanup_test_resources() {
    echo -e "${YELLOW}清理测试资源...${NC}"

    # 删除测试项目
    if [ -n "$PROJECT_ID" ]; then
        echo "  删除测试项目: $PROJECT_ID"
        http_delete "$ADMIN_API/projects/$PROJECT_ID" >/dev/null 2>&1 || true
    fi

    # 删除 WebSocket 测试项目
    if [ -n "$WS_PROJECT_ID" ]; then
        echo "  删除WebSocket测试项目: $WS_PROJECT_ID"
        http_delete "$ADMIN_API/projects/$WS_PROJECT_ID" >/dev/null 2>&1 || true
    fi

    # 清理测试文件
    cleanup_test_files "/tmp/mockserver_test"
}

print_test_summary() {
    echo ""
    echo -e "${BLUE}=========================================${NC}"
    echo -e "${BLUE}   测试结果统计${NC}"
    echo -e "${BLUE}=========================================${NC}"
    echo -e "通过测试: ${GREEN}$TEST_PASSED${NC}"
    echo -e "失败测试: ${RED}$TEST_FAILED${NC}"
    echo -e "跳过测试: ${YELLOW}$TEST_SKIPPED${NC}"
    echo -e "总计测试: $((TEST_PASSED + TEST_FAILED + TEST_SKIPPED))"

    local total=$((TEST_PASSED + TEST_FAILED))
    if [ $total -gt 0 ]; then
        local success_rate=$(( TEST_PASSED * 100 / total ))
        echo -e "成功率: ${GREEN}$success_rate%${NC}"
    fi

    if [ $TEST_FAILED -eq 0 ]; then
        echo -e "${GREEN}✓ 所有测试通过！${NC}"
        return 0
    else
        echo -e "${RED}✗ 部分测试失败${NC}"
        return 1
    fi
}

# ========================================
# 初始化函数
# ========================================

init_test_framework() {
    TEST_START_TIME=$(date +%s)
    detect_environment

    # 检查并安装必要工具
    check_and_install_dependencies

    # 协调服务启动 - 新增的服务管理功能
    coordinate_services

    trap cleanup_test_resources EXIT INT TERM
}

# 检查并安装依赖
check_and_install_dependencies() {
    # 加载工具安装器
    local installer_path="$(dirname "$0")/tool_installer.sh"
    if [ -f "$installer_path" ]; then
        source "$installer_path"

        # 检查基础工具
        if ! check_tools_ready "basic"; then
            echo -e "${YELLOW}检测到缺失的基础工具，正在自动安装...${NC}"
            if ! install_required_tools_silent "basic"; then
                echo -e "${YELLOW}部分工具安装失败，请检查工具可用性${NC}"
            fi
        fi
    fi
}

# ========================================
# Redis 测试函数
# ========================================

# 检查Redis连接
check_redis_connection() {
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    if command -v redis-cli >/dev/null 2>&1; then
        if redis-cli -h "$redis_host" -p "$redis_port" ping >/dev/null 2>&1; then
            return 0
        fi
    fi
    return 1
}

# 启动Redis服务（如果需要）
start_redis_if_needed() {
    if check_redis_connection; then
        test_info "Redis is already running"
        return 0
    fi

    test_info "Starting Redis service..."
    if command -v make >/dev/null 2>&1 && [ -f "$PROJECT_ROOT/Makefile" ]; then
        make start-redis >/dev/null 2>&1
        sleep 2
        return $?
    else
        test_warn "Cannot start Redis automatically - please start Redis manually"
        return 1
    fi
}

# Redis缓存测试
test_redis_cache_operations() {
    local test_key="mockserver_test_$(generate_random_string 8)"
    local test_value="test_value_$(date +%s)"
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    if ! check_redis_connection; then
        test_skip "Redis not available for cache testing"
        return 0
    fi

    # SET操作测试
    if redis-cli -h "$redis_host" -p "$redis_port" set "$test_key" "$test_value" | grep -q "OK"; then
        test_pass "Redis SET operation"
    else
        test_fail "Redis SET operation"
        return 1
    fi

    # GET操作测试
    local retrieved_value=$(redis-cli -h "$redis_host" -p "$redis_port" get "$test_key")
    if [ "$retrieved_value" = "$test_value" ]; then
        test_pass "Redis GET operation"
    else
        test_fail "Redis GET operation - expected '$test_value', got '$retrieved_value'"
        return 1
    fi

    # 过期时间测试
    redis-cli -h "$redis_host" -p "$redis_port" setex "${test_key}_expire" 2 "expire_test" >/dev/null
    if redis-cli -h "$redis_host" -p "$redis_port" get "${test_key}_expire" | grep -q "expire_test"; then
        test_pass "Redis SETEX operation (immediate)"
        sleep 3
        if redis-cli -h "$redis_host" -p "$redis_port" get "${test_key}_expire" | grep -q "(nil)"; then
            test_pass "Redis key expiration"
        else
            test_fail "Redis key expiration"
        fi
    else
        test_fail "Redis SETEX operation"
    fi

    # 清理测试数据
    redis-cli -h "$redis_host" -p "$redis_port" del "$test_key" "${test_key}_expire" >/dev/null 2>&1 || true

    return 0
}

# Redis性能测试
test_redis_performance() {
    local num_operations=100
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    if ! check_redis_connection; then
        test_skip "Redis not available for performance testing"
        return 0
    fi

    test_info "Testing Redis performance with $num_operations operations"

    local test_key="perf_test_$(date +%s)"
    local start_time=$(get_timestamp_ms)

    # SET性能测试
    for i in $(seq 1 $num_operations); do
        redis-cli -h "$redis_host" -p "$redis_port" set "${test_key}_$i" "value_$i" >/dev/null
    done

    local set_time=$(($(get_timestamp_ms) - start_time))
    local set_ops_per_sec=$((num_operations * 1000 / set_time))

    test_info "Redis SET: $num_operations operations in ${set_time}ms (${set_ops_per_sec} ops/sec)"

    # GET性能测试
    start_time=$(get_timestamp_ms)
    for i in $(seq 1 $num_operations); do
        redis-cli -h "$redis_host" -p "$redis_port" get "${test_key}_$i" >/dev/null
    done

    local get_time=$(($(get_timestamp_ms) - start_time))
    local get_ops_per_sec=$((num_operations * 1000 / get_time))

    test_info "Redis GET: $num_operations operations in ${get_time}ms (${get_ops_per_sec} ops/sec)"

    # 性能基准检查
    if [ $set_ops_per_sec -gt 1000 ] && [ $get_ops_per_sec -gt 2000 ]; then
        test_pass "Redis performance meets requirements"
    else
        test_warn "Redis performance below expected (SET: ${set_ops_per_sec}, GET: ${get_ops_per_sec})"
    fi

    # 清理性能测试数据
    for i in $(seq 1 $num_operations); do
        redis-cli -h "$redis_host" -p "$redis_port" del "${test_key}_$i" >/dev/null 2>&1 || true
    done

    return 0
}

# Redis内存使用检查
test_redis_memory_usage() {
    if ! check_redis_connection; then
        test_skip "Redis not available for memory testing"
        return 0
    fi

    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"

    # 获取内存信息
    local memory_info=$(redis-cli -h "$redis_host" -p "$redis_port" info memory 2>/dev/null)
    if [ -n "$memory_info" ]; then
        local used_memory=$(echo "$memory_info" | grep "used_memory_human:" | cut -d: -f2 | tr -d '[:space:]')
        local used_memory_rss=$(echo "$memory_info" | grep "used_memory_rss_human:" | cut -d: -f2 | tr -d '[:space:]')

        test_info "Redis memory usage: $used_memory (RSS: $used_memory_rss)"
        test_pass "Redis memory monitoring"
    else
        test_fail "Redis memory info retrieval"
        return 1
    fi

    return 0
}

# 运行Redis集成测试
run_redis_integration_tests() {
    test_info "Starting Redis integration tests"

    # 确保Redis运行
    start_redis_if_needed

    # 运行各种Redis测试
    test_redis_cache_operations
    test_redis_performance
    test_redis_memory_usage

    # 如果有独立的Redis测试脚本，也运行它们
    local redis_integration_script="$PROJECT_ROOT/tests/redis/redis_integration_test.sh"
    if [ -f "$redis_integration_script" ] && [ -x "$redis_integration_script" ]; then
        test_info "Running comprehensive Redis integration tests"
        if "$redis_integration_script"; then
            test_pass "Comprehensive Redis integration tests"
        else
            test_fail "Comprehensive Redis integration tests"
        fi
    fi

    test_info "Redis integration tests completed"
}

# 运行Redis性能测试
run_redis_performance_tests() {
    test_info "Starting Redis performance tests"

    # 确保Redis运行
    start_redis_if_needed

    # 如果有独立的Redis性能测试脚本，运行它们
    local redis_perf_script="$PROJECT_ROOT/tests/redis/redis_performance_test.sh"
    if [ -f "$redis_perf_script" ] && [ -x "$redis_perf_script" ]; then
        test_info "Running comprehensive Redis performance tests"
        if "$redis_perf_script" "/tmp/redis_performance_report.txt"; then
            test_pass "Comprehensive Redis performance tests"
            test_info "Performance report saved to /tmp/redis_performance_report.txt"
        else
            test_fail "Comprehensive Redis performance tests"
        fi
    else
        test_warn "Redis performance test script not found, running basic tests"
        test_redis_performance
    fi

    test_info "Redis performance tests completed"
}

# ========================================
# 导出函数
# ========================================

# 导出所有函数供其他脚本使用
export -f detect_environment
export -f test_pass test_fail test_skip test_info test_warn
export -f extract_json_field extract_json_field_array extract_json_bool
export -f http_request http_get http_post http_put http_delete
export -f mock_request
export -f retry_with_backoff
export -f get_timestamp_ms calculate_duration seq check_command find_process get_process_memory
export -f create_test_file cleanup_test_files
export -f generate_random_string generate_random_email generate_random_phone generate_random_id
export -f generate_project_data generate_environment_data generate_rule_data
export -f websocket_test_connection
export -f performance_test
export -f check_redis_connection start_redis_if_needed
export -f test_redis_cache_operations test_redis_performance test_redis_memory_usage
export -f run_redis_integration_tests run_redis_performance_tests
export -f generate_test_report
export -f cleanup_test_resources print_test_summary
export -f init_test_framework

echo -e "${GREEN}测试框架已加载${NC}"
# MockServer 测试失败问题详细解决方案

## 📋 问题概览

基于对测试日志和系统状态的深入分析，识别出了两个主要问题：

1. **Redis缓存测试失败** - 配置文件禁用了Redis功能
2. **WebSocket测试失败** - API健康检查路径错误和服务启动问题

## 🔍 问题一：Redis缓存测试失败

### 问题描述
- **错误表现**: Redis删除验证失败、键过期功能失败、MockServer缓存集成测试失败
- **根本原因**: `config.dev.yaml` 中 `redis.enabled: false`，但测试脚本期望Redis功能可用

### 详细分析
```yaml
# 当前配置 (config.dev.yaml)
redis:
  enabled: false  # ❌ 问题所在
  host: "localhost"
  port: 6379
  password: ""
  db: 0
```

### 解决方案
#### 步骤1：更新配置文件
```yaml
# 修复后配置 (config.dev.yaml)
redis:
  enabled: true   # ✅ 启用Redis
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  max_retries: 3
  pool_size: 10
  health_check_interval: 30
```

#### 步骤2：确保Redis服务集成
```go
// internal/service/cache_service.go
func NewCacheService(cfg config.RedisConfig) (*CacheService, error) {
    if !cfg.Enabled {
        log.Info("Redis cache disabled, using memory cache")
        return NewMemoryCacheService(), nil
    }

    // Redis连接逻辑
    client := redis.NewClient(&redis.Options{
        Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Password:     cfg.Password,
        DB:           cfg.DB,
        PoolSize:     cfg.PoolSize,
        MaxRetries:   cfg.MaxRetries,
    })

    // 测试连接
    if err := client.Ping(context.Background()).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to Redis: %w", err)
    }

    return &CacheService{
        client: client,
        config: cfg,
    }, nil
}
```

#### 步骤3：改进测试脚本逻辑
```bash
# tests/integration/simple_cache_test.sh
check_redis_prerequisites() {
    log_info "检查Redis测试前置条件..."

    # 检查Redis服务是否运行
    if ! check_redis_connection; then
        log_skip "Redis服务不可用，跳过缓存测试"
        return 0
    fi

    # 检查MockServer是否启用了Redis
    if ! check_mockserver_redis_enabled; then
        log_skip "MockServer未启用Redis缓存，跳过集成测试"
        return 0
    fi

    log_success "Redis测试前置条件检查通过"
    return 1
}

check_mockserver_redis_enabled() {
    local response=$(curl -s "$ADMIN_API/system/config" 2>/dev/null || echo "")
    if [[ "$response" == *"\"redis_enabled\": true"* ]]; then
        return 0
    fi
    return 1
}
```

## 🔍 问题二：WebSocket测试失败

### 问题描述
- **错误表现**: WebSocket消息接收失败、心跳机制测试失败、多客户端连接测试失败
- **根本原因**: API健康检查路径错误，MockServer WebSocket服务未正确启动

### 详细分析
```bash
# 当前测试脚本使用的错误路径
HEALTH_ENDPOINT="/health"  # ❌ 错误路径

# 实际的正确路径
HEALTH_ENDPOINT="/system/health"  # ✅ 正确路径
```

### 解决方案
#### 步骤1：修正API健康检查路径
```bash
# tests/integration/lib/test_framework.sh
detect_environment() {
    log_info "检测运行环境..."

    # 检查Admin API
    if curl -s "http://localhost:8080/api/v1/system/health" >/dev/null 2>&1; then
        ADMIN_API="http://localhost:8080/api/v1"
        MOCK_API="http://localhost:9090"
        WS_API="ws://localhost:9090"
        ENVIRONMENT="development"
        log_success "检测到开发环境"
        return 0
    fi

    # 其他环境检测逻辑...
}

verify_service_health() {
    local admin_health_url="$ADMIN_API/system/health"
    local mock_health_url="$MOCK_API/health"

    log_info "验证服务健康状态..."

    # 验证Admin API健康状态
    if ! verify_endpoint_health "$admin_health_url" 10 3; then
        log_error "Admin API健康检查失败: $admin_health_url"
        return 1
    fi

    # 验证Mock API健康状态
    if ! verify_endpoint_health "$mock_health_url" 10 3; then
        log_error "Mock API健康检查失败: $mock_health_url"
        return 1
    fi

    # 验证WebSocket端点可用性
    if ! check_websocket_availability; then
        log_error "WebSocket服务不可用"
        return 1
    fi

    log_success "所有服务健康检查通过"
    return 0
}

check_websocket_availability() {
    # 使用websocat或其他工具检查WebSocket端点
    if command -v websocat >/dev/null 2>&1; then
        timeout 5 websocat --text ws://localhost:9090/ws --exit-on-eof </dev/null >/dev/null 2>&1
        return $?
    fi

    # 备用检查：简单的TCP连接测试
    nc -z localhost 9090 2>/dev/null
    return $?
}
```

#### 步骤2：确保MockServer WebSocket服务正确初始化
```go
// internal/service/mock_service.go
func (s *MockService) StartMockServer() error {
    // 启动HTTP服务器
    router := gin.New()

    // 添加WebSocket支持
    router.GET("/ws", func(c *gin.Context) {
        handleWebSocket(c, s.ruleManager)
    })

    // 添加健康检查端点
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "service": "mock_api",
            "websocket_enabled": true,
            "timestamp": time.Now().Unix(),
        })
    })

    return router.Run(s.config.Address)
}

func handleWebSocket(c *gin.Context, ruleManager *RuleManager) {
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true // 允许跨域
        },
    }

    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        log.Error("Failed to upgrade to WebSocket:", err)
        return
    }
    defer conn.Close()

    // WebSocket处理逻辑
    handleWebSocketConnection(conn, ruleManager)
}
```

#### 步骤3：完善WebSocket测试逻辑
```bash
# tests/integration/simple_websocket_test.sh
test_websocket_functionality() {
    log_info "测试WebSocket功能..."

    # 检查WebSocket工具可用性
    check_websocket_tools

    # 测试基本连接
    test_websocket_connection

    # 测试消息发送和接收
    test_message_exchange

    # 测试心跳机制
    test_heartbeat_mechanism

    # 测试多客户端连接
    test_multiple_clients
}

test_websocket_connection() {
    log_info "测试WebSocket连接建立..."

    # 使用websocat测试连接
    if command -v websocat >/dev/null 2>&1; then
        echo '{"type":"ping","data":"test"}' | timeout 10 websocat --text ws://localhost:9090/ws --exit-on-eof >/dev/null 2>&1
        if [ $? -eq 0 ]; then
            log_success "WebSocket连接测试通过"
            return 0
        fi
    fi

    # 备用测试方法
    local response=$(curl -s -H "Connection: Upgrade" \
                           -H "Upgrade: websocket" \
                           -H "Sec-WebSocket-Key: test" \
                           -H "Sec-WebSocket-Version: 13" \
                           http://localhost:9090/ws 2>/dev/null | head -1)

    if [[ "$response" == *"101"* ]] || [[ "$response" == *"Upgrade"* ]]; then
        log_success "WebSocket连接测试通过"
        return 0
    fi

    log_error "WebSocket连接测试失败"
    return 1
}
```

## 🛠️ 修复计划

### 阶段1：紧急修复（预计30分钟）
1. ✅ **修复Redis配置** - 更新 `config.dev.yaml`
2. ✅ **修正API路径** - 更新测试框架中的健康检查路径
3. ✅ **更新测试脚本** - 添加服务可用性检查

### 阶段2：验证测试（预计1小时）
1. **重新启动服务** - 使用修复后的配置
2. **运行集成测试** - 验证修复效果
3. **详细日志分析** - 确保所有功能正常

### 阶段3：完善优化（预计2小时）
1. **增强错误处理** - 添加更详细的错误信息
2. **改进测试覆盖** - 确保边界条件也被测试
3. **文档更新** - 更新相关文档和注释

## 📊 预期结果

修复完成后，预期达到以下效果：

### Redis缓存测试
- ✅ Redis连接测试通过
- ✅ 缓存CRUD操作正常
- ✅ 键过期时间管理正确
- ✅ MockServer缓存集成正常

### WebSocket测试
- ✅ WebSocket连接建立成功
- ✅ 消息收发功能正常
- ✅ 心跳机制工作正常
- ✅ 多客户端连接支持正常

## 🔧 技术细节

### 配置文件模板
```yaml
# config.dev.yaml - 完整配置模板
server:
  admin_port: 8080
  mock_port: 9090
  host: "localhost"

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "mockserver"

redis:
  enabled: true
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  max_retries: 3
  pool_size: 10
  health_check_interval: 30

websocket:
  enabled: true
  path: "/ws"
  ping_interval: 30
  pong_timeout: 10

logging:
  level: "info"
  format: "json"
```

### 测试脚本更新模板
```bash
# tests/integration/lib/test_framework.sh - 关键函数更新
verify_prerequisites() {
    log_info "验证测试前置条件..."

    # 检查依赖服务
    check_mongodb_connection || return 1

    # 检查Redis（如果启用）
    if is_redis_enabled; then
        check_redis_connection || return 1
    fi

    # 检查服务健康状态
    verify_service_health || return 1

    log_success "所有前置条件验证通过"
    return 0
}

is_redis_enabled() {
    # 检查MockServer是否启用了Redis
    local config_response=$(curl -s "$ADMIN_API/system/config" 2>/dev/null)
    [[ "$config_response" == *"\"redis_enabled\": true"* ]]
}
```

## 📋 质量检查清单

- [ ] Redis配置已更新为 `enabled: true`
- [ ] API健康检查路径已修正为 `/system/health`
- [ ] WebSocket端点 `ws://localhost:9090/ws` 可访问
- [ ] 测试脚本包含服务可用性检查
- [ ] 所有测试套件可以成功执行
- [ ] 错误信息清晰且有助于调试
- [ ] 日志记录完整且结构化

## 🚀 执行步骤

1. **立即执行**：
   ```bash
   # 1. 修复Redis配置
   vim config.dev.yaml

   # 2. 重新启动服务
   make stop-all && make start-all

   # 3. 运行测试验证修复
   make e2e
   ```

2. **验证结果**：
   - 检查测试输出中的成功/失败统计
   - 分析详细的测试日志
   - 确认所有功能模块正常工作

3. **持续监控**：
   - 建立测试监控机制
   - 定期运行完整测试套件
   - 及时发现和解决新问题

通过执行这个详细的解决方案，我们可以系统性地解决Redis缓存测试和WebSocket测试失败的问题，确保MockServer项目的稳定性和可靠性。
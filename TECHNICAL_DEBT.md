# 技术债务清单

**项目**: gomockserver  
**创建日期**: 2025-01-21  
**最后更新**: 2025-01-21  
**负责人**: 开发团队

---

## 📋 概述

本文档跟踪 gomockserver 项目中的技术债务，包括待实现的功能、已知的限制和计划中的改进。所有技术债务按优先级分类，并指定目标版本。

### 统计信息

| 类别 | 数量 | 状态 |
|------|------|------|
| P1 - 高优先级 | 2 | 🔴 待处理 |
| P2 - 中优先级 | 5 | 🟡 计划中 |
| P3 - 低优先级 | 2 | 🟢 远期规划 |
| **总计** | **9** | - |

---

## 🔴 P1 - 高优先级 (v0.2.0)

### TD-001: CIDR 格式 IP 段匹配

**文件**: `internal/engine/match_engine.go:234`

**描述**: 当前只支持单个 IP 地址匹配，不支持 CIDR 格式的 IP 段匹配（如 `192.168.1.0/24`）。

**当前实现**:
```go
// TODO: 支持 CIDR 格式的 IP 段匹配
if ruleIP == clientIP {
    return true
}
```

**影响**:
- 无法对 IP 段进行批量匹配
- 需要为每个 IP 创建单独的规则
- 影响企业内网场景的使用

**建议实现**:
```go
import "net"

func matchCIDR(ruleIP, clientIP string) bool {
    // 检查是否是 CIDR 格式
    _, ipNet, err := net.ParseCIDR(ruleIP)
    if err == nil {
        // CIDR 格式
        ip := net.ParseIP(clientIP)
        return ipNet.Contains(ip)
    }
    // 单个 IP 匹配
    return ruleIP == clientIP
}
```

**工作量**: 小 (1-2小时)  
**目标版本**: v0.2.0  
**负责人**: 待分配  
**状态**: 🔴 未开始

---

### TD-002: 正则表达式匹配

**文件**: `internal/engine/match_engine.go:140`

**描述**: 当前匹配引擎只支持精确匹配和通配符，不支持正则表达式匹配。

**当前实现**:
```go
case "regex":
    // TODO: 阶段三实现
    return false
```

**影响**:
- 无法实现复杂的路径模式匹配
- 无法匹配动态参数（如 `/api/user/\d+`）
- 限制了规则的灵活性

**建议实现**:
```go
import "regexp"

// 添加正则缓存避免重复编译
var regexCache sync.Map

func matchRegex(pattern, value string) bool {
    var re *regexp.Regexp
    
    // 从缓存获取
    if cached, ok := regexCache.Load(pattern); ok {
        re = cached.(*regexp.Regexp)
    } else {
        var err error
        re, err = regexp.Compile(pattern)
        if err != nil {
            return false
        }
        regexCache.Store(pattern, re)
    }
    
    return re.MatchString(value)
}
```

**工作量**: 中等 (3-4小时)  
**目标版本**: v0.2.0  
**负责人**: 待分配  
**状态**: 🔴 未开始

**安全考虑**:
- 需要防止 ReDoS 攻击（正则拒绝服务）
- 添加正则表达式复杂度限制
- 添加匹配超时机制

---

## 🟡 P2 - 中优先级 (v0.2.0 - v0.3.0)

### TD-003: 二进制数据处理

**文件**: `internal/executor/mock_executor.go:88`

**描述**: 当前只支持文本响应，不支持二进制数据（图片、文件等）。

**当前实现**:
```go
// TODO: 支持二进制数据
body = []byte(response.Body)
```

**影响**:
- 无法 Mock 文件下载接口
- 无法返回图片、PDF 等二进制内容
- 限制了适用场景

**建议实现**:
```go
type ResponseData struct {
    Type     string `json:"type"`      // "text", "binary", "file"
    Content  string `json:"content"`   // 文本内容或 Base64 编码
    FilePath string `json:"file_path"` // 文件路径（用于大文件）
}

func (e *MockExecutor) getResponseBody(response *ResponseData) ([]byte, error) {
    switch response.Type {
    case "binary":
        return base64.StdEncoding.DecodeString(response.Content)
    case "file":
        return os.ReadFile(response.FilePath)
    default:
        return []byte(response.Content), nil
    }
}
```

**工作量**: 中等 (4-6小时)  
**目标版本**: v0.2.0  
**负责人**: 待分配  
**状态**: 🟡 计划中

---

### TD-004: 正态分布延迟策略

**文件**: `internal/executor/mock_executor.go:127`

**描述**: 当前只支持固定延迟和随机延迟，不支持正态分布延迟（更贴近真实场景）。

**当前实现**:
```go
case "normal":
    // TODO: 实现正态分布延迟
    delay = time.Duration(delayConfig.Value) * time.Millisecond
```

**影响**:
- 无法模拟真实的网络延迟分布
- 压力测试场景不够真实

**建议实现**:
```go
import "math/rand"

type NormalDelayConfig struct {
    Mean   int `json:"mean"`   // 平均延迟 (ms)
    StdDev int `json:"stddev"` // 标准差 (ms)
}

func normalDelay(mean, stdDev float64) time.Duration {
    // Box-Muller 变换生成正态分布随机数
    u1 := rand.Float64()
    u2 := rand.Float64()
    z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
    
    delay := mean + z*stdDev
    if delay < 0 {
        delay = 0
    }
    
    return time.Duration(delay) * time.Millisecond
}
```

**工作量**: 小 (2-3小时)  
**目标版本**: v0.2.0  
**负责人**: 待分配  
**状态**: 🟡 计划中

---

### TD-005: 阶梯延迟策略

**文件**: `internal/executor/mock_executor.go:130`

**描述**: 不支持阶梯式延迟（根据请求次数递增延迟），无法模拟服务降级场景。

**当前实现**:
```go
case "step":
    // TODO: 实现阶梯延迟
    delay = time.Duration(delayConfig.Value) * time.Millisecond
```

**影响**:
- 无法模拟服务过载场景
- 无法测试客户端的退避策略

**建议实现**:
```go
type StepDelayConfig struct {
    InitialDelay int `json:"initial_delay"` // 初始延迟 (ms)
    Step         int `json:"step"`          // 每次增加 (ms)
    MaxDelay     int `json:"max_delay"`     // 最大延迟 (ms)
}

// 在 MockExecutor 中维护请求计数
func (e *MockExecutor) stepDelay(ruleID string, config StepDelayConfig) time.Duration {
    e.mu.Lock()
    count := e.requestCounts[ruleID]
    e.requestCounts[ruleID]++
    e.mu.Unlock()
    
    delay := config.InitialDelay + count*config.Step
    if delay > config.MaxDelay {
        delay = config.MaxDelay
    }
    
    return time.Duration(delay) * time.Millisecond
}
```

**工作量**: 中等 (3-4小时)  
**目标版本**: v0.2.0  
**负责人**: 待分配  
**状态**: 🟡 计划中

---

### TD-006: 脚本匹配

**文件**: `internal/engine/match_engine.go:146`

**描述**: 不支持使用脚本进行动态匹配（JavaScript、Lua 等）。

**当前实现**:
```go
case "script":
    // TODO: 阶段三实现
    return false
```

**影响**:
- 无法实现复杂的业务逻辑匹配
- 无法动态计算匹配条件

**建议实现**:
- 集成 JavaScript 引擎（如 goja）
- 或集成 Lua 引擎（如 gopher-lua）
- 提供安全沙箱环境

**安全风险**: ⚠️ 高
- 需要严格的脚本沙箱
- 需要资源限制（CPU、内存、执行时间）
- 需要审计日志

**工作量**: 大 (10-15小时)  
**目标版本**: v0.3.0  
**负责人**: 待分配  
**状态**: 🟡 计划中

---

### TD-007: WebSocket 支持

**文件**: `internal/executor/mock_executor.go:38`

**描述**: 当前只支持 HTTP/HTTPS 协议，不支持 WebSocket。

**当前实现**:
```go
case "websocket":
    // TODO: 阶段三实现
    return nil, fmt.Errorf("WebSocket not implemented yet")
```

**影响**:
- 无法 Mock WebSocket 接口
- 无法测试实时通信功能

**建议实现**:
- 使用 gorilla/websocket 库
- 支持消息推送
- 支持双向通信

**工作量**: 大 (12-16小时)  
**目标版本**: v0.3.0  
**负责人**: 待分配  
**状态**: 🟡 计划中

---

## 🟢 P3 - 低优先级 (v0.4.0+)

### TD-008: gRPC 支持

**文件**: `internal/executor/mock_executor.go:41`

**描述**: 不支持 gRPC 协议的 Mock。

**当前实现**:
```go
case "grpc":
    // TODO: 阶段三实现
    return nil, fmt.Errorf("gRPC not implemented yet")
```

**影响**:
- 无法 Mock gRPC 服务
- 限制了微服务场景的使用

**建议实现**:
- 使用 grpc-go 库
- 支持动态 proto 解析
- 支持流式 RPC

**工作量**: 大 (15-20小时)  
**目标版本**: v0.4.0  
**负责人**: 待分配  
**状态**: 🟢 远期规划

---

### TD-009: TCP 协议支持

**文件**: `internal/executor/mock_executor.go:44`

**描述**: 不支持原始 TCP 协议的 Mock。

**当前实现**:
```go
case "tcp":
    // TODO: 阶段三实现
    return nil, fmt.Errorf("TCP not implemented yet")
```

**影响**:
- 无法 Mock TCP 服务
- 无法测试底层网络协议

**建议实现**:
- 实现 TCP Server
- 支持自定义协议解析
- 支持长连接

**工作量**: 大 (10-15小时)  
**目标版本**: v0.4.0  
**负责人**: 待分配  
**状态**: 🟢 远期规划

---

## 📊 优先级评估矩阵

| ID | 技术债务 | 影响范围 | 用户需求 | 实现难度 | 综合优先级 |
|----|---------|---------|---------|---------|-----------|
| TD-001 | CIDR IP 匹配 | 高 | 高 | 低 | **P1** |
| TD-002 | 正则匹配 | 高 | 高 | 中 | **P1** |
| TD-003 | 二进制数据 | 中 | 中 | 中 | **P2** |
| TD-004 | 正态分布延迟 | 中 | 中 | 低 | **P2** |
| TD-005 | 阶梯延迟 | 中 | 中 | 中 | **P2** |
| TD-006 | 脚本匹配 | 中 | 低 | 高 | **P2** |
| TD-007 | WebSocket | 中 | 中 | 高 | **P2** |
| TD-008 | gRPC 支持 | 低 | 低 | 高 | **P3** |
| TD-009 | TCP 支持 | 低 | 低 | 高 | **P3** |

---

## 📅 实施计划

### v0.2.0 (2025-02)
- [ ] TD-001: CIDR IP 匹配
- [ ] TD-002: 正则表达式匹配
- [ ] TD-003: 二进制数据处理
- [ ] TD-004: 正态分布延迟
- [ ] TD-005: 阶梯延迟

### v0.3.0 (2025-03)
- [ ] TD-006: 脚本匹配（需安全评估）
- [ ] TD-007: WebSocket 支持

### v0.4.0 (2025-Q2)
- [ ] TD-008: gRPC 支持
- [ ] TD-009: TCP 协议支持

---

## 🔄 变更记录

| 日期 | 变更内容 | 操作人 |
|------|---------|--------|
| 2025-01-21 | 创建技术债务清单，录入 9 个 TODO 项 | AI Code Review |

---

## 📝 备注

### 安全考虑

对于以下功能需要特别注意安全性：
- **TD-006 脚本匹配**: 需要严格的沙箱环境
- **TD-008 gRPC**: 需要防止恶意 proto 定义
- **TD-009 TCP**: 需要防止资源耗尽攻击

### 性能考虑

- **TD-002 正则匹配**: 需要缓存编译后的正则表达式
- **TD-007 WebSocket**: 需要连接数限制
- **TD-008 gRPC**: 需要流量控制

### 兼容性

所有新功能应保持向后兼容，不影响现有规则的运行。

---

**文档维护**: 每次 Sprint 结束后更新此文档  
**审查周期**: 每 2 周审查一次  
**联系方式**: 开发团队

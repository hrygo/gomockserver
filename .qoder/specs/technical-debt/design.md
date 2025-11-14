# gomockserver 技术债务项技术设计文档

## 概述

本文档详细描述了gomockserver项目中五个技术债务项的技术设计方案：
1. TD-001: 实现CIDR IP匹配功能
2. TD-002: 实现正则表达式匹配
3. TD-003: 添加二进制数据支持
4. TD-004: 实现正态分布延迟
5. TD-005: 实现阶梯式延迟策略

## 技术架构

### 整体架构
gomockserver采用分层架构设计，主要包括以下组件：
- API层：处理HTTP请求和响应
- Service层：业务逻辑处理
- Engine层：规则匹配引擎
- Executor层：Mock响应执行器
- Adapter层：协议适配器
- Repository层：数据访问层
- Model层：数据模型定义

### 核心组件交互
1. HTTP请求通过gin框架进入MockService
2. MockService使用HTTPAdapter将请求转换为统一Request模型
3. MatchEngine根据项目ID和环境ID查找匹配规则
4. 匹配成功的规则交给MockExecutor执行生成响应
5. HTTPAdapter将响应转换为HTTP响应返回给客户端

## 组件设计

### 1. TD-001: CIDR IP匹配功能

#### 当前实现分析
当前IP白名单匹配仅支持精确IP地址匹配，不支持CIDR格式的IP段匹配。

#### 技术方案
- 使用标准库`net`包中的`net.ParseCIDR`函数解析CIDR格式
- 遍历CIDR白名单，检查请求IP是否在任一CIDR范围内
- 保持向后兼容性，同时支持精确IP和CIDR格式

#### 实现位置
- 文件：`internal/engine/match_engine.go`
- 函数：`matchIPWhitelist`

#### 数据模型变更
无需变更数据模型，`HTTPMatchCondition.IPWhitelist`字段继续使用字符串数组。

### 2. TD-002: 正则表达式匹配

#### 当前实现分析
当前仅实现了简单字符串匹配，正则表达式匹配功能尚未实现。

#### 技术方案
- 实现基于Go标准库`regexp`的正则表达式匹配
- 支持对Path、Query参数、Headers等进行正则匹配
- 在规则匹配条件中增加正则表达式标识

#### 实现位置
- 文件：`internal/engine/match_engine.go`
- 函数：`regexMatch`

#### 数据模型变更
无需变更数据模型，通过MatchCondition中的配置区分正则匹配和简单匹配。

### 3. TD-003: 二进制数据支持

#### 当前实现分析
当前仅支持文本格式的响应体，二进制数据处理尚未实现。

#### 技术方案
- 在HTTP响应处理中增加二进制数据支持
- 支持Base64编码的二进制数据存储和传输
- 在响应构建时正确处理二进制数据

#### 实现位置
- 文件：`internal/executor/mock_executor.go`
- 函数：`staticResponse`

#### 数据模型变更
无需变更数据模型，通过ContentType字段标识二进制数据类型。

### 4. TD-004: 正态分布延迟

#### 当前实现分析
当前延迟策略仅支持固定延迟和随机延迟，正态分布延迟尚未实现。

#### 技术方案
- 使用Go标准库`math/rand`生成正态分布随机数
- 实现Marsaglia polar method算法生成正态分布随机数
- 支持均值和标准差配置

#### 实现位置
- 文件：`internal/executor/mock_executor.go`
- 函数：`calculateDelay`

#### 数据模型变更
无需变更数据模型，通过DelayConfig中的Mean和StdDev字段配置正态分布参数。

### 5. TD-005: 阶梯式延迟策略

#### 当前实现分析
当前延迟策略不支持基于请求次数的阶梯式延迟。

#### 技术方案
- 实现基于请求计数的阶梯延迟算法
- 支持步长和上限配置
- 使用内存存储或Redis存储请求计数（根据性能要求）

#### 实现位置
- 文件：`internal/executor/mock_executor.go`
- 函数：`calculateDelay`

#### 数据模型变更
无需变更数据模型，通过DelayConfig中的Step和Limit字段配置阶梯参数。

## 数据模型

### HTTPMatchCondition 结构
```go
type HTTPMatchCondition struct {
    Method      interface{}            `json:"method"` // string 或 []string
    Path        string                 `json:"path"`
    Query       map[string]string      `json:"query,omitempty"`
    Headers     map[string]string      `json:"headers,omitempty"`
    Body        map[string]interface{} `json:"body,omitempty"`
    IPWhitelist []string               `json:"ip_whitelist,omitempty"`
}
```

### DelayConfig 结构
```go
type DelayConfig struct {
    Type   string `bson:"type" json:"type"` // fixed, random, normal, step
    Min    int    `bson:"min,omitempty" json:"min,omitempty"`
    Max    int    `bson:"max,omitempty" json:"max,omitempty"`
    Fixed  int    `bson:"fixed,omitempty" json:"fixed,omitempty"`
    Mean   int    `bson:"mean,omitempty" json:"mean,omitempty"`
    StdDev int    `bson:"std_dev,omitempty" json:"std_dev,omitempty"`
    Step   int    `bson:"step,omitempty" json:"step,omitempty"`
    Limit  int    `bson:"limit,omitempty" json:"limit,omitempty"`
}
```

## API规范

### 规则创建API
```http
POST /api/projects/{projectID}/environments/{environmentID}/rules
Content-Type: application/json

{
  "name": "示例规则",
  "protocol": "HTTP",
  "match_type": "Simple",
  "priority": 100,
  "enabled": true,
  "match_condition": {
    "method": "GET",
    "path": "/api/test",
    "ip_whitelist": ["192.168.1.0/24", "10.0.0.1"]
  },
  "response": {
    "type": "Static",
    "delay": {
      "type": "normal",
      "mean": 1000,
      "std_dev": 200
    },
    "content": {
      "status_code": 200,
      "headers": {
        "Content-Type": "application/json"
      },
      "body": "eyJtc2ciOiJoZWxsbyJ9",
      "content_type": "Binary"
    }
  }
}
```

### 规则更新API
```http
PUT /api/rules/{ruleID}
Content-Type: application/json

{
  "match_condition": {
    "method": ["GET", "POST"],
    "path": "/api/.*",
    "ip_whitelist": ["192.168.0.0/16"]
  },
  "response": {
    "delay": {
      "type": "step",
      "fixed": 100,
      "step": 50,
      "limit": 2000
    }
  }
}
```

## 错误处理策略

### CIDR解析错误
- 当CIDR格式无效时，记录警告日志并跳过该条目
- 不影响其他白名单条目的正常处理

### 正则表达式编译错误
- 编译失败时返回明确错误信息
- 在规则创建/更新时进行预编译验证

### 二进制数据处理错误
- Base64解码失败时返回400错误
- 提供详细的错误信息帮助用户排查问题

### 延迟计算错误
- 参数配置错误时使用默认值或返回错误
- 确保不会因延迟计算异常导致服务不可用

## 测试策略

### 单元测试覆盖
1. CIDR IP匹配功能测试
   - 精确IP匹配测试
   - CIDR格式IP匹配测试
   - 混合格式IP白名单测试
   - 边界条件测试

2. 正则表达式匹配测试
   - 简单正则表达式测试
   - 复杂正则表达式测试
   - 性能测试
   - 错误处理测试

3. 二进制数据支持测试
   - Base64编码数据测试
   - 大文件二进制数据测试
   - 不同内容类型测试

4. 正态分布延迟测试
   - 分布统计测试
   - 参数边界测试
   - 性能影响测试

5. 阶梯式延迟测试
   - 计数准确性测试
   - 步长计算测试
   - 上限控制测试

### 集成测试覆盖
1. 完整规则匹配流程测试
2. 多种匹配类型组合测试
3. 延迟策略与匹配引擎集成测试
4. 性能压力测试

## 实现注意事项

### 性能考虑
1. CIDR匹配优化：预编译CIDR块以提高匹配效率
2. 正则表达式优化：缓存编译后的正则表达式对象
3. 延迟策略优化：避免阻塞主线程，合理使用goroutine
4. 内存使用优化：及时释放临时对象，避免内存泄漏

### 安全考虑
1. 正则表达式安全：防止ReDoS攻击，限制正则表达式复杂度
2. 二进制数据安全：限制文件大小，防止恶意大文件上传
3. IP白名单安全：验证CIDR格式有效性，防止无效配置

### 兼容性考虑
1. 向后兼容：确保新功能不影响现有规则的正常工作
2. 配置兼容：保持API接口一致性，新增字段可选
3. 数据兼容：不修改现有数据结构，通过扩展方式实现

## 实施计划

### 优先级排序
1. **高优先级**：TD-001(CIDR IP匹配)、TD-002(正则表达式匹配)
   - 这些是核心匹配功能，对规则灵活性至关重要

2. **中优先级**：TD-004(正态分布延迟)、TD-005(阶梯式延迟)
   - 增强Mock服务的真实性和测试能力

3. **低优先级**：TD-003(二进制数据支持)
   - 特定场景需求，可后续完善

### 依赖关系
- TD-002依赖于TD-001的基础IP匹配框架
- TD-004和TD-005都依赖于延迟计算框架的扩展
- 所有功能都可以独立实现，无强制依赖关系

### 风险评估
1. 正则表达式性能风险：复杂正则可能导致性能下降
2. 内存使用风险：大量规则可能增加内存消耗
3. 并发安全风险：延迟计数器需要考虑并发访问
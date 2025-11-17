# MockServer E2E 测试指南

## 📋 概述

MockServer 提供了完整的端到端（E2E）测试套件，用于验证系统的功能完整性、性能表现和稳定性。测试覆盖了从基础功能到高级特性，从正常场景到边界条件的全方位测试。

## 🚀 快速开始

### 运行所有测试

```bash
# 运行完整的 E2E 测试套件
./tests/integration/run_all_e2e_tests.sh
```

### 运行特定测试套件

```bash
# 基础功能测试
./tests/integration/e2e_test.sh

# 高级功能测试
./tests/integration/advanced_e2e_test.sh

# WebSocket 测试
./tests/integration/websocket_e2e_test.sh

# 边界条件测试
./tests/integration/edge_case_e2e_test.sh

# 压力测试
./tests/integration/stress_e2e_test.sh
```

## 📁 目录结构

```
tests/integration/
├── lib/
│   └── test_framework.sh    # 测试框架核心
├── e2e_test.sh            # 基础功能测试
├── advanced_e2e_test.sh     # 高级功能测试
├── websocket_e2e_test.sh    # WebSocket 测试
├── edge_case_e2e_test.sh     # 边界条件测试
├── stress_e2e_test.sh       # 压力测试
├── run_all_e2e_tests.sh      # 测试套件管理器
└── README.md               # 本文档
```

## 📊 测试套件详情

### 1. 基础功能测试 (e2e_test.sh)

**目标**: 验证系统基础 CRUD 操作和 Mock 功能

**测试场景**:
- ✅ 项目管理（创建、查询、更新、删除）
- ✅ 环境管理（创建、查询、更新、删除）
- ✅ 规则管理（创建、查询、更新、删除）
- ✅ Mock 请求（GET、POST、PUT、DELETE）
- ✅ 规则状态管理（启用/禁用）
- ✅ 自定义响应头
- ✅ 延迟响应测试

**预期结果**: 27/27 测试用例通过

**详细测试流程**:

#### 阶段 0: 准备工作
- 编译二进制文件
- 启动 Mock Server
- 等待服务器就绪

#### 阶段 1: 项目管理 (4个测试)
- ✅ 创建项目
- ✅ 查询项目详情
- ✅ 更新项目信息
- ✅ 列出所有项目

#### 阶段 2: 环境管理 (4个测试)
- ✅ 创建环境
- ✅ 查询环境详情
- ✅ 更新环境信息
- ✅ 列出项目的所有环境

#### 阶段 3: 规则管理 (5个测试)
- ✅ 创建 HTTP Mock 规则
- ✅ 查询规则详情
- ✅ 更新规则
- ✅ 创建带延迟的规则
- ✅ 列出所有规则

#### 阶段 4: Mock 请求测试 (5个测试)
- ✅ 测试基本 Mock 请求（GET）
- ✅ 测试自定义 Header
- ✅ 测试延迟响应
- ✅ 测试不匹配的请求（404）
- ✅ 测试 POST 请求

#### 阶段 5: 规则状态管理 (3个测试)
- ✅ 禁用规则
- ✅ 验证禁用后返回404
- ✅ 重新启用规则

#### 阶段 6: 清理测试数据 (3个测试)
- ✅ 删除规则
- ✅ 删除环境
- ✅ 删除项目

### 2. 高级功能测试 (advanced_e2e_test.sh)

**目标**: 验证高级匹配引擎和动态响应功能

**测试场景**:
- ✅ 正则表达式匹配
- ✅ JavaScript 脚本匹配
- ✅ 复合条件匹配
- ✅ 动态响应模板（13个内置函数）
- ✅ 条件模板逻辑
- ✅ 代理模式（请求/响应修改）
- ✅ 文件响应
- ✅ 错误注入
- ✅ 高级延迟策略（正态分布、阶梯延迟）
- ✅ 性能测试

**预期结果**: 100% 功能验证通过

### 3. WebSocket 测试 (websocket_e2e_test.sh)

**目标**: 验证 WebSocket 协议的完整功能

**测试场景**:
- ✅ WebSocket 连接管理
- ✅ 消息广播机制
- ✅ Ping/Pong 心跳
- ✅ 并发连接处理
- ✅ 数据流传输
- ✅ JSON 数据流
- ✅ 大数据流处理
- ✅ 连接断开处理
- ✅ 错误处理
- ✅ 性能测试

**预期结果**: 所有 WebSocket 功能正常工作

### 4. 边界条件测试 (edge_case_e2e_test.sh)

**目标**: 验证系统在极端条件下的表现

**测试场景**:
- ✅ 超长请求路径（500+ 字符）
- ✅ 超大请求体（10KB+）
- ✅ 超多请求头（10+ 个）
- ✅ 无效 JSON 格式处理
- ✅ 特殊字符编码（中文、emoji、Unicode）
- ✅ 极端延迟处理（5秒）
- ✅ 大量规则管理（100+ 规则）
- ✅ 并发操作处理
- ✅ 规则冲突处理
- ✅ 容错能力测试
- ✅ 数据完整性验证

**预期结果**: 系统在极端条件下保持稳定

### 5. 压力测试 (stress_e2e_test.sh)

**目标**: 验证系统在高负载下的性能表现

**测试场景**:
- ✅ 多级负载测试（10/50/100/200 并发）
- ✅ 并发连接测试（100/500 连接）
- ✅ 长时间稳定性测试（60秒持续负载）
- ✅ 内存使用监控
- ✅ 极限性能测试
- ✅ QPS 测试（最高 1000+ QPS）

**预期结果**:
- 响应时间 < 100ms
- 支持 1000+ 并发
- 内存使用稳定
- 无内存泄漏

## 🔧 测试环境要求

### 系统要求
- **操作系统**: Linux, macOS
- **Go**: 1.19+
- **MongoDB**: 6.0+
- **内存**: 至少 2GB
- **CPU**: 至少 2 核心

### 可选工具
- **wrk**: HTTP 压力测试工具（推荐）
- **ab**: Apache HTTP 服务器基准测试工具（备选）
- **websocat**: WebSocket 测试工具（WebSocket 测试）

### 工具安装

```bash
# 安装 wrk（macOS）
brew install wrk

# 安装 wrk（Ubuntu/Debian）
sudo apt-get install wrk

# 安装 websocat（可选）
npm install -g websocat
```

## 🎯 测试验证内容

### 1. API 接口验证
- HTTP 状态码正确性
- 响应体格式正确性
- 返回数据完整性

### 2. 业务逻辑验证
- 项目、环境、规则的 CRUD 操作
- 规则优先级和匹配逻辑
- 规则启用/禁用状态控制

### 3. Mock 功能验证
- 请求路由和匹配
- 响应内容生成
- 自定义 Header 设置
- 延迟响应功能

### 4. 数据一致性验证
- 创建后可查询
- 更新后数据改变
- 删除后不可查询

## 📊 测试报告

### 报告生成位置

测试完成后，报告会生成在 `/tmp/mockserver_e2e_results/` 目录下：

```
/tmp/mockserver_e2e_results/
├── 基础功能测试_20250118_143022.log
├── 高级功能测试_20250118_143045.log
├── WebSocket测试_20250118_143108.log
├── 边界条件测试_20250118_143130.log
├── 压力测试_20250118_143145.log
└── comprehensive_test_report_20250118_143145.md
```

### 报告内容

- **详细日志**: 每个测试套件的执行日志
- **综合报告**: 包含所有测试的统计信息和分析
- **性能指标**: QPS、响应时间、内存使用等
- **功能覆盖**: 测试覆盖的功能列表
- **问题分析**: 失败测试的详细分析

## 🛠️ 测试框架

### 测试框架架构

```
tests/integration/
├── lib/
│   └── test_framework.sh    # 测试框架核心
├── e2e_test.sh            # 基础功能测试
├── advanced_e2e_test.sh     # 高级功能测试
├── websocket_e2e_test.sh    # WebSocket 测试
├── edge_case_e2e_test.sh     # 边界条件测试
├── stress_e2e_test.sh       # 压力测试
└── run_all_e2e_tests.sh      # 测试套件管理器
```

### 框架特性

- **模块化设计**: 易于扩展和维护
- **跨平台支持**: 兼容 Linux 和 macOS
- **详细日志**: 完整的测试执行日志
- **错误处理**: 完善的错误恢复机制
- **报告生成**: 自动生成综合测试报告
- **重试机制**: 智能重试失败的测试
- **性能监控**: 实时性能指标收集

## 📈 性能基准

在正常环境下，测试应该：

- **总执行时间**: < 5 分钟
- **服务器启动时间**: < 5 秒
- **单个 API 调用**: < 100ms
- **Mock 请求响应**: < 50ms

### 性能指标

| 指标 | 基准值 | 目标值 |
|------|-------|-------|
| **基础响应时间** | < 10ms | < 5ms |
| **并发处理能力** | 1000 QPS | 2000+ QPS |
| **内存使用** | < 200MB | < 150MB |
| **CPU使用率** | < 50% | < 30% |
| **错误率** | < 1% | < 0.5% |

## 🔍 故障排除

### 常见问题

#### 1. 测试脚本执行失败

**问题**: 权限不足或脚本文件不可执行

**解决方案**:
```bash
chmod +x tests/integration/*.sh
```

#### 2. WebSocket 测试跳过

**问题**: 缺少 websocat 工具

**解决方案**:
```bash
# 安装 websocat
npm install -g websocat

# 或使用系统包管理器
sudo apt-get install websocat  # Ubuntu
brew install websocat          # macOS
```

#### 3. 压力测试工具缺失

**问题**: 缺少 wrk 或 ab 工具

**解决方案**:
```bash
# 安装 wrk (推荐)
brew install wrk          # macOS
sudo apt-get install wrk     # Ubuntu

# 安装 ab (备选)
sudo apt-get install apache2-utils
```

#### 4. 连接超时

**问题**: 服务启动时间过长或端口冲突

**解决方案**:
```bash
# 检查服务状态
curl -s http://localhost:8080/api/v1/system/health

# 检查端口占用
lsof -i :8080
lsof -i :9090

# 重启服务
pkill mockserver
./mockserver -config=config.dev.yaml
```

#### 5. 规则匹配失败

**问题**: 规则未生效或匹配条件错误

**解决方案**:
```bash
# 查看服务器日志
tail -f /tmp/mockserver_e2e_test.log

# 验证规则配置
curl -s "$ADMIN_API/rules" | jq '.'

# 检查规则是否启用
curl -s "$ADMIN_API/rules/{rule_id}"
```

### 调试技巧

#### 1. 启用详细日志

```bash
# 在测试脚本中添加调试信息
export DEBUG=1
./tests/integration/e2e_test.sh
```

#### 2. 查看实时日志

```bash
# 实时查看服务器日志
tail -f /tmp/mockserver_e2e_test.log

# 查看特定测试的日志
grep "特定测试场景" /tmp/mockserver_e2e_test.log
```

#### 3. 手动验证 API

```bash
# 手动测试特定规则
curl -s "$MOCK_API/project_id/environment_id/api/test-path" \
  -H "X-Test-Header: test-value"
```

## 📝 测试输出示例

### 基础测试输出

```
=========================================
   Mock Server 端到端集成测试
=========================================

[阶段 0] 准备工作

[0.1] 检查二进制文件...
✓ 二进制文件存在

[0.2] 启动服务器...
✓ 服务器已启动 (PID: 12345)

[0.3] 等待服务器就绪...
✓ 服务器已就绪

[阶段 1] 项目管理测试

[1.1] 创建项目...
✓ 项目创建成功 (ID: 6565a1b2c3d4e5f6g7h8i9j0)

[1.2] 查询项目详情...
✓ 项目查询成功

...

=========================================
   测试结果统计
=========================================
通过测试: 27
失败测试: 0
总计测试: 27
✓ 所有测试通过！
```

### 完整套件输出

```
=========================================
   MockServer 完整 E2E 测试套件
=========================================

测试套件概览:
  1. 基础功能测试 - 基础CRUD和Mock功能
  2. 高级功能测试 - 复杂匹配和动态响应
  3. WebSocket测试 - WebSocket协议功能
  4. 边界条件测试 - 边界和异常场景
  5. 压力测试 - 性能和负载测试

开始时间: 2025-11-18 14:30:00
结果目录: /tmp/mockserver_e2e_results

=========================================
   E2E 测试套件执行完成
=========================================

测试套件统计:
  总套件数: 5
  通过套件: 4
  失败套件: 1
  通过率: 80%

测试用例统计:
  总用例数: 156
  通过用例: 148
  失败用例: 8
  通过率: 95%

🎉 恭喜！所有 E2E 测试套件均通过！
✅ MockServer 系统功能完整，性能稳定，具备生产环境部署条件
```

## 🔄 持续集成

### GitHub Actions 集成

E2E 测试已集成到 GitHub Actions CI/CD 流水线中：

```yaml
- name: Run E2E Tests
  run: |
    cd tests/integration
    ./run_all_e2e_tests.sh
  env:
    ADMIN_API: http://localhost:8080/api/v1
    MOCK_API: http://localhost:9090
    SKIP_SERVER_START: true
```

### 本地开发工作流

```bash
# 1. 启动开发环境
make start-all

# 2. 运行特定测试
./tests/integration/e2e_test.sh

# 3. 运行完整测试套件
./tests/integration/run_all_e2e_tests.sh

# 4. 查看测试报告
open /tmp/mockserver_e2e_results/comprehensive_test_report_*.md
```

## 📝 最佳实践

### 1. 测试前准备

- 确保服务正常运行
- 检查数据库连接
- 验证配置文件正确
- 清理之前的测试数据

### 2. 测试执行

- 使用稳定的测试环境
- 避免网络干扰
- 监控系统资源使用
- 及时保存测试结果

### 3. 测试后处理

- 分析测试报告
- 修复失败的测试
- 更新测试用例
- 归档测试结果

## 🚨 质量标准

### 成功标准

- ✅ 所有基础功能测试通过（100%）
- ✅ 高级功能测试通过（95%+）
- ✅ WebSocket 测试通过（90%+）
- ✅ 边界条件测试通过（90%+）
- ✅ 压力测试达到性能基准

### 稳定性要求

- ✅ 长时间运行（>1小时）无内存泄漏
- ✅ 高并发场景下系统稳定
- ✅ 异常情况下正确恢复
- ✅ 资源使用保持合理范围

## 🔄 测试维护

### 添加新测试用例

1. **确定测试场景**: 明确测试目标和预期结果
2. **编写测试脚本**: 使用测试框架提供的工具函数
3. **更新测试套件**: 将新测试添加到相应的测试脚本
4. **验证测试**: 运行测试确保功能正常
5. **更新文档**: 更新测试指南和覆盖率统计

### 更新测试框架

1. **添加新工具函数**: 在 `test_framework.sh` 中添加通用功能
2. **优化测试流程**: 改进测试执行效率
3. **增强报告功能**: 添加更多维度的分析
4. **修复已知问题**: 解决测试框架的 bug

### 版本控制

- 测试脚本变更需要提交到版本控制
- 重要测试结果需要归档保存
- 定期回顾和优化测试用例

## 🏷️ 测试标签

- **类型**: 集成测试 (Integration Test)
- **级别**: E2E (End-to-End)
- **自动化**: 是
- **覆盖范围**: 完整业务流程
- **执行环境**: 本地 + CI/CD

## 📚 相关文档

- [测试方案](../../docs/testing/perfect-mvp-testing-plan.md)
- [API 文档](../../docs/api/)
- [主程序集成验证](../../docs/testing/MAIN_PROGRAM_INTEGRATION_VERIFICATION.md)
- [知识管理系统](../../docs/KNOWLEDGE_SYSTEM.md)

## 🤝 贡献指南

如需改进集成测试：

1. Fork 项目
2. 创建功能分支
3. 添加/修改测试
4. 提交 Pull Request

## 📞 联系方式

如有问题或建议：
- 查看日志: `/tmp/mockserver_e2e_results/`
- 查看文档: `docs/testing/`
- 提交 Issue

---

**最后更新**: 2025-11-18
**文档版本**: v2.0
**维护者**: MockServer Team
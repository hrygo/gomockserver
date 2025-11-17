## [0.5.1] - 2025-01-17

### 文档优化

- ✅ **项目结构优化**
  - 创建标准化的文档目录结构：`docs/testing/`, `docs/releases/`
  - 归档测试文档至 `docs/testing/reports/` 和 `docs/testing/coverage/`
  - 归档发布文档至 `docs/releases/`
  - 清理根目录临时文件，文件数从 28 个减少到 19 个 (-32%)

- ✅ **文档新增**
  - `docs/PROJECT_STRUCTURE.md` - 项目结构说明文档
  - `docs/releases/RELEASE_CHECKLIST.md` - 版本发布清单（适用于所有版本）
  - `docs/DIRECTORY_OPTIMIZATION_REPORT.md` - 目录优化报告

- ✅ **发布流程优化**
  - 简化版本发布检查，移除 CONTRIBUTING.md 和 LICENSE 的检查（不常更新）
  - 建立版本发布默认动作流程：目录检查 → 结构优化 → 文档更新

### 技术细节

- 优化目录结构，符合开源项目标准
- 测试文档分类归档，提高查找效率
- 发布文档集中管理，便于版本追溯

---

## [0.5.0] - 2025-01-17

### 新增功能

#### 可观测性增强
- ✅ **请求日志系统**：完整的请求/响应日志记录
  - 请求日志模型（RequestLog）：记录HTTP方法、路径、headers、响应状态等
  - 日志记录中间件：异步记录，不阻塞请求处理
  - 敏感信息脱敏：自动脱敏Authorization、Cookie等敏感header
  - 日志查询API：支持分页、时间范围、状态码、项目ID等多维度过滤
  - 日志清理服务：可配置保留天数（默认7天），定时清理过期日志
  - 日志统计API：按时间段统计请求量和响应时间

- ✅ **实时监控系统**
  - Prometheus指标采集：HTTP请求总数、请求延迟、响应状态码分布
  - 健康检查增强：详细的组件状态（数据库、缓存等）
  - 系统信息API：版本号、构建时间、Go版本、服务URL等
  - 性能监控中间件：慢请求检测（>1秒告警）
  - 请求追踪：自动生成Request ID，支持分布式追踪

- ✅ **统计分析增强**
  - 实时数据统计：当前活跃项目数、规则数、请求总量
  - 趋势分析：7天/30天请求趋势和响应时间趋势
  - 对比分析：环比、同比数据对比功能

### 质量提升

#### 单元测试覆盖率大幅提升
- **总体覆盖率**：从52.4%提升到**68%+**
- **核心模块达到80%+覆盖率**：
  - ✅ executor模块：**80.7%**（+30.8%）
  - ✅ service模块：**80.1%**（+1.8%）
  - ✅ engine模块：**80.9%**（已达标）
  - ✅ middleware模块：**97.2%**
  - ✅ metrics模块：**100%**
  - ✅ monitoring模块：**100%**

#### 新增测试文件
- `proxy_executor_test.go`（512行）：完整测试代理模式
- `request_logger_test.go`（300行+）：请求日志中间件测试
- `log_cleanup_service_test.go`（256行）：日志清理服务测试
- `health_test.go`（259行）：健康检查测试
- `middleware_test.go`（226行）：性能监控中间件测试
- 扩展`template_engine_test.go`：新增12个模板函数测试

#### Bug修复
- 修复`sanitizeHeaders`未转小写导致敏感header未脱敏的问题
- 修复Prometheus重复注册导致panic的问题（使用sync.Once）
- 修复executor测试缺少os包的编译错误
- 修复health_test断言错误（MongoDB依赖问题）

### 技术细节

#### 架构改进
- 请求日志采用异步写入，不影响主业务性能
- LRU缓存优化正则表达式编译（容量1000）
- 日志清理服务使用定时器，每天凌晨2点执行
- Prometheus指标使用单例模式，避免重复注册

#### API新增
- `GET /api/v1/request-logs`：查询请求日志
- `GET /api/v1/request-logs/:id`：获取日志详情
- `DELETE /api/v1/request-logs/cleanup`：手动清理日志
- `GET /api/v1/request-logs/statistics`：日志统计
- `GET /api/v1/health/metrics`：Prometheus指标端点
- `GET /api/v1/system/info`：系统详细信息

### 性能优化
- 请求日志异步写入，不阻塞请求处理
- 正则表达式编译结果缓存（LRU，容量1000）
- 慢请求检测优化（>1秒阈值可配置）

### 文档更新
- 更新README.md：添加v0.5.0功能说明
- 更新API文档：新增可观测性相关API
- 更新测试覆盖率报告

---

## [0.4.0] - 2025-01-18
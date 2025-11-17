## [0.6.1] - 2025-11-17

### 修复

- ✅ **GitHub Actions CI/CD 修复**
  - 升级 `actions/upload-artifact` 从 v3 到 v4，解决 GitHub 弃用警告
  - 修复 Docker 构建错误：显式复制关键目录（cmd/, internal/, pkg/）确保构建可靠性
  - 更新 `.github/workflows/docker.yml` 和 `.github/workflows/ci.yml`

- ✅ **Docker 构建优化**
  - 修改 Dockerfile 使用显式目录复制，提高构建稳定性
  - 优化 `.dockerignore` 配置，确保关键文件不被排除

- ✅ **Admin Service 路由修复**
  - 修复环境管理 API 路由配置，确保正确嵌套在项目路由下
  - 优化路由组结构，提高代码可读性

### 文档更新

- ✅ **新增文档**
  - `docs/summaries/docker-build-fix-summary.md` - Docker 构建错误修复总结
  - `docs/summaries/github-actions-artifact-upgrade.md` - GitHub Actions 升级修复总结

### 技术细节

- ✅ **依赖管理**
  - 保持所有依赖版本稳定，无破坏性变更
  - 确保向后兼容性

---

## [0.6.0] - 2025-11-17

### 新增功能

#### 前后端分离支持
- ✅ **CORS 中间件**（`internal/middleware/cors.go`）
  - 支持开发环境（localhost:5173）和生产环境（localhost:8080）
  - 预留 Authorization 头部支持（用于 v0.9.0 认证功能）
  - 支持请求追踪 ID（X-Request-ID）
  - 可配置的跨域策略（origins、methods、headers）
  - 完整的单元测试覆盖（245行测试代码）

- ✅ **前端环境配置**
  - `.env.development`：开发环境配置（Vite 代理 /api/v1）
  - `.env.production`：生产环境配置（集成部署）
  - 环境变量：API_BASE_URL、MOCK_BASE_URL、DEBUG、LOG_LEVEL

#### 配置导入导出功能
- ✅ **导出功能**
  - 导出单个项目（含规则、环境）：`GET /api/v1/import-export/projects/:id/export`
  - 批量导出规则：`POST /api/v1/import-export/rules/export`
  - 支持元数据导出（创建时间、更新时间、版本号）
  - 数据格式：JSON，易于版本控制和迁移

- ✅ **导入功能**
  - 智能导入：`POST /api/v1/import-export/import`
  - 数据验证：`POST /api/v1/import-export/validate`
  - **三种冲突策略**：
    - `skip`：跳过已存在的记录（默认，安全）
    - `overwrite`：覆盖已存在的记录
    - `append`：保留已存在的记录，仅导入新记录
  - 完整的结果反馈：成功数、跳过数、失败数、错误详情
  - 支持部分成功（HTTP 207 Multi-Status）

- ✅ **数据验证**
  - 必填字段验证（project.name、rule.name、environment.name）
  - 数据类型验证（priority、enabled、status_code）
  - 引用完整性检查（project_id、environment_id）
  - 重复性检查（基于 name 字段）
  - 详细的验证错误提示

#### 统计分析增强
- ✅ **协议分布统计**
  - 新增 `getProtocolDistribution()` 方法
  - 使用 MongoDB 聚合查询（$group 管道）
  - 返回每种协议的请求计数（HTTP、HTTPS、WebSocket）

- ✅ **Top 项目统计**
  - 新增 `getTopProjects()` 方法
  - 按请求量排序，支持 Top N（默认 Top 10）
  - 计算成功率（status_code 2xx/3xx）
  - 使用复杂聚合查询优化性能

### 构建和部署增强

#### Makefile 优化
- ✅ **前端构建命令**
  - `make build-frontend`：构建前端（npm run build）
  - `make build-fullstack`：构建完整应用（前端+后端）
  - `make build-platforms`：跨平台编译（重命名自 build-all）

- ✅ **Docker 多阶段构建**
  - `make docker-build`：后端 Docker 镜像（仅后端）
  - `make docker-build-full`：完整 Docker 镜像（前端+后端）
  - 新增 `Dockerfile.fullstack`（103行，三阶段构建）
    - Stage 1：前端构建（Node.js 18-alpine）
    - Stage 2：后端构建（Go 1.21-alpine）
    - Stage 3：运行时镜像（Alpine latest）
  - 镜像优化：使用 npm 镜像源、Go 代理加速构建
  - 安全加固：非 root 用户运行、健康检查
  - 版本信息注入：VERSION、BUILD_TIME、GIT_COMMIT

- ✅ **Docker 构建优化**
  - 更新 `.dockerignore`：排除前端 node_modules、dist
  - 多阶段构建减少镜像体积（预计 <50MB）

### 架构改进

#### 循环依赖解决
- ✅ 避免 `api` 和 `service` 包的循环导入
- ✅ 在 `AdminService` 中定义 `ImportExportService` 接口
- ✅ 直接在 `AdminService` 中实现导入导出 HTTP 处理函数
- ✅ 保持代码组织清晰，职责分离

### 文档更新

- ✅ 更新 `README.md`
  - 新增 v0.6.0 功能说明
  - 更新 API 文档（导入导出、统计分析）
  - 更新开发计划（v0.6.0 已完成，v0.7.0-v0.9.0 规划）
  - 调整版本规划（认证功能移至 v0.9.0）

- ✅ 创建 `v0.6.0-backend-implementation-summary.md`（696行）
  - 完整的实现总结
  - API 使用示例
  - 技术亮点说明
  - 测试验证方法

### 技术细节

#### 依赖更新
- ✅ 添加 `github.com/gin-contrib/cors` v1.7.6

#### 测试覆盖
- ✅ CORS 中间件测试：245行，覆盖所有场景
  - GET/POST/OPTIONS 请求测试
  - Headers 验证测试
  - Credentials 测试
  - 预检请求测试

### 版本规划调整

**重要变更**：将用户认证和权限功能降低优先级

- 🗓️ **v0.6.0**（当前版本）：聚焦基础设施和配置管理
- 🗓️ **v0.7.0**：性能优化（Redis 缓存、数据库优化）
- 🗓️ **v0.8.0**：企业级特性（用户认证、权限体系、版本控制）
- 🗓️ **v0.9.0**：协议扩展（gRPC、TCP/UDP）

### 向后兼容性

- ✅ 所有现有 API 保持向后兼容
- ✅ 新增 API 不影响现有功能
- ✅ CORS 中间件预留 Authorization 头部，为 v0.9.0 做准备
- ✅ 导入导出支持元数据，便于版本迁移

---

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
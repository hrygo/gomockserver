# Mock Server 功能测试执行指南

## 执行概要

本文档提供详细的功能测试执行步骤，帮助测试人员快速上手并完成功能测试。

## 测试准备（5分钟）

### 步骤1: 启动Mock Server

**选项A: 使用Docker Compose（推荐）**
```bash
cd /Users/huangzhonghui/aicoding/gomockserver
docker-compose up -d
```

**选项B: 本地运行**
```bash
cd /Users/huangzhonghui/aicoding/gomockserver

# 确保MongoDB正在运行
docker ps | grep mongo

# 编译并运行
go build -o mockserver ./cmd/mockserver
./mockserver -config=config.yaml
```

### 步骤2: 验证服务状态

```bash
# 检查管理API
curl http://localhost:8080/api/v1/system/health

# 预期响应: {"status":"ok"} 或类似的健康状态
```

### 步骤3: 验证测试工具

```bash
# 检查curl
curl --version

# 检查jq（可选，用于JSON格式化）
jq --version

# 如果jq未安装
brew install jq  # macOS
```

## 第一层：交互式自动化测试（30-60分钟）

### 执行步骤

1. **进入测试目录**
   ```bash
   cd /Users/huangzhonghui/aicoding/gomockserver/tests/functional
   ```

2. **运行测试脚本**
   ```bash
   ./functional_test.sh
   ```

3. **环境检查**
   - 脚本会自动检查必要命令和服务状态
   - 如果检查失败，按提示修复问题

4. **选择测试模块**
   ```
   ========================================
       Mock Server 功能测试菜单
   ========================================
   
     1. 系统管理功能测试
     2. 项目管理功能测试
     3. 环境管理功能测试
     4. 规则管理功能测试
     5. Mock服务功能测试
   
     0. 执行完整测试流程
   
     q. 退出
   
   ========================================
   
   请选择测试项 [0-5, q]:
   ```

5. **推荐执行顺序**
   - 首次测试：选择 `0`（执行完整测试流程）
   - 单模块测试：选择 `1-5`
   
6. **测试交互流程**
   
   对于每个测试用例：
   
   a. **阅读测试说明**
   ```
   [SYS-001 & SYS-002: 健康检查测试]
   
   测试目的: 验证健康检查接口正常响应且返回正确状态
   测试步骤:
     1. 调用健康检查API
     2. 验证HTTP状态码为200
     3. 验证响应包含'ok'或'healthy'状态
   预期结果: 返回200状态码，响应体包含健康状态信息
   
   是否执行此测试 [y/N]:
   ```
   
   b. **确认执行**
   - 输入 `y` 执行测试
   - 输入 `n` 跳过测试
   
   c. **观察执行结果**
   ```
   正在执行测试...
   
   响应内容:
   {
     "status": "ok",
     "timestamp": "2025-11-14T17:00:00Z"
   }
   
   ✓ 健康检查接口返回正常
   ```
   
   d. **确认测试结果**
   ```
   请确认测试结果：SYS-001 & SYS-002: 健康检查测试
     Y - 通过
     N - 失败
     S - 跳过
   请输入 [Y/N/S]:
   ```
   - 输入 `Y` 如果测试符合预期
   - 输入 `N` 如果发现问题
   - 输入 `S` 如果不确定或需要稍后验证

7. **测试完成**
   
   所有测试执行完毕后：
   
   ```
   ========================================
       测试清理
   ========================================
   
   是否清理测试数据 [y/N]:
   ```
   
   - 输入 `y`: 删除测试创建的项目、环境、规则
   - 输入 `n`: 保留数据用于进一步验证

8. **生成测试报告**
   
   ```
   是否生成测试报告 [y/N]: y
   测试报告已生成: ./reports/functional_test_report_20251114_170530.md
   
   是否生成HTML报告 [y/N]: y
   HTML报告已生成: ./reports/functional_test_report_20251114_170530.html
   ```

9. **查看测试结果**
   
   ```bash
   # 查看Markdown报告
   cat reports/functional_test_report_*.md
   
   # 在浏览器中打开HTML报告
   open reports/functional_test_report_*.html
   ```

### 预期时长

| 测试模块 | 预计时长 |
|---------|---------|
| 系统管理测试 | 3-5分钟 |
| 项目管理测试 | 5-8分钟 |
| 环境管理测试 | 5-8分钟 |
| 规则管理测试 | 10-15分钟 |
| Mock服务测试 | 10-15分钟 |
| **完整流程** | **30-50分钟** |

## 第二层：手工检查清单测试（2-4小时）

### 执行步骤

1. **打开检查清单**
   ```bash
   # 使用编辑器打开
   code functional_test_checklist.md
   # 或
   vim functional_test_checklist.md
   ```

2. **填写文档信息**
   - 测试人员姓名
   - 测试环境信息
   - 测试开始时间

3. **按优先级执行测试**
   
   **P0测试（必须100%通过）**:
   - 重点测试核心功能
   - 仔细验证每个步骤
   - 详细记录任何异常
   
   **P1测试（建议90%以上通过）**:
   - 测试重要功能和边界场景
   - 记录用户体验问题
   
   **P2测试（建议80%以上通过）**:
   - 测试极端场景和优化项

4. **记录测试结果**
   
   对于每个测试项：
   
   | 测试项编号 | 实际结果 | 通过状态 | 备注 |
   |-----------|---------|---------|------|
   | PRJ-007 | 返回错误提示"项目名称已存在" | **通过** | 错误信息清晰 |
   | PRJ-008 | 返回400错误，提示"名称不能为空" | **通过** | - |
   | MOCK-024 | 100个并发请求全部成功响应 | **通过** | 响应时间稳定 |

5. **记录缺陷**
   
   发现问题时在"缺陷记录"表中记录：
   
   | 缺陷ID | 发现日期 | 测试项 | 缺陷描述 | 严重程度 |
   |-------|---------|--------|---------|---------|
   | BUG-20251114-001 | 2025-11-14 | PRJ-011 | 删除包含环境的项目时未提示，直接删除 | P1 |

6. **更新统计信息**
   
   完成测试后更新：
   - 已执行数量
   - 通过数量
   - 失败数量
   - 通过率

7. **填写测试结论**
   
   根据测试结果填写：
   - 测试是否通过
   - 主要发现
   - 需要修复的问题
   - 是否建议发布

### 测试技巧

1. **使用curl进行API测试**
   ```bash
   # 创建项目（名称为空）
   curl -X POST http://localhost:8080/api/v1/projects \
     -H "Content-Type: application/json" \
     -d '{"name":"","workspace_id":"test"}'
   
   # 预期：返回400错误
   ```

2. **使用Postman（如果可用）**
   - 导入API集合
   - 批量执行测试
   - 保存测试结果

3. **使用MongoDB Compass验证数据**
   ```bash
   # 连接MongoDB
   mongo mockserver
   
   # 查询项目
   db.projects.find({name: "TestProject"})
   ```

## 第三层：探索性测试（1-2小时）

### 执行步骤

1. **选择测试会话主题**
   - 新手用户首次使用
   - 高级用户复杂场景
   - 异常操作路径
   - 性能压力测试
   - 数据一致性验证

2. **复制模板文件**
   ```bash
   cp exploratory_test_template.md exploratory_test_20251114_session1.md
   ```

3. **填写会话信息**
   - 会话ID、测试人员、时间、主题

4. **自由探索**
   
   示例探索路径：
   
   **会话：新手用户首次使用**
   ```
   操作1 (17:00): 访问系统，尝试创建第一个项目
   - 预期：顺利创建
   - 实际：不知道workspace_id应该填什么
   - 观察：缺少字段说明，新手困惑
   
   操作2 (17:05): 创建环境
   - 预期：成功创建
   - 实际：base_url填错格式，系统接受了
   - 观察：缺少URL格式验证
   
   操作3 (17:10): 创建规则
   - 预期：配置Mock响应
   - 实际：match_condition配置复杂，不知道如何填写
   - 观察：需要更多示例和文档
   ```

5. **记录发现的问题**
   - 功能缺陷
   - 性能问题
   - 用户体验问题
   - 文档问题

6. **提出改进建议**
   - 功能改进
   - 性能优化
   - 交互优化
   - 文档完善

## 测试完成与报告

### 测试完成标准

确认以下条件：

- [ ] 所有P0测试用例已执行并通过
- [ ] P1测试用例通过率≥90%
- [ ] P2测试用例通过率≥80%
- [ ] 无P0级别缺陷
- [ ] P1级别缺陷≤2个且有解决方案
- [ ] 测试报告已生成并审核

### 提交测试报告

1. **收集所有测试产出物**
   ```bash
   # 测试报告
   reports/functional_test_report_*.md
   reports/functional_test_report_*.html
   
   # 手工测试检查清单
   functional_test_checklist.md
   
   # 探索性测试记录
   exploratory_test_*.md
   
   # 测试日志
   /tmp/functional_test_*.log
   ```

2. **归档到测试目录**
   ```bash
   # 创建归档目录（按日期）
   mkdir -p /Users/huangzhonghui/aicoding/gomockserver/docs/testing/functional/20251114
   
   # 复制测试产出物
   cp functional_test_checklist.md docs/testing/functional/20251114/
   cp exploratory_test_*.md docs/testing/functional/20251114/
   cp reports/* docs/testing/functional/20251114/
   ```

3. **提交代码库**
   ```bash
   git add docs/testing/functional/20251114/
   git commit -m "Add functional test report for 2025-11-14"
   ```

## 常见问题处理

### 问题1: 测试中断

**场景**: 测试执行到一半需要中断

**解决方案**:
- 测试脚本支持随时退出（Ctrl+C）
- 下次运行时可以重新开始
- 测试数据可能需要手动清理

### 问题2: 测试数据残留

**场景**: 多次测试后数据库中有很多测试数据

**解决方案**:
```bash
# 方式1: 通过API删除
curl -X DELETE http://localhost:8080/api/v1/projects/<project_id>

# 方式2: 通过MongoDB清理
mongo mockserver --eval '
  db.projects.deleteMany({workspace_id: /test|functional/i});
  db.environments.deleteMany({});
  db.rules.deleteMany({});
'
```

### 问题3: 服务响应慢

**场景**: API调用超时或响应很慢

**解决方案**:
1. 检查MongoDB状态
2. 检查系统资源
3. 增加超时时间：`export API_TIMEOUT=30`

## 下一步

测试完成后：

1. **回归测试**: 缺陷修复后执行回归测试
2. **性能测试**: 参考 `tests/performance/` 执行性能测试
3. **发布评审**: 准备发布前评审会议

---

**文档版本**: 1.0  
**最后更新**: 2025-11-14  
**维护者**: AI Agent

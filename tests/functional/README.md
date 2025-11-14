# Mock Server 功能测试套件

## 概述

本目录包含 Mock Server 的完整功能测试套件，采用**人机交互**的方式进行测试，结合自动化执行和手工验证，确保系统功能的完整性和正确性。

## 测试层级

### 第一层：交互式自动化测试
- **执行方式**: 运行交互式测试脚本
- **交互形式**: 用户选择测试场景、确认执行、验证结果
- **覆盖范围**: 核心业务流程（系统管理、项目管理、环境管理、规则管理、Mock服务）

### 第二层：手工功能测试
- **执行方式**: 使用测试检查清单
- **交互形式**: 用户按清单逐项执行并记录
- **覆盖范围**: 边界场景、异常场景、复杂业务场景

### 第三层：探索性测试
- **执行方式**: 自由探索
- **交互形式**: 用户自由操作、记录问题
- **覆盖范围**: 用户体验、隐藏缺陷、性能问题

## 目录结构

```
tests/functional/
├── functional_test.sh              # 主测试脚本（交互式）
├── test_cases/                     # 测试用例脚本
│   ├── system_tests.sh             # 系统管理测试
│   ├── project_tests.sh            # 项目管理测试
│   ├── environment_tests.sh        # 环境管理测试
│   ├── rule_tests.sh               # 规则管理测试
│   └── mock_tests.sh               # Mock服务测试
├── lib/                            # 公共库
│   ├── common.sh                   # 通用函数（颜色、日志、统计）
│   ├── api_client.sh               # API调用封装
│   └── report_generator.sh         # 报告生成器
├── data/                           # 测试数据（待扩展）
├── reports/                        # 测试报告输出目录
├── functional_test_checklist.md    # 手工测试检查清单
├── exploratory_test_template.md    # 探索性测试模板
└── README.md                       # 本文档
```

## 快速开始

### 前置条件

1. **确保 Mock Server 正在运行**
   ```bash
   # 方式1：使用Docker Compose
   cd /path/to/gomockserver
   docker-compose up -d
   
   # 方式2：直接运行
   ./mockserver -config=config.yaml
   ```

2. **验证服务可访问**
   ```bash
   curl http://localhost:8080/api/v1/system/health
   ```

3. **安装必要工具**
   - `curl`: HTTP请求工具
   - `jq`: JSON处理工具（推荐）
   
   ```bash
   # macOS
   brew install jq
   
   # Linux
   sudo apt-get install jq
   ```

### 执行交互式自动化测试

```bash
# 进入测试目录
cd /Users/huangzhonghui/aicoding/gomockserver/tests/functional

# 运行主测试脚本
./functional_test.sh
```

**测试流程**：
1. 脚本自动检查环境（服务状态、必要工具）
2. 显示测试菜单，选择要执行的测试模块
3. 对每个测试场景：
   - 显示测试目的、步骤和预期结果
   - 询问是否执行
   - 自动调用API并显示响应
   - 用户确认测试结果（Y/N/S）
4. 测试结束后生成测试报告

**菜单选项**：
- `1`: 系统管理功能测试
- `2`: 项目管理功能测试
- `3`: 环境管理功能测试
- `4`: 规则管理功能测试
- `5`: Mock服务功能测试
- `0`: 执行完整测试流程
- `q`: 退出

### 使用手工测试检查清单

1. **打开检查清单文件**
   ```bash
   # 使用Markdown编辑器打开
   code functional_test_checklist.md
   # 或
   vim functional_test_checklist.md
   ```

2. **按优先级执行测试**
   - P0（核心功能）→ P1（重要功能）→ P2（一般功能）
   - 填写"实际结果"、"通过状态"、"备注"列

3. **记录缺陷**
   - 在"缺陷记录"表格中记录发现的问题

4. **统计结果**
   - 更新"测试统计"部分
   - 填写"测试结论"和"测试建议"

### 进行探索性测试

1. **复制模板文件**
   ```bash
   cp exploratory_test_template.md exploratory_test_$(date +%Y%m%d).md
   ```

2. **填写会话信息**
   - 会话ID、测试人员、时间、主题

3. **记录探索过程**
   - 按时间顺序记录每个操作
   - 记录观察和想法

4. **总结发现**
   - 记录发现的问题
   - 提出改进建议

## 测试报告

### 自动生成的报告

运行交互式测试后，会自动生成以下报告：

- **Markdown报告**: `reports/functional_test_report_<timestamp>.md`
- **HTML报告**: `reports/functional_test_report_<timestamp>.html`（可选）

### 报告内容

- 测试概要（执行时间、环境信息、测试人员）
- 测试统计（总数、通过数、失败数、通过率）
- 测试详情（每个场景的执行结果）
- 失败详情（失败场景的日志和错误信息）
- 下一步建议

## 测试数据管理

### 自动清理

交互式测试脚本会在测试结束时询问是否清理测试数据：
- 选择"是"：删除测试创建的项目、环境、规则
- 选择"否"：保留数据，便于问题排查

### 手动清理

如果需要手动清理测试数据：

```bash
# 删除项目（会级联删除关联的环境和规则）
curl -X DELETE http://localhost:8080/api/v1/projects/<project_id>

# 或通过MongoDB直接清理
mongo mockserver --eval "db.projects.deleteMany({workspace_id: 'functional-test'})"
```

## 配置

### 环境变量

可以通过环境变量自定义配置：

```bash
# API地址
export ADMIN_API="http://localhost:8080/api/v1"
export MOCK_API="http://localhost:9090"

# 超时时间（秒）
export API_TIMEOUT=10

# 运行测试
./functional_test.sh
```

### 日志文件

测试日志自动保存到：
```
/tmp/functional_test_<timestamp>.log
```

可以查看详细的执行日志：
```bash
tail -f /tmp/functional_test_*.log
```

## 测试用例说明

### 系统管理测试（system_tests.sh）

- SYS-001 & SYS-002: 健康检查测试
- SYS-003 & SYS-004: 版本信息测试
- SYS-005: 服务启动时间测试（手工）
- SYS-006: 服务重启后数据完整性测试（手工）

### 项目管理测试（project_tests.sh）

- PRJ-001 & PRJ-002: 创建项目测试
- PRJ-003: 查询项目详情测试
- PRJ-004: 更新项目信息测试

### 环境管理测试（environment_tests.sh）

- ENV-001 & ENV-002: 创建环境测试
- ENV-003: 查询环境详情测试

### 规则管理测试（rule_tests.sh）

- RULE-001 & RULE-002: 创建HTTP规则测试
- RULE-007: 启用规则测试
- RULE-008: 禁用规则测试
- RULE-009: 创建带延迟的规则测试

### Mock服务测试（mock_tests.sh）

- MOCK-001: GET请求Mock响应测试
- MOCK-006: 响应延迟功能测试
- MOCK-012: 未匹配规则返回404测试

## 扩展测试

### 添加新的测试用例

1. **在相应的测试用例文件中添加函数**
   ```bash
   # 编辑 test_cases/project_tests.sh
   test_prj_new_feature() {
       subtitle "PRJ-XXX: 新功能测试"
       # 测试逻辑
   }
   ```

2. **在测试套件函数中调用**
   ```bash
   run_project_tests() {
       # ... 现有测试
       test_prj_new_feature
   }
   ```

### 创建新的测试模块

1. **创建测试文件**
   ```bash
   touch test_cases/new_module_tests.sh
   chmod +x test_cases/new_module_tests.sh
   ```

2. **实现测试函数**
   ```bash
   #!/bin/bash
   
   run_new_module_tests() {
       title "新模块功能测试"
       # 测试实现
   }
   
   export -f run_new_module_tests
   ```

3. **在主脚本中加载**
   ```bash
   # 编辑 functional_test.sh
   source "$SCRIPT_DIR/test_cases/new_module_tests.sh"
   ```

## 常见问题

### 1. 服务连接失败

**问题**: 提示"管理API服务不可访问"

**解决方案**:
```bash
# 检查服务是否运行
docker-compose ps
# 或
ps aux | grep mockserver

# 检查端口是否正确
lsof -i :8080
lsof -i :9090

# 查看服务日志
docker-compose logs mockserver
```

### 2. jq 命令未找到

**问题**: JSON响应格式化失败

**解决方案**:
```bash
# 安装 jq
brew install jq  # macOS
sudo apt-get install jq  # Linux

# 或者忽略jq，测试仍可正常运行
```

### 3. 权限问题

**问题**: "Permission denied"

**解决方案**:
```bash
chmod +x functional_test.sh
chmod +x lib/*.sh
chmod +x test_cases/*.sh
```

### 4. MongoDB 连接失败

**问题**: 创建数据失败

**解决方案**:
```bash
# 启动MongoDB
docker run -d -p 27017:27017 mongo:6.0

# 或使用 Docker Compose
docker-compose up -d mongodb
```

## 最佳实践

1. **测试前准备**
   - 确保服务运行正常
   - 清理旧的测试数据
   - 准备好MongoDB连接

2. **执行测试**
   - 优先执行P0测试
   - 仔细阅读测试说明
   - 认真验证每个测试结果

3. **问题记录**
   - 详细记录失败场景
   - 保存截图和日志
   - 记录重现步骤

4. **报告编写**
   - 及时生成测试报告
   - 清晰描述发现的问题
   - 提供改进建议

## 贡献指南

欢迎贡献新的测试用例和改进！

1. Fork 项目
2. 创建功能分支
3. 添加测试用例
4. 提交 Pull Request

## 许可证

MIT License

---

**维护者**: AI Agent  
**最后更新**: 2025-11-14  
**版本**: 1.0

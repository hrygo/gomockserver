# 功能测试快速开始指南

## 一分钟快速开始

```bash
# 1. 确保Mock Server正在运行
docker-compose up -d

# 2. 进入测试目录
cd tests/functional

# 3. 运行交互式测试（推荐首次使用）
./functional_test.sh

# 4. 在菜单中选择 "0" 执行完整测试流程

# 5. 按提示确认每个测试结果
#    - Y: 测试通过
#    - N: 测试失败
#    - S: 跳过测试

# 6. 测试完成后生成报告
#    - 选择 "y" 生成测试报告
#    - 查看报告: cat reports/functional_test_report_*.md
```

## 文档导航

| 文档 | 用途 | 适合人群 |
|------|------|---------|
| [README.md](README.md) | 测试套件介绍和使用说明 | 所有用户 |
| [EXECUTION_GUIDE.md](EXECUTION_GUIDE.md) | 详细的分层测试执行指南 | 测试人员 |
| [TEST_IMPLEMENTATION_SUMMARY.md](TEST_IMPLEMENTATION_SUMMARY.md) | 测试方案实施总结 | 项目经理、技术负责人 |
| [functional_test_checklist.md](functional_test_checklist.md) | 手工测试检查清单（50项） | 测试人员 |
| [exploratory_test_template.md](exploratory_test_template.md) | 探索性测试记录模板 | 测试人员 |

## 测试场景选择

### 场景1: 快速验证（5-10分钟）

**目的**: 验证核心功能是否正常

```bash
./functional_test.sh

# 选择菜单项：
# 1 - 系统管理功能测试
# 2 - 项目管理功能测试
# 5 - Mock服务功能测试
```

### 场景2: 完整自动化测试（30-50分钟）

**目的**: 执行所有自动化测试用例

```bash
./functional_test.sh

# 选择菜单项：
# 0 - 执行完整测试流程
```

### 场景3: 全面功能测试（2-4小时）

**目的**: 完整的功能验证，包括边界场景

**步骤**:
1. 执行交互式自动化测试（30-50分钟）
2. 使用手工测试检查清单（1-2小时）
3. 进行探索性测试（30-60分钟）
4. 整理和提交测试报告（10-20分钟）

## 测试前检查

```bash
# 检查服务状态
curl http://localhost:8080/api/v1/system/health
# 预期响应: {"status":"ok"}

# 检查工具
which curl  # 必需
which jq    # 可选，用于JSON格式化

# 检查权限
ls -la functional_test.sh
# 应该有执行权限（-rwxr-xr-x）
```

## 常见问题速查

| 问题 | 解决方案 |
|------|---------|
| 服务连接失败 | `docker-compose up -d` 启动服务 |
| 权限被拒绝 | `chmod +x functional_test.sh lib/*.sh test_cases/*.sh` |
| jq未找到 | `brew install jq` （可选，不影响测试） |
| 端口占用 | `lsof -i :8080` 查看占用进程 |

## 测试报告位置

```bash
# 自动生成的测试报告
ls -lh reports/

# 查看最新报告
cat reports/functional_test_report_*.md | head -50

# 在浏览器中查看HTML报告（如果生成）
open reports/functional_test_report_*.html
```

## 获取帮助

1. **阅读详细文档**: `cat EXECUTION_GUIDE.md`
2. **查看实施总结**: `cat TEST_IMPLEMENTATION_SUMMARY.md`
3. **查看测试日志**: `tail -f /tmp/functional_test_*.log`

---

**提示**: 首次使用建议先阅读 [EXECUTION_GUIDE.md](EXECUTION_GUIDE.md) 了解详细流程。

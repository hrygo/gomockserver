# Mock Server 单元测试总结报告

**生成时间**: 2025-11-14 10:16:00  
**测试结果**: PASS

## 📊 测试统计

| 指标 | 数值 |
|------|------|
| 总测试数 | 276 |
| 通过测试 | 92 |
| 源文件数 | 13 |
| 测试文件数 | 11 |

## 📈 覆盖率详情

### 总体覆盖率
```
total:											(statements)			54.9%
```

### 各模块覆盖率

| 模块 | 覆盖率 | 测试文件 |
|------|--------|---------|
| adapter | 96.3% | 1 |
| api | 89.5% | 2 |
| engine | 89.8% | 2 |
| executor | 71.9% | 1 |
| repository | 0.0% | 2 |
| service | 45.6% | 2 |

## 🎯 测试覆盖模块

### adapter

- http_adapter_test.go: 9 个测试函数

### api

- rule_handler_test.go: 7 个测试函数
- project_handler_test.go: 9 个测试函数

### config

- config_test.go: 11 个测试函数

### engine

- match_engine_simple_test.go: 5 个测试函数
- match_engine_test.go: 7 个测试函数

### executor

- mock_executor_test.go: 10 个测试函数

### models

- 无测试文件

### repository

- repository_real_test.go: 6 个测试函数
- repository_test.go: 18 个测试函数

### service

- admin_service_test.go: 7 个测试函数
- mock_service_test.go: 9 个测试函数

## 📁 生成文件

- 覆盖率数据: `/Users/huangzhonghui/aicoding/gomockserver/docs/testing/coverage/unit-coverage-all.out`
- HTML 报告: `/Users/huangzhonghui/aicoding/gomockserver/docs/testing/coverage/unit-coverage-all.html`
- 测试输出: `/Users/huangzhonghui/aicoding/gomockserver/docs/testing/reports/unit_test_output_20251114_101553.txt`
- 覆盖率分析: `/Users/huangzhonghui/aicoding/gomockserver/docs/testing/reports/coverage_analysis_20251114_101553.txt`

## 🔍 低覆盖率文件（< 80%）

```
github.com/gomockserver/mockserver/internal/api/rule_handler.go:99: 77.8%
github.com/gomockserver/mockserver/internal/api/rule_handler.go:191: 55.6%
github.com/gomockserver/mockserver/internal/engine/match_engine.go:29: 73.3%
github.com/gomockserver/mockserver/internal/executor/mock_executor.go:52: 56.7%
github.com/gomockserver/mockserver/internal/executor/mock_executor.go:117: 75.0%
github.com/gomockserver/mockserver/internal/repository/database.go:19: 0.0%
github.com/gomockserver/mockserver/internal/repository/database.go:53: 0.0%
github.com/gomockserver/mockserver/internal/repository/database.go:196: 0.0%
github.com/gomockserver/mockserver/internal/repository/database.go:201: 0.0%
github.com/gomockserver/mockserver/internal/repository/database.go:206: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:29: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:36: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:53: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:75: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:87: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:108: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:126: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:169: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:176: 0.0%
github.com/gomockserver/mockserver/internal/repository/project_repository.go:193: 0.0%
```


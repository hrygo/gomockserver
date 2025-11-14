# 功能测试 报告

## 测试概要

| 项目 | 内容 |
|------|------|
| 报告生成时间 | 2025-11-14 17:28:45 |
| 测试执行时长 | 1m 49s |
| 测试人员 | huangzhonghui |
| 操作系统 | Darwin 25.0.0 |
| Go 版本 | go version go1.25.4 darwin/arm64 |
| MongoDB 版本 |  |

## 测试统计

| 统计项 | 数量 | 占比 |
|-------|------|------|
| **总测试数** | **14** | **100%** |
| ✓ 通过测试 | 13 | 13/14 |
| ✗ 失败测试 | 1 | 1/14 |
| ⊙ 跳过测试 | 0 | 0/14 |
| **通过率** | **92.86%** | - |


## 测试结论

❌ **测试失败** - 发现 1 个失败用例，需要进一步排查和修复。


## 详细日志

完整的测试执行日志请查看：`/tmp/functional_test_20251114_172656.log`

### 日志摘要

```
[2025-11-14 17:28:19] SUCCESS: 延迟规则创建成功，ID: 6916f633feabb8011c44a387
[2025-11-14 17:28:19] SUCCESS: RULE-009: 创建带延迟的规则测试
[2025-11-14 17:28:19] TEST PASSED: RULE-009: 创建带延迟的规则测试
[2025-11-14 17:28:19] SUBTITLE: RULE-008: 禁用规则测试
[2025-11-14 17:28:19] INFO: 正在禁用规则...
[2025-11-14 17:28:19] API POST: http://localhost:8080/api/v1/rules/6916f633feabb8011c44a386/disable
[2025-11-14 17:28:19] Request Data: {}
[2025-11-14 17:28:19] Response Code: 200
[2025-11-14 17:28:19] Response Body: {"message":"Rule disabled successfully"}
[2025-11-14 17:28:19] SUCCESS: RULE-008: 禁用规则测试
[2025-11-14 17:28:19] TEST PASSED: RULE-008: 禁用规则测试
[2025-11-14 17:28:19] SUBTITLE: RULE-007: 启用规则测试
[2025-11-14 17:28:19] INFO: 正在启用规则...
[2025-11-14 17:28:19] API POST: http://localhost:8080/api/v1/rules/6916f633feabb8011c44a386/enable
[2025-11-14 17:28:19] Request Data: {}
[2025-11-14 17:28:19] Response Code: 200
[2025-11-14 17:28:19] Response Body: {"message":"Rule enabled successfully"}
[2025-11-14 17:28:19] SUCCESS: RULE-007: 启用规则测试
[2025-11-14 17:28:19] TEST PASSED: RULE-007: 启用规则测试
[2025-11-14 17:28:19] TITLE: Mock服务功能测试
[2025-11-14 17:28:20] SUBTITLE: MOCK-001: GET请求Mock响应测试
[2025-11-14 17:28:20] INFO: 正在发送Mock请求...
[2025-11-14 17:28:20] MOCK REQUEST: GET http://localhost:9090/6916f625feabb8011c44a384/6916f631feabb8011c44a385/api/users/1
[2025-11-14 17:28:20] API GET: http://localhost:9090/6916f625feabb8011c44a384/6916f631feabb8011c44a385/api/users/1
[2025-11-14 17:28:20] Response Code: 404
[2025-11-14 17:28:20] Response Body: {"error": "No matching rule found"}
[2025-11-14 17:28:20] ERROR: 期望状态码 ��实际 404
[2025-11-14 17:28:20] WARNING: Mock响应内容可能不正确
[2025-11-14 17:28:30] ERROR: MOCK-001: GET请求Mock响应测试
[2025-11-14 17:28:30] TEST FAILED: MOCK-001: GET请求Mock响应测试
[2025-11-14 17:28:30] TITLE: 测试清理
[2025-11-14 17:28:39] INFO: 正在清理测试数据...
[2025-11-14 17:28:39] API DELETE: http://localhost:8080/api/v1/rules/6916f633feabb8011c44a386
[2025-11-14 17:28:39] Response Code: 200
[2025-11-14 17:28:39] Response Body: {"message":"Rule deleted successfully"}
[2025-11-14 17:28:39] SUCCESS: 规则 6916f633feabb8011c44a386 已删除
[2025-11-14 17:28:39] API DELETE: http://localhost:8080/api/v1/rules/6916f633feabb8011c44a387
[2025-11-14 17:28:39] Response Code: 200
[2025-11-14 17:28:39] Response Body: {"message":"Rule deleted successfully"}
[2025-11-14 17:28:39] SUCCESS: 规则 6916f633feabb8011c44a387 已删除
[2025-11-14 17:28:39] API DELETE: http://localhost:8080/api/v1/environments/6916f631feabb8011c44a385
[2025-11-14 17:28:39] Response Code: 200
[2025-11-14 17:28:39] Response Body: {"message":"Environment deleted successfully"}
[2025-11-14 17:28:39] SUCCESS: 环境 6916f631feabb8011c44a385 已删除
[2025-11-14 17:28:39] API DELETE: http://localhost:8080/api/v1/projects/6916f625feabb8011c44a384
[2025-11-14 17:28:39] Response Code: 200
[2025-11-14 17:28:39] Response Body: {"message":"Project deleted successfully"}
[2025-11-14 17:28:39] SUCCESS: 项目 6916f625feabb8011c44a384 已删除
[2025-11-14 17:28:39] TITLE: 测试结果统计
[2025-11-14 17:28:39] TEST SUMMARY: Total=14, Passed=13, Failed=1, Skipped=0, Rate=92.86%
```


## 下一步建议

1. 查看测试日志文件，定位失败原因
2. 修复发现的缺陷
3. 执行回归测试验证修复效果
4. 更新相关文档


---

**报告生成器版本**: 1.0  
**报告文件路径**: /Users/huangzhonghui/aicoding/gomockserver/tests/functional/reports/functional_test_report_20251114_172845.md

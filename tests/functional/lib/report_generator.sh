#!/bin/bash

# 测试报告生成器
# 生成 Markdown 格式的测试报告

# ==================== 报告生成 ====================

# 生成测试报告
generate_test_report() {
    local report_file="${1:-./reports/functional_test_report_$(date +%Y%m%d_%H%M%S).md}"
    local test_name="${2:-功能测试}"
    
    # 获取统计数据
    local stats=($(get_test_stats))
    local total=${stats[0]}
    local passed=${stats[1]}
    local failed=${stats[2]}
    local skipped=${stats[3]}
    local pass_rate=${stats[4]}
    
    # 获取测试时长
    local duration=$(get_duration)
    local formatted_duration=$(format_duration $duration)
    
    # 获取环境信息
    local os_info=$(uname -s)
    local os_version=$(uname -r)
    local go_version=$(go version 2>/dev/null || echo "未安装")
    local mongo_version=$(mongod --version 2>/dev/null | head -n1 || echo "未安装")
    
    # 创建报告目录
    mkdir -p $(dirname "$report_file")
    
    # 确保使用UTF-8编码生成报告
    export LC_ALL=zh_CN.UTF-8
    export LANG=zh_CN.UTF-8
    
    # 生成报告内容
    cat > "$report_file" << EOF
# $test_name 报告

## 测试概要

| 项目 | 内容 |
|------|------|
| 报告生成时间 | $(date '+%Y-%m-%d %H:%M:%S') |
| 测试执行时长 | $formatted_duration |
| 测试人员 | $(whoami) |
| 操作系统 | $os_info $os_version |
| Go 版本 | $go_version |
| MongoDB 版本 | $mongo_version |

## 测试统计

| 统计项 | 数量 | 占比 |
|-------|------|------|
| **总测试数** | **$total** | **100%** |
| ✓ 通过测试 | $passed | ${passed}/${total} |
| ✗ 失败测试 | $failed | ${failed}/${total} |
| ⊙ 跳过测试 | $skipped | ${skipped}/${total} |
| **通过率** | **${pass_rate}%** | - |

EOF

    # 添加结论
    if [ $failed -eq 0 ] && [ $passed -gt 0 ]; then
        cat >> "$report_file" << EOF

## 测试结论

✅ **测试通过** - 所有测试用例均通过验证，系统功能正常。

EOF
    elif [ $failed -gt 0 ]; then
        cat >> "$report_file" << EOF

## 测试结论

❌ **测试失败** - 发现 $failed 个失败用例，需要进一步排查和修复。

EOF
    else
        cat >> "$report_file" << EOF

## 测试结论

⚠️  **测试未完成** - 未执行任何测试或所有测试均被跳过。

EOF
    fi
    
    # 添加详细日志引用
    if [ -f "$TEST_LOG_FILE" ]; then
        cat >> "$report_file" << EOF

## 详细日志

完整的测试执行日志请查看：\`$TEST_LOG_FILE\`

### 日志摘要

\`\`\`
$(tail -50 "$TEST_LOG_FILE")
\`\`\`

EOF
    fi
    
    # 添加建议
    cat >> "$report_file" << EOF

## 下一步建议

EOF

    if [ $failed -gt 0 ]; then
        cat >> "$report_file" << EOF
1. 查看测试日志文件，定位失败原因
2. 修复发现的缺陷
3. 执行回归测试验证修复效果
4. 更新相关文档

EOF
    elif [ $skipped -gt 0 ]; then
        cat >> "$report_file" << EOF
1. 分析跳过的测试用例，确定是否需要执行
2. 补充完整测试覆盖
3. 执行完整的测试流程

EOF
    else
        cat >> "$report_file" << EOF
1. 继续执行其他测试场景
2. 进行探索性测试
3. 准备发布前的最终验证

EOF
    fi
    
    cat >> "$report_file" << EOF

---

**报告生成器版本**: 1.0  
**报告文件路径**: $report_file
EOF
    
    success "测试报告已生成: $report_file"
    log "Test report generated: $report_file"
    
    echo "$report_file"
}

# 生成简要报告
generate_summary_report() {
    local stats=($(get_test_stats))
    local total=${stats[0]}
    local passed=${stats[1]}
    local failed=${stats[2]}
    local skipped=${stats[3]}
    local pass_rate=${stats[4]}
    
    echo ""
    echo "========================================="
    echo "           测试简要报告"
    echo "========================================="
    echo "总测试数: $total"
    echo "通过: $passed | 失败: $failed | 跳过: $skipped"
    echo "通过率: $pass_rate%"
    echo "========================================="
    echo ""
}

# 生成HTML报告（基础版）
generate_html_report() {
    local report_file="${1:-./reports/functional_test_report_$(date +%Y%m%d_%H%M%S).html}"
    
    local stats=($(get_test_stats))
    local total=${stats[0]}
    local passed=${stats[1]}
    local failed=${stats[2]}
    local skipped=${stats[3]}
    local pass_rate=${stats[4]}
    
    mkdir -p $(dirname "$report_file")
    
    # 确保使用UTF-8编码
    export LC_ALL=zh_CN.UTF-8
    export LANG=zh_CN.UTF-8
    
    cat > "$report_file" << EOF
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>功能测试报告</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 3px solid #4CAF50;
            padding-bottom: 10px;
        }
        .stats {
            display: flex;
            justify-content: space-around;
            margin: 30px 0;
        }
        .stat-box {
            text-align: center;
            padding: 20px;
            border-radius: 8px;
            flex: 1;
            margin: 0 10px;
        }
        .stat-total { background-color: #2196F3; color: white; }
        .stat-pass { background-color: #4CAF50; color: white; }
        .stat-fail { background-color: #f44336; color: white; }
        .stat-skip { background-color: #FF9800; color: white; }
        .stat-number { font-size: 48px; font-weight: bold; }
        .stat-label { font-size: 14px; margin-top: 10px; }
        .progress-bar {
            width: 100%;
            height: 30px;
            background-color: #e0e0e0;
            border-radius: 15px;
            overflow: hidden;
            margin: 20px 0;
        }
        .progress-fill {
            height: 100%;
            background-color: #4CAF50;
            transition: width 0.3s ease;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-weight: bold;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 12px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Mock Server 功能测试报告</h1>
        
        <h2>测试概要</h2>
        <table>
            <tr><th>项目</th><th>内容</th></tr>
            <tr><td>报告生成时间</td><td>$(date '+%Y-%m-%d %H:%M:%S')</td></tr>
            <tr><td>测试人员</td><td>$(whoami)</td></tr>
            <tr><td>操作系统</td><td>$(uname -s) $(uname -r)</td></tr>
        </table>
        
        <h2>测试统计</h2>
        <div class="stats">
            <div class="stat-box stat-total">
                <div class="stat-number">$total</div>
                <div class="stat-label">总测试数</div>
            </div>
            <div class="stat-box stat-pass">
                <div class="stat-number">$passed</div>
                <div class="stat-label">通过</div>
            </div>
            <div class="stat-box stat-fail">
                <div class="stat-number">$failed</div>
                <div class="stat-label">失败</div>
            </div>
            <div class="stat-box stat-skip">
                <div class="stat-number">$skipped</div>
                <div class="stat-label">跳过</div>
            </div>
        </div>
        
        <h2>通过率</h2>
        <div class="progress-bar">
            <div class="progress-fill" style="width: ${pass_rate}%">${pass_rate}%</div>
        </div>
        
        <h2>详细日志</h2>
        <p>完整日志文件: <code>$TEST_LOG_FILE</code></p>
    </div>
</body>
</html>
EOF
    
    success "HTML报告已生成: $report_file"
    echo "$report_file"
}

# ==================== 导出函数 ====================

export -f generate_test_report generate_summary_report generate_html_report

-- 性能测试 Lua 脚本
-- 用于 wrk 的自定义请求脚本

-- 在每个请求前设置必要的 headers
request = function()
    path = "/perf/test"
    headers = {}
    headers["X-Project-ID"] = os.getenv("PROJECT_ID") or "test-project"
    headers["X-Environment-ID"] = os.getenv("ENVIRONMENT_ID") or "test-env"
    headers["Content-Type"] = "application/json"
    return wrk.format("GET", path, headers, nil)
end

-- 响应处理
response = function(status, headers, body)
    if status ~= 200 then
        print("Error: received status " .. status)
    end
end

-- 测试完成后的统计
done = function(summary, latency, requests)
    io.write("=======================================\n")
    io.write("性能测试结果汇总\n")
    io.write("=======================================\n")
    io.write(string.format("总请求数: %d\n", summary.requests))
    io.write(string.format("总时长: %.2f 秒\n", summary.duration / 1000000))
    io.write(string.format("总数据量: %.2f MB\n", summary.bytes / 1024 / 1024))
    io.write(string.format("平均 QPS: %.2f\n", summary.requests / (summary.duration / 1000000)))
    io.write(string.format("平均延迟: %.2f ms\n", latency.mean / 1000))
    io.write(string.format("最大延迟: %.2f ms\n", latency.max / 1000))
    io.write(string.format("错误数: %d\n", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.timeout))
    io.write("=======================================\n")
    io.write("延迟分布:\n")
    io.write(string.format("  50%%: %.2f ms\n", latency:percentile(50) / 1000))
    io.write(string.format("  75%%: %.2f ms\n", latency:percentile(75) / 1000))
    io.write(string.format("  90%%: %.2f ms\n", latency:percentile(90) / 1000))
    io.write(string.format("  95%%: %.2f ms\n", latency:percentile(95) / 1000))
    io.write(string.format("  99%%: %.2f ms\n", latency:percentile(99) / 1000))
    io.write("=======================================\n")
end

package metrics

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	testMetrics *Metrics
	testOnce    sync.Once
)

func getTestMetrics() *Metrics {
	testOnce.Do(func() {
		testMetrics = New()
	})
	return testMetrics
}

func TestNew(t *testing.T) {
	m := getTestMetrics()

	assert.NotNil(t, m)
	assert.NotNil(t, m.HTTPRequestsTotal)
	assert.NotNil(t, m.HTTPRequestDuration)
	assert.NotNil(t, m.HTTPRequestsInFlight)
	assert.NotNil(t, m.WebSocketConnections)
	assert.NotNil(t, m.WebSocketMessagesTotal)
	assert.NotNil(t, m.RuleMatchesTotal)
	assert.NotNil(t, m.RuleMatchDuration)
	assert.NotNil(t, m.RulesTotal)
	assert.NotNil(t, m.GoroutinesCount)
	assert.NotNil(t, m.MemoryUsage)
	assert.NotNil(t, m.CPUUsage)
	assert.NotNil(t, m.DBQueryDuration)
	assert.NotNil(t, m.DBConnectionsTotal)
	assert.NotNil(t, m.ErrorsTotal)
}

func TestMetrics_RecordHTTPRequest(t *testing.T) {
	m := getTestMetrics()

	// 测试正常记录
	m.RecordHTTPRequest("GET", "/api/test", 200, 0.150)
	m.RecordHTTPRequest("POST", "/api/users", 201, 0.250)
	m.RecordHTTPRequest("GET", "/api/error", 500, 0.050)

	// 测试不同状态码
	m.RecordHTTPRequest("GET", "/api/test", 404, 0.010)
	m.RecordHTTPRequest("GET", "/api/test", 301, 0.020)
}

func TestMetrics_RecordRuleMatch(t *testing.T) {
	m := getTestMetrics()

	// 测试匹配成功
	m.RecordRuleMatch("rule-001", "proj-001", true, 0.005)

	// 测试匹配失败
	m.RecordRuleMatch("rule-002", "proj-001", false, 0.003)

	// 测试不同项目
	m.RecordRuleMatch("rule-003", "proj-002", true, 0.008)
}

func TestMetrics_RecordDBQuery(t *testing.T) {
	m := getTestMetrics()

	// 测试不同操作
	m.RecordDBQuery("find", "rules", 0.010)
	m.RecordDBQuery("insert", "projects", 0.015)
	m.RecordDBQuery("update", "environments", 0.012)
	m.RecordDBQuery("delete", "request_logs", 0.008)
}

func TestMetrics_RecordError(t *testing.T) {
	m := getTestMetrics()

	// 测试不同错误类型
	m.RecordError("database_error", "repository")
	m.RecordError("validation_error", "api")
	m.RecordError("timeout_error", "executor")
	m.RecordError("network_error", "adapter")
}

func TestMetrics_WebSocketOperations(t *testing.T) {
	m := getTestMetrics()

	// 测试连接增加
	m.IncrementWSConnections()
	m.IncrementWSConnections()

	// 测试连接减少
	m.DecrementWSConnections()

	// 测试消息记录
	m.RecordWSMessage("send")
	m.RecordWSMessage("receive")
	m.RecordWSMessage("send")
}

func TestMetrics_UpdateSystemMetrics(t *testing.T) {
	m := getTestMetrics()

	// 测试系统指标更新
	m.UpdateSystemMetrics(100, 1024*1024*256, 45.5)
	m.UpdateSystemMetrics(150, 1024*1024*512, 60.2)
}

func TestMetrics_SetRulesTotal(t *testing.T) {
	m := getTestMetrics()

	m.SetRulesTotal(10)
	m.SetRulesTotal(25)
	m.SetRulesTotal(0)
}

func TestMetrics_SetDBConnections(t *testing.T) {
	m := getTestMetrics()

	m.SetDBConnections(5)
	m.SetDBConnections(10)
	m.SetDBConnections(3)
}

func TestMetrics_HTTPRequestsInFlight(t *testing.T) {
	m := getTestMetrics()

	// 测试进行中请求计数
	m.IncrementHTTPRequestsInFlight()
	m.IncrementHTTPRequestsInFlight()
	m.IncrementHTTPRequestsInFlight()

	m.DecrementHTTPRequestsInFlight()
	m.DecrementHTTPRequestsInFlight()
}

func TestMetrics_AllStatusCodes(t *testing.T) {
	m := getTestMetrics()

	// 测试所有状态码类别
	statusCodes := []int{100, 200, 201, 204, 301, 302, 400, 401, 403, 404, 500, 502, 503}

	for _, code := range statusCodes {
		m.RecordHTTPRequest("GET", "/test", code, 0.1)
	}
}

func TestMetrics_ConcurrentOperations(t *testing.T) {
	m := getTestMetrics()

	// 模拟并发操作
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			m.RecordHTTPRequest("GET", "/concurrent", 200, 0.05)
			m.IncrementHTTPRequestsInFlight()
			m.DecrementHTTPRequestsInFlight()
			m.RecordRuleMatch("rule-test", "proj-test", true, 0.001)
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMetrics_EdgeCases(t *testing.T) {
	m := getTestMetrics()

	// 测试边界情况
	m.RecordHTTPRequest("", "", 0, 0)
	m.RecordHTTPRequest("VERYLONGMETHOD", "/very/long/path/that/exceeds/normal/limits", 999, 999.999)

	m.RecordRuleMatch("", "", false, 0)
	m.RecordDBQuery("", "", 0)
	m.RecordError("", "")

	m.UpdateSystemMetrics(0, 0, 0)
	m.SetRulesTotal(0)
	m.SetDBConnections(0)
}

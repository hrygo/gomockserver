package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 系统指标收集器
type Metrics struct {
	// HTTP 请求指标
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// WebSocket 指标
	WebSocketConnections prometheus.Gauge
	WebSocketMessagesTotal *prometheus.CounterVec

	// 规则匹配指标
	RuleMatchesTotal   *prometheus.CounterVec
	RuleMatchDuration  *prometheus.HistogramVec
	RulesTotal         prometheus.Gauge

	// 系统资源指标
	GoroutinesCount prometheus.Gauge
	MemoryUsage     prometheus.Gauge
	CPUUsage        prometheus.Gauge

	// 数据库指标
	DBQueryDuration    *prometheus.HistogramVec
	DBConnectionsTotal prometheus.Gauge

	// 错误指标
	ErrorsTotal *prometheus.CounterVec
}

// New 创建指标收集器
func New() *Metrics {
	return &Metrics{
		// HTTP 请求指标
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mockserver",
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mockserver",
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mockserver",
				Subsystem: "http",
				Name:      "requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		),

		// WebSocket 指标
		WebSocketConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mockserver",
				Subsystem: "websocket",
				Name:      "connections",
				Help:      "Current number of WebSocket connections",
			},
		),
		WebSocketMessagesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mockserver",
				Subsystem: "websocket",
				Name:      "messages_total",
				Help:      "Total number of WebSocket messages",
			},
			[]string{"direction"}, // send, receive
		),

		// 规则匹配指标
		RuleMatchesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mockserver",
				Subsystem: "rule",
				Name:      "matches_total",
				Help:      "Total number of rule matches",
			},
			[]string{"rule_id", "project_id", "matched"},
		),
		RuleMatchDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mockserver",
				Subsystem: "rule",
				Name:      "match_duration_seconds",
				Help:      "Rule matching duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
			},
			[]string{"project_id"},
		),
		RulesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mockserver",
				Subsystem: "rule",
				Name:      "total",
				Help:      "Total number of active rules",
			},
		),

		// 系统资源指标
		GoroutinesCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mockserver",
				Subsystem: "system",
				Name:      "goroutines",
				Help:      "Number of goroutines",
			},
		),
		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mockserver",
				Subsystem: "system",
				Name:      "memory_bytes",
				Help:      "Memory usage in bytes",
			},
		),
		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mockserver",
				Subsystem: "system",
				Name:      "cpu_usage_percent",
				Help:      "CPU usage percentage",
			},
		),

		// 数据库指标
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "mockserver",
				Subsystem: "db",
				Name:      "query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5, 1.0},
			},
			[]string{"operation", "collection"},
		),
		DBConnectionsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "mockserver",
				Subsystem: "db",
				Name:      "connections",
				Help:      "Number of database connections",
			},
		),

		// 错误指标
		ErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "mockserver",
				Subsystem: "errors",
				Name:      "total",
				Help:      "Total number of errors",
			},
			[]string{"type", "component"},
		),
	}
}

// RecordHTTPRequest 记录 HTTP 请求
func (m *Metrics) RecordHTTPRequest(method, path string, status int, duration float64) {
	statusClass := fmt.Sprintf("%dxx", status/100)
	m.HTTPRequestsTotal.WithLabelValues(method, path, statusClass).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordRuleMatch 记录规则匹配
func (m *Metrics) RecordRuleMatch(ruleID, projectID string, matched bool, duration float64) {
	matchedStr := "false"
	if matched {
		matchedStr = "true"
	}
	m.RuleMatchesTotal.WithLabelValues(ruleID, projectID, matchedStr).Inc()
	m.RuleMatchDuration.WithLabelValues(projectID).Observe(duration)
}

// RecordDBQuery 记录数据库查询
func (m *Metrics) RecordDBQuery(operation, collection string, duration float64) {
	m.DBQueryDuration.WithLabelValues(operation, collection).Observe(duration)
}

// RecordError 记录错误
func (m *Metrics) RecordError(errorType, component string) {
	m.ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// IncrementWSConnections WebSocket 连接数增加
func (m *Metrics) IncrementWSConnections() {
	m.WebSocketConnections.Inc()
}

// DecrementWSConnections WebSocket 连接数减少
func (m *Metrics) DecrementWSConnections() {
	m.WebSocketConnections.Dec()
}

// RecordWSMessage 记录 WebSocket 消息
func (m *Metrics) RecordWSMessage(direction string) {
	m.WebSocketMessagesTotal.WithLabelValues(direction).Inc()
}

// UpdateSystemMetrics 更新系统指标
func (m *Metrics) UpdateSystemMetrics(goroutines int, memoryBytes uint64, cpuPercent float64) {
	m.GoroutinesCount.Set(float64(goroutines))
	m.MemoryUsage.Set(float64(memoryBytes))
	m.CPUUsage.Set(cpuPercent)
}

// SetRulesTotal 设置规则总数
func (m *Metrics) SetRulesTotal(count int) {
	m.RulesTotal.Set(float64(count))
}

// SetDBConnections 设置数据库连接数
func (m *Metrics) SetDBConnections(count int) {
	m.DBConnectionsTotal.Set(float64(count))
}

// IncrementHTTPRequestsInFlight HTTP 请求进行中数量增加
func (m *Metrics) IncrementHTTPRequestsInFlight() {
	m.HTTPRequestsInFlight.Inc()
}

// DecrementHTTPRequestsInFlight HTTP 请求进行中数量减少
func (m *Metrics) DecrementHTTPRequestsInFlight() {
	m.HTTPRequestsInFlight.Dec()
}

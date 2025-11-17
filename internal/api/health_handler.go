package api

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/metrics"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// HealthHandler 健康检查和系统指标处理器
type HealthHandler struct {
	db      *mongo.Database
	metrics *metrics.Metrics
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(db *mongo.Database, metrics *metrics.Metrics) *HealthHandler {
	return &HealthHandler{
		db:      db,
		metrics: metrics,
	}
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
	Uptime    int64             `json:"uptime"` // 秒
}

// SystemMetricsResponse 系统指标响应
type SystemMetricsResponse struct {
	Timestamp   time.Time       `json:"timestamp"`
	Runtime     RuntimeMetrics  `json:"runtime"`
	Memory      MemoryMetrics   `json:"memory"`
	Goroutines  int             `json:"goroutines"`
	Database    DatabaseMetrics `json:"database"`
}

// RuntimeMetrics 运行时指标
type RuntimeMetrics struct {
	Uptime      int64  `json:"uptime"`      // 秒
	GoVersion   string `json:"go_version"`
	NumCPU      int    `json:"num_cpu"`
	GOOS        string `json:"goos"`
	GOARCH      string `json:"goarch"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Alloc      uint64  `json:"alloc"`       // bytes
	TotalAlloc uint64  `json:"total_alloc"` // bytes
	Sys        uint64  `json:"sys"`         // bytes
	NumGC      uint32  `json:"num_gc"`
	HeapAlloc  uint64  `json:"heap_alloc"`  // bytes
	HeapSys    uint64  `json:"heap_sys"`    // bytes
	HeapInuse  uint64  `json:"heap_inuse"`  // bytes
	StackInuse uint64  `json:"stack_inuse"` // bytes
	GCCPUFraction float64 `json:"gc_cpu_fraction"` // GC CPU 占用率
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	Connected bool   `json:"connected"`
	Status    string `json:"status"`
	Latency   int64  `json:"latency"` // 毫秒
}

var startTime = time.Now()

// Health 健康检查接口
// @Summary 健康检查
// @Description 检查系统各组件的健康状态
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} ErrorResponse
// @Router /api/v1/health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)
	overallStatus := "healthy"

	// 检查数据库连接
	dbStatus := "healthy"
	if err := h.db.Client().Ping(ctx, nil); err != nil {
		logger.Error("database health check failed", zap.Error(err))
		dbStatus = "unhealthy"
		overallStatus = "unhealthy"
	}
	services["database"] = dbStatus

	// 检查内存使用
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memStatus := "healthy"
	// 如果内存使用超过 90%，标记为警告
	if m.Alloc > m.Sys*9/10 {
		memStatus = "warning"
	}
	services["memory"] = memStatus

	// 检查 Goroutine 数量
	goroutineStatus := "healthy"
	numGoroutines := runtime.NumGoroutine()
	// 如果 Goroutine 数量超过 10000，标记为警告
	if numGoroutines > 10000 {
		goroutineStatus = "warning"
	}
	services["goroutines"] = goroutineStatus

	uptime := int64(time.Since(startTime).Seconds())

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  services,
		Uptime:    uptime,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Metrics 系统指标接口
// @Summary 获取系统指标
// @Description 获取详细的系统运行指标
// @Tags Health
// @Produce json
// @Success 200 {object} SystemMetricsResponse
// @Router /api/v1/metrics [get]
func (h *HealthHandler) Metrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 获取内存统计
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取运行时信息
	runtimeMetrics := RuntimeMetrics{
		Uptime:    int64(time.Since(startTime).Seconds()),
		GoVersion: runtime.Version(),
		NumCPU:    runtime.NumCPU(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
	}

	// 获取内存指标
	memoryMetrics := MemoryMetrics{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		NumGC:         m.NumGC,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapInuse:     m.HeapInuse,
		StackInuse:    m.StackInuse,
		GCCPUFraction: m.GCCPUFraction,
	}

	// 获取 Goroutine 数量
	goroutines := runtime.NumGoroutine()

	// 获取数据库指标
	dbMetrics := DatabaseMetrics{
		Connected: true,
		Status:    "healthy",
	}

	// 测试数据库延迟
	startPing := time.Now()
	if err := h.db.Client().Ping(ctx, nil); err != nil {
		logger.Error("database ping failed", zap.Error(err))
		dbMetrics.Connected = false
		dbMetrics.Status = "unhealthy"
		dbMetrics.Latency = 0
	} else {
		dbMetrics.Latency = time.Since(startPing).Milliseconds()
	}

	// 更新 Prometheus 指标
	if h.metrics != nil {
		h.metrics.UpdateSystemMetrics(goroutines, m.Alloc, 0) // CPU 使用率需要另外计算
	}

	response := SystemMetricsResponse{
		Timestamp:  time.Now(),
		Runtime:    runtimeMetrics,
		Memory:     memoryMetrics,
		Goroutines: goroutines,
		Database:   dbMetrics,
	}

	c.JSON(http.StatusOK, response)
}

// Ready 就绪检查接口
// @Summary 就绪检查
// @Description 检查服务是否就绪（所有依赖服务都正常）
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} ErrorResponse
// @Router /api/v1/ready [get]
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	// 检查数据库连接
	if err := h.db.Client().Ping(ctx, nil); err != nil {
		logger.Error("database not ready", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "Service not ready",
			"message": "Database connection failed",
			"code":    "SERVICE_NOT_READY",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"timestamp": time.Now(),
	})
}

// Live 存活检查接口
// @Summary 存活检查
// @Description 检查服务是否存活（简单响应，无依赖检查）
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"timestamp": time.Now(),
	})
}

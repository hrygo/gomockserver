package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestNewStatisticsHandler 测试创建统计处理器
func TestNewStatisticsHandler(t *testing.T) {
	projectRepo := &MockProjectRepository{}
	ruleRepo := &MockRuleRepository{}
	envRepo := &MockEnvironmentRepository{}

	handler := NewStatisticsHandler(projectRepo, ruleRepo, envRepo)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.projectRepo)
	assert.NotNil(t, handler.ruleRepo)
	assert.NotNil(t, handler.environmentRepo)
}

// TestGetDashboardStatistics 测试获取仪表盘统计数据
func TestGetDashboardStatistics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	projectRepo := new(MockProjectRepository)
	ruleRepo := new(MockRuleRepository)
	envRepo := new(MockEnvironmentRepository)

	// Mock 设置
	projectRepo.On("List", mock.Anything, int64(0), int64(10000)).Return(
		[]*models.Project{
			{ID: "1", Name: "Project 1"},
			{ID: "2", Name: "Project 2"},
		},
		int64(2),
		nil,
	)

	envRepo.On("FindByProject", mock.Anything, "1").Return([]*models.Environment{{ID: "1"}}, nil)
	envRepo.On("FindByProject", mock.Anything, "2").Return([]*models.Environment{{ID: "2"}}, nil)

	ruleRepo.On("List", mock.Anything, mock.Anything, int64(0), int64(10000)).Return(
		[]*models.Rule{
			{ID: "1", Name: "Rule 1", Enabled: true},
			{ID: "2", Name: "Rule 2", Enabled: false},
			{ID: "3", Name: "Rule 3", Enabled: true},
		},
		int64(3),
		nil,
	)

	handler := NewStatisticsHandler(projectRepo, ruleRepo, envRepo)

	router := gin.New()
	router.GET("/statistics/dashboard", handler.GetDashboardStatistics)

	req := httptest.NewRequest(http.MethodGet, "/statistics/dashboard", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "total_projects")
	assert.Contains(t, w.Body.String(), "total_rules")
	assert.Contains(t, w.Body.String(), "enabled_rules")
	projectRepo.AssertExpectations(t)
	ruleRepo.AssertExpectations(t)
}

// TestGetRequestTrend 测试获取请求趋势
func TestGetRequestTrend(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStatisticsHandler(nil, nil, nil)

	router := gin.New()
	router.GET("/statistics/request-trend", handler.GetRequestTrend)

	req := httptest.NewRequest(http.MethodGet, "/statistics/request-trend", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "date")
	assert.Contains(t, w.Body.String(), "count")
}

// TestGetResponseTimeDistribution 测试获取响应时间分布
func TestGetResponseTimeDistribution(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewStatisticsHandler(nil, nil, nil)

	router := gin.New()
	router.GET("/statistics/response-time", handler.GetResponseTimeDistribution)

	req := httptest.NewRequest(http.MethodGet, "/statistics/response-time", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "range")
	assert.Contains(t, w.Body.String(), "count")
}

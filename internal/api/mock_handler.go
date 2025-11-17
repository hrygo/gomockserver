package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MockHandler Mock处理器
type MockHandler struct {
	// 如果需要存储历史记录，可以添加相关依赖
}

// NewMockHandler 创建Mock处理器
func NewMockHandler() *MockHandler {
	return &MockHandler{}
}

// MockTestHistory Mock测试历史记录
type MockTestHistory struct {
	ID            string                 `json:"id"`
	Request       map[string]interface{} `json:"request"`
	Response      map[string]interface{} `json:"response"`
	Timestamp     string                 `json:"timestamp"`
	ProjectID     string                 `json:"project_id"`
	EnvironmentID string                 `json:"environment_id"`
}

// SendMockRequest 发送Mock测试请求
// POST /api/v1/mock/test
func (h *MockHandler) SendMockRequest(c *gin.Context) {
	// TODO: 实现发送Mock测试请求的逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Mock request sent successfully",
	})
}

// GetMockHistory 获取Mock测试历史
// GET /api/v1/mock/history
func (h *MockHandler) GetMockHistory(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	// TODO: 实现获取Mock测试历史的逻辑
	// 目前返回空数组
	history := []MockTestHistory{}
	c.JSON(http.StatusOK, history)
}

// ClearMockHistory 清空Mock测试历史
// DELETE /api/v1/mock/history
func (h *MockHandler) ClearMockHistory(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	// TODO: 实现清空Mock测试历史的逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Mock history cleared successfully",
	})
}

// DeleteMockHistoryItem 删除单条Mock测试历史记录
// DELETE /api/v1/mock/history/:id
func (h *MockHandler) DeleteMockHistoryItem(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	// TODO: 实现删除单条Mock测试历史记录的逻辑
	c.JSON(http.StatusOK, gin.H{
		"message": "Mock history item deleted successfully",
	})
}

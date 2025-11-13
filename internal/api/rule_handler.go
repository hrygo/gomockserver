package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// RuleHandler 规则处理器
type RuleHandler struct {
	ruleRepo repository.RuleRepository
}

// NewRuleHandler 创建规则处理器
func NewRuleHandler(ruleRepo repository.RuleRepository) *RuleHandler {
	return &RuleHandler{
		ruleRepo: ruleRepo,
	}
}

// CreateRule 创建规则
func (h *RuleHandler) CreateRule(c *gin.Context) {
	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证必填字段
	if rule.Name == "" || rule.ProjectID == "" || rule.EnvironmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, project_id and environment_id are required"})
		return
	}

	// 创建规则
	if err := h.ruleRepo.Create(c.Request.Context(), &rule); err != nil {
		logger.Error("failed to create rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule"})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// GetRule 获取规则详情
func (h *RuleHandler) GetRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	rule, err := h.ruleRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rule"})
		return
	}

	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// UpdateRule 更新规则
func (h *RuleHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	var rule models.Rule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule.ID = id

	if err := h.ruleRepo.Update(c.Request.Context(), &rule); err != nil {
		logger.Error("failed to update rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteRule 删除规则
func (h *RuleHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	if err := h.ruleRepo.Delete(c.Request.Context(), id); err != nil {
		logger.Error("failed to delete rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}

// ListRules 列出规则
func (h *RuleHandler) ListRules(c *gin.Context) {
	// 分页参数
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 64)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	skip := (page - 1) * pageSize

	// 过滤条件
	filter := make(map[string]interface{})
	if projectID := c.Query("project_id"); projectID != "" {
		filter["project_id"] = projectID
	}
	if environmentID := c.Query("environment_id"); environmentID != "" {
		filter["environment_id"] = environmentID
	}
	if protocol := c.Query("protocol"); protocol != "" {
		filter["protocol"] = protocol
	}
	if enabled := c.Query("enabled"); enabled != "" {
		filter["enabled"] = enabled == "true"
	}

	rules, total, err := h.ruleRepo.List(c.Request.Context(), filter, skip, pageSize)
	if err != nil {
		logger.Error("failed to list rules", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list rules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      rules,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// EnableRule 启用规则
func (h *RuleHandler) EnableRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	rule, err := h.ruleRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rule"})
		return
	}

	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	rule.Enabled = true
	if err := h.ruleRepo.Update(c.Request.Context(), rule); err != nil {
		logger.Error("failed to enable rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule enabled successfully"})
}

// DisableRule 禁用规则
func (h *RuleHandler) DisableRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	rule, err := h.ruleRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rule"})
		return
	}

	if rule == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	rule.Enabled = false
	if err := h.ruleRepo.Update(c.Request.Context(), rule); err != nil {
		logger.Error("failed to disable rule", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable rule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule disabled successfully"})
}

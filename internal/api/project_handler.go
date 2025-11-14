package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// ProjectHandler 项目处理器
type ProjectHandler struct {
	projectRepo     repository.ProjectRepository
	environmentRepo repository.EnvironmentRepository
}

// NewProjectHandler 创建项目处理器
func NewProjectHandler(projectRepo repository.ProjectRepository, environmentRepo repository.EnvironmentRepository) *ProjectHandler {
	return &ProjectHandler{
		projectRepo:     projectRepo,
		environmentRepo: environmentRepo,
	}
}

// CreateProject 创建项目
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectRepo.Create(c.Request.Context(), &project); err != nil {
		logger.Error("failed to create project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// ListProjects 获取项目列表
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	projects, _, err := h.projectRepo.List(c.Request.Context(), 0, 10000)
	if err != nil {
		logger.Error("failed to list projects", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// GetProject 获取项目详情
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id := c.Param("id")

	project, err := h.projectRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	if project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject 更新项目
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id := c.Param("id")

	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project.ID = id

	if err := h.projectRepo.Update(c.Request.Context(), &project); err != nil {
		logger.Error("failed to update project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject 删除项目
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id := c.Param("id")

	if err := h.projectRepo.Delete(c.Request.Context(), id); err != nil {
		logger.Error("failed to delete project", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// CreateEnvironment 创建环境
func (h *ProjectHandler) CreateEnvironment(c *gin.Context) {
	var environment models.Environment
	if err := c.ShouldBindJSON(&environment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.environmentRepo.Create(c.Request.Context(), &environment); err != nil {
		logger.Error("failed to create environment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create environment"})
		return
	}

	c.JSON(http.StatusCreated, environment)
}

// GetEnvironment 获取环境详情
func (h *ProjectHandler) GetEnvironment(c *gin.Context) {
	id := c.Param("id")

	environment, err := h.environmentRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		logger.Error("failed to get environment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get environment"})
		return
	}

	if environment == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Environment not found"})
		return
	}

	c.JSON(http.StatusOK, environment)
}

// ListEnvironments 列出项目下的环境
func (h *ProjectHandler) ListEnvironments(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	environments, err := h.environmentRepo.FindByProject(c.Request.Context(), projectID)
	if err != nil {
		logger.Error("failed to list environments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list environments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": environments,
	})
}

// UpdateEnvironment 更新环境
func (h *ProjectHandler) UpdateEnvironment(c *gin.Context) {
	id := c.Param("id")

	var environment models.Environment
	if err := c.ShouldBindJSON(&environment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	environment.ID = id

	if err := h.environmentRepo.Update(c.Request.Context(), &environment); err != nil {
		logger.Error("failed to update environment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update environment"})
		return
	}

	c.JSON(http.StatusOK, environment)
}

// DeleteEnvironment 删除环境
func (h *ProjectHandler) DeleteEnvironment(c *gin.Context) {
	id := c.Param("id")

	if err := h.environmentRepo.Delete(c.Request.Context(), id); err != nil {
		logger.Error("failed to delete environment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete environment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment deleted successfully"})
}

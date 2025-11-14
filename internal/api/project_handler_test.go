package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProjectRepository Mock 项目仓库
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project *models.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Update(ctx context.Context, project *models.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) FindByID(ctx context.Context, id string) (*models.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockProjectRepository) FindByWorkspace(ctx context.Context, workspaceID string) ([]*models.Project, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockProjectRepository) List(ctx context.Context, skip, limit int64) ([]*models.Project, int64, error) {
	args := m.Called(ctx, skip, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Project), args.Get(1).(int64), args.Error(2)
}

// MockEnvironmentRepository Mock 环境仓库
type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, environment *models.Environment) error {
	args := m.Called(ctx, environment)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Update(ctx context.Context, environment *models.Environment) error {
	args := m.Called(ctx, environment)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) FindByID(ctx context.Context, id string) (*models.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) FindByProject(ctx context.Context, projectID string) ([]*models.Environment, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Environment), args.Error(1)
}

// TestProjectHandler_CreateProject 测试创建项目
func TestProjectHandler_CreateProject(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockProjectRepository)
		expectedStatus int
	}{
		{
			name: "成功创建项目",
			requestBody: models.Project{
				Name:        "测试项目",
				WorkspaceID: "workspace-001",
			},
			mockSetup: func(m *MockProjectRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Project")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "无效的JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockProjectRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "数据库错误",
			requestBody: models.Project{
				Name: "测试项目",
			},
			mockSetup: func(m *MockProjectRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Project")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockProjectRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.POST("/projects", handler.CreateProject)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockProjectRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_GetProject 测试获取项目
func TestProjectHandler_GetProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		mockSetup      func(*MockProjectRepository)
		expectedStatus int
	}{
		{
			name:      "成功获取项目",
			projectID: "project-001",
			mockSetup: func(m *MockProjectRepository) {
				m.On("FindByID", mock.Anything, "project-001").Return(&models.Project{
					ID:   "project-001",
					Name: "测试项目",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "项目不存在",
			projectID: "project-999",
			mockSetup: func(m *MockProjectRepository) {
				m.On("FindByID", mock.Anything, "project-999").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:      "数据库错误",
			projectID: "project-001",
			mockSetup: func(m *MockProjectRepository) {
				m.On("FindByID", mock.Anything, "project-001").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockProjectRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.GET("/projects/:id", handler.GetProject)

			req := httptest.NewRequest(http.MethodGet, "/projects/"+tt.projectID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockProjectRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_UpdateProject 测试更新项目
func TestProjectHandler_UpdateProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		requestBody    interface{}
		mockSetup      func(*MockProjectRepository)
		expectedStatus int
	}{
		{
			name:      "成功更新项目",
			projectID: "project-001",
			requestBody: models.Project{
				Name: "更新后的项目",
			},
			mockSetup: func(m *MockProjectRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Project")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的JSON",
			projectID:      "project-001",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockProjectRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "数据库错误",
			projectID: "project-001",
			requestBody: models.Project{
				Name: "更新后的项目",
			},
			mockSetup: func(m *MockProjectRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Project")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockProjectRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.PUT("/projects/:id", handler.UpdateProject)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPut, "/projects/"+tt.projectID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockProjectRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_DeleteProject 测试删除项目
func TestProjectHandler_DeleteProject(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		mockSetup      func(*MockProjectRepository)
		expectedStatus int
	}{
		{
			name:      "成功删除项目",
			projectID: "project-001",
			mockSetup: func(m *MockProjectRepository) {
				m.On("Delete", mock.Anything, "project-001").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "数据库错误",
			projectID: "project-001",
			mockSetup: func(m *MockProjectRepository) {
				m.On("Delete", mock.Anything, "project-001").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockProjectRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.DELETE("/projects/:id", handler.DeleteProject)

			req := httptest.NewRequest(http.MethodDelete, "/projects/"+tt.projectID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockProjectRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_CreateEnvironment 测试创建环境
func TestProjectHandler_CreateEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockEnvironmentRepository)
		expectedStatus int
	}{
		{
			name: "成功创建环境",
			requestBody: models.Environment{
				Name:      "开发环境",
				ProjectID: "project-001",
			},
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Environment")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "无效的JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockEnvironmentRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "数据库错误",
			requestBody: models.Environment{
				Name: "开发环境",
			},
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Environment")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockEnvRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.POST("/environments", handler.CreateEnvironment)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/environments", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockEnvRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_GetEnvironment 测试获取环境
func TestProjectHandler_GetEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		envID          string
		mockSetup      func(*MockEnvironmentRepository)
		expectedStatus int
	}{
		{
			name:  "成功获取环境",
			envID: "env-001",
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("FindByID", mock.Anything, "env-001").Return(&models.Environment{
					ID:   "env-001",
					Name: "开发环境",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "环境不存在",
			envID: "env-999",
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("FindByID", mock.Anything, "env-999").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "数据库错误",
			envID: "env-001",
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("FindByID", mock.Anything, "env-001").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockEnvRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.GET("/environments/:id", handler.GetEnvironment)

			req := httptest.NewRequest(http.MethodGet, "/environments/"+tt.envID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockEnvRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_ListEnvironments 测试列出环境
func TestProjectHandler_ListEnvironments(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockEnvironmentRepository)
		expectedStatus int
	}{
		{
			name:        "成功列出环境",
			queryParams: "?project_id=project-001",
			mockSetup: func(m *MockEnvironmentRepository) {
				envs := []*models.Environment{
					{ID: "env-001", Name: "开发环境"},
					{ID: "env-002", Name: "测试环境"},
				}
				m.On("FindByProject", mock.Anything, "project-001").Return(envs, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "缺少project_id参数",
			queryParams:    "",
			mockSetup:      func(m *MockEnvironmentRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "数据库错误",
			queryParams: "?project_id=project-001",
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("FindByProject", mock.Anything, "project-001").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockEnvRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.GET("/environments", handler.ListEnvironments)

			req := httptest.NewRequest(http.MethodGet, "/environments"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockEnvRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_UpdateEnvironment 测试更新环境
func TestProjectHandler_UpdateEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		envID          string
		requestBody    interface{}
		mockSetup      func(*MockEnvironmentRepository)
		expectedStatus int
	}{
		{
			name:  "成功更新环境",
			envID: "env-001",
			requestBody: models.Environment{
				Name: "更新后的环境",
			},
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Environment")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的JSON",
			envID:          "env-001",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockEnvironmentRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "数据库错误",
			envID: "env-001",
			requestBody: models.Environment{
				Name: "更新后的环境",
			},
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Environment")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockEnvRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.PUT("/environments/:id", handler.UpdateEnvironment)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPut, "/environments/"+tt.envID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockEnvRepo.AssertExpectations(t)
		})
	}
}

// TestProjectHandler_DeleteEnvironment 测试删除环境
func TestProjectHandler_DeleteEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		envID          string
		mockSetup      func(*MockEnvironmentRepository)
		expectedStatus int
	}{
		{
			name:  "成功删除环境",
			envID: "env-001",
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("Delete", mock.Anything, "env-001").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "数据库错误",
			envID: "env-001",
			mockSetup: func(m *MockEnvironmentRepository) {
				m.On("Delete", mock.Anything, "env-001").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockProjectRepo := new(MockProjectRepository)
			mockEnvRepo := new(MockEnvironmentRepository)
			tt.mockSetup(mockEnvRepo)

			handler := NewProjectHandler(mockProjectRepo, mockEnvRepo)
			router := setupTestRouter()
			router.DELETE("/environments/:id", handler.DeleteEnvironment)

			req := httptest.NewRequest(http.MethodDelete, "/environments/"+tt.envID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockEnvRepo.AssertExpectations(t)
		})
	}
}

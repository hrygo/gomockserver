package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRuleRepository Mock 规则仓库
type MockRuleRepository struct {
	mock.Mock
}

func (m *MockRuleRepository) Create(ctx context.Context, rule *models.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Update(ctx context.Context, rule *models.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockRuleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRuleRepository) FindByID(ctx context.Context, id string) (*models.Rule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rule), args.Error(1)
}

func (m *MockRuleRepository) FindByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error) {
	args := m.Called(ctx, projectID, environmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Rule), args.Error(1)
}

func (m *MockRuleRepository) FindEnabledByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error) {
	args := m.Called(ctx, projectID, environmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Rule), args.Error(1)
}

func (m *MockRuleRepository) List(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]*models.Rule, int64, error) {
	args := m.Called(ctx, filter, skip, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.Rule), args.Get(1).(int64), args.Error(2)
}

// setupTestRouter 创建测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// TestRuleHandler_CreateRule 测试创建规则
func TestRuleHandler_CreateRule(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockRuleRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "成功创建规则",
			requestBody: models.Rule{
				Name:          "测试规则",
				ProjectID:     "project-001",
				EnvironmentID: "env-001",
				Protocol:      models.ProtocolHTTP,
			},
			mockSetup: func(m *MockRuleRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Rule")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "无效的JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockRuleRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "缺少必填字段 - Name",
			requestBody: models.Rule{
				ProjectID:     "project-001",
				EnvironmentID: "env-001",
			},
			mockSetup:      func(m *MockRuleRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, project_id and environment_id are required",
		},
		{
			name: "缺少必填字段 - ProjectID",
			requestBody: models.Rule{
				Name:          "测试规则",
				EnvironmentID: "env-001",
			},
			mockSetup:      func(m *MockRuleRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name, project_id and environment_id are required",
		},
		{
			name: "数据库错误",
			requestBody: models.Rule{
				Name:          "测试规则",
				ProjectID:     "project-001",
				EnvironmentID: "env-001",
				Protocol:      models.ProtocolHTTP,
			},
			mockSetup: func(m *MockRuleRepository) {
				m.On("Create", mock.Anything, mock.AnythingOfType("*models.Rule")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to create rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 Mock Repository
			mockRepo := new(MockRuleRepository)
			tt.mockSetup(mockRepo)

			// 创建 Handler
			handler := NewRuleHandler(mockRepo)

			// 创建路由
			router := setupTestRouter()
			router.POST("/rules", handler.CreateRule)

			// 准备请求
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/rules", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// 执行请求
			router.ServeHTTP(w, req)

			// 验证结果
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRuleHandler_GetRule 测试获取规则
func TestRuleHandler_GetRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		mockSetup      func(*MockRuleRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功获取规则",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				m.On("FindByID", mock.Anything, "rule-001").Return(&models.Rule{
					ID:   "rule-001",
					Name: "测试规则",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "规则不存在",
			ruleID: "rule-999",
			mockSetup: func(m *MockRuleRepository) {
				m.On("FindByID", mock.Anything, "rule-999").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Rule not found",
		},
		{
			name:   "数据库错误",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				m.On("FindByID", mock.Anything, "rule-001").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRuleRepository)
			tt.mockSetup(mockRepo)

			handler := NewRuleHandler(mockRepo)
			router := setupTestRouter()
			router.GET("/rules/:id", handler.GetRule)

			req := httptest.NewRequest(http.MethodGet, "/rules/"+tt.ruleID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRuleHandler_UpdateRule 测试更新规则
func TestRuleHandler_UpdateRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		requestBody    interface{}
		mockSetup      func(*MockRuleRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功更新规则",
			ruleID: "rule-001",
			requestBody: models.Rule{
				Name:     "更新后的规则",
				Protocol: models.ProtocolHTTP,
			},
			mockSetup: func(m *MockRuleRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Rule")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "无效的JSON",
			ruleID:         "rule-001",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockRuleRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "数据库错误",
			ruleID: "rule-001",
			requestBody: models.Rule{
				Name: "更新后的规则",
			},
			mockSetup: func(m *MockRuleRepository) {
				m.On("Update", mock.Anything, mock.AnythingOfType("*models.Rule")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to update rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRuleRepository)
			tt.mockSetup(mockRepo)

			handler := NewRuleHandler(mockRepo)
			router := setupTestRouter()
			router.PUT("/rules/:id", handler.UpdateRule)

			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPut, "/rules/"+tt.ruleID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRuleHandler_DeleteRule 测试删除规则
func TestRuleHandler_DeleteRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		mockSetup      func(*MockRuleRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功删除规则",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				m.On("Delete", mock.Anything, "rule-001").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "数据库错误",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				m.On("Delete", mock.Anything, "rule-001").Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to delete rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRuleRepository)
			tt.mockSetup(mockRepo)

			handler := NewRuleHandler(mockRepo)
			router := setupTestRouter()
			router.DELETE("/rules/:id", handler.DeleteRule)

			req := httptest.NewRequest(http.MethodDelete, "/rules/"+tt.ruleID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRuleHandler_ListRules 测试列出规则
func TestRuleHandler_ListRules(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*MockRuleRepository)
		expectedStatus int
		expectedTotal  int64
	}{
		{
			name:        "成功列出规则 - 默认参数",
			queryParams: "",
			mockSetup: func(m *MockRuleRepository) {
				rules := []*models.Rule{
					{ID: "rule-001", Name: "规则1"},
					{ID: "rule-002", Name: "规则2"},
				}
				m.On("List", mock.Anything, mock.AnythingOfType("map[string]interface {}"), int64(0), int64(20)).
					Return(rules, int64(2), nil)
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  2,
		},
		{
			name:        "自定义分页参数",
			queryParams: "?page=2&page_size=10",
			mockSetup: func(m *MockRuleRepository) {
				rules := []*models.Rule{
					{ID: "rule-011", Name: "规则11"},
				}
				m.On("List", mock.Anything, mock.AnythingOfType("map[string]interface {}"), int64(10), int64(10)).
					Return(rules, int64(15), nil)
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  15,
		},
		{
			name:        "带过滤条件",
			queryParams: "?project_id=project-001&enabled=true",
			mockSetup: func(m *MockRuleRepository) {
				rules := []*models.Rule{
					{ID: "rule-001", Name: "启用的规则"},
				}
				m.On("List", mock.Anything, mock.MatchedBy(func(filter map[string]interface{}) bool {
					return filter["project_id"] == "project-001" && filter["enabled"] == true
				}), int64(0), int64(20)).Return(rules, int64(1), nil)
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  1,
		},
		{
			name:        "数据库错误",
			queryParams: "",
			mockSetup: func(m *MockRuleRepository) {
				m.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, int64(0), errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRuleRepository)
			tt.mockSetup(mockRepo)

			handler := NewRuleHandler(mockRepo)
			router := setupTestRouter()
			router.GET("/rules", handler.ListRules)

			req := httptest.NewRequest(http.MethodGet, "/rules"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, float64(tt.expectedTotal), response["total"])
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRuleHandler_EnableRule 测试启用规则
func TestRuleHandler_EnableRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		mockSetup      func(*MockRuleRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功启用规则",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				rule := &models.Rule{
					ID:      "rule-001",
					Name:    "测试规则",
					Enabled: false,
				}
				m.On("FindByID", mock.Anything, "rule-001").Return(rule, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(r *models.Rule) bool {
					return r.Enabled == true
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "规则不存在",
			ruleID: "rule-999",
			mockSetup: func(m *MockRuleRepository) {
				m.On("FindByID", mock.Anything, "rule-999").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Rule not found",
		},
		{
			name:   "查询时数据库错误",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				m.On("FindByID", mock.Anything, "rule-001").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to get rule",
		},
		{
			name:   "更新时数据库错误",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				rule := &models.Rule{
					ID:      "rule-001",
					Enabled: false,
				}
				m.On("FindByID", mock.Anything, "rule-001").Return(rule, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Failed to enable rule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRuleRepository)
			tt.mockSetup(mockRepo)

			handler := NewRuleHandler(mockRepo)
			router := setupTestRouter()
			router.POST("/rules/:id/enable", handler.EnableRule)

			req := httptest.NewRequest(http.MethodPost, "/rules/"+tt.ruleID+"/enable", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestRuleHandler_DisableRule 测试禁用规则
func TestRuleHandler_DisableRule(t *testing.T) {
	tests := []struct {
		name           string
		ruleID         string
		mockSetup      func(*MockRuleRepository)
		expectedStatus int
		expectedError  string
	}{
		{
			name:   "成功禁用规则",
			ruleID: "rule-001",
			mockSetup: func(m *MockRuleRepository) {
				rule := &models.Rule{
					ID:      "rule-001",
					Name:    "测试规则",
					Enabled: true,
				}
				m.On("FindByID", mock.Anything, "rule-001").Return(rule, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(r *models.Rule) bool {
					return r.Enabled == false
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "规则不存在",
			ruleID: "rule-999",
			mockSetup: func(m *MockRuleRepository) {
				m.On("FindByID", mock.Anything, "rule-999").Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Rule not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRuleRepository)
			tt.mockSetup(mockRepo)

			handler := NewRuleHandler(mockRepo)
			router := setupTestRouter()
			router.POST("/rules/:id/disable", handler.DisableRule)

			req := httptest.NewRequest(http.MethodPost, "/rules/"+tt.ruleID+"/disable", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Contains(t, response["error"], tt.expectedError)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

package service

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMatchEngine Mock 匹配引擎
type MockMatchEngine struct {
	mock.Mock
}

func (m *MockMatchEngine) Match(ctx context.Context, request *adapter.Request, projectID, environmentID string) (*models.Rule, error) {
	args := m.Called(ctx, request, projectID, environmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rule), args.Error(1)
}

// MockMockExecutor Mock 执行器
type MockMockExecutor struct {
	mock.Mock
}

func (m *MockMockExecutor) Execute(request *adapter.Request, rule *models.Rule) (*adapter.Response, error) {
	args := m.Called(request, rule)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*adapter.Response), args.Error(1)
}

func (m *MockMockExecutor) GetDefaultResponse() *adapter.Response {
	args := m.Called()
	return args.Get(0).(*adapter.Response)
}

// TestNewMockService 测试创建 Mock 服务
func TestNewMockService(t *testing.T) {
	mockEngine := new(MockMatchEngine)
	mockExecutor := new(MockMockExecutor)

	service := NewMockService(mockEngine, mockExecutor)

	assert.NotNil(t, service)
	assert.NotNil(t, service.httpAdapter)
	assert.NotNil(t, service.matchEngine)
	assert.NotNil(t, service.mockExecutor)
}

// TestMockService_HandleMockRequest_MissingParams 测试缺少参数
func TestMockService_HandleMockRequest_MissingParams(t *testing.T) {
	tests := []struct {
		name          string
		projectID     string
		environmentID string
		expectedError string
	}{
		{
			name:          "缺少 projectID",
			projectID:     "",
			environmentID: "env-001",
			expectedError: "projectID and environmentID are required",
		},
		{
			name:          "缺少 environmentID",
			projectID:     "project-001",
			environmentID: "",
			expectedError: "projectID and environmentID are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockEngine := new(MockMatchEngine)
			mockExecutor := new(MockMockExecutor)
			service := NewMockService(mockEngine, mockExecutor)

			router := setupTestRouter()
			router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

			req := httptest.NewRequest(http.MethodGet, "/"+tt.projectID+"/"+tt.environmentID+"/api/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedError)
		})
	}
}

// TestMockService_HandleMockRequest_MatchRuleError 测试匹配规则错误
func TestMockService_HandleMockRequest_MatchRuleError(t *testing.T) {
	mockEngine := new(MockMatchEngine)
	mockExecutor := new(MockMockExecutor)
	service := NewMockService(mockEngine, mockExecutor)

	// 模拟匹配规则失败
	mockEngine.On("Match", mock.Anything, mock.Anything, "project-001", "env-001").
		Return(nil, errors.New("database error"))

	router := setupTestRouter()
	router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

	req := httptest.NewRequest(http.MethodGet, "/project-001/env-001/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to match rule")
	mockEngine.AssertExpectations(t)
}

// TestMockService_HandleMockRequest_NoRuleMatched 测试无匹配规则
func TestMockService_HandleMockRequest_NoRuleMatched(t *testing.T) {
	mockEngine := new(MockMatchEngine)
	mockExecutor := new(MockMockExecutor)
	service := NewMockService(mockEngine, mockExecutor)

	// 模拟无匹配规则
	mockEngine.On("Match", mock.Anything, mock.Anything, "project-001", "env-001").
		Return(nil, nil)

	// 模拟返回默认响应
	defaultResponse := &adapter.Response{
		StatusCode: 404,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"error": "No matching rule found"}`),
	}
	mockExecutor.On("GetDefaultResponse").Return(defaultResponse)

	router := setupTestRouter()
	router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

	req := httptest.NewRequest(http.MethodGet, "/project-001/env-001/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "No matching rule found")
	mockEngine.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

// TestMockService_HandleMockRequest_ExecuteError 测试执行错误
func TestMockService_HandleMockRequest_ExecuteError(t *testing.T) {
	mockEngine := new(MockMatchEngine)
	mockExecutor := new(MockMockExecutor)
	service := NewMockService(mockEngine, mockExecutor)

	// 模拟匹配成功
	testRule := &models.Rule{
		ID:            "rule-001",
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
	}
	mockEngine.On("Match", mock.Anything, mock.Anything, "project-001", "env-001").
		Return(testRule, nil)

	// 模拟执行失败
	mockExecutor.On("Execute", mock.Anything, testRule).
		Return(nil, errors.New("execution error"))

	router := setupTestRouter()
	router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

	req := httptest.NewRequest(http.MethodGet, "/project-001/env-001/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to execute mock")
	mockEngine.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

// TestMockService_HandleMockRequest_Success 测试成功处理请求
func TestMockService_HandleMockRequest_Success(t *testing.T) {
	mockEngine := new(MockMatchEngine)
	mockExecutor := new(MockMockExecutor)
	service := NewMockService(mockEngine, mockExecutor)

	// 模拟匹配成功
	testRule := &models.Rule{
		ID:            "rule-001",
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
	}
	mockEngine.On("Match", mock.Anything, mock.Anything, "project-001", "env-001").
		Return(testRule, nil)

	// 模拟执行成功
	mockResponse := &adapter.Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"message": "success"}`),
	}
	mockExecutor.On("Execute", mock.Anything, testRule).
		Return(mockResponse, nil)

	router := setupTestRouter()
	router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

	req := httptest.NewRequest(http.MethodGet, "/project-001/env-001/api/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	mockEngine.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

// TestMockService_HandleMockRequest_WithBody 测试带请求体的请求
func TestMockService_HandleMockRequest_WithBody(t *testing.T) {
	mockEngine := new(MockMatchEngine)
	mockExecutor := new(MockMockExecutor)
	service := NewMockService(mockEngine, mockExecutor)

	// 模拟匹配成功
	testRule := &models.Rule{
		ID:            "rule-001",
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
	}
	mockEngine.On("Match", mock.Anything, mock.Anything, "project-001", "env-001").
		Return(testRule, nil)

	// 模拟执行成功
	mockResponse := &adapter.Response{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"id": "123", "message": "created"}`),
	}
	mockExecutor.On("Execute", mock.Anything, testRule).
		Return(mockResponse, nil)

	router := setupTestRouter()
	router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

	requestBody := `{"name": "test", "value": 123}`
	req := httptest.NewRequest(http.MethodPost, "/project-001/env-001/api/users", bytes.NewBufferString(requestBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "created")
	mockEngine.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

// TestMockService_HandleMockRequest_WithHeaders 测试带自定义头部的请求
func TestMockService_HandleMockRequest_WithHeaders(t *testing.T) {
	mockEngine := new(MockMatchEngine)
	mockExecutor := new(MockMockExecutor)
	service := NewMockService(mockEngine, mockExecutor)

	// 模拟匹配成功
	testRule := &models.Rule{
		ID:            "rule-001",
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
	}
	mockEngine.On("Match", mock.Anything, mock.Anything, "project-001", "env-001").
		Return(testRule, nil)

	// 模拟执行成功，返回自定义头部
	mockResponse := &adapter.Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":      "application/json",
			"X-Custom-Header":   "custom-value",
			"X-Request-ID":      "req-123",
		},
		Body: []byte(`{"message": "success"}`),
	}
	mockExecutor.On("Execute", mock.Anything, testRule).
		Return(mockResponse, nil)

	router := setupTestRouter()
	router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

	req := httptest.NewRequest(http.MethodGet, "/project-001/env-001/api/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-value", w.Header().Get("X-Custom-Header"))
	assert.Equal(t, "req-123", w.Header().Get("X-Request-ID"))
	mockEngine.AssertExpectations(t)
	mockExecutor.AssertExpectations(t)
}

// TestMockService_HandleMockRequest_DifferentMethods 测试不同的 HTTP 方法
func TestMockService_HandleMockRequest_DifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			mockEngine := new(MockMatchEngine)
			mockExecutor := new(MockMockExecutor)
			service := NewMockService(mockEngine, mockExecutor)

			// 模拟匹配成功
			testRule := &models.Rule{
				ID:            "rule-001",
				Name:          "测试规则",
				ProjectID:     "project-001",
				EnvironmentID: "env-001",
				Protocol:      models.ProtocolHTTP,
			}
			mockEngine.On("Match", mock.Anything, mock.Anything, "project-001", "env-001").
				Return(testRule, nil)

			// 模拟执行成功
			mockResponse := &adapter.Response{
				StatusCode: 200,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: []byte(`{"method": "` + method + `"}`),
			}
			mockExecutor.On("Execute", mock.Anything, testRule).
				Return(mockResponse, nil)

			router := setupTestRouter()
			router.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

			req := httptest.NewRequest(method, "/project-001/env-001/api/test", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), method)
			mockEngine.AssertExpectations(t)
			mockExecutor.AssertExpectations(t)
		})
	}
}

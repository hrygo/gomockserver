package engine

import (
	"context"
	"testing"

	"github.com/gomockserver/mockserver/internal/adapter"
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
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Rule), args.Get(1).(int64), args.Error(2)
}

// TestNewMatchEngine 测试创建匹配引擎
func TestNewMatchEngine(t *testing.T) {
	mockRepo := new(MockRuleRepository)
	engine := NewMatchEngine(mockRepo)

	assert.NotNil(t, engine)
	assert.Equal(t, mockRepo, engine.ruleRepo)
}

// TestMatch 测试完整的Match流程
func TestMatch(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockRuleRepository)
		request       *adapter.Request
		projectID     string
		environmentID string
		expectedRule  *models.Rule
		expectError   bool
	}{
		{
			name: "成功匹配第一条规则",
			setupMock: func(m *MockRuleRepository) {
				rules := []*models.Rule{
					{
						ID:        "rule-1",
						Name:      "规则1",
						Protocol:  models.ProtocolHTTP,
						MatchType: models.MatchTypeSimple,
						Priority:  100,
						Enabled:   true,
						MatchCondition: map[string]interface{}{
							"method": "GET",
							"path":   "/api/users",
						},
					},
				}
				m.On("FindEnabledByEnvironment", mock.Anything, "project-1", "env-1").
					Return(rules, nil)
			},
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			projectID:     "project-1",
			environmentID: "env-1",
			expectedRule: &models.Rule{
				ID:   "rule-1",
				Name: "规则1",
			},
			expectError: false,
		},
		{
			name: "匹配高优先级规则",
			setupMock: func(m *MockRuleRepository) {
				rules := []*models.Rule{
					{
						ID:        "rule-high",
						Name:      "高优先级规则",
						Protocol:  models.ProtocolHTTP,
						MatchType: models.MatchTypeSimple,
						Priority:  200,
						Enabled:   true,
						MatchCondition: map[string]interface{}{
							"method": "GET",
							"path":   "/api/test",
						},
					},
					{
						ID:        "rule-low",
						Name:      "低优先级规则",
						Protocol:  models.ProtocolHTTP,
						MatchType: models.MatchTypeSimple,
						Priority:  100,
						Enabled:   true,
						MatchCondition: map[string]interface{}{
							"method": "GET",
							"path":   "/api/test",
						},
					},
				}
				m.On("FindEnabledByEnvironment", mock.Anything, "project-1", "env-1").
					Return(rules, nil)
			},
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			projectID:     "project-1",
			environmentID: "env-1",
			expectedRule: &models.Rule{
				ID:   "rule-high",
				Name: "高优先级规则",
			},
			expectError: false,
		},
		{
			name: "没有匹配的规则",
			setupMock: func(m *MockRuleRepository) {
				rules := []*models.Rule{
					{
						ID:        "rule-1",
						Protocol:  models.ProtocolHTTP,
						MatchType: models.MatchTypeSimple,
						MatchCondition: map[string]interface{}{
							"method": "POST",
							"path":   "/api/users",
						},
					},
				}
				m.On("FindEnabledByEnvironment", mock.Anything, "project-1", "env-1").
					Return(rules, nil)
			},
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			projectID:     "project-1",
			environmentID: "env-1",
			expectedRule:  nil,
			expectError:   false,
		},
		{
			name: "空规则列表",
			setupMock: func(m *MockRuleRepository) {
				m.On("FindEnabledByEnvironment", mock.Anything, "project-1", "env-1").
					Return([]*models.Rule{}, nil)
			},
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			projectID:     "project-1",
			environmentID: "env-1",
			expectedRule:  nil,
			expectError:   false,
		},
		{
			name: "协议类型不匹配",
			setupMock: func(m *MockRuleRepository) {
				rules := []*models.Rule{
					{
						ID:        "rule-1",
						Protocol:  models.ProtocolWebSocket,
						MatchType: models.MatchTypeSimple,
					},
				}
				m.On("FindEnabledByEnvironment", mock.Anything, "project-1", "env-1").
					Return(rules, nil)
			},
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users",
			},
			projectID:     "project-1",
			environmentID: "env-1",
			expectedRule:  nil,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRuleRepository)
			tt.setupMock(mockRepo)

			engine := NewMatchEngine(mockRepo)
			rule, err := engine.Match(context.Background(), tt.request, tt.projectID, tt.environmentID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectedRule == nil {
				assert.Nil(t, rule)
			} else {
				assert.NotNil(t, rule)
				assert.Equal(t, tt.expectedRule.ID, rule.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestMatchIPWhitelist 测试IP白名单匹配
func TestMatchIPWhitelist(t *testing.T) {
	tests := []struct {
		name      string
		requestIP string
		whitelist []string
		expected  bool
	}{
		{
			name:      "IP在白名单内",
			requestIP: "192.168.1.100",
			whitelist: []string{"192.168.1.100", "192.168.1.101"},
			expected:  true,
		},
		{
			name:      "IP不在白名单内",
			requestIP: "192.168.1.200",
			whitelist: []string{"192.168.1.100", "192.168.1.101"},
			expected:  false,
		},
		{
			name:      "空白名单",
			requestIP: "192.168.1.100",
			whitelist: []string{},
			expected:  false,
		},
		{
			name:      "单个IP白名单匹配",
			requestIP: "10.0.0.1",
			whitelist: []string{"10.0.0.1"},
			expected:  true,
		},
		{
			name:      "localhost匹配",
			requestIP: "127.0.0.1",
			whitelist: []string{"127.0.0.1", "::1"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchIPWhitelist(tt.requestIP, tt.whitelist)
			assert.Equal(t, tt.expected, result, tt.name)
		})
	}
}

// TestSimpleMatch_WithIPWhitelist 测试带IP白名单的简单匹配
func TestSimpleMatch_WithIPWhitelist(t *testing.T) {
	engine := &MatchEngine{}

	tests := []struct {
		name     string
		request  *adapter.Request
		rule     *models.Rule
		expected bool
	}{
		{
			name: "IP白名单匹配成功",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/admin",
				SourceIP: "192.168.1.100",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/admin",
					"ip_whitelist": []interface{}{
						"192.168.1.100",
						"192.168.1.101",
					},
				},
			},
			expected: true,
		},
		{
			name: "IP白名单匹配失败",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/admin",
				SourceIP: "192.168.1.200",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/admin",
					"ip_whitelist": []interface{}{
						"192.168.1.100",
						"192.168.1.101",
					},
				},
			},
			expected: false,
		},
		{
			name: "无IP白名单限制",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/public",
				SourceIP: "0.0.0.0",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/public",
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := engine.simpleMatch(tt.request, tt.rule)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matched)
		})
	}
}

// TestMatchRule 测试matchRule函数
func TestMatchRule(t *testing.T) {
	engine := &MatchEngine{}

	tests := []struct {
		name        string
		request     *adapter.Request
		rule        *models.Rule
		expectMatch bool
		expectError bool
	}{
		{
			name: "简单匹配类型",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/test",
				},
			},
			expectMatch: true,
			expectError: false,
		},
		{
			name: "正则匹配类型(未实现)",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeRegex,
			},
			expectMatch: false,
			expectError: true,
		},
		{
			name: "脚本匹配类型(未实现)",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeScript,
			},
			expectMatch: false,
			expectError: true,
		},
		{
			name: "不支持的匹配类型",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: "unknown",
			},
			expectMatch: false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := engine.matchRule(tt.request, tt.rule)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectMatch, matched)
		})
	}
}

// TestSimpleMatch_ComplexConditions 测试复杂条件组合
func TestSimpleMatch_ComplexConditions(t *testing.T) {
	engine := &MatchEngine{}

	tests := []struct {
		name     string
		request  *adapter.Request
		rule     *models.Rule
		expected bool
	}{
		{
			name: "所有条件都匹配",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users/123",
				SourceIP: "192.168.1.100",
				Headers: map[string]string{
					"Content-Type": "application/json",
					"X-API-Key":    "secret123",
				},
				Metadata: map[string]interface{}{
					"method": "GET",
					"query": map[string]string{
						"status": "active",
						"role":   "admin",
					},
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/users/:id",
					"query": map[string]interface{}{
						"status": "active",
						"role":   "admin",
					},
					"headers": map[string]interface{}{
						"Content-Type": "application/json",
						"X-API-Key":    "secret123",
					},
					"ip_whitelist": []interface{}{
						"192.168.1.100",
					},
				},
			},
			expected: true,
		},
		{
			name: "Query参数不完全匹配",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/users",
				Metadata: map[string]interface{}{
					"method": "GET",
					"query": map[string]string{
						"status": "inactive",
					},
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/users",
					"query": map[string]interface{}{
						"status": "active",
					},
				},
			},
			expected: false,
		},
		{
			name: "Header不匹配",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
				Headers: map[string]string{
					"Content-Type": "text/plain",
				},
				Metadata: map[string]interface{}{
					"method": "POST",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "POST",
					"path":   "/api/test",
					"headers": map[string]interface{}{
						"Content-Type": "application/json",
					},
				},
			},
			expected: false,
		},
		{
			name: "IP白名单不匹配导致整体不匹配",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/secure",
				SourceIP: "10.0.0.1",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
					"path":   "/api/secure",
					"ip_whitelist": []interface{}{
						"192.168.1.100",
					},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := engine.simpleMatch(tt.request, tt.rule)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, matched)
		})
	}
}

// TestSimpleMatch_EdgeCases 测试边界场景
func TestSimpleMatch_EdgeCases(t *testing.T) {
	engine := &MatchEngine{}

	tests := []struct {
		name     string
		request  *adapter.Request
		rule     *models.Rule
		expected bool
		hasError bool
	}{
		{
			name: "空路径条件",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/any/path",
				Metadata: map[string]interface{}{
					"method": "GET",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"method": "GET",
				},
			},
			expected: true,
			hasError: false,
		},
		{
			name: "空Method条件",
			request: &adapter.Request{
				Protocol: models.ProtocolHTTP,
				Path:     "/api/test",
				Metadata: map[string]interface{}{
					"method": "POST",
				},
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolHTTP,
				MatchType: models.MatchTypeSimple,
				MatchCondition: map[string]interface{}{
					"path": "/api/test",
				},
			},
			expected: true,
			hasError: false,
		},
		{
			name: "非HTTP协议",
			request: &adapter.Request{
				Protocol: models.ProtocolWebSocket,
				Path:     "/ws",
			},
			rule: &models.Rule{
				Protocol:  models.ProtocolWebSocket,
				MatchType: models.MatchTypeSimple,
			},
			expected: false,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := engine.simpleMatch(tt.request, tt.rule)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, matched)
		})
	}
}

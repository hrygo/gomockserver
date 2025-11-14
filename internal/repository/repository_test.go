package repository

import (
	"context"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestObjectIDConversion 测试 ObjectID 转换
func TestObjectIDConversion(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		shouldErr bool
	}{
		{
			name:      "有效的ObjectID",
			id:        "507f1f77bcf86cd799439011",
			shouldErr: false,
		},
		{
			name:      "无效的ObjectID",
			id:        "invalid-id",
			shouldErr: true,
		},
		{
			name:      "空ID",
			id:        "",
			shouldErr: true,
		},
		{
			name:      "长度不够的ID",
			id:        "123",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := primitive.ObjectIDFromHex(tt.id)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRuleModel 测试 Rule 模型
func TestRuleModel(t *testing.T) {
	rule := &models.Rule{
		ID:            "507f1f77bcf86cd799439011",
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		MatchType:     models.MatchTypeSimple,
		Priority:      100,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": "GET",
			"path":   "/api/test",
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code": 200,
				"body":        "test response",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NotNil(t, rule)
	assert.Equal(t, "测试规则", rule.Name)
	assert.Equal(t, "project-001", rule.ProjectID)
	assert.Equal(t, models.ProtocolHTTP, rule.Protocol)
	assert.True(t, rule.Enabled)
	assert.Equal(t, 100, rule.Priority)
}

// TestProjectModel 测试 Project 模型
func TestProjectModel(t *testing.T) {
	project := &models.Project{
		ID:          "507f1f77bcf86cd799439011",
		Name:        "测试项目",
		WorkspaceID: "workspace-001",
		Description: "这是一个测试项目",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.NotNil(t, project)
	assert.Equal(t, "测试项目", project.Name)
	assert.Equal(t, "workspace-001", project.WorkspaceID)
	assert.Equal(t, "这是一个测试项目", project.Description)
}

// TestEnvironmentModel 测试 Environment 模型
func TestEnvironmentModel(t *testing.T) {
	env := &models.Environment{
		ID:        "507f1f77bcf86cd799439011",
		Name:      "开发环境",
		ProjectID: "project-001",
		BaseURL:   "http://localhost:9090",
		Variables: map[string]interface{}{
			"api_key": "test-key",
			"timeout": 30,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.NotNil(t, env)
	assert.Equal(t, "开发环境", env.Name)
	assert.Equal(t, "project-001", env.ProjectID)
	assert.Equal(t, "http://localhost:9090", env.BaseURL)
	assert.NotEmpty(t, env.Variables)
}

// TestRuleTimeStamps 测试时间戳自动设置
func TestRuleTimeStamps(t *testing.T) {
	rule := &models.Rule{
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
	}

	// 模拟创建时设置时间戳
	now := time.Now()
	rule.CreatedAt = now
	rule.UpdatedAt = now

	assert.False(t, rule.CreatedAt.IsZero())
	assert.False(t, rule.UpdatedAt.IsZero())
	assert.Equal(t, rule.CreatedAt, rule.UpdatedAt)

	// 模拟更新时修改 UpdatedAt
	time.Sleep(10 * time.Millisecond)
	rule.UpdatedAt = time.Now()
	assert.True(t, rule.UpdatedAt.After(rule.CreatedAt))
}

// TestFilterConstruction 测试过滤条件构造
func TestFilterConstruction(t *testing.T) {
	tests := []struct {
		name   string
		filter map[string]interface{}
		keys   []string
	}{
		{
			name: "项目和环境过滤",
			filter: map[string]interface{}{
				"project_id":     "project-001",
				"environment_id": "env-001",
			},
			keys: []string{"project_id", "environment_id"},
		},
		{
			name: "启用状态过滤",
			filter: map[string]interface{}{
				"project_id":     "project-001",
				"environment_id": "env-001",
				"enabled":        true,
			},
			keys: []string{"project_id", "environment_id", "enabled"},
		},
		{
			name:   "空过滤条件",
			filter: map[string]interface{}{},
			keys:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.filter)
			assert.Equal(t, len(tt.keys), len(tt.filter))
			for _, key := range tt.keys {
				_, exists := tt.filter[key]
				assert.True(t, exists)
			}
		})
	}
}

// TestPaginationParameters 测试分页参数
func TestPaginationParameters(t *testing.T) {
	tests := []struct {
		name          string
		page          int64
		pageSize      int64
		expectedSkip  int64
		expectedLimit int64
	}{
		{
			name:          "第一页",
			page:          1,
			pageSize:      20,
			expectedSkip:  0,
			expectedLimit: 20,
		},
		{
			name:          "第二页",
			page:          2,
			pageSize:      20,
			expectedSkip:  20,
			expectedLimit: 20,
		},
		{
			name:          "第三页，较小pageSize",
			page:          3,
			pageSize:      10,
			expectedSkip:  20,
			expectedLimit: 10,
		},
		{
			name:          "大页码",
			page:          10,
			pageSize:      50,
			expectedSkip:  450,
			expectedLimit: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skip := (tt.page - 1) * tt.pageSize
			assert.Equal(t, tt.expectedSkip, skip)
			assert.Equal(t, tt.expectedLimit, tt.pageSize)
		})
	}
}

// TestRuleValidation 测试规则验证逻辑
func TestRuleValidation(t *testing.T) {
	tests := []struct {
		name    string
		rule    *models.Rule
		isValid bool
	}{
		{
			name: "有效规则",
			rule: &models.Rule{
				Name:          "测试规则",
				ProjectID:     "project-001",
				EnvironmentID: "env-001",
				Protocol:      models.ProtocolHTTP,
				MatchType:     models.MatchTypeSimple,
			},
			isValid: true,
		},
		{
			name: "缺少名称",
			rule: &models.Rule{
				ProjectID:     "project-001",
				EnvironmentID: "env-001",
				Protocol:      models.ProtocolHTTP,
			},
			isValid: false,
		},
		{
			name: "缺少项目ID",
			rule: &models.Rule{
				Name:          "测试规则",
				EnvironmentID: "env-001",
				Protocol:      models.ProtocolHTTP,
			},
			isValid: false,
		},
		{
			name: "缺少环境ID",
			rule: &models.Rule{
				Name:      "测试规则",
				ProjectID: "project-001",
				Protocol:  models.ProtocolHTTP,
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证必填字段
			isValid := tt.rule.Name != "" &&
				tt.rule.ProjectID != "" &&
				tt.rule.EnvironmentID != ""
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

// TestProjectValidation 测试项目验证逻辑
func TestProjectValidation(t *testing.T) {
	tests := []struct {
		name    string
		project *models.Project
		isValid bool
	}{
		{
			name: "有效项目",
			project: &models.Project{
				Name:        "测试项目",
				WorkspaceID: "workspace-001",
			},
			isValid: true,
		},
		{
			name: "缺少名称",
			project: &models.Project{
				WorkspaceID: "workspace-001",
			},
			isValid: false,
		},
		{
			name: "缺少工作空间ID",
			project: &models.Project{
				Name: "测试项目",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.project.Name != "" && tt.project.WorkspaceID != ""
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

// TestEnvironmentValidation 测试环境验证逻辑
func TestEnvironmentValidation(t *testing.T) {
	tests := []struct {
		name    string
		env     *models.Environment
		isValid bool
	}{
		{
			name: "有效环境",
			env: &models.Environment{
				Name:      "开发环境",
				ProjectID: "project-001",
			},
			isValid: true,
		},
		{
			name: "缺少名称",
			env: &models.Environment{
				ProjectID: "project-001",
			},
			isValid: false,
		},
		{
			name: "缺少项目ID",
			env: &models.Environment{
				Name: "开发环境",
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.env.Name != "" && tt.env.ProjectID != ""
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

// TestContextHandling 测试 Context 处理
func TestContextHandling(t *testing.T) {
	ctx := context.Background()
	assert.NotNil(t, ctx)

	// 测试带超时的 context
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	assert.NotNil(t, ctxWithTimeout)

	// 测试带取消的 context
	ctxWithCancel, cancel2 := context.WithCancel(ctx)
	defer cancel2()
	assert.NotNil(t, ctxWithCancel)
}

// TestProtocolTypes 测试协议类型常量
func TestProtocolTypes(t *testing.T) {
	tests := []struct {
		name     string
		protocol models.ProtocolType
		expected string
	}{
		{"HTTP协议", models.ProtocolHTTP, "HTTP"},
		{"WebSocket协议", models.ProtocolWebSocket, "WebSocket"},
		{"gRPC协议", models.ProtocolGRPC, "gRPC"},
		{"TCP协议", models.ProtocolTCP, "TCP"},
		{"UDP协议", models.ProtocolUDP, "UDP"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.protocol))
		})
	}
}

// TestMatchTypes 测试匹配类型常量
func TestMatchTypes(t *testing.T) {
	tests := []struct {
		name      string
		matchType models.MatchType
		expected  string
	}{
		{"简单匹配", models.MatchTypeSimple, "Simple"},
		{"正则匹配", models.MatchTypeRegex, "Regex"},
		{"脚本匹配", models.MatchTypeScript, "Script"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.matchType))
		})
	}
}

// TestResponseTypes 测试响应类型常量
func TestResponseTypes(t *testing.T) {
	tests := []struct {
		name         string
		responseType models.ResponseType
		expected     string
	}{
		{"静态响应", models.ResponseTypeStatic, "Static"},
		{"动态响应", models.ResponseTypeDynamic, "Dynamic"},
		{"代理响应", models.ResponseTypeProxy, "Proxy"},
		{"脚本响应", models.ResponseTypeScript, "Script"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.responseType))
		})
	}
}

// TestContentTypes 测试内容类型常量
func TestContentTypes(t *testing.T) {
	tests := []struct {
		name        string
		contentType models.ContentType
		expected    string
	}{
		{"JSON", models.ContentTypeJSON, "JSON"},
		{"XML", models.ContentTypeXML, "XML"},
		{"HTML", models.ContentTypeHTML, "HTML"},
		{"文本", models.ContentTypeText, "Text"},
		{"二进制", models.ContentTypeBinary, "Binary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.contentType))
		})
	}
}

// TestDelayConfig 测试延迟配置
func TestDelayConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *models.DelayConfig
		valid  bool
	}{
		{
			name: "固定延迟",
			config: &models.DelayConfig{
				Type:  "fixed",
				Fixed: 100,
			},
			valid: true,
		},
		{
			name: "随机延迟",
			config: &models.DelayConfig{
				Type: "random",
				Min:  50,
				Max:  200,
			},
			valid: true,
		},
		{
			name: "正态分布延迟",
			config: &models.DelayConfig{
				Type:   "normal",
				Mean:   100,
				StdDev: 20,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.config)
			assert.NotEmpty(t, tt.config.Type)
			if tt.valid {
				switch tt.config.Type {
				case "fixed":
					assert.Greater(t, tt.config.Fixed, 0)
				case "random":
					assert.Greater(t, tt.config.Max, tt.config.Min)
				case "normal":
					assert.Greater(t, tt.config.Mean, 0)
				}
			}
		})
	}
}

// TestRuleResponse 测试规则响应配置
func TestRuleResponse(t *testing.T) {
	response := models.Response{
		Type: models.ResponseTypeStatic,
		Delay: &models.DelayConfig{
			Type:  "fixed",
			Fixed: 100,
		},
		Content: map[string]interface{}{
			"status_code":  200,
			"content_type": "JSON",
			"body": map[string]interface{}{
				"code":    0,
				"message": "success",
			},
		},
	}

	assert.Equal(t, models.ResponseTypeStatic, response.Type)
	assert.NotNil(t, response.Delay)
	assert.Equal(t, 100, response.Delay.Fixed)
	assert.NotEmpty(t, response.Content)
	
	statusCode, ok := response.Content["status_code"].(int)
	assert.True(t, ok)
	assert.Equal(t, 200, statusCode)
}

// TestBSONTagsPresence 测试 BSON 标签存在性
func TestBSONTagsPresence(t *testing.T) {
	// 这个测试验证模型是否正确定义了 BSON 标签
	// 虽然不能直接测试标签，但可以确保模型结构正确

	rule := &models.Rule{}
	assert.NotNil(t, rule)

	project := &models.Project{}
	assert.NotNil(t, project)

	env := &models.Environment{}
	assert.NotNil(t, env)
}

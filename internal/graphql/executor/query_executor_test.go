package executor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/graphql/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryExecutor_ExecuteQuery(t *testing.T) {
	executor := NewQueryExecutor()

	tests := []struct {
		name    string
		query   string
		op      string
		wantErr bool
	}{
		{
			name:    "简单查询",
			query:   "{ hello }",
			op:      "QUERY",
			wantErr: false,
		},
		{
			name:    "多个字段查询",
			query:   "{ hello status }",
			op:      "QUERY",
			wantErr: false,
		},
		{
			name:    "查询用户",
			query:   "{ user }",
			op:      "QUERY",
			wantErr: false,
		},
		{
			name:    "查询用户列表",
			query:   "{ users }",
			op:      "QUERY",
			wantErr: false,
		},
		{
			name:    "空查询",
			query:   "",
			op:      "QUERY",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			execCtx := &types.ExecutionContext{
				RequestID: "test-" + tt.name,
				Query: &types.GraphQLQuery{
					ID:        "test-query",
					Query:     tt.query,
					Operation: tt.op,
					Variables: make(map[string]interface{}),
					Timestamp: time.Now(),
				},
				Variables: make(map[string]interface{}),
				Operation: tt.op,
				Headers:   make(map[string]string),
				Metadata:  make(map[string]interface{}),
				StartTime: time.Now(),
			}

			result, err := executor.ExecuteQuery(ctx, execCtx)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.Data)
			assert.NotNil(t, result.Extensions)

			// 验证扩展信息存在
			assert.Contains(t, result.Extensions, "executionTime")
			assert.Contains(t, result.Extensions, "timestamp")
		})
	}
}

func TestQueryExecutor_ExecuteMutation(t *testing.T) {
	executor := NewQueryExecutor()

	ctx := context.Background()

	execCtx := &types.ExecutionContext{
		RequestID: "test-mutation",
		Query: &types.GraphQLQuery{
			ID:        "test-mutation",
			Query:     "{ createUser }",
			Operation: "MUTATION",
			Variables: make(map[string]interface{}),
			Timestamp: time.Now(),
		},
		Variables: make(map[string]interface{}),
		Operation: "MUTATION",
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		StartTime: time.Now(),
	}

	result, err := executor.ExecuteQuery(ctx, execCtx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Data)
}

func TestQueryExecutor_ParseSimpleQuery(t *testing.T) {
	executor := NewQueryExecutor()

	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "简单单字段",
			query:    "{ hello }",
			expected: []string{"hello"},
		},
		{
			name:     "多个字段",
			query:    "{ hello status user }",
			expected: []string{"hello", "status", "user"},
		},
		{
			name:     "带操作名的查询",
			query:    "query { hello status }",
			expected: []string{"hello", "status"},
		},
		{
			name:     "复杂查询格式",
			query:    "query GetUserList {\n  users\n  hello\n  status\n}",
			expected: []string{"users", "hello", "status"},
		},
		{
			name:     "空查询",
			query:    "",
			expected: nil,
		},
		{
			name:     "只有大括号",
			query:    "{}",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.parseSimpleQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMockResolver(t *testing.T) {
	resolver := NewMockResolver()

	tests := []struct {
		name         string
		fieldName    string
		parentType   string
		expectError  bool
		expectTypename string
	}{
		{
			name:            "hello字段",
			fieldName:       "hello",
			parentType:      "Query",
			expectError:     false,
			expectTypename:  "HelloResponse",
		},
		{
			name:            "status字段",
			fieldName:       "status",
			parentType:      "Query",
			expectError:     false,
			expectTypename:  "ServerStatus",
		},
		{
			name:            "user字段",
			fieldName:       "user",
			parentType:      "Query",
			expectError:     false,
			expectTypename:  "User",
		},
		{
			name:            "users字段",
			fieldName:       "users",
			parentType:      "Query",
			expectError:     false,
			expectTypename:  "",
		}, // users返回数组
		{
			name:            "未知字段",
			fieldName:       "unknown",
			parentType:      "Query",
			expectError:     true,
			expectTypename:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fieldCtx := &types.FieldContext{
				ParentType: tt.parentType,
				FieldName:  tt.fieldName,
				Arguments:  make(map[string]interface{}),
				Alias:      tt.fieldName,
				Path:       []string{tt.fieldName},
			}

			result, err := resolver.Resolve(ctx, fieldCtx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// 检查返回值是否包含__typename（如果是对象）
				if resultMap, ok := result.(map[string]interface{}); ok {
					if tt.expectTypename != "" {
						assert.Equal(t, tt.expectTypename, resultMap["__typename"])
					}
				}
			}
		})
	}
}

func TestResolverManager(t *testing.T) {
	rm := NewResolverManager()

	// 注册静态解析器
	staticResolver := NewStaticResolver("test-value")
	rm.RegisterResolver("TestType", "testField", staticResolver)

	// 注册Mock解析器
	mockResolver := NewMockResolver()
	rm.RegisterResolver("Query", "hello", mockResolver)

	tests := []struct {
		name        string
		typeName    string
		fieldName   string
		expectError bool
		expectValue interface{}
	}{
		{
			name:        "注册的静态解析器",
			typeName:    "TestType",
			fieldName:   "testField",
			expectError: false,
			expectValue: "test-value",
		},
		{
			name:        "注册的Mock解析器",
			typeName:    "Query",
			fieldName:   "hello",
			expectError: false,
			expectValue: nil, // Mock解析器返回复杂对象
		},
		{
			name:        "未注册的解析器",
			typeName:    "UnknownType",
			fieldName:   "unknownField",
			expectError: true, // Mock解析器对未知字段会返回错误
			expectValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fieldCtx := &types.FieldContext{
				ParentType: tt.typeName,
				FieldName:  tt.fieldName,
				Arguments:  make(map[string]interface{}),
				Alias:      tt.fieldName,
				Path:       []string{tt.fieldName},
			}

			result, err := rm.ResolveField(ctx, fieldCtx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				if tt.expectValue != nil {
					assert.Equal(t, tt.expectValue, result)
				}
			}
		})
	}
}

func TestStaticResolver(t *testing.T) {
	resolver := NewStaticResolver("static-test-value")

	ctx := context.Background()
	fieldCtx := &types.FieldContext{
		ParentType: "TestType",
		FieldName:  "testField",
		Arguments:  make(map[string]interface{}),
		Alias:      "testField",
		Path:       []string{"testField"},
	}

	result, err := resolver.Resolve(ctx, fieldCtx)
	assert.NoError(t, err)
	assert.Equal(t, "static-test-value", result)
}

func TestDynamicResolver(t *testing.T) {
	data := map[string]interface{}{
		"field1": "value1",
		"field2": 42,
		"field3": map[string]interface{}{
			"nested": "data",
		},
	}

	resolver := NewDynamicResolver(data)

	tests := []struct {
		name         string
		fieldName    string
		expectError  bool
		expectValue  interface{}
	}{
		{
			name:        "存在的字段1",
			fieldName:   "field1",
			expectError: false,
			expectValue: "value1",
		},
		{
			name:        "存在的字段2",
			fieldName:   "field2",
			expectError: false,
			expectValue: 42,
		},
		{
			name:        "存在的字段3",
			fieldName:   "field3",
			expectError: false,
			expectValue: map[string]interface{}{"nested": "data"},
		},
		{
			name:        "不存在的字段",
			fieldName:   "unknown",
			expectError: true,
			expectValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fieldCtx := &types.FieldContext{
				ParentType: "TestType",
				FieldName:  tt.fieldName,
				Arguments:  make(map[string]interface{}),
				Alias:      tt.fieldName,
				Path:       []string{tt.fieldName},
			}

			result, err := resolver.Resolve(ctx, fieldCtx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectValue, result)
			}
		})
	}
}

func TestQueryExecutor_AddValidatorAndMiddleware(t *testing.T) {
	executor := NewQueryExecutor()

	// 测试添加验证器
	validator := &MockValidator{}
	executor.AddValidator(validator)

	assert.Len(t, executor.validators, 1)

	// 测试添加中间件
	middleware := &MockMiddleware{}
	executor.AddMiddleware(middleware)

	assert.Len(t, executor.middleware, 1)
}

func TestQueryExecutor_ExecuteSubscription(t *testing.T) {
	executor := NewQueryExecutor()

	ctx := context.Background()
	execCtx := &types.ExecutionContext{
		RequestID: "test-subscription",
		Query: &types.GraphQLQuery{
			ID:        "test-subscription",
			Query:     "{ userUpdated }",
			Operation: "SUBSCRIPTION",
			Variables: make(map[string]interface{}),
			Timestamp: time.Now(),
		},
		Variables: make(map[string]interface{}),
		Operation: "SUBSCRIPTION",
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		StartTime: time.Now(),
	}

	result, err := executor.ExecuteQuery(ctx, execCtx)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.Data)
}

func TestQueryExecutor_SchemaValidator(t *testing.T) {
	validator := NewSchemaValidator()

	tests := []struct {
		name        string
		execCtx     *types.ExecutionContext
		expectError bool
	}{
		{
			name: "有效的执行上下文",
			execCtx: &types.ExecutionContext{
				Schema: &types.GraphQLSchema{},
				Query: &types.GraphQLQuery{},
				Operation: "QUERY",
			},
			expectError: false,
		},
		{
			name: "缺少schema",
			execCtx: &types.ExecutionContext{
				Query: &types.GraphQLQuery{},
				Operation: "QUERY",
			},
			expectError: true,
		},
		{
			name: "缺少query",
			execCtx: &types.ExecutionContext{
				Schema: &types.GraphQLSchema{},
				Operation: "QUERY",
			},
			expectError: true,
		},
		{
			name: "缺少operation",
			execCtx: &types.ExecutionContext{
				Schema: &types.GraphQLSchema{},
				Query: &types.GraphQLQuery{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := validator.Validate(ctx, tt.execCtx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQueryExecutor_SecurityValidator(t *testing.T) {
	validator := NewSecurityValidator()

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "正常查询",
			query:       "{ hello }",
			expectError: false,
		},
		{
			name:        "查询过长",
			query:       strings.Repeat("{ field }", 1000),
			expectError: true,
		},
		{
			name:        "空查询",
			query:       "",
			expectError: false, // SecurityValidator不检查空查询
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			execCtx := &types.ExecutionContext{
				Query: &types.GraphQLQuery{
					Query: tt.query,
				},
			}

			err := validator.Validate(ctx, execCtx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "查询过于复杂")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestQueryExecutor_Middleware(t *testing.T) {
	// 测试日志中间件
	loggingMiddleware := NewLoggingMiddleware()
	assert.NotNil(t, loggingMiddleware)

	// 测试指标中间件
	metricsMiddleware := NewMetricsMiddleware()
	assert.NotNil(t, metricsMiddleware)

	// 测试超时中间件
	timeoutMiddleware := NewTimeoutMiddleware(time.Second)
	assert.NotNil(t, timeoutMiddleware)
}

func TestQueryExecutor_MiddlewareChain(t *testing.T) {
	executor := NewQueryExecutor()
	executor.AddMiddleware(&MockMiddleware{})

	ctx := context.Background()
	execCtx := &types.ExecutionContext{
		RequestID: "test-middleware-chain",
		Query: &types.GraphQLQuery{
			ID:        "test-middleware-chain",
			Query:     "{ hello }",
			Operation: "QUERY",
			Variables: make(map[string]interface{}),
			Timestamp: time.Now(),
		},
		Variables: make(map[string]interface{}),
		Operation: "QUERY",
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		StartTime: time.Now(),
	}

	result, err := executor.ExecuteQuery(ctx, execCtx)
	require.NoError(t, err)
	assert.NotNil(t, result)
}

func TestProxyResolver(t *testing.T) {
	resolver := NewProxyResolver("http://example.com")
	assert.NotNil(t, resolver)

	ctx := context.Background()
	fieldCtx := &types.FieldContext{
		ParentType: "TestType",
		FieldName:  "testField",
		Arguments:  make(map[string]interface{}),
		Alias:      "testField",
		Path:       []string{"testField"},
	}

	result, err := resolver.Resolve(ctx, fieldCtx)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// 检查返回的结果结构
	if resultMap, ok := result.(map[string]interface{}); ok {
		assert.Equal(t, true, resultMap["proxied"])
		assert.Equal(t, "testField", resultMap["field"])
		assert.Contains(t, resultMap, "timestamp")
		assert.Equal(t, "ProxyResult", resultMap["__typename"])
	}
}

// MockValidator 用于测试的模拟验证器
type MockValidator struct{}

func (v *MockValidator) Validate(ctx context.Context, execCtx *types.ExecutionContext) error {
	return nil
}

// MockMiddleware 用于测试的模拟中间件
type MockMiddleware struct{}

func (m *MockMiddleware) Handle(ctx context.Context, execCtx *types.ExecutionContext, next QueryHandler) *types.GraphQLResult {
	// 直接调用下一个处理器
	return next(ctx, execCtx)
}
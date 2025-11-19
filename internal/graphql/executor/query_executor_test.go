package executor

import (
	"context"
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
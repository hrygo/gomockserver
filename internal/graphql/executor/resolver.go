package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/gomockserver/mockserver/internal/graphql/types"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// Resolver 基础解析器接口
type Resolver interface {
	Resolve(ctx context.Context, fieldCtx *types.FieldContext) (interface{}, error)
}

// StaticResolver 静态解析器 - 返回预定义值
type StaticResolver struct {
	Value interface{}
}

// NewStaticResolver 创建静态解析器
func NewStaticResolver(value interface{}) *StaticResolver {
	return &StaticResolver{Value: value}
}

// Resolve 执行静态解析
func (r *StaticResolver) Resolve(ctx context.Context, fieldCtx *types.FieldContext) (interface{}, error) {
	return r.Value, nil
}

// DynamicResolver 动态解析器 - 基于字段名返回值
type DynamicResolver struct {
	Data map[string]interface{}
}

// NewDynamicResolver 创建动态解析器
func NewDynamicResolver(data map[string]interface{}) *DynamicResolver {
	return &DynamicResolver{Data: data}
}

// Resolve 执行动态解析
func (r *DynamicResolver) Resolve(ctx context.Context, fieldCtx *types.FieldContext) (interface{}, error) {
	if value, exists := r.Data[fieldCtx.FieldName]; exists {
		return value, nil
	}
	return nil, fmt.Errorf("字段 '%s' 未找到", fieldCtx.FieldName)
}

// MockResolver Mock解析器 - 为测试和演示提供模拟数据
type MockResolver struct {
	logger *zap.Logger
}

// NewMockResolver 创建Mock解析器
func NewMockResolver() *MockResolver {
	return &MockResolver{
		logger: logger.Get().Named("mock-resolver"),
	}
}

// Resolve 执行Mock解析
func (r *MockResolver) Resolve(ctx context.Context, fieldCtx *types.FieldContext) (interface{}, error) {
	r.logger.Debug("执行Mock解析",
		zap.String("field", fieldCtx.FieldName),
		zap.String("parent_type", fieldCtx.ParentType))

	switch fieldCtx.FieldName {
	case "user":
		return map[string]interface{}{
			"id":         "mock-user-1",
			"name":       "Mock User",
			"email":      "mock@example.com",
			"createdAt":  time.Now().Format(time.RFC3339),
			"__typename": "User",
		}, nil
	case "users":
		return []map[string]interface{}{
			{
				"id":         "mock-user-1",
				"name":       "Mock User 1",
				"email":      "mock1@example.com",
				"createdAt":  time.Now().Format(time.RFC3339),
				"__typename": "User",
			},
			{
				"id":         "mock-user-2",
				"name":       "Mock User 2",
				"email":      "mock2@example.com",
				"createdAt":  time.Now().Add(-time.Hour).Format(time.RFC3339),
				"__typename": "User",
			},
		}, nil
	case "hello":
		return map[string]interface{}{
			"message":    "Hello from MockServer GraphQL!",
			"timestamp":  time.Now().Unix(),
			"__typename": "HelloResponse",
		}, nil
	case "status":
		return map[string]interface{}{
			"status":     "healthy",
			"version":    "0.8.0",
			"timestamp":  time.Now().Format(time.RFC3339),
			"__typename": "ServerStatus",
		}, nil
	case "_service":
		return map[string]interface{}{
			"sdl": "type Query { user(id: ID!): User users: [User!]! hello: HelloResponse status: ServerStatus } type User { id: ID! name: String! email: String createdAt: String! } type HelloResponse { message: String! timestamp: Int! } type ServerStatus { status: String! version: String! timestamp: String! }",
			"__typename": "Service",
		}, nil
	default:
		return nil, fmt.Errorf("未知的Mock字段: %s", fieldCtx.FieldName)
	}
}

// ProxyResolver 代理解析器 - 转发到其他服务
type ProxyResolver struct {
	BaseURL string
	logger  *zap.Logger
}

// NewProxyResolver 创建代理解析器
func NewProxyResolver(baseURL string) *ProxyResolver {
	return &ProxyResolver{
		BaseURL: baseURL,
		logger:  logger.Get().Named("proxy-resolver"),
	}
}

// Resolve 执行代理解析
func (r *ProxyResolver) Resolve(ctx context.Context, fieldCtx *types.FieldContext) (interface{}, error) {
	r.logger.Debug("执行代理解析",
		zap.String("field", fieldCtx.FieldName),
		zap.String("base_url", r.BaseURL))

	// 简化实现 - 将在Phase 3中完善HTTP代理功能
	// 目前返回一个模拟的代理结果
	return map[string]interface{}{
		"proxied":   true,
		"field":     fieldCtx.FieldName,
		"timestamp": time.Now().Unix(),
		"__typename": "ProxyResult",
	}, nil
}

// ResolverManager 解析器管理器
type ResolverManager struct {
	resolvers map[string]Resolver
	logger    *zap.Logger
}

// NewResolverManager 创建解析器管理器
func NewResolverManager() *ResolverManager {
	return &ResolverManager{
		resolvers: make(map[string]Resolver),
		logger:    logger.Get().Named("resolver-manager"),
	}
}

// RegisterResolver 注册解析器
func (rm *ResolverManager) RegisterResolver(typeName, fieldName string, resolver Resolver) {
	key := fmt.Sprintf("%s.%s", typeName, fieldName)
	rm.resolvers[key] = resolver
	rm.logger.Debug("注册解析器", zap.String("key", key))
}

// GetResolver 获取解析器
func (rm *ResolverManager) GetResolver(typeName, fieldName string) (Resolver, bool) {
	key := fmt.Sprintf("%s.%s", typeName, fieldName)
	resolver, exists := rm.resolvers[key]
	return resolver, exists
}

// ResolveField 解析字段
func (rm *ResolverManager) ResolveField(ctx context.Context, fieldCtx *types.FieldContext) (interface{}, error) {
	resolver, exists := rm.GetResolver(fieldCtx.ParentType, fieldCtx.FieldName)
	if !exists {
		// 如果没有注册特定解析器，使用默认Mock解析器
		mockResolver := NewMockResolver()
		return mockResolver.Resolve(ctx, fieldCtx)
	}

	return resolver.Resolve(ctx, fieldCtx)
}

// DefaultResolverManager 创建默认解析器管理器
func DefaultResolverManager() *ResolverManager {
	rm := NewResolverManager()

	// 注册Query类型的解析器
	mockResolver := NewMockResolver()
	rm.RegisterResolver("Query", "user", mockResolver)
	rm.RegisterResolver("Query", "users", mockResolver)
	rm.RegisterResolver("Query", "hello", mockResolver)
	rm.RegisterResolver("Query", "status", mockResolver)
	rm.RegisterResolver("Query", "_service", mockResolver)

	// 注册Mutation类型的解析器
	rm.RegisterResolver("Mutation", "createUser", mockResolver)
	rm.RegisterResolver("Mutation", "updateUser", mockResolver)

	return rm
}
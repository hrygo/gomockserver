package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/graphql/executor"
	"github.com/gomockserver/mockserver/internal/graphql/parser"
	"github.com/gomockserver/mockserver/internal/graphql/types"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// GraphQLHandler GraphQL HTTP处理器
type GraphQLHandler struct {
	queryExecutor  *executor.QueryExecutor
	schemaParser  *parser.SchemaParser
	queryParser   *parser.QueryParser
}

// NewGraphQLHandler 创建GraphQL处理器
func NewGraphQLHandler() *GraphQLHandler {
	return &GraphQLHandler{
		queryExecutor: executor.NewQueryExecutor(),
		schemaParser: parser.NewSchemaParser(),
		queryParser:  parser.NewQueryParser(),
	}
}

// GraphQLRequest GraphQL请求体
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse GraphQL响应体
type GraphQLResponse struct {
	Data       interface{}            `json:"data"`
	Errors     []*types.GraphQLErrorWrapper `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// RegisterRoutes 注册GraphQL路由
func (h *GraphQLHandler) RegisterRoutes(router *gin.Engine) {
	// GraphQL端点 - 支持GET和POST
	router.Any("/graphql", h.HandleGraphQL)

	// GraphQL Playground (开发环境)
	router.GET("/graphql-playground", h.HandlePlayground)

	// GraphQL Schema Introspection
	router.GET("/graphql/schema", h.HandleSchemaIntrospection)

	// GraphQL健康检查
	router.GET("/graphql/health", h.HandleHealth)
}

// HandleGraphQL 处理GraphQL请求
func (h *GraphQLHandler) HandleGraphQL(c *gin.Context) {
	startTime := time.Now()

	// 获取请求ID
	requestID := c.GetString("requestId")
	if requestID == "" {
		requestID = fmt.Sprintf("graphql-%d", startTime.UnixNano())
	}

	logger.Info("GraphQL请求开始",
		zap.String("request_id", requestID),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("remote_addr", c.ClientIP()))

	// 解析请求
	var graphqlReq GraphQLRequest
	var err error

	switch c.Request.Method {
	case http.MethodGet:
		// GET请求从查询参数解析
		graphqlReq = GraphQLRequest{
			Query:         c.Query("query"),
			Variables:     h.parseVariables(c.Query("variables")),
			OperationName: c.Query("operationName"),
		}
	case http.MethodPost:
		// POST请求从请求体解析
		err = c.ShouldBindJSON(&graphqlReq)
		if err != nil {
			logger.Error("解析GraphQL请求体失败",
				zap.String("request_id", requestID),
				zap.Error(err))
			h.sendError(c, http.StatusBadRequest, "无效的JSON格式: "+err.Error(), requestID)
			return
		}
	default:
		h.sendError(c, http.StatusMethodNotAllowed, "只支持GET和POST请求", requestID)
		return
	}

	// 验证请求
	if graphqlReq.Query == "" {
		h.sendError(c, http.StatusBadRequest, "查询不能为空", requestID)
		return
	}

	// 设置执行上下文
	execCtx := &types.ExecutionContext{
		RequestID: requestID,
		Query: &types.GraphQLQuery{
			ID:        fmt.Sprintf("query-%d", startTime.UnixNano()),
			Query:     graphqlReq.Query,
			Variables: graphqlReq.Variables,
			Operation: graphqlReq.OperationName,
			Timestamp: startTime,
		},
		Variables: graphqlReq.Variables,
		Operation: graphqlReq.OperationName,
		Headers:   h.getHeaders(c),
		Metadata:  h.getMetadata(c),
		StartTime: startTime,
	}

	// 如果没有指定操作类型，尝试从查询中推断
	if execCtx.Operation == "" {
		execCtx.Operation = h.inferOperationType(graphqlReq.Query)
	}

	// 执行GraphQL查询
	result, err := h.queryExecutor.ExecuteQuery(c.Request.Context(), execCtx)
	if err != nil {
		logger.Error("GraphQL查询执行失败",
			zap.String("request_id", requestID),
			zap.Error(err))
		h.sendError(c, http.StatusInternalServerError, "查询执行失败: "+err.Error(), requestID)
		return
	}

	// 转换响应格式
	response := GraphQLResponse{
		Data:       result.Data,
		Errors:     result.Errors,
		Extensions: result.Extensions,
	}

	// 记录执行时间
	executionTime := time.Since(startTime)
	logger.Info("GraphQL请求完成",
		zap.String("request_id", requestID),
		zap.Duration("execution_time", executionTime),
		zap.Int("error_count", len(result.Errors)))

	// 设置响应头
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}

// HandlePlayground 处理GraphQL Playground
func (h *GraphQLHandler) HandlePlayground(c *gin.Context) {
	playgroundHTML := `<!DOCTYPE html>
<html>
<head>
    <title>GraphQL Playground</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; margin: 0; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { text-align: center; margin-bottom: 30px; }
        .playground { border: 1px solid #ddd; border-radius: 8px; overflow: hidden; }
        .playground-iframe { width: 100%; height: 600px; border: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 MockServer GraphQL Playground</h1>
            <p>使用此界面测试您的GraphQL查询和变更</p>
        </div>
        <div class="playground">
            <iframe
                src="https://graphql.github.io/playground"
                class="playground-iframe"
                frameborder="0">
            </iframe>
        </div>
        <div style="margin-top: 20px; padding: 15px; background-color: #f5f5f5; border-radius: 5px;">
            <h3>📚 示例查询：</h3>
            <pre><code># 基础查询
query {
  hello
  status
}

# 查询用户
query {
  user {
    id
    name
    email
    createdAt
  }
}

# 查询用户列表
query {
  users {
    id
    name
    email
    createdAt
  }
}</code></pre>
        </div>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, playgroundHTML)
}

// HandleSchemaIntrospection 处理Schema内省
func (h *GraphQLHandler) HandleSchemaIntrospection(c *gin.Context) {
	schemaSDL := `
# MockServer GraphQL Schema

type Query {
  "Hello world问候"
  hello: HelloResponse

  "服务器状态"
  status: ServerStatus

  "获取单个用户"
  user(id: ID!): User

  "获取用户列表"
  users: [User!]!

  "服务信息"
  _service: Service
}

type Mutation {
  "创建用户"
  createUser(input: CreateUserInput!): User

  "更新用户"
  updateUser(id: ID!, input: UpdateUserInput!): User
}

type Subscription {
  "用户更新订阅"
  userUpdated(id: ID!): User
}

type HelloResponse {
  "问候消息"
  message: String!
  "时间戳"
  timestamp: Int!
}

type ServerStatus {
  "服务器状态"
  status: String!
  "版本号"
  version: String!
  "时间戳"
  timestamp: String!
}

type User {
  "用户ID"
  id: ID!
  "用户名"
  name: String!
  "用户邮箱"
  email: String
  "创建时间"
  createdAt: String!
}

type CreateUserInput {
  "用户名"
  name: String!
  "用户邮箱"
  email: String
}

type UpdateUserInput {
  "用户名"
  name: String
  "用户邮箱"
  email: String
}

type Service {
  "SDL定义"
  sdl: String!
}

# 内省类型
type __Schema {
  types: [__Type!]!
}

type __Type {
  kind: __TypeKind!
  name: String
  description: String
}

enum __TypeKind {
  SCALAR
  OBJECT
  INTERFACE
  UNION
  ENUM
  INPUT_OBJECT
  LIST
  NON_NULL
}
`

	response := map[string]interface{}{
		"data": map[string]interface{}{
			"__schema": map[string]interface{}{
				"types": []interface{}{}, // 这里可以进一步解析SDL
			},
		},
	}

	// 如果需要完整的Schema，可以返回SDL
	if c.Query("sdl") == "true" {
		response = map[string]interface{}{
			"sdl": schemaSDL,
		}
	}

	c.JSON(http.StatusOK, response)
}

// HandleHealth 处理GraphQL健康检查
func (h *GraphQLHandler) HandleHealth(c *gin.Context) {
	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "MockServer GraphQL",
		"version":   "0.8.0",
		"timestamp": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// sendError 发送错误响应
func (h *GraphQLHandler) sendError(c *gin.Context, statusCode int, message string, requestID string) {
	response := GraphQLResponse{
		Errors: []*types.GraphQLErrorWrapper{
			{
				Kind:    types.ErrorKindInternal,
				Message: message,
			},
		},
		Extensions: map[string]interface{}{
			"requestId": requestID,
			"timestamp": time.Now().Unix(),
		},
	}

	logger.Error("GraphQL错误",
		zap.String("request_id", requestID),
		zap.String("message", message),
		zap.Int("status_code", statusCode))

	c.JSON(statusCode, response)
}

// parseVariables 解析变量JSON字符串
func (h *GraphQLHandler) parseVariables(variablesStr string) map[string]interface{} {
	if variablesStr == "" {
		return make(map[string]interface{})
	}

	var variables map[string]interface{}
	err := json.Unmarshal([]byte(variablesStr), &variables)
	if err != nil {
		logger.Error("解析变量失败", zap.String("variables", variablesStr), zap.Error(err))
		return make(map[string]interface{})
	}

	return variables
}

// getHeaders 获取请求头
func (h *GraphQLHandler) getHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	return headers
}

// getMetadata 获取元数据
func (h *GraphQLHandler) getMetadata(c *gin.Context) map[string]interface{} {
	metadata := make(map[string]interface{})

	// 添加客户端信息
	metadata["clientIP"] = c.ClientIP()
	metadata["userAgent"] = c.Request.UserAgent()
	metadata["method"] = c.Request.Method
	metadata["path"] = c.Request.URL.Path

	return metadata
}

// inferOperationType 从查询字符串推断操作类型
func (h *GraphQLHandler) inferOperationType(query string) string {
	// 简单的启发式推断
	if containsIgnoreCase(query, "mutation") {
		return "MUTATION"
	}
	if containsIgnoreCase(query, "subscription") {
		return "SUBSCRIPTION"
	}
	return "QUERY"
}

// containsIgnoreCase 检查字符串是否包含子字符串（忽略大小写）
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    len(s) > len(substr) &&
		    (s[:len(substr)] == substr ||
		     s[len(s)-len(substr):] == substr ||
		     containsIgnoreCaseRec(s[1:], substr)))
}

func containsIgnoreCaseRec(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsIgnoreCaseRec(s[1:], substr)
}
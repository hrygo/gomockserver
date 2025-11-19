package executor

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gomockserver/mockserver/internal/graphql/types"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// QueryExecutor GraphQL查询执行器
type QueryExecutor struct {
	logger     *zap.Logger
	validators []QueryValidator
	middleware []QueryMiddleware
	resolver   *ResolverManager
}

// NewQueryExecutor 创建查询执行器
func NewQueryExecutor() *QueryExecutor {
	return &QueryExecutor{
		logger:     logger.Get().Named("graphql-executor"),
		validators: make([]QueryValidator, 0),
		middleware: make([]QueryMiddleware, 0),
		resolver:   DefaultResolverManager(),
	}
}

// AddValidator 添加查询验证器
func (e *QueryExecutor) AddValidator(validator QueryValidator) {
	e.validators = append(e.validators, validator)
}

// AddMiddleware 添加查询中间件
func (e *QueryExecutor) AddMiddleware(middleware QueryMiddleware) {
	e.middleware = append(e.middleware, middleware)
}

// ExecuteQuery 执行GraphQL查询
func (e *QueryExecutor) ExecuteQuery(ctx context.Context, execCtx *types.ExecutionContext) (*types.GraphQLResult, error) {
	startTime := time.Now()

	e.logger.Info("开始执行GraphQL查询",
		zap.String("request_id", execCtx.RequestID),
		zap.String("operation", execCtx.Operation))

	// 执行中间件
	result := e.executeMiddlewareChain(ctx, execCtx, e.executeQueryInternal)

	// 计算执行时间
	executionTime := time.Since(startTime)
	if result.Data == nil {
		result.Data = map[string]interface{}{}
	}

	// 添加执行扩展信息
	if result.Extensions == nil {
		result.Extensions = make(map[string]interface{})
	}
	result.Extensions["executionTime"] = executionTime.String()
	result.Extensions["timestamp"] = time.Now().Unix()

	e.logger.Info("查询执行完成",
		zap.String("request_id", execCtx.RequestID),
		zap.Duration("execution_time", executionTime),
		zap.Int("error_count", len(result.Errors)))

	return result, nil
}

// executeQueryInternal 内部查询执行
func (e *QueryExecutor) executeQueryInternal(ctx context.Context, execCtx *types.ExecutionContext) *types.GraphQLResult {
	result := &types.GraphQLResult{
		Data:   make(map[string]interface{}),
		Errors: make([]*types.GraphQLErrorWrapper, 0),
	}

	// 验证查询
	if err := e.validateQuery(ctx, execCtx); err != nil {
		result.Errors = append(result.Errors, e.wrapError(err, types.ErrorKindValidation))
		return result
	}

	// 根据操作类型执行查询
	switch execCtx.Operation {
	case string(types.Query):
		e.executeSelectQuery(ctx, execCtx, result)
	case string(types.Mutation):
		e.executeMutationQuery(ctx, execCtx, result)
	case string(types.Subscription):
		e.executeSubscriptionQuery(ctx, execCtx, result)
	default:
		result.Errors = append(result.Errors, &types.GraphQLErrorWrapper{
			Kind:    types.ErrorKindSyntax,
			Message: fmt.Sprintf("不支持的操作类型: %s", execCtx.Operation),
		})
	}

	return result
}

// executeSelectQuery 执行SELECT查询
func (e *QueryExecutor) executeSelectQuery(ctx context.Context, execCtx *types.ExecutionContext, result *types.GraphQLResult) {
	e.logger.Debug("执行SELECT查询", zap.String("request_id", execCtx.RequestID))

	// 解析查询字符串，提取请求的字段
	queryFields := e.parseSimpleQuery(execCtx.Query.Query)
	if len(queryFields) == 0 {
		// 如果无法解析，返回默认结果
		result.Data = map[string]interface{}{
			"__typename": "QueryResult",
			"status":     "success",
			"timestamp":  time.Now().Unix(),
		}
		return
	}

	// 执行每个字段的解析
	data := make(map[string]interface{})
	for _, fieldName := range queryFields {
		// 解析字段参数（如果有的话）
		fieldArguments := make(map[string]interface{})

		// 从查询字符串中提取字段参数
		if strings.Contains(execCtx.Query.Query, fieldName+"(") {
			fieldArgs := e.extractFieldArguments(execCtx.Query.Query, fieldName, execCtx.Variables)
			fieldArguments = fieldArgs
		} else {
			// 如果没有字段参数，传递全部变量（兼容性）
			fieldArguments = execCtx.Variables
		}

		fieldCtx := &types.FieldContext{
			ParentType: "Query",
			FieldName:  fieldName,
			Arguments:  fieldArguments,
			Alias:      fieldName,
			Path:       []string{fieldName},
		}

		e.logger.Debug("解析字段",
			zap.String("field", fieldName),
			zap.Any("arguments", fieldArguments))

		fieldResult, err := e.resolver.ResolveField(ctx, fieldCtx)
		if err != nil {
			e.logger.Error("字段解析失败",
				zap.String("field", fieldName),
				zap.Error(err))
			result.Errors = append(result.Errors, &types.GraphQLErrorWrapper{
				Kind:    types.ErrorKindExecution,
				Message: fmt.Sprintf("字段 %s 解析失败: %v", fieldName, err),
				Path:    []interface{}{fieldName},
			})
			continue
		}

		data[fieldName] = fieldResult
	}

	result.Data = data
}

// executeMutationQuery 执行Mutation查询
func (e *QueryExecutor) executeMutationQuery(ctx context.Context, execCtx *types.ExecutionContext, result *types.GraphQLResult) {
	e.logger.Debug("执行Mutation查询", zap.String("request_id", execCtx.RequestID))

	// 解析Mutation查询字符串，提取请求的字段
	mutationFields := e.parseSimpleQuery(execCtx.Query.Query)
	if len(mutationFields) == 0 {
		// 如果无法解析，返回默认结果
		result.Data = map[string]interface{}{
			"__typename": "MutationResult",
			"success":    true,
			"timestamp":  time.Now().Unix(),
		}
		return
	}

	// 执行每个Mutation字段的解析
	data := make(map[string]interface{})
	for _, fieldName := range mutationFields {
		fieldCtx := &types.FieldContext{
			ParentType: "Mutation",
			FieldName:  fieldName,
			Arguments:  execCtx.Variables,
			Alias:      fieldName,
			Path:       []string{fieldName},
		}

		fieldResult, err := e.resolver.ResolveField(ctx, fieldCtx)
		if err != nil {
			e.logger.Error("Mutation字段解析失败",
				zap.String("field", fieldName),
				zap.Error(err))
			result.Errors = append(result.Errors, &types.GraphQLErrorWrapper{
				Kind:    types.ErrorKindExecution,
				Message: fmt.Sprintf("Mutation字段 %s 解析失败: %v", fieldName, err),
				Path:    []interface{}{fieldName},
			})
			continue
		}

		data[fieldName] = fieldResult
	}

	result.Data = data
}

// executeSubscriptionQuery 执行Subscription查询
func (e *QueryExecutor) executeSubscriptionQuery(ctx context.Context, execCtx *types.ExecutionContext, result *types.GraphQLResult) {
	e.logger.Debug("执行Subscription查询", zap.String("request_id", execCtx.RequestID))

	// 这里简化处理，实际应该建立WebSocket连接
	// 目前返回一个示例结果
	result.Data = map[string]interface{}{
		"__typename": "SubscriptionResult",
		"connected":  true,
		"timestamp":  time.Now().Unix(),
	}
}

// validateQuery 验证查询
func (e *QueryExecutor) validateQuery(ctx context.Context, execCtx *types.ExecutionContext) error {
	for _, validator := range e.validators {
		if err := validator.Validate(ctx, execCtx); err != nil {
			return err
		}
	}
	return nil
}

// executeMiddlewareChain 执行中间件链
func (e *QueryExecutor) executeMiddlewareChain(ctx context.Context, execCtx *types.ExecutionContext, finalFunc func(ctx context.Context, execCtx *types.ExecutionContext) *types.GraphQLResult) *types.GraphQLResult {
	// 如果没有中间件，直接执行最终函数
	if len(e.middleware) == 0 {
		return finalFunc(ctx, execCtx)
	}

	// 创建中间件链
	var handler QueryHandler = func(ctx context.Context, execCtx *types.ExecutionContext) *types.GraphQLResult {
		return finalFunc(ctx, execCtx)
	}

	// 从后向前构建中间件链
	for i := len(e.middleware) - 1; i >= 0; i-- {
		currentHandler := handler
		middleware := e.middleware[i]
		handler = func(ctx context.Context, execCtx *types.ExecutionContext) *types.GraphQLResult {
			return middleware.Handle(ctx, execCtx, currentHandler)
		}
	}

	return handler(ctx, execCtx)
}

// wrapError 包装错误
func (e *QueryExecutor) wrapError(err error, kind types.ErrorKind) *types.GraphQLErrorWrapper {
	return &types.GraphQLErrorWrapper{
		Kind:     kind,
		Message:  err.Error(),
		Internal: err,
	}
}

// QueryValidator 查询验证器接口
type QueryValidator interface {
	Validate(ctx context.Context, execCtx *types.ExecutionContext) error
}

// QueryMiddleware 查询中间件接口
type QueryMiddleware interface {
	Handle(ctx context.Context, execCtx *types.ExecutionContext, next QueryHandler) *types.GraphQLResult
}

// QueryHandler 查询处理器函数类型
type QueryHandler func(ctx context.Context, execCtx *types.ExecutionContext) *types.GraphQLResult

// 基础验证器实现

// SchemaValidator Schema验证器
type SchemaValidator struct {
	logger *zap.Logger
}

// NewSchemaValidator 创建Schema验证器
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		logger: logger.Get().Named("schema-validator"),
	}
}

// Validate 验证Schema
func (v *SchemaValidator) Validate(ctx context.Context, execCtx *types.ExecutionContext) error {
	if execCtx.Schema == nil {
		return fmt.Errorf("schema不能为空")
	}

	if execCtx.Query == nil {
		return fmt.Errorf("query不能为空")
	}

	if execCtx.Operation == "" {
		return fmt.Errorf("operation不能为空")
	}

	v.logger.Debug("Schema验证通过", zap.String("request_id", execCtx.RequestID))
	return nil
}

// 安全验证器
type SecurityValidator struct {
	logger    *zap.Logger
	maxDepth  int
	maxTokens int
}

// NewSecurityValidator 创建安全验证器
func NewSecurityValidator() *SecurityValidator {
	return &SecurityValidator{
		logger:    logger.Get().Named("security-validator"),
		maxDepth:  10,  // 最大查询深度
		maxTokens: 100, // 最大token数
	}
}

// Validate 安全验证
func (v *SecurityValidator) Validate(ctx context.Context, execCtx *types.ExecutionContext) error {
	if execCtx.Query == nil {
		return fmt.Errorf("query不能为空")
	}

	// 简化的查询复杂度检查
	queryLength := len(execCtx.Query.Query)
	if queryLength > v.maxTokens*10 { // 简化的token估算
		return fmt.Errorf("查询过于复杂，超过最大限制")
	}

	v.logger.Debug("安全验证通过", zap.String("request_id", execCtx.RequestID))
	return nil
}

// 基础中间件实现

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger.Get().Named("logging-middleware"),
	}
}

// Handle 处理日志中间件
func (m *LoggingMiddleware) Handle(ctx context.Context, execCtx *types.ExecutionContext, next QueryHandler) *types.GraphQLResult {
	startTime := time.Now()

	m.logger.Info("GraphQL查询开始",
		zap.String("request_id", execCtx.RequestID),
		zap.String("operation", execCtx.Operation))

	result := next(ctx, execCtx)

	duration := time.Since(startTime)
	m.logger.Info("GraphQL查询完成",
		zap.String("request_id", execCtx.RequestID),
		zap.Duration("duration", duration),
		zap.Int("error_count", len(result.Errors)))

	return result
}

// MetricsMiddleware 指标收集中间件
type MetricsMiddleware struct {
	logger *zap.Logger
}

// NewMetricsMiddleware 创建指标收集中间件
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{
		logger: logger.Get().Named("metrics-middleware"),
	}
}

// Handle 处理指标收集中间件
func (m *MetricsMiddleware) Handle(ctx context.Context, execCtx *types.ExecutionContext, next QueryHandler) *types.GraphQLResult {
	startTime := time.Now()
	result := next(ctx, execCtx)
	duration := time.Since(startTime)

	// 简化的指标记录
	m.logger.Debug("记录查询指标",
		zap.String("operation", execCtx.Operation),
		zap.Duration("duration", duration),
		zap.Bool("has_errors", len(result.Errors) > 0))

	return result
}

// TimeoutMiddleware 超时中间件
type TimeoutMiddleware struct {
	timeout time.Duration
	logger  *zap.Logger
}

// NewTimeoutMiddleware 创建超时中间件
func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		timeout: timeout,
		logger:  logger.Get().Named("timeout-middleware"),
	}
}

// Handle 处理超时中间件
func (m *TimeoutMiddleware) Handle(ctx context.Context, execCtx *types.ExecutionContext, next QueryHandler) *types.GraphQLResult {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	resultChan := make(chan *types.GraphQLResult, 1)
	go func() {
		resultChan <- next(ctx, execCtx)
	}()

	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		return &types.GraphQLResult{
			Data: nil,
			Errors: []*types.GraphQLErrorWrapper{
				{
					Kind:    types.ErrorKindExecution,
					Message: "查询执行超时",
				},
			},
		}
	}
}

// parseSimpleQuery 简单的GraphQL查询解析器
// 这是一个基础实现，用于从查询字符串中提取字段名
func (e *QueryExecutor) parseSimpleQuery(query string) []string {
	if query == "" {
		return nil
	}

	// 移除注释和多余空白
	cleanQuery := strings.TrimSpace(query)

	// 处理不同类型的GraphQL查询格式
	var fields []string

	// 1. 处理带变量和参数的查询，如: query GetUser($id: ID!) { user(id: $id) { id name } }
	// 提取顶层字段名（在第一个 { 块中，但不在嵌套块中的字段）
	if strings.Contains(cleanQuery, "(") && strings.Contains(cleanQuery, "$") {
		// 对于带变量的查询，查找第一个 { 和 } 之间的顶层字段
		firstBrace := strings.Index(cleanQuery, "{")
		lastBrace := strings.LastIndex(cleanQuery, "}")

		if firstBrace != -1 && lastBrace != -1 && lastBrace > firstBrace {
			// 提取第一个查询块
			firstBlock := cleanQuery[firstBrace+1 : lastBrace]

			// 找到嵌套的开始位置（第一个内嵌的 {）
			nestedStart := strings.Index(firstBlock, "{")
			if nestedStart != -1 {
				// 只处理嵌套之前的部分（顶层字段）
				topLevelContent := strings.TrimSpace(firstBlock[:nestedStart])
				fields = e.extractTopLevelFields(topLevelContent)
			} else {
				// 没有嵌套，直接处理整个块
				fields = e.extractTopLevelFields(firstBlock)
			}
		}
	} else {
		// 2. 处理简单查询，如: { hello status user }
		queryBlockRegex := regexp.MustCompile(`\{([^{}]+)\}`)
		matches := queryBlockRegex.FindAllStringSubmatch(cleanQuery, -1)

		for _, match := range matches {
			if len(match) > 1 {
				blockContent := strings.TrimSpace(match[1])
				extractedFields := e.extractTopLevelFields(blockContent)
				fields = append(fields, extractedFields...)
			}
		}
	}

	// 去重
	uniqueFields := make([]string, 0)
	seen := make(map[string]bool)
	for _, field := range fields {
		if !seen[field] {
			seen[field] = true
			uniqueFields = append(uniqueFields, field)
		}
	}

	e.logger.Debug("解析GraphQL查询字段",
		zap.Strings("fields", uniqueFields),
		zap.String("original_query", cleanQuery))

	return uniqueFields
}

// extractTopLevelFields 从查询内容中提取顶层字段名
func (e *QueryExecutor) extractTopLevelFields(content string) []string {
	var fields []string

	// 使用更简单的正则表达式匹配字段名和参数
	// 首先匹配带参数的字段: fieldName(argument: value)
	fieldWithArgsRegex := regexp.MustCompile(`(\w+)\s*\([^)]*\)`)
	matches := fieldWithArgsRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 1 {
			field := strings.TrimSpace(match[1])
			// 过滤掉GraphQL关键字
			if field != "" &&
				!strings.EqualFold(field, "query") &&
				!strings.EqualFold(field, "mutation") &&
				!strings.EqualFold(field, "subscription") &&
				!strings.EqualFold(field, "fragment") {
				fields = append(fields, field)
			}
		}
	}

	// 然后匹配不带参数的字段: fieldName
	// 移除已经处理的带参数的字段
	contentWithoutArgs := fieldWithArgsRegex.ReplaceAllString(content, "")

	// 匹配简单的字段名
	simpleFieldRegex := regexp.MustCompile(`(\w+)`)
	simpleMatches := simpleFieldRegex.FindAllStringSubmatch(contentWithoutArgs, -1)

	for _, match := range simpleMatches {
		if len(match) > 1 {
			field := strings.TrimSpace(match[1])
			// 过滤掉GraphQL关键字
			if field != "" &&
				!strings.EqualFold(field, "query") &&
				!strings.EqualFold(field, "mutation") &&
				!strings.EqualFold(field, "subscription") &&
				!strings.EqualFold(field, "fragment") {
				fields = append(fields, field)
			}
		}
	}

	return fields
}

// extractFieldArguments 从查询字符串中提取字段参数
func (e *QueryExecutor) extractFieldArguments(query, fieldName string, variables map[string]interface{}) map[string]interface{} {
	args := make(map[string]interface{})

	// 构建字段参数的正则表达式
	// 匹配模式: fieldName(arg1: $var1, arg2: "value2")
	pattern := fmt.Sprintf(`%s\s*\(([^)]+)\)`, regexp.QuoteMeta(fieldName))
	argRegex := regexp.MustCompile(pattern)

	matches := argRegex.FindStringSubmatch(query)
	if len(matches) > 1 {
		argsString := strings.TrimSpace(matches[1])

		// 解析参数字符串 "id: $id, name: \"test\""
		argPairs := strings.Split(argsString, ",")
		for _, pair := range argPairs {
			pair = strings.TrimSpace(pair)
			if parts := strings.SplitN(pair, ":", 2); len(parts) == 2 {
				argName := strings.TrimSpace(parts[0])
				argValue := strings.TrimSpace(parts[1])

				// 处理变量引用 ($variable)
				if strings.HasPrefix(argValue, "$") {
					varName := strings.TrimPrefix(argValue, "$")
					if varValue, exists := variables[varName]; exists {
						args[argName] = varValue
					}
				} else {
					// 处理字面量值（移除引号）
					if strings.HasPrefix(argValue, "\"") && strings.HasSuffix(argValue, "\"") {
						args[argName] = strings.Trim(argValue, "\"")
					} else if argValue == "true" || argValue == "false" {
						args[argName] = argValue == "true"
					} else {
						// 尝试解析为数字或其他类型
						args[argName] = argValue
					}
				}
			}
		}
	}

	e.logger.Debug("提取字段参数",
		zap.String("field", fieldName),
		zap.Any("extracted_args", args),
		zap.Any("available_variables", variables))

	return args
}

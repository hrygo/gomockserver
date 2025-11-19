package parser

import (
	"fmt"
	"time"

	"github.com/gomockserver/mockserver/internal/graphql/types"
	"github.com/gomockserver/mockserver/pkg/logger"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"
)

// QueryParser GraphQL查询解析器
type QueryParser struct {
	schemaParser *SchemaParser
	logger       *zap.Logger
}

// NewQueryParser 创建查询解析器
func NewQueryParser() *QueryParser {
	return &QueryParser{
		schemaParser: NewSchemaParser(),
		logger:       logger.Get().Named("graphql-query-parser"),
	}
}

// ParseQuery 解析GraphQL查询
func (p *QueryParser) ParseQuery(query string, schema *types.SchemaDocument) (*types.GraphQLQuery, error) {
	p.logger.Info("开始解析GraphQL查询",
		zap.String("query_length", fmt.Sprintf("%d", len(query))),
		zap.String("request_id", types.GenerateID()))

	// 转换内部schema为gqlparser schema
	gqlSchema, err := p.convertToGqlSchema(schema)
	if err != nil {
		p.logger.Error("转换Schema失败", zap.Error(err))
		return nil, fmt.Errorf("failed to convert schema: %w", err)
	}

	// 解析查询
	queryDoc, err := gqlparser.LoadQuery(gqlSchema, query)
	if err != nil {
		p.logger.Error("解析查询失败", zap.Error(err))
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// 转换为内部查询对象
	internalQuery := p.convertToInternalQuery(queryDoc)

	p.logger.Info("查询解析成功",
		zap.String("query_id", internalQuery.ID),
		zap.String("operation", internalQuery.Operation))

	return internalQuery, nil
}

// ValidateQuery 验证GraphQL查询
func (p *QueryParser) ValidateQuery(query *types.GraphQLQuery, schema *types.SchemaDocument) error {
	p.logger.Info("开始验证GraphQL查询", zap.String("query_id", query.ID))

	var errors []string

	// 基础验证
	if query.Query == "" {
		errors = append(errors, "查询不能为空")
	}

	if query.Operation == "" {
		errors = append(errors, "操作类型不能为空")
	}

	// 检查操作类型
	if !types.IsOperationType(query.Operation) {
		errors = append(errors, fmt.Sprintf("无效的操作类型: %s", query.Operation))
	}

	// 验证变量
	if query.Variables != nil {
		for varName, varValue := range query.Variables {
			if varName == "" {
				errors = append(errors, "变量名不能为空")
			}
			if varValue == nil {
				p.logger.Warn("变量值为空", zap.String("variable", varName))
			}
		}
	}

	if len(errors) > 0 {
		errMsg := "查询验证失败:\n" + fmt.Sprintf("%v", errors)
		p.logger.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	p.logger.Info("查询验证通过", zap.String("query_id", query.ID))
	return nil
}

// convertToGqlSchema 将内部schema转换为gqlparser schema
func (p *QueryParser) convertToGqlSchema(schema *types.SchemaDocument) (*ast.Schema, error) {
	// 创建基础schema
	gqlSchema := &ast.Schema{
		Types: make(map[string]*ast.Definition),
	}

	// 添加内置类型
	builtins := []string{
		"ID", "String", "Int", "Float", "Boolean",
		"Query", "Mutation", "Subscription",
	}
	for _, builtin := range builtins {
		gqlSchema.Types[builtin] = &ast.Definition{
			Kind: ast.Scalar,
			Name: builtin,
			BuiltIn: true,
		}
	}

	// 转换类型定义 - 简化实现
	for range schema.Definitions {
		// 简化实现 - 直接处理而不进行复杂的类型断言
		// 这部分功能将在后续版本中完善
		// 目前先创建一个基础的schema结构
	}

	return gqlSchema, nil
}

// GenerateID 生成唯一ID
func (p *QueryParser) GenerateID() string {
	return fmt.Sprintf("graphql_%d", time.Now().UnixNano())
}

// convertToInternalQuery 将gqlparser query转换为内部查询
func (p *QueryParser) convertToInternalQuery(queryDoc *ast.QueryDocument) *types.GraphQLQuery {
	query := &types.GraphQLQuery{
		ID:        p.GenerateID(),
		Query:     "", // Will be populated from the original query string
		Variables: make(map[string]interface{}),
		Operation: "",
		Timestamp: time.Now(),
	}

	// 转换操作
	for _, operation := range queryDoc.Operations {
		query.Operation = string(operation.Operation)
		break // 目前只支持单个操作
	}

	// 转换变量
	if queryDoc.Operations != nil && len(queryDoc.Operations) > 0 {
		op := queryDoc.Operations[0]
		for _, varDef := range op.VariableDefinitions {
			// 这里简化处理，实际应该转换变量类型
			query.Variables[varDef.Variable] = nil
		}
	}

	return query
}

// convertToGqlFields 转换字段定义 - 简化实现
func (p *QueryParser) convertToGqlFields(fields []types.FieldDefinition) []*ast.FieldDefinition {
	// 简化实现，将在Phase 2中完善
	return make([]*ast.FieldDefinition, 0)
}

// convertToGqlInputValues 转换输入值定义 - 简化实现
func (p *QueryParser) convertToGqlInputValues(values []types.InputValueDefinition) []*ast.ArgumentDefinition {
	// 简化实现，将在Phase 2中完善
	return make([]*ast.ArgumentDefinition, 0)
}

// convertToGqlArguments 转换参数定义 - 简化实现
func (p *QueryParser) convertToGqlArguments(args []types.InputValueDefinition) []*ast.ArgumentDefinition {
	return p.convertToGqlInputValues(args)
}

// convertToGqlDirectives 转换指令 - 简化实现
func (p *QueryParser) convertToGqlDirectives(directives []types.Directive) []*ast.Directive {
	// 简化实现，将在Phase 2中完善
	return make([]*ast.Directive, 0)
}

// convertToGqlArgumentValues 转换参数值 - 简化实现
func (p *QueryParser) convertToGqlArgumentValues(args []types.ArgumentValue) []*ast.Argument {
	// 简化实现，将在Phase 2中完善
	return make([]*ast.Argument, 0)
}

// convertToGqlValue 转换值 - 简化实现
func (p *QueryParser) convertToGqlValue(value interface{}) *ast.Value {
	if value == nil {
		return &ast.Value{Kind: ast.NullValue}
	}

	// 简化实现，将在Phase 2中完善
	switch v := value.(type) {
	case string:
		return &ast.Value{
			Kind: ast.StringValue,
			Raw:  v,
		}
	case int, int32, int64:
		return &ast.Value{
			Kind: ast.IntValue,
			Raw:  fmt.Sprintf("%d", v),
		}
	case float32, float64:
		return &ast.Value{
			Kind: ast.FloatValue,
			Raw:  fmt.Sprintf("%f", v),
		}
	case bool:
		return &ast.Value{
			Kind: ast.BooleanValue,
			Raw:  fmt.Sprintf("%t", v),
		}
	default:
		return &ast.Value{
			Kind: ast.StringValue,
			Raw:  fmt.Sprintf("%v", v),
		}
	}
}

// convertToGqlType 转换类型
func (p *QueryParser) convertToGqlType(typ types.Type) *ast.Type {
	switch t := typ.(type) {
	case *types.NamedType:
		return &ast.Type{
			NamedType: t.Name,
		}
	case *types.ListType:
		return &ast.Type{
			Elem: p.convertToGqlType(t.Type),
		}
	case *types.NonNullType:
		elem := p.convertToGqlType(t.Type)
		if elem != nil {
			elem.NonNull = true
		}
		return elem
	default:
		return nil
	}
}

// convertToGqlEnumValues 转换枚举值 - 简化实现
func (p *QueryParser) convertToGqlEnumValues(values []types.EnumValueDefinition) []*ast.EnumValueDefinition {
	// 简化实现，将在Phase 2中完善
	return make([]*ast.EnumValueDefinition, 0)
}
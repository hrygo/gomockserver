package parser

import (
	"fmt"

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
	queryDoc, err := gqlparser.LoadQuery(gqlSchema, &ast.Source{
		Name:    "query.graphql",
		Input:   query,
		BuiltIn: false,
	})
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

	// 转换类型定义
	for _, def := range schema.Definitions {
		switch t := def.(type) {
		case *types.ObjectTypeDefinition:
			gqlDef := &ast.Definition{
				Kind:        ast.Object,
				Name:        t.Name,
				Description: t.Description,
				Fields:      p.convertToGqlFields(t.Fields),
				Interfaces:  t.Implements,
				Directives:  p.convertToGqlDirectives(t.Directives),
				Position:    ast.Position{Line: t.Position.Line, Column: t.Position.Column},
			}
			gqlSchema.Types[t.Name] = gqlDef

		case *types.InterfaceTypeDefinition:
			gqlDef := &ast.Definition{
				Kind:        ast.Interface,
				Name:        t.Name,
				Description: t.Description,
				Fields:      p.convertToGqlFields(t.Fields),
				Directives:  p.convertToGqlDirectives(t.Directives),
				Position:    ast.Position{Line: t.Position.Line, Column: t.Position.Column},
			}
			gqlSchema.Types[t.Name] = gqlDef

		case *types.UnionTypeDefinition:
			gqlDef := &ast.Definition{
				Kind:        ast.Union,
				Name:        t.Name,
				Description: t.Description,
				Types:       t.Types,
				Directives:  p.convertToGqlDirectives(t.Directives),
				Position:    ast.Position{Line: t.Position.Line, Column: t.Position.Column},
			}
			gqlSchema.Types[t.Name] = gqlDef

		case *types.ScalarTypeDefinition:
			gqlDef := &ast.Definition{
				Kind:        ast.Scalar,
				Name:        t.Name,
				Description: t.Description,
				Directives:  p.convertToGqlDirectives(t.Directives),
				Position:    ast.Position{Line: t.Position.Line, Column: t.Position.Column},
			}
			gqlSchema.Types[t.Name] = gqlDef

		case *types.EnumTypeDefinition:
			gqlDef := &ast.Definition{
				Kind:        ast.Enum,
				Name:        t.Name,
				Description: t.Description,
				EnumValues:  p.convertToGqlEnumValues(t.Values),
				Directives:  p.convertToGqlDirectives(t.Directives),
				Position:    ast.Position{Line: t.Position.Line, Column: t.Position.Column},
			}
			gqlSchema.Types[t.Name] = gqlDef

		case *types.InputObjectTypeDefinition:
			gqlDef := &ast.Definition{
				Kind:        ast.InputObject,
				Name:        t.Name,
				Description: t.Description,
				Fields:      p.convertToGqlInputValues(t.Fields),
				Directives:  p.convertToGqlDirectives(t.Directives),
				Position:    ast.Position{Line: t.Position.Line, Column: t.Position.Column},
			}
			gqlSchema.Types[t.Name] = gqlDef
		}
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
		Variables: make(map[string]interface{}),
		Timestamp: queryDoc.Position.Loc.StartTime,
	}

	// 转换操作
	for _, operation := range queryDoc.Operations {
		query.Operation = operation.Operation
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

// convertToGqlFields 转换字段定义
func (p *QueryParser) convertToGqlFields(fields []types.FieldDefinition) ast.FieldDefinitionList {
	result := make(ast.FieldDefinitionList, 0)
	for _, field := range fields {
		gqlField := &ast.FieldDefinition{
			Name:        field.Name,
			Description: field.Description,
			Type:        p.convertToGqlType(field.Type),
			Arguments:   p.convertToGqlArguments(field.Arguments),
			Directives:  p.convertToGqlDirectives(field.Directives),
			Position:    ast.Position{Line: field.Position.Line, Column: field.Position.Column},
		}
		result = append(result, gqlField)
	}
	return result
}

// convertToGqlInputValues 转换输入值定义
func (p *QueryParser) convertToGqlInputValues(values []types.InputValueDefinition) ast.ArgumentDefinitionList {
	result := make(ast.ArgumentDefinitionList, 0)
	for _, value := range values {
		gqlValue := &ast.ArgumentDefinition{
			Name:         value.Name,
			Description:  value.Description,
			Type:         p.convertToGqlType(value.Type),
			DefaultValue: p.convertToGqlValue(value.DefaultValue),
			Directives:   p.convertToGqlDirectives(value.Directives),
			Position:     ast.Position{Line: value.Position.Line, Column: value.Position.Column},
		}
		result = append(result, gqlValue)
	}
	return result
}

// convertToGqlArguments 转换参数定义
func (p *QueryParser) convertToGqlArguments(args []types.InputValueDefinition) ast.ArgumentDefinitionList {
	return p.convertToGqlInputValues(args)
}

// convertToGqlDirectives 转换指令
func (p *QueryParser) convertToGqlDirectives(directives []types.Directive) ast.DirectiveList {
	result := make(ast.DirectiveList, 0)
	for _, dir := range directives {
		gqlDir := &ast.Directive{
			Name:       dir.Name,
			Arguments:  p.convertToGqlArgumentValues(dir.Arguments),
			Position:   ast.Position{Line: dir.Position.Line, Column: dir.Position.Column},
		}
		result = append(result, gqlDir)
	}
	return result
}

// convertToGqlArgumentValues 转换参数值
func (p *QueryParser) convertToGqlArgumentValues(args []types.ArgumentValue) ast.ArgumentList {
	result := make(ast.ArgumentList, 0)
	for _, arg := range args {
		gqlArg := &ast.Argument{
			Name:  arg.Name,
			Value: p.convertToGqlValue(arg.Value),
		}
		result = append(result, gqlArg)
	}
	return result
}

// convertToGqlValue 转换值
func (p *QueryParser) convertToGqlValue(value interface{}) *ast.Value {
	if value == nil {
		return &ast.Value{Kind: ast.NullValue}
	}

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
	case []interface{}:
		children := make(ast.ValueList, 0)
		for _, item := range v {
			children = append(children, p.convertToGqlValue(item))
		}
		return &ast.Value{
			Kind:     ast.ListValue,
			Children: children,
		}
	case map[string]interface{}:
		children := make(ast.ValueList, 0)
		for key, val := range v {
			childValue := p.convertToGqlValue(val)
			childValue.Name = key
			children = append(children, childValue)
		}
		return &ast.Value{
			Kind:     ast.ObjectValue,
			Children: children,
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

// convertToGqlEnumValues 转换枚举值
func (p *QueryParser) convertToGqlEnumValues(values []types.EnumValueDefinition) ast.EnumValueDefinitionList {
	result := make(ast.EnumValueDefinitionList, 0)
	for _, value := range values {
		gqlValue := &ast.EnumValueDefinition{
			Name:        value.Name,
			Description: value.Description,
			Directives:  p.convertToGqlDirectives(value.Directives),
			Position:    ast.Position{Line: value.Position.Line, Column: value.Position.Column},
		}
		result = append(result, gqlValue)
	}
	return result
}
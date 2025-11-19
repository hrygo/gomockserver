package parser

import (
	"fmt"
	"strings"

	"github.com/gomockserver/mockserver/internal/graphql/types"
	"github.com/gomockserver/mockserver/pkg/logger"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"go.uber.org/zap"
)

// SchemaParser GraphQL Schema解析器
type SchemaParser struct {
	logger *zap.Logger
}

// NewSchemaParser 创建Schema解析器
func NewSchemaParser() *SchemaParser {
	return &SchemaParser{
		logger: logger.Get().Named("graphql-parser"),
	}
}

// ParseSchema 解析GraphQL Schema
func (p *SchemaParser) ParseSchema(sdl string) (*types.SchemaDocument, error) {
	p.logger.Info("开始解析GraphQL Schema", zap.String("sdl_length", fmt.Sprintf("%d", len(sdl))))

	// 使用gqlparser解析SDL
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:    "schema.graphql",
		Input:   sdl,
		BuiltIn: false,
	})
	if err != nil {
		p.logger.Error("解析Schema失败", zap.Error(err))
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	p.logger.Info("Schema解析成功",
		zap.Int("types_count", len(schema.Types)),
		zap.Strings("operation_types", []string{schema.Query.Name, schema.Mutation.Name, schema.Subscription.Name}))

	// 转换为内部类型
	doc := p.convertToInternalSchema(schema)

	p.logger.Info("Schema转换完成")
	return doc, nil
}

// ValidateSchema 验证GraphQL Schema
func (p *SchemaParser) ValidateSchema(doc *types.SchemaDocument) error {
	p.logger.Info("开始验证GraphQL Schema")

	var errors []string

	// 检查类型定义
	typeNames := make(map[string]bool)
	for _, def := range doc.Definitions {
		switch t := def.(type) {
		case *types.ObjectTypeDefinition:
			if typeNames[t.Name] {
				errors = append(errors, fmt.Sprintf("重复的类型定义: %s", t.Name))
			}
			typeNames[t.Name] = true

			// 验证字段
			fieldNames := make(map[string]bool)
			for _, field := range t.Fields {
				if fieldNames[field.Name] {
					errors = append(errors, fmt.Sprintf("对象类型 %s 中重复的字段定义: %s", t.Name, field.Name))
				}
				fieldNames[field.Name] = true

				// 验证字段类型
				if err := p.validateType(field.Type); err != nil {
					errors = append(errors, fmt.Sprintf("对象类型 %s 字段 %s 类型错误: %v", t.Name, field.Name, err))
				}
			}

		case *types.InterfaceTypeDefinition:
			if typeNames[t.Name] {
				errors = append(errors, fmt.Sprintf("重复的类型定义: %s", t.Name))
			}
			typeNames[t.Name] = true

		case *types.UnionTypeDefinition:
			if typeNames[t.Name] {
				errors = append(errors, fmt.Sprintf("重复的类型定义: %s", t.Name))
			}
			typeNames[t.Name] = true

		case *types.ScalarTypeDefinition:
			if typeNames[t.Name] {
				errors = append(errors, fmt.Sprintf("重复的类型定义: %s", t.Name))
			}
			typeNames[t.Name] = true

		case *types.EnumTypeDefinition:
			if typeNames[t.Name] {
				errors = append(errors, fmt.Sprintf("重复的类型定义: %s", t.Name))
			}
			typeNames[t.Name] = true

		case *types.InputObjectTypeDefinition:
			if typeNames[t.Name] {
				errors = append(errors, fmt.Sprintf("重复的类型定义: %s", t.Name))
			}
			typeNames[t.Name] = true
		}
	}

	if len(errors) > 0 {
		errMsg := "Schema验证失败:\n" + strings.Join(errors, "\n")
		p.logger.Error(errMsg)
		return fmt.Errorf(errMsg)
	}

	p.logger.Info("Schema验证通过")
	return nil
}

// convertToInternalSchema 将gqlparser schema转换为内部schema
func (p *SchemaParser) convertToInternalSchema(schema *ast.Schema) *types.SchemaDocument {
	doc := &types.SchemaDocument{
		Definitions: make([]types.Definition, 0),
	}

	// 转换类型定义
	for name, typ := range schema.Types {
		if typ.IsBuiltin() {
			continue
		}

		switch typ.Kind {
		case ast.Object:
			objDef := p.convertObjectType(typ)
			doc.Definitions = append(doc.Definitions, objDef)

		case ast.Interface:
			ifaceDef := p.convertInterfaceType(typ)
			doc.Definitions = append(doc.Definitions, ifaceDef)

		case ast.Union:
			unionDef := p.convertUnionType(typ)
			doc.Definitions = append(doc.Definitions, unionDef)

		case ast.Scalar:
			scalarDef := p.convertScalarType(typ)
			doc.Definitions = append(doc.Definitions, scalarDef)

		case ast.Enum:
			enumDef := p.convertEnumType(typ)
			doc.Definitions = append(doc.Definitions, enumDef)

		case ast.InputObject:
			inputDef := p.convertInputObjectType(typ)
			doc.Definitions = append(doc.Definitions, inputDef)
		}
	}

	// 转换操作类型定义
	if schema.Query != nil && !schema.Query.IsBuiltin() {
		queryDef := p.convertObjectType(schema.Query)
		if schemaDef := p.convertToSchemaDefinition(queryDef, types.Query); schemaDef != nil {
			doc.Definitions = append(doc.Definitions, schemaDef)
		}
	}

	if schema.Mutation != nil && !schema.Mutation.IsBuiltin() {
		mutationDef := p.convertObjectType(schema.Mutation)
		if schemaDef := p.convertToSchemaDefinition(mutationDef, types.Mutation); schemaDef != nil {
			doc.Definitions = append(doc.Definitions, schemaDef)
		}
	}

	if schema.Subscription != nil && !schema.Subscription.IsBuiltin() {
		subscriptionDef := p.convertObjectType(schema.Subscription)
		if schemaDef := p.convertToSchemaDefinition(subscriptionDef, types.Subscription); schemaDef != nil {
			doc.Definitions = append(doc.Definitions, schemaDef)
		}
	}

	return doc
}

// convertObjectType 转换对象类型
func (p *SchemaParser) convertObjectType(typ *ast.Definition) *types.ObjectTypeDefinition {
	fields := make([]types.FieldDefinition, 0)
	for _, field := range typ.Fields {
		fieldDef := types.FieldDefinition{
			Description: p.formatDescription(field.Description),
			Name:        field.Name,
			Arguments:   p.convertArguments(field.Arguments),
			Type:        p.convertType(field.Type),
			Directives:  p.convertDirectives(field.Directives),
			Position: types.Position{
				Line:   field.Position.Line,
				Column: field.Position.Column,
			},
		}
		fields = append(fields, fieldDef)
	}

	return &types.ObjectTypeDefinition{
		Description: p.formatDescription(typ.Description),
		Name:        typ.Name,
		Implements:  p.getImplements(typ),
		Directives:  p.convertDirectives(typ.Directives),
		Fields:      fields,
		Position: types.Position{
			Line:   typ.Position.Line,
			Column: typ.Position.Column,
		},
	}
}

// convertInterfaceType 转换接口类型
func (p *SchemaParser) convertInterfaceType(typ *ast.Definition) *types.InterfaceTypeDefinition {
	fields := make([]types.FieldDefinition, 0)
	for _, field := range typ.Fields {
		fieldDef := types.FieldDefinition{
			Description: p.formatDescription(field.Description),
			Name:        field.Name,
			Arguments:   p.convertArguments(field.Arguments),
			Type:        p.convertType(field.Type),
			Directives:  p.convertDirectives(field.Directives),
			Position: types.Position{
				Line:   field.Position.Line,
				Column: field.Position.Column,
			},
		}
		fields = append(fields, fieldDef)
	}

	return &types.InterfaceTypeDefinition{
		Description: p.formatDescription(typ.Description),
		Name:        typ.Name,
		Directives:  p.convertDirectives(typ.Directives),
		Fields:      fields,
		Position: types.Position{
			Line:   typ.Position.Line,
			Column: typ.Position.Column,
		},
	}
}

// convertUnionType 转换联合类型
func (p *SchemaParser) convertUnionType(typ *ast.Definition) *types.UnionTypeDefinition {
	unionTypes := make([]string, 0)
	for _, unionType := range typ.Types {
		unionTypes = append(unionTypes, unionType)
	}

	return &types.UnionTypeDefinition{
		Description: p.formatDescription(typ.Description),
		Name:        typ.Name,
		Directives:  p.convertDirectives(typ.Directives),
		Types:       unionTypes,
		Position: types.Position{
			Line:   typ.Position.Line,
			Column: typ.Position.Column,
		},
	}
}

// convertScalarType 转换标量类型
func (p *SchemaParser) convertScalarType(typ *ast.Definition) *types.ScalarTypeDefinition {
	return &types.ScalarTypeDefinition{
		Description: p.formatDescription(typ.Description),
		Name:        typ.Name,
		Directives:  p.convertDirectives(typ.Directives),
		Position: types.Position{
			Line:   typ.Position.Line,
			Column: typ.Position.Column,
		},
	}
}

// convertEnumType 转换枚举类型
func (p *SchemaParser) convertEnumType(typ *ast.Definition) *types.EnumTypeDefinition {
	values := make([]types.EnumValueDefinition, 0)
	for _, value := range typ.EnumValues {
		valueDef := types.EnumValueDefinition{
			Description: p.formatDescription(value.Description),
			Name:        value.Name,
			Directives:  p.convertDirectives(value.Directives),
			Position: types.Position{
				Line:   value.Position.Line,
				Column: value.Position.Column,
			},
		}
		values = append(values, valueDef)
	}

	return &types.EnumTypeDefinition{
		Description: p.formatDescription(typ.Description),
		Name:        typ.Name,
		Directives:  p.convertDirectives(typ.Directives),
		Values:      values,
		Position: types.Position{
			Line:   typ.Position.Line,
			Column: typ.Position.Column,
		},
	}
}

// convertInputObjectType 转换输入对象类型
func (p *SchemaParser) convertInputObjectType(typ *ast.Definition) *types.InputObjectTypeDefinition {
	fields := make([]types.InputValueDefinition, 0)
	for _, field := range typ.Fields {
		fieldDef := types.InputValueDefinition{
			Description:  p.formatDescription(field.Description),
			Name:         field.Name,
			Type:         p.convertType(field.Type),
			DefaultValue: p.convertValue(field.DefaultValue),
			Directives:   p.convertDirectives(field.Directives),
			Position: types.Position{
				Line:   field.Position.Line,
				Column: field.Position.Column,
			},
		}
		fields = append(fields, fieldDef)
	}

	return &types.InputObjectTypeDefinition{
		Description: p.formatDescription(typ.Description),
		Name:        typ.Name,
		Directives:  p.convertDirectives(typ.Directives),
		Fields:      fields,
		Position: types.Position{
			Line:   typ.Position.Line,
			Column: typ.Position.Column,
		},
	}
}

// convertType 转换类型引用
func (p *SchemaParser) convertType(typ *ast.Type) types.Type {
	if typ.NonNull {
		return &types.NonNullType{
			Type: p.convertType(typ.Elem),
		}
	}

	if typ.Elem != nil {
		return &types.ListType{
			Type: p.convertType(typ.Elem),
		}
	}

	return &types.NamedType{
		Name: typ.NamedType,
	}
}

// convertArguments 转换参数定义
func (p *SchemaParser) convertArguments(args ast.ArgumentDefinitionList) []types.InputValueDefinition {
	result := make([]types.InputValueDefinition, 0)
	for _, arg := range args {
		argDef := types.InputValueDefinition{
			Description:  p.formatDescription(arg.Description),
			Name:         arg.Name,
			Type:         p.convertType(arg.Type),
			DefaultValue: p.convertValue(arg.DefaultValue),
			Directives:   p.convertDirectives(arg.Directives),
			Position: types.Position{
				Line:   arg.Position.Line,
				Column: arg.Position.Column,
			},
		}
		result = append(result, argDef)
	}
	return result
}

// convertDirectives 转换指令
func (p *SchemaParser) convertDirectives(directives ast.DirectiveList) []types.Directive {
	result := make([]types.Directive, 0)
	for _, dir := range directives {
		dirDef := types.Directive{
			Name: dir.Name,
			Arguments: p.convertArgumentValues(dir.Arguments),
			Position: types.Position{
				Line:   dir.Position.Line,
				Column: dir.Position.Column,
			},
		}
		result = append(result, dirDef)
	}
	return result
}

// convertArgumentValues 转换参数值
func (p *SchemaParser) convertArgumentValues(args ast.ArgumentList) []types.ArgumentValue {
	result := make([]types.ArgumentValue, 0)
	for _, arg := range args {
		argValue := types.ArgumentValue{
			Name:  arg.Name,
			Value: p.convertValue(arg.Value),
		}
		result = append(result, argValue)
	}
	return result
}

// convertValue 转换值
func (p *SchemaParser) convertValue(val *ast.Value) interface{} {
	if val == nil {
		return nil
	}

	switch val.Kind {
	case ast.StringValue:
		return val.Raw
	case ast.IntValue:
		return val.Raw
	case ast.FloatValue:
		return val.Raw
	case ast.BooleanValue:
		return val.Raw
	case ast.NullValue:
		return nil
	case ast.EnumValue:
		return val.Raw
	case ast.ListValue:
		values := make([]interface{}, 0)
		for _, item := range val.Children {
			values = append(values, p.convertValue(item))
		}
		return values
	case ast.ObjectValue:
		obj := make(map[string]interface{})
		for _, field := range val.Children {
			obj[field.Name] = p.convertValue(field.Value)
		}
		return obj
	default:
		return val.Raw
	}
}

// getImplements 获取实现的接口
func (p *SchemaParser) getImplements(typ *ast.Definition) []string {
	if typ.Interfaces == nil {
		return nil
	}

	interfaces := make([]string, 0)
	for _, iface := range typ.Interfaces {
		interfaces = append(interfaces, iface)
	}
	return interfaces
}

// convertToSchemaDefinition 转换为Schema定义
func (p *SchemaParser) convertToSchemaDefinition(objDef *types.ObjectTypeDefinition, opType types.OperationType) *types.SchemaDefinition {
	if objDef == nil {
		return nil
	}

	return &types.SchemaDefinition{
		OperationTypes: []types.OperationTypeDefinition{
			{
				Operation: opType,
				Type:      objDef.Name,
			},
		},
		Position: objDef.Position,
	}
}

// validateType 验证类型
func (p *SchemaParser) validateType(typ types.Type) error {
	switch t := typ.(type) {
	case *types.NamedType:
		if t.Name == "" {
			return fmt.Errorf("named type cannot be empty")
		}
	case *types.ListType:
		return p.validateType(t.Type)
	case *types.NonNullType:
		return p.validateType(t.Type)
	default:
		return fmt.Errorf("unknown type kind: %T", typ)
	}
	return nil
}

// formatDescription 格式化描述
func (p *SchemaParser) formatDescription(desc string) string {
	if desc == "" {
		return ""
	}

	// 移除前后引号和空格
	desc = strings.TrimSpace(desc)
	if strings.HasPrefix(desc, `"""`) && strings.HasSuffix(desc, `"""`) {
		desc = desc[3 : len(desc)-3]
	} else if strings.HasPrefix(desc, `"`) && strings.HasSuffix(desc, `"`) {
		desc = desc[1 : len(desc)-1]
	}

	return strings.TrimSpace(desc)
}
package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaParser_ParseSchema(t *testing.T) {
	parser := NewSchemaParser()

	// 测试基础Schema解析
	sdl := `
		type Query {
			hello: String!
			user(id: ID!): User
		}

		type User {
			id: ID!
			name: String!
			email: String
		}

		type Mutation {
			createUser(input: CreateUserInput!): User!
		}

		type Subscription {
			userUpdated(id: ID!): User
		}
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)
	assert.NotNil(t, doc)

	// 验证类型定义
	assert.GreaterOrEqual(t, len(doc.Definitions), 4)

	// 验证Schema验证
	err = parser.ValidateSchema(doc)
	require.NoError(t, err)
}

func TestSchemaParser_ParseObjectTypes(t *testing.T) {
	parser := NewSchemaParser()

	sdl := `
		type User {
			"""用户ID"""
			id: ID!
			"""用户名"""
			name: String!
			email: String
			age: Int
		}
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)

	// 查找User类型
	var userType *types.ObjectTypeDefinition
	for _, def := range doc.Definitions {
		if objDef, ok := def.(*types.ObjectTypeDefinition); ok && objDef.Name == "User" {
			userType = objDef
			break
		}
	}

	require.NotNil(t, userType)
	assert.Equal(t, "User", userType.Name)
	assert.Equal(t, 4, len(userType.Fields))

	// 验证字段
	fields := make(map[string]types.FieldDefinition)
	for _, field := range userType.Fields {
		fields[field.Name] = field
	}

	// 验证id字段
	idField, exists := fields["id"]
	assert.True(t, exists)
	assert.Equal(t, "ID!", idField.Type.String())

	// 验证name字段
	nameField, exists := fields["name"]
	assert.True(t, exists)
	assert.Equal(t, "String!", nameField.Type.String())
}

func TestSchemaParser_ParseInvalidSchema(t *testing.T) {
	parser := NewSchemaParser()

	// 测试无效的Schema
	sdl := `
		type User {
			id: ID!
			# 无效的重复字段
			id: String!
		}
	`

	_, err := parser.ParseSchema(sdl)
	assert.Error(t, err)
}

func TestSchemaParser_ValidateDuplicateTypes(t *testing.T) {
	parser := NewSchemaParser()

	// 创建包含重复类型的Schema
	doc := &types.SchemaDocument{
		Definitions: []types.Definition{
			&types.ObjectTypeDefinition{
				Name: "User",
				Fields: []types.FieldDefinition{
					{Name: "id", Type: &types.NamedType{Name: "ID!"}},
				},
				Position: types.Position{Line: 1, Column: 1},
			},
			&types.ObjectTypeDefinition{
				Name: "User", // 重复类型
				Fields: []types.FieldDefinition{
					{Name: "name", Type: &types.NamedType{Name: "String!"}},
				},
				Position: types.Position{Line: 5, Column: 1},
			},
		},
	}

	err := parser.ValidateSchema(doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "重复的类型定义")
}

func TestSchemaParser_ValidateDuplicateFields(t *testing.T) {
	parser := NewSchemaParser()

	// 创建包含重复字段的Schema
	doc := &types.SchemaDocument{
		Definitions: []types.Definition{
			&types.ObjectTypeDefinition{
				Name: "User",
				Fields: []types.FieldDefinition{
					{Name: "id", Type: &types.NamedType{Name: "ID!"}},
					{Name: "id", Type: &types.NamedType{Name: "String!"}}, // 重复字段
				},
				Position: types.Position{Line: 1, Column: 1},
			},
		},
	}

	err := parser.ValidateSchema(doc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "重复的字段定义")
}

func TestSchemaParser_ParseInterface(t *testing.T) {
	parser := NewSchemaParser()

	sdl := `
		interface Node {
			id: ID!
		}

		type User implements Node {
			id: ID!
			name: String!
		}
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)

	// 验证接口定义
	var interfaceDef *types.InterfaceTypeDefinition
	for _, def := range doc.Definitions {
		if ifaceDef, ok := def.(*types.InterfaceTypeDefinition); ok && ifaceDef.Name == "Node" {
			interfaceDef = ifaceDef
			break
		}
	}

	require.NotNil(t, interfaceDef)
	assert.Equal(t, "Node", interfaceDef.Name)
	assert.Equal(t, 1, len(interfaceDef.Fields))
}

func TestSchemaParser_ParseUnion(t *testing.T) {
	parser := NewSchemaParser()

	sdl := `
		union SearchResult = User | Post
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)

	// 验证联合类型定义
	var unionDef *types.UnionTypeDefinition
	for _, def := range doc.Definitions {
		if union, ok := def.(*types.UnionTypeDefinition); ok && union.Name == "SearchResult" {
			unionDef = union
			break
		}
	}

	require.NotNil(t, unionDef)
	assert.Equal(t, "SearchResult", unionDef.Name)
	assert.Equal(t, []string{"User", "Post"}, unionDef.Types)
}

func TestSchemaParser_ParseEnum(t *testing.T) {
	parser := NewSchemaParser()

	sdl := `
		enum Status {
			ACTIVE
			INACTIVE
			PENDING
		}
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)

	// 验证枚举类型定义
	var enumDef *types.EnumTypeDefinition
	for _, def := range doc.Definitions {
		if enum, ok := def.(*types.EnumTypeDefinition); ok && enum.Name == "Status" {
			enumDef = enum
			break
		}
	}

	require.NotNil(t, enumDef)
	assert.Equal(t, "Status", enumDef.Name)
	assert.Equal(t, 3, len(enumDef.Values))
}

func TestSchemaParser_ParseInputObject(t *testing.T) {
	parser := NewSchemaParser()

	sdl := `
		input CreateUserInput {
			name: String!
			email: String
			age: Int
		}
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)

	// 验证输入对象类型定义
	var inputDef *types.InputObjectTypeDefinition
	for _, def := range doc.Definitions {
		if input, ok := def.(*types.InputObjectTypeDefinition); ok && input.Name == "CreateUserInput" {
			inputDef = input
			break
		}
	}

	require.NotNil(t, inputDef)
	assert.Equal(t, "CreateUserInput", inputDef.Name)
	assert.Equal(t, 3, len(inputDef.Fields))
}

func TestSchemaParser_ParseScalar(t *testing.T) {
	parser := NewSchemaParser()

	sdl := `
		scalar DateTime
		scalar JSON
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)

	// 验证标量类型定义
	scalarTypes := make(map[string]*types.ScalarTypeDefinition)
	for _, def := range doc.Definitions {
		if scalar, ok := def.(*types.ScalarTypeDefinition); ok {
			scalarTypes[scalar.Name] = scalar
		}
	}

	assert.Equal(t, 2, len(scalarTypes))
	assert.Contains(t, scalarTypes, "DateTime")
	assert.Contains(t, scalarTypes, "JSON")
}

func TestSchemaParser_ComplexSchema(t *testing.T) {
	parser := NewSchemaParser()

	// 测试复杂的Schema
	sdl := `
		"""
		用户查询接口
		"""
		type Query {
			"""获取用户列表"""
			users(first: Int, after: String): UserConnection!
			"""根据ID获取用户"""
			user(id: ID!): User
		}

		"""
		用户变更接口
		"""
		type Mutation {
			"""创建用户"""
			createUser(input: CreateUserInput!): CreateUserPayload!
			"""更新用户"""
			updateUser(id: ID!, input: UpdateUserInput!): UpdateUserPayload!
			"""删除用户"""
			deleteUser(id: ID!): DeleteUserPayload!
		}

		"""
		用户订阅接口
		"""
		type Subscription {
			"""用户更新订阅"""
			userUpdated(id: ID!): User!
		}

		type User {
			"""用户ID"""
			id: ID!
			"""用户名"""
			name: String!
			"""用户邮箱"""
			email: String
			"""创建时间"""
			createdAt: String!
		}

		type UserConnection {
			edges: [UserEdge!]!
			pageInfo: PageInfo!
		}

		type UserEdge {
			node: User!
			cursor: String!
		}

		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
		}

		input CreateUserInput {
			name: String!
			email: String
		}

		input UpdateUserInput {
			name: String
			email: String
		}

		type CreateUserPayload {
			user: User
			errors: [String!]
		}

		type UpdateUserPayload {
			user: User
			errors: [String!]
		}

		type DeleteUserPayload {
			deletedUserId: ID
			errors: [String!]
		}
	`

	doc, err := parser.ParseSchema(sdl)
	require.NoError(t, err)

	// 验证复杂Schema结构
	typeCount := 0
	for _, def := range doc.Definitions {
		switch def.(type) {
		case *types.ObjectTypeDefinition:
			typeCount++
		}
	}

	assert.GreaterOrEqual(t, typeCount, 10) // 至少包含10个对象类型

	// 验证Schema有效性
	err = parser.ValidateSchema(doc)
	require.NoError(t, err)
}
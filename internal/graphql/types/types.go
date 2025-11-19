package types

import (
	"time"

	"github.com/google/uuid"
)

// GraphQL基础类型定义

// OperationType GraphQL操作类型
type OperationType string

const (
	Query        OperationType = "QUERY"
	Mutation     OperationType = "MUTATION"
	Subscription OperationType = "SUBSCRIPTION"
)

// ResolverType 解析器类型
type ResolverType string

const (
	ResolverStatic   ResolverType = "STATIC"
	ResolverDynamic  ResolverType = "DYNAMIC"
	ResolverProxy    ResolverType = "PROXY"
	ResolverScript   ResolverType = "SCRIPT"
	ResolverTemplate ResolverType = "TEMPLATE"
	ResolverPlugin   ResolverType = "PLUGIN"
)

// GraphQL Schema 相关类型

// GraphQLSchema GraphQL Schema定义
type GraphQLSchema struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Operations  []*GraphQLOperation    `json:"operations"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// GraphQLOperation GraphQL操作定义
type GraphQLOperation struct {
	ID          string           `json:"id"`
	Type        OperationType    `json:"type"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Schema      string           `json:"schema"`
	Resolver    *GraphQLResolver `json:"resolver"`
	Config      *OperationConfig `json:"config"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// GraphQLResolver GraphQL解析器定义
type GraphQLResolver struct {
	ID           string                 `json:"id"`
	Type         ResolverType           `json:"type"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies"`
	Plugin       string                 `json:"plugin,omitempty"`
	Script       string                 `json:"script,omitempty"`
}

// OperationConfig 操作配置
type OperationConfig struct {
	Timeout    time.Duration `json:"timeout"`
	RetryCount int           `json:"retry_count"`
	Cache      bool          `json:"cache"`
	CacheTTL   time.Duration `json:"cache_ttl"`
	Middleware []string      `json:"middleware"`
}

// GraphQL 查询相关类型

// GraphQLQuery GraphQL查询
type GraphQLQuery struct {
	ID        string                 `json:"id"`
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
	Operation string                 `json:"operation"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
}

// GraphQLResult GraphQL查询结果
type GraphQLResult struct {
	ID         string                 `json:"id"`
	Data       interface{}            `json:"data"`
	Errors     []*GraphQLErrorWrapper `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError GraphQL错误
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []SourceLocation       `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// SourceLocation 源码位置
type SourceLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// GraphQL Schema AST相关类型

// Schema AST节点类型
type SchemaDocument struct {
	Definitions []Definition `json:"definitions"`
	Position    Position     `json:"position"`
}

type Definition interface {
	isDefinition()
	GetPosition() Position
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Schema定义类型
type SchemaDefinition struct {
	Description    string                    `json:"description"`
	Directives     []Directive               `json:"directives"`
	OperationTypes []OperationTypeDefinition `json:"operation_types"`
	Position       Position                  `json:"position"`
}

func (s *SchemaDefinition) isDefinition()         {}
func (s *SchemaDefinition) GetPosition() Position { return s.Position }

type OperationTypeDefinition struct {
	Operation OperationType `json:"operation"`
	Type      string        `json:"type"`
	Position  Position      `json:"position"`
}

// 类型定义
type TypeDefinition interface {
	isTypeDefinition()
	GetName() string
	GetPosition() Position
}

type ObjectTypeDefinition struct {
	Description string            `json:"description"`
	Name        string            `json:"name"`
	Implements  []string          `json:"implements"`
	Directives  []Directive       `json:"directives"`
	Fields      []FieldDefinition `json:"fields"`
	Position    Position          `json:"position"`
}

func (o *ObjectTypeDefinition) isTypeDefinition()     {}
func (o *ObjectTypeDefinition) GetName() string       { return o.Name }
func (o *ObjectTypeDefinition) GetPosition() Position { return o.Position }

type InterfaceTypeDefinition struct {
	Description string            `json:"description"`
	Name        string            `json:"name"`
	Directives  []Directive       `json:"directives"`
	Fields      []FieldDefinition `json:"fields"`
	Position    Position          `json:"position"`
}

func (i *InterfaceTypeDefinition) isTypeDefinition()     {}
func (i *InterfaceTypeDefinition) GetName() string       { return i.Name }
func (i *InterfaceTypeDefinition) GetPosition() Position { return i.Position }

type UnionTypeDefinition struct {
	Description string      `json:"description"`
	Name        string      `json:"name"`
	Directives  []Directive `json:"directives"`
	Types       []string    `json:"types"`
	Position    Position    `json:"position"`
}

func (u *UnionTypeDefinition) isTypeDefinition()     {}
func (u *UnionTypeDefinition) GetName() string       { return u.Name }
func (u *UnionTypeDefinition) GetPosition() Position { return u.Position }

type ScalarTypeDefinition struct {
	Description string      `json:"description"`
	Name        string      `json:"name"`
	Directives  []Directive `json:"directives"`
	Position    Position    `json:"position"`
}

func (s *ScalarTypeDefinition) isTypeDefinition()     {}
func (s *ScalarTypeDefinition) GetName() string       { return s.Name }
func (s *ScalarTypeDefinition) GetPosition() Position { return s.Position }

type EnumTypeDefinition struct {
	Description string                `json:"description"`
	Name        string                `json:"name"`
	Directives  []Directive           `json:"directives"`
	Values      []EnumValueDefinition `json:"values"`
	Position    Position              `json:"position"`
}

func (e *EnumTypeDefinition) isTypeDefinition()     {}
func (e *EnumTypeDefinition) GetName() string       { return e.Name }
func (e *EnumTypeDefinition) GetPosition() Position { return e.Position }

type InputObjectTypeDefinition struct {
	Description string                 `json:"description"`
	Name        string                 `json:"name"`
	Directives  []Directive            `json:"directives"`
	Fields      []InputValueDefinition `json:"fields"`
	Position    Position               `json:"position"`
}

func (i *InputObjectTypeDefinition) isTypeDefinition()     {}
func (i *InputObjectTypeDefinition) GetName() string       { return i.Name }
func (i *InputObjectTypeDefinition) GetPosition() Position { return i.Position }

// 字段定义
type FieldDefinition struct {
	Description string                 `json:"description"`
	Name        string                 `json:"name"`
	Arguments   []InputValueDefinition `json:"arguments"`
	Type        Type                   `json:"type"`
	Directives  []Directive            `json:"directives"`
	Position    Position               `json:"position"`
}

type InputValueDefinition struct {
	Description  string      `json:"description"`
	Name         string      `json:"name"`
	Type         Type        `json:"type"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Directives   []Directive `json:"directives"`
	Position     Position    `json:"position"`
}

type EnumValueDefinition struct {
	Description string      `json:"description"`
	Name        string      `json:"name"`
	Directives  []Directive `json:"directives"`
	Position    Position    `json:"position"`
}

// 类型引用
type Type interface {
	isType()
	String() string
}

type NamedType struct {
	Name string `json:"name"`
}

func (n *NamedType) isType()        {}
func (n *NamedType) String() string { return n.Name }

type ListType struct {
	Type Type `json:"type"`
}

func (l *ListType) isType()        {}
func (l *ListType) String() string { return "[" + l.Type.String() + "]" }

type NonNullType struct {
	Type Type `json:"type"`
}

func (n *NonNullType) isType()        {}
func (n *NonNullType) String() string { return n.Type.String() + "!" }

// 指令
type Directive struct {
	Name      string          `json:"name"`
	Arguments []ArgumentValue `json:"arguments"`
	Position  Position        `json:"position"`
}

type ArgumentValue struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// GraphQL 执行相关类型

// ExecutionContext 执行上下文
type ExecutionContext struct {
	RequestID string                 `json:"request_id"`
	Schema    *GraphQLSchema         `json:"schema"`
	Query     *GraphQLQuery          `json:"query"`
	Variables map[string]interface{} `json:"variables"`
	Operation string                 `json:"operation"`
	Fragments map[string]interface{} `json:"fragments"`
	Headers   map[string]string      `json:"headers"`
	Metadata  map[string]interface{} `json:"metadata"`
	StartTime time.Time              `json:"start_time"`
}

// FieldContext 字段执行上下文
type FieldContext struct {
	ParentType string                 `json:"parent_type"`
	FieldName  string                 `json:"field_name"`
	Arguments  map[string]interface{} `json:"arguments"`
	Alias      string                 `json:"alias"`
	Path       []string               `json:"path"`
	Resolver   *GraphQLResolver       `json:"resolver"`
}

// FieldExecutionResult 字段执行结果
type FieldExecutionResult struct {
	Value      interface{}            `json:"value"`
	Errors     []error                `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQL 订阅相关类型

// GraphQLSubscription GraphQL订阅
type GraphQLSubscription struct {
	ID          string                   `json:"id"`
	Query       string                   `json:"query"`
	Variables   map[string]interface{}   `json:"variables"`
	Connections []SubscriptionConnection `json:"connections"`
	CreatedAt   time.Time                `json:"created_at"`
}

// SubscriptionConnection 订阅连接
type SubscriptionConnection struct {
	ID             string                 `json:"id"`
	SubscriptionID string                 `json:"subscription_id"`
	ClientID       string                 `json:"client_id"`
	Context        map[string]interface{} `json:"context"`
	ConnectedAt    time.Time              `json:"connected_at"`
	LastPing       time.Time              `json:"last_ping"`
}

// SubscriptionEvent 订阅事件
type SubscriptionEvent struct {
	ID             string                 `json:"id"`
	SubscriptionID string                 `json:"subscription_id"`
	Type           string                 `json:"type"`
	Data           interface{}            `json:"data"`
	Context        map[string]interface{} `json:"context"`
	Timestamp      time.Time              `json:"timestamp"`
}

// GraphQL Introspection 相关类型

// IntrospectionQuery 内省查询
type IntrospectionQuery struct {
	Schema SchemaIntrospection `json:"__schema"`
}

type SchemaIntrospection struct {
	Types            []TypeIntrospection      `json:"types"`
	QueryType        TypeIntrospection        `json:"queryType"`
	MutationType     *TypeIntrospection       `json:"mutationType,omitempty"`
	SubscriptionType *TypeIntrospection       `json:"subscriptionType,omitempty"`
	Directives       []DirectiveIntrospection `json:"directives"`
}

type TypeIntrospection struct {
	Kind          string                    `json:"kind"`
	Name          string                    `json:"name"`
	Description   string                    `json:"description"`
	Fields        []FieldIntrospection      `json:"fields,omitempty"`
	InputFields   []InputValueIntrospection `json:"inputFields,omitempty"`
	Interfaces    []TypeIntrospection       `json:"interfaces,omitempty"`
	PossibleTypes []TypeIntrospection       `json:"possibleTypes,omitempty"`
	EnumValues    []EnumValueIntrospection  `json:"enumValues,omitempty"`
	InputTypes    []TypeIntrospection       `json:"inputTypes,omitempty"`
	OfType        *TypeIntrospection        `json:"ofType,omitempty"`
}

type FieldIntrospection struct {
	Name              string                    `json:"name"`
	Description       string                    `json:"description"`
	Args              []InputValueIntrospection `json:"args"`
	Type              TypeIntrospection         `json:"type"`
	IsDeprecated      bool                      `json:"isDeprecated"`
	DeprecationReason *string                   `json:"deprecationReason"`
}

type InputValueIntrospection struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Type         TypeIntrospection `json:"type"`
	DefaultValue *string           `json:"defaultValue,omitempty"`
}

type EnumValueIntrospection struct {
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	IsDeprecated      bool    `json:"isDeprecated"`
	DeprecationReason *string `json:"deprecationReason"`
}

type DirectiveIntrospection struct {
	Name              string                    `json:"name"`
	Description       string                    `json:"description"`
	Locations         []string                  `json:"locations"`
	Args              []InputValueIntrospection `json:"args"`
	IsDeprecated      bool                      `json:"isDeprecated"`
	DeprecationReason *string                   `json:"deprecationReason"`
}

// GraphQL 错误类型

// ErrorKind 错误类型
type ErrorKind string

const (
	ErrorKindSyntax     ErrorKind = "SYNTAX_ERROR"
	ErrorKindValidation ErrorKind = "VALIDATION_ERROR"
	ErrorKindExecution  ErrorKind = "EXECUTION_ERROR"
	ErrorKindInternal   ErrorKind = "INTERNAL_ERROR"
)

// GraphQLErrorWrapper GraphQL错误包装器
type GraphQLErrorWrapper struct {
	Kind       ErrorKind              `json:"kind"`
	Message    string                 `json:"message"`
	Locations  []SourceLocation       `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
	Internal   error                  `json:"-"`
}

func (e *GraphQLErrorWrapper) Error() string {
	return e.Message
}

// GraphQL 工具类型

// ValidationContext 验证上下文
type ValidationContext struct {
	Schema    *GraphQLSchema         `json:"schema"`
	Fragment  map[string]interface{} `json:"fragments"`
	Variables map[string]interface{} `json:"variables"`
	Errors    []*GraphQLErrorWrapper `json:"errors"`
}

// ExecutionPlan 执行计划
type ExecutionPlan struct {
	Operation    string               `json:"operation"`
	Fields       []FieldExecutionPlan `json:"fields"`
	Dependencies []string             `json:"dependencies"`
}

type FieldExecutionPlan struct {
	Field     string                 `json:"field"`
	Alias     string                 `json:"alias"`
	Arguments map[string]interface{} `json:"arguments"`
	Type      string                 `json:"type"`
	Resolver  string                 `json:"resolver"`
	SubFields []FieldExecutionPlan   `json:"sub_fields"`
}

// GraphQL 工具函数

// NewGraphQLSchema 创建GraphQL Schema
func NewGraphQLSchema(name, description, version string) *GraphQLSchema {
	now := time.Now()
	return &GraphQLSchema{
		ID:          generateID(),
		Name:        name,
		Description: description,
		Version:     version,
		Operations:  make([]*GraphQLOperation, 0),
		Metadata:    make(map[string]interface{}),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewGraphQLOperation 创建GraphQL操作
func NewGraphQLOperation(name string, opType OperationType, schema string) *GraphQLOperation {
	now := time.Now()
	return &GraphQLOperation{
		ID:        generateID(),
		Type:      opType,
		Name:      name,
		Schema:    schema,
		Resolver:  &GraphQLResolver{},
		Config:    &OperationConfig{},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// 辅助函数
func generateID() string {
	// 使用UUID生成ID
	return uuid.New().String()
}

// GenerateID 公共ID生成函数
func GenerateID() string {
	return generateID()
}

// IsOperationType 检查是否为有效的操作类型
func IsOperationType(opType string) bool {
	switch OperationType(opType) {
	case Query, Mutation, Subscription:
		return true
	default:
		return false
	}
}

// IsResolverType 检查是否为有效的解析器类型
func IsResolverType(resolverType string) bool {
	switch ResolverType(resolverType) {
	case ResolverStatic, ResolverDynamic, ResolverProxy, ResolverScript, ResolverTemplate, ResolverPlugin:
		return true
	default:
		return false
	}
}

// 为 ObjectTypeDefinition 实现 Definition 接口
func (o *ObjectTypeDefinition) isDefinition() {}

// 为 InputObjectTypeDefinition 实现 Definition 接口
func (i *InputObjectTypeDefinition) isDefinition() {}

// 为 InputValueDefinition 实现 Definition 接口
func (i *InputValueDefinition) isDefinition() {}

// 为 FieldDefinition 实现 Definition 接口
func (f *FieldDefinition) isDefinition() {}

// 为 EnumTypeDefinition 实现 Definition 接口
func (e *EnumTypeDefinition) isDefinition() {}

// 为 EnumValueDefinition 实现 Definition 接口
func (e *EnumValueDefinition) isDefinition() {}

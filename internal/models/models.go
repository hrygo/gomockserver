package models

import (
	"time"
)

// ProtocolType 协议类型
type ProtocolType string

const (
	ProtocolHTTP      ProtocolType = "HTTP"
	ProtocolWebSocket ProtocolType = "WebSocket"
	ProtocolGRPC      ProtocolType = "gRPC"
	ProtocolTCP       ProtocolType = "TCP"
	ProtocolUDP       ProtocolType = "UDP"
)

// MatchType 匹配类型
type MatchType string

const (
	MatchTypeSimple MatchType = "Simple"
	MatchTypeRegex  MatchType = "Regex"
	MatchTypeScript MatchType = "Script"
)

// ResponseType 响应类型
type ResponseType string

const (
	ResponseTypeStatic  ResponseType = "Static"
	ResponseTypeDynamic ResponseType = "Dynamic"
	ResponseTypeProxy   ResponseType = "Proxy"
	ResponseTypeScript  ResponseType = "Script"
)

// ContentType 内容类型
type ContentType string

const (
	ContentTypeJSON   ContentType = "JSON"
	ContentTypeXML    ContentType = "XML"
	ContentTypeHTML   ContentType = "HTML"
	ContentTypeText   ContentType = "Text"
	ContentTypeBinary ContentType = "Binary"
)

// Rule Mock 规则模型
type Rule struct {
	ID             string                 `bson:"_id,omitempty" json:"id"`
	Name           string                 `bson:"name" json:"name"`
	ProjectID      string                 `bson:"project_id" json:"project_id"`
	EnvironmentID  string                 `bson:"environment_id" json:"environment_id"`
	Protocol       ProtocolType           `bson:"protocol" json:"protocol"`
	MatchType      MatchType              `bson:"match_type" json:"match_type"`
	Priority       int                    `bson:"priority" json:"priority"`
	Enabled        bool                   `bson:"enabled" json:"enabled"`
	MatchCondition map[string]interface{} `bson:"match_condition" json:"match_condition"`
	Response       Response               `bson:"response" json:"response"`
	Tags           []string               `bson:"tags,omitempty" json:"tags,omitempty"`
	Creator        string                 `bson:"creator,omitempty" json:"creator,omitempty"`
	CreatedAt      time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time              `bson:"updated_at" json:"updated_at"`
}

// HTTPMatchCondition HTTP 匹配条件
type HTTPMatchCondition struct {
	Method      interface{}            `json:"method"` // string 或 []string
	Path        string                 `json:"path"`
	PathRegex   string                 `json:"path_regex,omitempty"`
	Query       map[string]string      `json:"query,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Body        map[string]interface{} `json:"body,omitempty"`
	IPWhitelist []string               `json:"ip_whitelist,omitempty"`
}

// Response 响应配置
type Response struct {
	Type    ResponseType           `bson:"type" json:"type"`
	Delay   *DelayConfig           `bson:"delay,omitempty" json:"delay,omitempty"`
	Content map[string]interface{} `bson:"content" json:"content"`
}

// DelayConfig 延迟配置
type DelayConfig struct {
	Type   string `bson:"type" json:"type"` // fixed, random, normal, step
	Min    int    `bson:"min,omitempty" json:"min,omitempty"`
	Max    int    `bson:"max,omitempty" json:"max,omitempty"`
	Fixed  int    `bson:"fixed,omitempty" json:"fixed,omitempty"`
	Mean   int    `bson:"mean,omitempty" json:"mean,omitempty"`
	StdDev int    `bson:"std_dev,omitempty" json:"std_dev,omitempty"`
	Step   int    `bson:"step,omitempty" json:"step,omitempty"`
	Limit  int    `bson:"limit,omitempty" json:"limit,omitempty"`
}

// HTTPResponse HTTP 响应配置
type HTTPResponse struct {
	StatusCode  int               `json:"status_code"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        interface{}       `json:"body"`
	ContentType ContentType       `json:"content_type"`
}

// Project 项目模型
type Project struct {
	ID          string    `bson:"_id,omitempty" json:"id"`
	Name        string    `bson:"name" json:"name"`
	WorkspaceID string    `bson:"workspace_id" json:"workspace_id"`
	Description string    `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// Environment 环境模型
type Environment struct {
	ID        string                 `bson:"_id,omitempty" json:"id"`
	Name      string                 `bson:"name" json:"name"`
	ProjectID string                 `bson:"project_id" json:"project_id"`
	BaseURL   string                 `bson:"base_url,omitempty" json:"base_url,omitempty"`
	Variables map[string]interface{} `bson:"variables,omitempty" json:"variables,omitempty"`
	CreatedAt time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time              `bson:"updated_at" json:"updated_at"`
}

// Workspace 工作空间模型
type Workspace struct {
	ID          string                 `bson:"_id,omitempty" json:"id"`
	Name        string                 `bson:"name" json:"name"`
	Owner       string                 `bson:"owner" json:"owner"`
	Members     []string               `bson:"members,omitempty" json:"members,omitempty"`
	Permissions map[string]interface{} `bson:"permissions,omitempty" json:"permissions,omitempty"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// RequestLog 请求日志模型
type RequestLog struct {
	ID            string                 `bson:"_id,omitempty" json:"id"`
	RequestID     string                 `bson:"request_id" json:"request_id"`
	ProjectID     string                 `bson:"project_id" json:"project_id"`
	EnvironmentID string                 `bson:"environment_id" json:"environment_id"`
	RuleID        string                 `bson:"rule_id,omitempty" json:"rule_id,omitempty"`
	Protocol      ProtocolType           `bson:"protocol" json:"protocol"`
	Method        string                 `bson:"method,omitempty" json:"method,omitempty"`
	Path          string                 `bson:"path,omitempty" json:"path,omitempty"`
	Request       map[string]interface{} `bson:"request" json:"request"`
	Response      map[string]interface{} `bson:"response" json:"response"`
	StatusCode    int                    `bson:"status_code,omitempty" json:"status_code,omitempty"`
	Duration      int64                  `bson:"duration" json:"duration"` // 毫秒
	SourceIP      string                 `bson:"source_ip" json:"source_ip"`
	Timestamp     time.Time              `bson:"timestamp" json:"timestamp"`
}

// Version 版本记录模型
type Version struct {
	ID          string                 `bson:"_id,omitempty" json:"id"`
	RuleID      string                 `bson:"rule_id" json:"rule_id"`
	Version     string                 `bson:"version" json:"version"`
	ChangeType  string                 `bson:"change_type" json:"change_type"` // Create, Update, Delete
	Changes     map[string]interface{} `bson:"changes" json:"changes"`
	Operator    string                 `bson:"operator" json:"operator"`
	Description string                 `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
}

// User 用户模型
type User struct {
	ID           string    `bson:"_id,omitempty" json:"id"`
	Username     string    `bson:"username" json:"username"`
	Email        string    `bson:"email" json:"email"`
	PasswordHash string    `bson:"password_hash" json:"-"`
	Role         string    `bson:"role" json:"role"` // Admin, Developer, Viewer
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}

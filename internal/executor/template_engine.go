package executor

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
)

// TemplateEngine 模板引擎
type TemplateEngine struct {
	funcMap template.FuncMap
	counter int64 // 用于计数器功能
}

// NewTemplateEngine 创建模板引擎
func NewTemplateEngine() *TemplateEngine {
	engine := &TemplateEngine{}
	engine.funcMap = engine.buildFuncMap()
	return engine
}

// buildFuncMap 构建模板函数映射
func (e *TemplateEngine) buildFuncMap() template.FuncMap {
	return template.FuncMap{
		// 时间相关函数
		"timestamp": func() int64 {
			return time.Now().Unix()
		},
		"timestampMilli": func() int64 {
			return time.Now().UnixMilli()
		},
		"now": func(format string) string {
			if format == "" {
				format = "2006-01-02 15:04:05"
			}
			return time.Now().Format(format)
		},
		"date": func(format string) string {
			if format == "" {
				format = "2006-01-02"
			}
			return time.Now().Format(format)
		},
		"date_format": func(format string) string {
			if format == "" {
				format = "2006-01-02"
			}
			return time.Now().Format(format)
		},
		"time": func(format string) string {
			if format == "" {
				format = "15:04:05"
			}
			return time.Now().Format(format)
		},

		// 随机数相关函数
		"uuid": func() string {
			return uuid.New().String()
		},
		"uuidShort": func() string {
			id := uuid.New().String()
			return id[:8]
		},
		"random": func(min, max int) int {
			if max <= min {
				return min
			}
			return min + rand.Intn(max-min)
		},
		"randomString": func(length int) string {
			const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			result := make([]byte, length)
			for i := range result {
				result[i] = charset[rand.Intn(len(charset))]
			}
			return string(result)
		},
		"randomInt": func() int {
			return rand.Int()
		},
		"randomFloat": func() float64 {
			return rand.Float64()
		},

		// 计数器相关函数
		"counter": func() int64 {
			e.counter++
			return e.counter
		},

		// 编码相关函数
		"base64": func(s string) string {
			return base64.StdEncoding.EncodeToString([]byte(s))
		},
		"base64Decode": func(s string) (string, error) {
			decoded, err := base64.StdEncoding.DecodeString(s)
			if err != nil {
				return "", err
			}
			return string(decoded), nil
		},

		// 字符串相关函数
		"concat": func(strs ...string) string {
			var result string
			for _, s := range strs {
				result += s
			}
			return result
		},
		"quote": func(s string) string {
			return fmt.Sprintf("\"%s\"", s)
		},

		// JSON 相关函数
		"toJSON": func(v interface{}) (string, error) {
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		},
		"toJSONPretty": func(v interface{}) (string, error) {
			bytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		},
	}
}

// TemplateContext 模板上下文
type TemplateContext struct {
	Request     *RequestContext     `json:"request"`
	Rule        *RuleContext        `json:"rule"`
	Environment *EnvironmentContext `json:"environment"`
}

// RequestContext 请求上下文
type RequestContext struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Headers map[string]string      `json:"headers"`
	Query   map[string]string      `json:"query"`
	Body    interface{}            `json:"body"`
	IP      string                 `json:"ip"`
}

// RuleContext 规则上下文
type RuleContext struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
}

// EnvironmentContext 环境上下文
type EnvironmentContext struct {
	Variables map[string]interface{} `json:"variables"`
}

// BuildContext 构建模板上下文
func (e *TemplateEngine) BuildContext(request *adapter.Request, rule *models.Rule, env *models.Environment) *TemplateContext {
	// 提取HTTP特定信息
	method := ""
	query := make(map[string]string)
	if request.Metadata != nil {
		if m, ok := request.Metadata["method"].(string); ok {
			method = m
		}
		if q, ok := request.Metadata["query"].(map[string]string); ok {
			query = q
		}
	}

	ctx := &TemplateContext{
		Request: &RequestContext{
			Method:  method,
			Path:    request.Path,
			Headers: request.Headers,
			Query:   query,
			IP:      request.SourceIP,
		},
		Rule: &RuleContext{
			ID:       rule.ID,
			Name:     rule.Name,
			Priority: rule.Priority,
		},
		Environment: &EnvironmentContext{
			Variables: make(map[string]interface{}),
		},
	}

	// 解析请求体
	if len(request.Body) > 0 {
		var body interface{}
		if err := json.Unmarshal(request.Body, &body); err == nil {
			ctx.Request.Body = body
		} else {
			ctx.Request.Body = string(request.Body)
		}
	}

	// 设置环境变量
	if env != nil && env.Variables != nil {
		ctx.Environment.Variables = env.Variables
	}

	return ctx
}

// Render 渲染模板
func (e *TemplateEngine) Render(templateStr string, context *TemplateContext) (string, error) {
	// 创建模板
	tmpl, err := template.New("response").Funcs(e.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderJSON 渲染JSON模板
func (e *TemplateEngine) RenderJSON(templateObj interface{}, context *TemplateContext) (interface{}, error) {
	return e.renderJSONRecursive(templateObj, context)
}

// renderJSONRecursive 递归渲染JSON对象
func (e *TemplateEngine) renderJSONRecursive(obj interface{}, context *TemplateContext) (interface{}, error) {
	switch v := obj.(type) {
	case string:
		// 如果是字符串，检查是否包含模板语法
		if len(v) > 0 && (v[0] == '{' && v[len(v)-1] == '}' || contains(v, "{{")) {
			rendered, err := e.Render(v, context)
			if err != nil {
				return v, nil // 如果渲染失败，返回原始字符串
			}
			return rendered, nil
		}
		return v, nil
	
	case map[string]interface{}:
		// 递归处理map
		result := make(map[string]interface{})
		for key, val := range v {
			rendered, err := e.renderJSONRecursive(val, context)
			if err != nil {
				return nil, err
			}
			result[key] = rendered
		}
		return result, nil
	
	case []interface{}:
		// 递归处理数组
		result := make([]interface{}, len(v))
		for i, val := range v {
			rendered, err := e.renderJSONRecursive(val, context)
			if err != nil {
				return nil, err
			}
			result[i] = rendered
		}
		return result, nil
	
	default:
		// 其他类型直接返回
		return v, nil
	}
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

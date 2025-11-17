package engine

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// ScriptEngine 脚本匹配引擎
type ScriptEngine struct {
	// 资源限制
	maxExecutionTime time.Duration
	maxMemory        int64
	// 审计日志
	auditLog bool
}

// ScriptMatchConfig 脚本匹配配置
type ScriptMatchConfig struct {
	Script string `json:"script"` // JavaScript 脚本代码
}

// ScriptContext 脚本执行上下文
type ScriptContext struct {
	Request *adapter.Request
	Rule    *models.Rule
	Env     map[string]interface{}
}

// NewScriptEngine 创建脚本引擎
func NewScriptEngine() *ScriptEngine {
	return &ScriptEngine{
		maxExecutionTime: 5 * time.Second, // 默认最大执行时间 5 秒
		maxMemory:        10 * 1024 * 1024, // 默认最大内存 10MB
		auditLog:         true,
	}
}

// Match 执行脚本匹配
func (e *ScriptEngine) Match(request *adapter.Request, rule *models.Rule) (bool, error) {
	// 检查规则类型
	if rule.MatchType != models.MatchTypeScript {
		return false, errors.New("not a script match rule")
	}

	// 提取脚本配置
	script, ok := rule.MatchCondition["script"].(string)
	if !ok || script == "" {
		return false, errors.New("script not found in match condition")
	}

	// 创建脚本上下文
	ctx := &ScriptContext{
		Request: request,
		Rule:    rule,
		Env:     make(map[string]interface{}),
	}

	// 执行脚本
	startTime := time.Now()
	result, err := e.executeScript(script, ctx)
	duration := time.Since(startTime)

	// 审计日志
	if e.auditLog {
		logger.Info("script execution",
			zap.String("rule_id", rule.ID),
			zap.String("rule_name", rule.Name),
			zap.Duration("duration", duration),
			zap.Bool("result", result),
			zap.Error(err),
		)
	}

	return result, err
}

// executeScript 执行 JavaScript 脚本
func (e *ScriptEngine) executeScript(script string, ctx *ScriptContext) (bool, error) {
	// 创建超时上下文
	execCtx, cancel := context.WithTimeout(context.Background(), e.maxExecutionTime)
	defer cancel()

	// 创建 goja 运行时
	vm := goja.New()
	
	// 设置中断处理
	done := make(chan struct{})
	defer close(done)
	
	go func() {
		select {
		case <-execCtx.Done():
			vm.Interrupt("execution timeout")
		case <-done:
		}
	}()

	// 注入安全 API
	e.injectSecureAPI(vm, ctx)

	// 执行脚本
	result, err := vm.RunString(script)
	if err != nil {
		logger.Error("script execution error", 
			zap.String("script", script),
			zap.Error(err),
		)
		return false, fmt.Errorf("script execution failed: %w", err)
	}

	// 转换结果为布尔值
	matched, ok := result.Export().(bool)
	if !ok {
		return false, errors.New("script must return a boolean value")
	}

	return matched, nil
}

// injectSecureAPI 注入安全的 API 到脚本环境
func (e *ScriptEngine) injectSecureAPI(vm *goja.Runtime, ctx *ScriptContext) {
	// 注入 request 对象
	vm.Set("request", map[string]interface{}{
		"id":       ctx.Request.ID,
		"protocol": string(ctx.Request.Protocol),
		"path":     ctx.Request.Path,
		"headers":  ctx.Request.Headers,
		"body":     string(ctx.Request.Body),
		"sourceIP": ctx.Request.SourceIP,
		"metadata": ctx.Request.Metadata,
	})

	// 注入 rule 对象（只读）
	vm.Set("rule", map[string]interface{}{
		"id":          ctx.Rule.ID,
		"name":        ctx.Rule.Name,
		"project_id":  ctx.Rule.ProjectID,
		"environment": ctx.Rule.EnvironmentID,
		"priority":    ctx.Rule.Priority,
	})

	// 注入工具函数
	vm.Set("log", func(msg string) {
		logger.Info("script log", zap.String("message", msg))
	})

	vm.Set("match", func(pattern string, text string) bool {
		// 简单的字符串匹配
		return contains(text, pattern)
	})

	vm.Set("hasHeader", func(key string) bool {
		_, exists := ctx.Request.Headers[key]
		return exists
	})

	vm.Set("getHeader", func(key string) string {
		return ctx.Request.Headers[key]
	})

	vm.Set("hasQuery", func(key string) bool {
		if query, ok := ctx.Request.Metadata["query"].(map[string]string); ok {
			_, exists := query[key]
			return exists
		}
		return false
	})

	vm.Set("getQuery", func(key string) string {
		if query, ok := ctx.Request.Metadata["query"].(map[string]string); ok {
			return query[key]
		}
		return ""
	})

	// 禁用危险功能
	vm.Set("require", goja.Undefined())
	vm.Set("import", goja.Undefined())
	vm.Set("eval", goja.Undefined())
	vm.Set("Function", goja.Undefined())
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

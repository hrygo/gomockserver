package engine

import (
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewScriptEngine(t *testing.T) {
	engine := NewScriptEngine()
	assert.NotNil(t, engine)
	assert.Equal(t, 5*time.Second, engine.maxExecutionTime)
	assert.True(t, engine.auditLog)
}

func TestScriptEngine_Match_SimpleScript(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
		Path:     "/api/users",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: []byte(`{"name":"test"}`),
		Metadata: map[string]interface{}{
			"method": "POST",
			"query": map[string]string{
				"id": "123",
			},
		},
	}

	rule := &models.Rule{
		ID:        "rule-1",
		Name:      "Test Rule",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `request.path === "/api/users"`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.NoError(t, err)
	assert.True(t, matched)
}

func TestScriptEngine_Match_HeaderCheck(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
		Path:     "/api/test",
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"Content-Type":  "application/json",
		},
	}

	rule := &models.Rule{
		ID:        "rule-2",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `hasHeader("Authorization") && getHeader("Authorization").indexOf("Bearer") === 0`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.NoError(t, err)
	assert.True(t, matched)
}

func TestScriptEngine_Match_QueryCheck(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
		Path:     "/api/search",
		Metadata: map[string]interface{}{
			"query": map[string]string{
				"keyword": "golang",
				"page":    "1",
			},
		},
	}

	rule := &models.Rule{
		ID:        "rule-3",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `hasQuery("keyword") && getQuery("keyword") === "golang"`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.NoError(t, err)
	assert.True(t, matched)
}

func TestScriptEngine_Match_ComplexLogic(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
		Path:     "/api/users/123",
		Headers: map[string]string{
			"X-User-Role": "admin",
		},
		Metadata: map[string]interface{}{
			"method": "DELETE",
		},
	}

	rule := &models.Rule{
		ID:        "rule-4",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `
				// 只有管理员可以删除用户
				request.path.indexOf("/api/users/") === 0 &&
				getHeader("X-User-Role") === "admin"
			`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.NoError(t, err)
	assert.True(t, matched)
}

func TestScriptEngine_Match_NoMatch(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
		Path:     "/api/products",
	}

	rule := &models.Rule{
		ID:        "rule-5",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `request.path === "/api/users"`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.NoError(t, err)
	assert.False(t, matched)
}

func TestScriptEngine_Match_InvalidScript(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
	}

	rule := &models.Rule{
		ID:        "rule-6",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `this is not valid javascript code {{{`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.Error(t, err)
	assert.False(t, matched)
}

func TestScriptEngine_Match_NonBooleanReturn(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
	}

	rule := &models.Rule{
		ID:        "rule-7",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `"this returns a string, not a boolean"`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must return a boolean")
	assert.False(t, matched)
}

func TestScriptEngine_Match_Timeout(t *testing.T) {
	engine := NewScriptEngine()
	engine.maxExecutionTime = 100 * time.Millisecond

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
	}

	rule := &models.Rule{
		ID:        "rule-8",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			"script": `
				// 无限循环导致超时
				while(true) {}
				true
			`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.Error(t, err)
	assert.False(t, matched)
}

func TestScriptEngine_Match_WrongRuleType(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
	}

	rule := &models.Rule{
		ID:        "rule-9",
		MatchType: models.MatchTypeSimple, // 不是 Script 类型
		MatchCondition: map[string]interface{}{
			"script": `true`,
		},
	}

	matched, err := engine.Match(request, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a script match rule")
	assert.False(t, matched)
}

func TestScriptEngine_Match_MissingScript(t *testing.T) {
	engine := NewScriptEngine()

	request := &adapter.Request{
		ID:       "test-123",
		Protocol: models.ProtocolHTTP,
	}

	rule := &models.Rule{
		ID:        "rule-10",
		MatchType: models.MatchTypeScript,
		MatchCondition: map[string]interface{}{
			// 没有 script 字段
		},
	}

	matched, err := engine.Match(request, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "script not found")
	assert.False(t, matched)
}

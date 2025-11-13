package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// MatchEngine 规则匹配引擎
type MatchEngine struct {
	ruleRepo repository.RuleRepository
}

// NewMatchEngine 创建匹配引擎
func NewMatchEngine(ruleRepo repository.RuleRepository) *MatchEngine {
	return &MatchEngine{
		ruleRepo: ruleRepo,
	}
}

// Match 匹配规则
func (e *MatchEngine) Match(ctx context.Context, request *adapter.Request, projectID, environmentID string) (*models.Rule, error) {
	// 加载环境下所有启用的规则
	rules, err := e.ruleRepo.FindEnabledByEnvironment(ctx, projectID, environmentID)
	if err != nil {
		logger.Error("failed to load rules", zap.Error(err))
		return nil, err
	}

	// 按优先级已排序，逐条匹配
	for _, rule := range rules {
		// 检查协议类型
		if rule.Protocol != request.Protocol {
			continue
		}

		// 根据匹配类型执行匹配
		matched, err := e.matchRule(request, rule)
		if err != nil {
			logger.Warn("rule match error",
				zap.String("rule_id", rule.ID),
				zap.Error(err))
			continue
		}

		if matched {
			logger.Info("rule matched",
				zap.String("rule_id", rule.ID),
				zap.String("rule_name", rule.Name))
			return rule, nil
		}
	}

	// 没有匹配的规则
	return nil, nil
}

// matchRule 匹配单条规则
func (e *MatchEngine) matchRule(request *adapter.Request, rule *models.Rule) (bool, error) {
	switch rule.MatchType {
	case models.MatchTypeSimple:
		return e.simpleMatch(request, rule)
	case models.MatchTypeRegex:
		return e.regexMatch(request, rule)
	case models.MatchTypeScript:
		return e.scriptMatch(request, rule)
	default:
		return false, fmt.Errorf("unsupported match type: %s", rule.MatchType)
	}
}

// simpleMatch 简单匹配
func (e *MatchEngine) simpleMatch(request *adapter.Request, rule *models.Rule) (bool, error) {
	if rule.Protocol != models.ProtocolHTTP {
		return false, nil
	}

	// 解析 HTTP 匹配条件
	conditionBytes, err := json.Marshal(rule.MatchCondition)
	if err != nil {
		return false, err
	}

	var condition models.HTTPMatchCondition
	if err := json.Unmarshal(conditionBytes, &condition); err != nil {
		return false, err
	}

	// 获取请求方法
	method, _ := request.Metadata["method"].(string)

	// 匹配 Method
	if condition.Method != nil {
		if !matchMethod(method, condition.Method) {
			return false, nil
		}
	}

	// 匹配 Path
	if condition.Path != "" {
		if !matchPath(request.Path, condition.Path) {
			return false, nil
		}
	}

	// 匹配 Query 参数
	if len(condition.Query) > 0 {
		query, _ := request.Metadata["query"].(map[string]string)
		if !matchQuery(query, condition.Query) {
			return false, nil
		}
	}

	// 匹配 Headers
	if len(condition.Headers) > 0 {
		if !matchHeaders(request.Headers, condition.Headers) {
			return false, nil
		}
	}

	// 匹配 IP 白名单
	if len(condition.IPWhitelist) > 0 {
		if !matchIPWhitelist(request.SourceIP, condition.IPWhitelist) {
			return false, nil
		}
	}

	return true, nil
}

// regexMatch 正则表达式匹配（阶段一暂不实现）
func (e *MatchEngine) regexMatch(request *adapter.Request, rule *models.Rule) (bool, error) {
	// TODO: 阶段三实现
	return false, fmt.Errorf("regex match not implemented yet")
}

// scriptMatch 脚本匹配（阶段一暂不实现）
func (e *MatchEngine) scriptMatch(request *adapter.Request, rule *models.Rule) (bool, error) {
	// TODO: 阶段三实现
	return false, fmt.Errorf("script match not implemented yet")
}

// matchMethod 匹配请求方法
func matchMethod(requestMethod string, conditionMethod interface{}) bool {
	switch v := conditionMethod.(type) {
	case string:
		return strings.EqualFold(requestMethod, v)
	case []interface{}:
		for _, m := range v {
			if method, ok := m.(string); ok {
				if strings.EqualFold(requestMethod, method) {
					return true
				}
			}
		}
		return false
	default:
		return false
	}
}

// matchPath 匹配路径（支持简单通配符）
func matchPath(requestPath, conditionPath string) bool {
	// 精确匹配
	if requestPath == conditionPath {
		return true
	}

	// 支持简单的路径参数匹配，如 /api/users/:id
	conditionParts := strings.Split(conditionPath, "/")
	requestParts := strings.Split(requestPath, "/")

	if len(conditionParts) != len(requestParts) {
		return false
	}

	for i, part := range conditionParts {
		if strings.HasPrefix(part, ":") {
			// 路径参数，匹配任意值
			continue
		}
		if part != requestParts[i] {
			return false
		}
	}

	return true
}

// matchQuery 匹配查询参数
func matchQuery(requestQuery, conditionQuery map[string]string) bool {
	for key, value := range conditionQuery {
		if requestQuery[key] != value {
			return false
		}
	}
	return true
}

// matchHeaders 匹配请求头
func matchHeaders(requestHeaders, conditionHeaders map[string]string) bool {
	for key, value := range conditionHeaders {
		// Header 名称不区分大小写
		found := false
		for reqKey, reqValue := range requestHeaders {
			if strings.EqualFold(reqKey, key) {
				if reqValue != value {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// matchIPWhitelist 匹配 IP 白名单（简单实现）
func matchIPWhitelist(requestIP string, whitelist []string) bool {
	for _, ip := range whitelist {
		if requestIP == ip {
			return true
		}
		// TODO: 支持 CIDR 格式的 IP 段匹配
	}
	return false
}

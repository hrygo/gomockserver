package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// RegexCacheStats 缓存统计信息
type RegexCacheStats struct {
	Hits   int64
	Misses int64
	Size   int
}

// MatchEngine 规则匹配引擎
type MatchEngine struct {
	ruleRepo     repository.RuleRepository
	regexCache   *LRURegexCache
	cacheStats   RegexCacheStats
	statsMu      sync.RWMutex
}

// NewMatchEngine 创建匹配引擎
func NewMatchEngine(ruleRepo repository.RuleRepository) *MatchEngine {
	return &MatchEngine{
		ruleRepo:   ruleRepo,
		regexCache: NewLRURegexCache(1000), // 默认缓存容量1000
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

// compileRegex 编译正则表达式并缓存
func (e *MatchEngine) compileRegex(pattern string) (*regexp.Regexp, error) {
	// 先尝试从缓存中获取
	if re, exists := e.regexCache.Get(pattern); exists {
		e.statsMu.Lock()
		e.cacheStats.Hits++
		e.statsMu.Unlock()
		return re, nil
	}

	e.statsMu.Lock()
	e.cacheStats.Misses++
	e.statsMu.Unlock()

	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	e.regexCache.Put(pattern, re)

	return re, nil
}

// GetCacheStats 获取缓存统计信息
func (e *MatchEngine) GetCacheStats() RegexCacheStats {
	e.statsMu.RLock()
	defer e.statsMu.RUnlock()
	stats := e.cacheStats
	stats.Size = e.regexCache.Size()
	return stats
}

// regexMatch 正则表达式匹配
func (e *MatchEngine) regexMatch(request *adapter.Request, rule *models.Rule) (bool, error) {
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

	// 匹配 Path (支持正则表达式)
	if condition.Path != "" {
		// 编译正则表达式
		re, err := e.compileRegex(condition.Path)
		if err != nil {
			logger.Warn("failed to compile regex pattern for path",
				zap.String("pattern", condition.Path),
				zap.Error(err))
			return false, nil
		}

		// 执行正则匹配
		if !re.MatchString(request.Path) {
			return false, nil
		}
	}

	// 匹配 Query 参数 (支持正则表达式)
	if len(condition.Query) > 0 {
		query, _ := request.Metadata["query"].(map[string]string)
		for key, pattern := range condition.Query {
			value, exists := query[key]
			if !exists {
				return false, nil
			}

			// 编译正则表达式
			re, err := e.compileRegex(pattern)
			if err != nil {
				logger.Warn("failed to compile regex pattern for query",
					zap.String("key", key),
					zap.String("pattern", pattern),
					zap.Error(err))
				return false, nil
			}

			// 执行正则匹配
			if !re.MatchString(value) {
				return false, nil
			}
		}
	}

	// 匹配 Headers (支持正则表达式)
	if len(condition.Headers) > 0 {
		for key, pattern := range condition.Headers {
			// Header 名称不区分大小写查找
			found := false
			var value string
			for reqKey, reqValue := range request.Headers {
				if strings.EqualFold(reqKey, key) {
					value = reqValue
					found = true
					break
				}
			}

			if !found {
				return false, nil
			}

			// 编译正则表达式
			re, err := e.compileRegex(pattern)
			if err != nil {
				logger.Warn("failed to compile regex pattern for header",
					zap.String("key", key),
					zap.String("pattern", pattern),
					zap.Error(err))
				return false, nil
			}

			// 执行正则匹配
			if !re.MatchString(value) {
				return false, nil
			}
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

// matchIPWhitelist 匹配 IP 白名单（支持精确IP和CIDR格式）
func matchIPWhitelist(requestIP string, whitelist []string) bool {
	// 解析请求IP
	requestIPAddr := net.ParseIP(requestIP)
	if requestIPAddr == nil {
		logger.Warn("invalid request IP address", zap.String("ip", requestIP))
		return false
	}

	for _, ipEntry := range whitelist {
		// 尝试解析为CIDR格式
		if _, ipNet, err := net.ParseCIDR(ipEntry); err == nil {
			// CIDR格式匹配
			if ipNet.Contains(requestIPAddr) {
				return true
			}
			continue
		}

		// 尝试解析为精确IP格式
		if ip := net.ParseIP(ipEntry); ip != nil {
			// 精确IP匹配
			if ip.Equal(requestIPAddr) {
				return true
			}
			continue
		}

		// 无效的IP格式，记录警告但不中断其他白名单条目的处理
		logger.Warn("invalid IP format in whitelist", zap.String("ip", ipEntry))
	}

	return false
}

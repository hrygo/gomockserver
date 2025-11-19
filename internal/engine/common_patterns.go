// Package engine provides the rule matching engine for the mock server.
package engine

import (
	"regexp"

	"github.com/gomockserver/mockserver/internal/monitoring"
)

// CommonRegexPatterns 定义常用的正则表达式模式
var CommonRegexPatterns = []string{
	// HTTP路径相关
	`^/api/.*$`,          // API路径
	`^/api/v[0-9]+/.*$`,  // 带版本的API路径
	`^/users/[^/]+$`,     // 用户路径
	`^/users/[0-9]+$`,    // 数字用户ID路径
	`^/users/[^/]+/.*$`,  // 用户子路径
	`^/products/[^/]+$`,  // 产品路径
	`^/products/[0-9]+$`, // 数字产品ID路径
	`^/orders/[^/]+$`,    // 订单路径
	`^/orders/[0-9]+$`,   // 数字订单ID路径

	// 查询参数相关
	`^[0-9]+$`,         // 数字
	`^[a-zA-Z0-9_-]+$`, // 字母数字下划线横线
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`, // UUID
	`^\d{4}-\d{2}-\d{2}$`,                    // 日期 (YYYY-MM-DD)
	`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`, // ISO 8601 日期时间

	// HTTP头相关
	`^[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, // 邮箱
	`^Bearer\s+[a-zA-Z0-9._-]+$`,                     // Bearer Token
	`^Basic\s+[a-zA-Z0-9+/=]+$`,                      // Basic Auth
	`^[A-Za-z0-9+/]*={0,2}$`,                         // Base64

	// 通用模式
	`.*`,           // 任意字符
	`^.*$`,         // 任意字符（锚定）
	`[a-zA-Z0-9]+`, // 字母数字
	`[0-9]+`,       // 数字
	`[a-zA-Z]+`,    // 字母
}

// PrecompileCommonPatterns 预编译常用正则表达式模式
func (e *MatchEngine) PrecompileCommonPatterns() {
	for _, pattern := range CommonRegexPatterns {
		// 验证模式
		if err := validateRegexPattern(pattern); err != nil {
			continue // 跳过无效模式
		}

		// 编译并缓存
		if re, err := regexp.Compile(pattern); err == nil {
			e.regexCache.Put(pattern, re)
		}
	}

	// 更新缓存统计
	monitoring.SetRegexCacheSize(int64(e.regexCache.Size()))
}

package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

// CORSConfig CORS 配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// DefaultCORSConfig 默认 CORS 配置
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173", // 前端开发服务器
			"http://localhost:8080", // 前端生产环境（集成部署）
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
		},
		AllowHeaders: []string{
			"Content-Type",
			"Authorization",  // 预留用于 v0.9.0
			"X-Request-ID",   // 前端添加的请求追踪 ID
		},
		ExposeHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// CORS 创建 CORS 中间件
func CORS(config ...CORSConfig) gin.HandlerFunc {
	// 如果没有提供配置，使用默认配置
	cfg := DefaultCORSConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}

	return cors.New(corsConfig)
}

// CORSWithDynamicOrigin 创建支持动态源的 CORS 中间件
// 用于支持更灵活的 CORS 配置（例如：允许所有 localhost 端口）
func CORSWithDynamicOrigin() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// 允许所有 localhost 源
			if origin == "http://localhost:5173" || origin == "http://localhost:8080" {
				return true
			}
			// 生产环境可以根据配置文件或环境变量添加允许的域名
			return false
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
		},
		AllowHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Request-ID",
		},
		ExposeHeaders: []string{
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

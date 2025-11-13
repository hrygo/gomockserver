package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	Security    SecurityConfig    `mapstructure:"security"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Performance PerformanceConfig `mapstructure:"performance"`
	Features    FeaturesConfig    `mapstructure:"features"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Admin AdminServerConfig `mapstructure:"admin"`
	Mock  MockServerConfig  `mapstructure:"mock"`
}

// AdminServerConfig 管理 API 服务配置
type AdminServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// MockServerConfig Mock 服务配置
type MockServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	MongoDB MongoDBConfig `mapstructure:"mongodb"`
}

// MongoDBConfig MongoDB 配置
type MongoDBConfig struct {
	URI      string         `mapstructure:"uri"`
	Database string         `mapstructure:"database"`
	Timeout  time.Duration  `mapstructure:"timeout"`
	Pool     ConnectionPool `mapstructure:"pool"`
}

// ConnectionPool 连接池配置
type ConnectionPool struct {
	Min int `mapstructure:"min"`
	Max int `mapstructure:"max"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Enabled  bool           `mapstructure:"enabled"`
	Host     string         `mapstructure:"host"`
	Port     int            `mapstructure:"port"`
	Password string         `mapstructure:"password"`
	DB       int            `mapstructure:"db"`
	Pool     ConnectionPool `mapstructure:"pool"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWT         JWTConfig         `mapstructure:"jwt"`
	APIKey      APIKeyConfig      `mapstructure:"api_key"`
	IPWhitelist IPWhitelistConfig `mapstructure:"ip_whitelist"`
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Expiration int    `mapstructure:"expiration"`
}

// APIKeyConfig API Key 配置
type APIKeyConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// IPWhitelistConfig IP 白名单配置
type IPWhitelistConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	IPs     []string `mapstructure:"ips"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string        `mapstructure:"level"`
	Format string        `mapstructure:"format"`
	Output string        `mapstructure:"output"`
	File   LogFileConfig `mapstructure:"file"`
}

// LogFileConfig 日志文件配置
type LogFileConfig struct {
	Path       string `mapstructure:"path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	LogRetentionDays int             `mapstructure:"log_retention_days"`
	Cache            CacheConfig     `mapstructure:"cache"`
	RateLimit        RateLimitConfig `mapstructure:"rate_limit"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	RuleTTL   int `mapstructure:"rule_ttl"`
	ConfigTTL int `mapstructure:"config_ttl"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled     bool `mapstructure:"enabled"`
	IPLimit     int  `mapstructure:"ip_limit"`
	GlobalLimit int  `mapstructure:"global_limit"`
}

// FeaturesConfig 功能开关
type FeaturesConfig struct {
	VersionControl bool `mapstructure:"version_control"`
	AuditLog       bool `mapstructure:"audit_log"`
	Metrics        bool `mapstructure:"metrics"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
	}

	// 读取环境变量
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// GetAdminAddress 获取管理服务地址
func (c *Config) GetAdminAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Admin.Host, c.Server.Admin.Port)
}

// GetMockAddress 获取 Mock 服务地址
func (c *Config) GetMockAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Mock.Host, c.Server.Mock.Port)
}

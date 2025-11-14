package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoad_ValidConfig 测试加载有效配置文件
func TestLoad_ValidConfig(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `
server:
  admin:
    host: "0.0.0.0"
    port: 8080
  mock:
    host: "0.0.0.0"
    port: 9090

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "mockserver_test"
    timeout: 10s
    pool:
      min: 5
      max: 100

redis:
  enabled: false
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool:
    min: 5
    max: 50

security:
  jwt:
    secret: "test-secret"
    expiration: 3600
  api_key:
    enabled: false
  ip_whitelist:
    enabled: false
    ips: []

logging:
  level: "info"
  format: "json"
  output: "stdout"
  file:
    path: "/var/log/mockserver.log"
    max_size: 100
    max_backups: 3
    max_age: 7

performance:
  log_retention_days: 30
  cache:
    rule_ttl: 300
    config_ttl: 600
  rate_limit:
    enabled: false
    ip_limit: 100
    global_limit: 1000

features:
  version_control: false
  audit_log: false
  metrics: false
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)
	
	// 加载配置
	cfg, err := Load(configPath)
	
	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	
	// 验证服务器配置
	assert.Equal(t, "0.0.0.0", cfg.Server.Admin.Host)
	assert.Equal(t, 8080, cfg.Server.Admin.Port)
	assert.Equal(t, "0.0.0.0", cfg.Server.Mock.Host)
	assert.Equal(t, 9090, cfg.Server.Mock.Port)
	
	// 验证数据库配置
	assert.Equal(t, "mongodb://localhost:27017", cfg.Database.MongoDB.URI)
	assert.Equal(t, "mockserver_test", cfg.Database.MongoDB.Database)
	assert.Equal(t, 5, cfg.Database.MongoDB.Pool.Min)
	assert.Equal(t, 100, cfg.Database.MongoDB.Pool.Max)
	
	// 验证日志配置
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.Equal(t, "stdout", cfg.Logging.Output)
	
	// 验证全局配置已设置
	assert.Equal(t, cfg, Get())
}

// TestLoad_FileNotFound 测试配置文件不存在
func TestLoad_FileNotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/config.yaml")
	
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

// TestLoad_InvalidYAML 测试无效的YAML格式
func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	
	// 写入无效的YAML内容
	invalidContent := `
server:
  admin
    host: "invalid
    port: abc
`
	
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	assert.NoError(t, err)
	
	cfg, err := Load(configPath)
	
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

// TestLoad_WithoutPath 测试不指定路径加载配置
func TestLoad_WithoutPath(t *testing.T) {
	// 在当前目录创建config.yaml
	configContent := `
server:
  admin:
    host: "127.0.0.1"
    port: 8888
  mock:
    host: "127.0.0.1"
    port: 9999

database:
  mongodb:
    uri: "mongodb://test:27017"
    database: "test_db"
    timeout: 5s
    pool:
      min: 1
      max: 10

redis:
  enabled: true
  host: "redis-host"
  port: 6379

logging:
  level: "debug"
  format: "text"
`
	
	// 保存当前工作目录
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	
	// 切换到临时目录
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	
	// 创建配置文件
	err := os.WriteFile("config.yaml", []byte(configContent), 0644)
	assert.NoError(t, err)
	
	// 不指定路径加载
	cfg, err := Load("")
	
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "127.0.0.1", cfg.Server.Admin.Host)
	assert.Equal(t, 8888, cfg.Server.Admin.Port)
	assert.Equal(t, "debug", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
}

// TestGet 测试获取全局配置
func TestGet(t *testing.T) {
	// 先创建一个配置
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `
server:
  admin:
    host: "0.0.0.0"
    port: 7777
  mock:
    host: "0.0.0.0"
    port: 8888

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "test"
    timeout: 10s

logging:
  level: "info"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)
	
	// 加载配置
	cfg, err := Load(configPath)
	assert.NoError(t, err)
	
	// 使用Get获取
	globalCfg := Get()
	assert.NotNil(t, globalCfg)
	assert.Equal(t, cfg, globalCfg)
	assert.Equal(t, 7777, globalCfg.Server.Admin.Port)
}

// TestConfig_GetAdminAddress 测试获取管理服务地址
func TestConfig_GetAdminAddress(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "默认地址",
			config: &Config{
				Server: ServerConfig{
					Admin: AdminServerConfig{
						Host: "0.0.0.0",
						Port: 8080,
					},
				},
			},
			expected: "0.0.0.0:8080",
		},
		{
			name: "自定义地址",
			config: &Config{
				Server: ServerConfig{
					Admin: AdminServerConfig{
						Host: "127.0.0.1",
						Port: 3000,
					},
				},
			},
			expected: "127.0.0.1:3000",
		},
		{
			name: "域名地址",
			config: &Config{
				Server: ServerConfig{
					Admin: AdminServerConfig{
						Host: "api.example.com",
						Port: 443,
					},
				},
			},
			expected: "api.example.com:443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address := tt.config.GetAdminAddress()
			assert.Equal(t, tt.expected, address)
		})
	}
}

// TestConfig_GetMockAddress 测试获取Mock服务地址
func TestConfig_GetMockAddress(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "默认地址",
			config: &Config{
				Server: ServerConfig{
					Mock: MockServerConfig{
						Host: "0.0.0.0",
						Port: 9090,
					},
				},
			},
			expected: "0.0.0.0:9090",
		},
		{
			name: "自定义地址",
			config: &Config{
				Server: ServerConfig{
					Mock: MockServerConfig{
						Host: "localhost",
						Port: 8000,
					},
				},
			},
			expected: "localhost:8000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address := tt.config.GetMockAddress()
			assert.Equal(t, tt.expected, address)
		})
	}
}

// TestLoad_CompleteConfig 测试加载完整配置
func TestLoad_CompleteConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "complete.yaml")
	
	configContent := `
server:
  admin:
    host: "0.0.0.0"
    port: 8080
  mock:
    host: "0.0.0.0"
    port: 9090

database:
  mongodb:
    uri: "mongodb://user:pass@localhost:27017"
    database: "mockserver"
    timeout: 30s
    pool:
      min: 10
      max: 200

redis:
  enabled: true
  host: "redis.example.com"
  port: 6379
  password: "secret"
  db: 2
  pool:
    min: 10
    max: 100

security:
  jwt:
    secret: "super-secret-key"
    expiration: 7200
  api_key:
    enabled: true
  ip_whitelist:
    enabled: true
    ips:
      - "192.168.1.0/24"
      - "10.0.0.1"

logging:
  level: "debug"
  format: "json"
  output: "file"
  file:
    path: "/var/log/mockserver/app.log"
    max_size: 200
    max_backups: 10
    max_age: 30

performance:
  log_retention_days: 90
  cache:
    rule_ttl: 600
    config_ttl: 1200
  rate_limit:
    enabled: true
    ip_limit: 200
    global_limit: 5000

features:
  version_control: true
  audit_log: true
  metrics: true
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)
	
	cfg, err := Load(configPath)
	
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	
	// 验证 Redis 配置
	assert.True(t, cfg.Redis.Enabled)
	assert.Equal(t, "redis.example.com", cfg.Redis.Host)
	assert.Equal(t, 6379, cfg.Redis.Port)
	assert.Equal(t, "secret", cfg.Redis.Password)
	assert.Equal(t, 2, cfg.Redis.DB)
	assert.Equal(t, 10, cfg.Redis.Pool.Min)
	assert.Equal(t, 100, cfg.Redis.Pool.Max)
	
	// 验证安全配置
	assert.Equal(t, "super-secret-key", cfg.Security.JWT.Secret)
	assert.Equal(t, 7200, cfg.Security.JWT.Expiration)
	assert.True(t, cfg.Security.APIKey.Enabled)
	assert.True(t, cfg.Security.IPWhitelist.Enabled)
	assert.Len(t, cfg.Security.IPWhitelist.IPs, 2)
	assert.Contains(t, cfg.Security.IPWhitelist.IPs, "192.168.1.0/24")
	
	// 验证日志文件配置
	assert.Equal(t, "file", cfg.Logging.Output)
	assert.Equal(t, "/var/log/mockserver/app.log", cfg.Logging.File.Path)
	assert.Equal(t, 200, cfg.Logging.File.MaxSize)
	assert.Equal(t, 10, cfg.Logging.File.MaxBackups)
	assert.Equal(t, 30, cfg.Logging.File.MaxAge)
	
	// 验证性能配置
	assert.Equal(t, 90, cfg.Performance.LogRetentionDays)
	assert.Equal(t, 600, cfg.Performance.Cache.RuleTTL)
	assert.Equal(t, 1200, cfg.Performance.Cache.ConfigTTL)
	assert.True(t, cfg.Performance.RateLimit.Enabled)
	assert.Equal(t, 200, cfg.Performance.RateLimit.IPLimit)
	assert.Equal(t, 5000, cfg.Performance.RateLimit.GlobalLimit)
	
	// 验证功能开关
	assert.True(t, cfg.Features.VersionControl)
	assert.True(t, cfg.Features.AuditLog)
	assert.True(t, cfg.Features.Metrics)
}

// TestLoad_MinimalConfig 测试最小配置
func TestLoad_MinimalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")
	
	// 只包含必需的配置项
	configContent := `
server:
  admin:
    host: "localhost"
    port: 8080
  mock:
    host: "localhost"
    port: 9090

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "mockserver"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)
	
	cfg, err := Load(configPath)
	
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "localhost", cfg.Server.Admin.Host)
	assert.Equal(t, 8080, cfg.Server.Admin.Port)
	assert.Equal(t, "mockserver", cfg.Database.MongoDB.Database)
}

// TestLoad_PartialConfig 测试部分配置（验证零值）
func TestLoad_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.yaml")
	
	configContent := `
server:
  admin:
    host: "0.0.0.0"
    port: 8080
  mock:
    host: "0.0.0.0"
    port: 9090

database:
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "test"

# Redis 配置不提供，应该使用零值
# Security 配置不提供
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)
	
	cfg, err := Load(configPath)
	
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	
	// 验证未配置的项使用零值
	assert.False(t, cfg.Redis.Enabled)
	assert.Equal(t, "", cfg.Redis.Host)
	assert.Equal(t, 0, cfg.Redis.Port)
	
	assert.Equal(t, "", cfg.Security.JWT.Secret)
	assert.Equal(t, 0, cfg.Security.JWT.Expiration)
	assert.False(t, cfg.Security.APIKey.Enabled)
	
	assert.Equal(t, "", cfg.Logging.Level)
	assert.Equal(t, "", cfg.Logging.Format)
}

// TestConfigStructure 测试配置结构完整性
func TestConfigStructure(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Admin: AdminServerConfig{Host: "test", Port: 8080},
			Mock:  MockServerConfig{Host: "test", Port: 9090},
		},
		Database: DatabaseConfig{
			MongoDB: MongoDBConfig{
				URI:      "mongodb://test",
				Database: "test",
				Pool:     ConnectionPool{Min: 5, Max: 50},
			},
		},
		Redis: RedisConfig{
			Enabled: true,
			Host:    "redis",
			Port:    6379,
			Pool:    ConnectionPool{Min: 5, Max: 50},
		},
		Security: SecurityConfig{
			JWT:         JWTConfig{Secret: "secret", Expiration: 3600},
			APIKey:      APIKeyConfig{Enabled: true},
			IPWhitelist: IPWhitelistConfig{Enabled: true, IPs: []string{"127.0.0.1"}},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
			File:   LogFileConfig{Path: "/tmp/log", MaxSize: 100},
		},
		Performance: PerformanceConfig{
			LogRetentionDays: 30,
			Cache:            CacheConfig{RuleTTL: 300, ConfigTTL: 600},
			RateLimit:        RateLimitConfig{Enabled: true, IPLimit: 100, GlobalLimit: 1000},
		},
		Features: FeaturesConfig{
			VersionControl: true,
			AuditLog:       true,
			Metrics:        true,
		},
	}
	
	// 验证所有字段都可以正确访问
	assert.Equal(t, "test", cfg.Server.Admin.Host)
	assert.Equal(t, 8080, cfg.Server.Admin.Port)
	assert.Equal(t, "test", cfg.Server.Mock.Host)
	assert.Equal(t, 9090, cfg.Server.Mock.Port)
	
	assert.Equal(t, "mongodb://test", cfg.Database.MongoDB.URI)
	assert.Equal(t, "test", cfg.Database.MongoDB.Database)
	assert.Equal(t, 5, cfg.Database.MongoDB.Pool.Min)
	
	assert.True(t, cfg.Redis.Enabled)
	assert.Equal(t, "redis", cfg.Redis.Host)
	
	assert.Equal(t, "secret", cfg.Security.JWT.Secret)
	assert.True(t, cfg.Security.APIKey.Enabled)
	assert.True(t, cfg.Security.IPWhitelist.Enabled)
	
	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	
	assert.Equal(t, 30, cfg.Performance.LogRetentionDays)
	assert.True(t, cfg.Performance.RateLimit.Enabled)
	
	assert.True(t, cfg.Features.VersionControl)
	assert.True(t, cfg.Features.AuditLog)
	assert.True(t, cfg.Features.Metrics)
}

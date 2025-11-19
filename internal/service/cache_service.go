package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gomockserver/mockserver/internal/cache"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/internal/models"
	"go.uber.org/zap"
)

// CacheService 缓存服务
type CacheService struct {
	manager  cache.Manager
	logger   *zap.Logger
}

// NewCacheService 创建缓存服务
func NewCacheService(
	redisConfig *cache.RedisConfig,
	cacheStrategy *cache.CacheStrategy,
	logger *zap.Logger,
) (*CacheService, error) {
	if cacheStrategy == nil {
		cacheStrategy = cache.DefaultCacheStrategy()
	}

	// 创建L1内存缓存
	l1Cache := cache.NewMemoryL1Cache(
		cacheStrategy.L1MaxEntries,
		100, // 100MB内存限制
		cacheStrategy.L1CleanupInterval,
		logger.Named("l1_cache"),
	)

	// 创建L2 Redis缓存
	l2Cache, err := cache.NewRedisL2Cache(redisConfig, logger.Named("l2_redis"))
	if err != nil {
		return nil, fmt.Errorf("failed to create L2 cache: %w", err)
	}

	// 创建频率跟踪器
	tracker := cache.NewSimpleFrequencyTracker(cacheStrategy.AccessFreqWindow)

	// 创建三级缓存管理器
	manager := cache.NewThreeLevelCacheManager(
		l1Cache,
		l2Cache,
		tracker,
		cacheStrategy,
		logger.Named("cache_manager"),
	)

	service := &CacheService{
		manager: manager,
		logger:  logger.Named("cache_service"),
	}

	logger.Info("Cache service initialized successfully")
	return service, nil
}

// Start 启动缓存服务
func (s *CacheService) Start(ctx context.Context) error {
	s.logger.Info("Starting cache service")
	return s.manager.Start(ctx)
}

// Stop 停止缓存服务
func (s *CacheService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping cache service")
	return s.manager.Stop(ctx)
}

// GetManager 获取缓存管理器
func (s *CacheService) GetManager() cache.Manager {
	return s.manager
}

// CacheRuleService 规则缓存服务
type CacheRuleService struct {
	cacheService *CacheService
	ruleRepo     repository.RuleRepository
	logger       *zap.Logger
}

// NewCacheRuleService 创建规则缓存服务
func NewCacheRuleService(
	cacheService *CacheService,
	ruleRepo repository.RuleRepository,
	logger *zap.Logger,
) *CacheRuleService {
	return &CacheRuleService{
		cacheService: cacheService,
		ruleRepo:     ruleRepo,
		logger:       logger.Named("cache_rule_service"),
	}
}

// GetRulesByProject 获取项目规则（带缓存）
func (s *CacheRuleService) GetRulesByProject(ctx context.Context, projectID string) ([]*models.Rule, error) {
	cacheKey := fmt.Sprintf("rules:project:%s", projectID)

	// 尝试从缓存获取
	if cached, err := s.cacheService.GetManager().Get(ctx, cacheKey); err == nil && cached != nil {
		s.logger.Debug("Cache hit for project rules", zap.String("project_id", projectID))
		if rules, ok := cached.([]*models.Rule); ok {
			return rules, nil
		}
	}

	// 缓存未命中，从数据库获取
	rules, err := s.ruleRepo.FindByEnvironment(ctx, projectID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get rules from database: %w", err)
	}

	// 存入缓存
	if err := s.cacheService.GetManager().Set(ctx, cacheKey, rules, 5*time.Minute); err != nil {
		s.logger.Warn("Failed to cache project rules",
			zap.String("project_id", projectID),
			zap.Error(err),
		)
	}

	s.logger.Debug("Cache miss for project rules, loaded from database",
		zap.String("project_id", projectID),
		zap.Int("rules_count", len(rules)),
	)
	return rules, nil
}

// GetEnabledRulesByEnvironment 获取环境启用规则（带缓存）
func (s *CacheRuleService) GetEnabledRulesByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error) {
	cacheKey := fmt.Sprintf("rules:enabled:%s:%s", projectID, environmentID)

	// 尝试从缓存获取
	if cached, err := s.cacheService.GetManager().Get(ctx, cacheKey); err == nil && cached != nil {
		s.logger.Debug("Cache hit for environment rules",
			zap.String("project_id", projectID),
			zap.String("environment_id", environmentID),
		)
		if rules, ok := cached.([]*models.Rule); ok {
			return rules, nil
		}
	}

	// 缓存未命中，从数据库获取
	rules, err := s.ruleRepo.FindEnabledByEnvironment(ctx, projectID, environmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled rules from database: %w", err)
	}

	// 存入缓存
	if err := s.cacheService.GetManager().Set(ctx, cacheKey, rules, 3*time.Minute); err != nil {
		s.logger.Warn("Failed to cache environment rules",
			zap.String("project_id", projectID),
			zap.String("environment_id", environmentID),
			zap.Error(err),
		)
	}

	s.logger.Debug("Cache miss for environment rules, loaded from database",
		zap.String("project_id", projectID),
		zap.String("environment_id", environmentID),
		zap.Int("rules_count", len(rules)),
	)
	return rules, nil
}

// GetRuleByID 根据ID获取规则（带缓存）
func (s *CacheRuleService) GetRuleByID(ctx context.Context, ruleID string) (*models.Rule, error) {
	cacheKey := fmt.Sprintf("rule:id:%s", ruleID)

	// 尝试从缓存获取
	if cached, err := s.cacheService.GetManager().Get(ctx, cacheKey); err == nil && cached != nil {
		s.logger.Debug("Cache hit for rule", zap.String("rule_id", ruleID))
		if rule, ok := cached.(*models.Rule); ok {
			return rule, nil
		}
	}

	// 缓存未命中，从数据库获取
	rule, err := s.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule from database: %w", err)
	}

	// 存入缓存
	if err := s.cacheService.GetManager().Set(ctx, cacheKey, rule, 10*time.Minute); err != nil {
		s.logger.Warn("Failed to cache rule",
			zap.String("rule_id", ruleID),
			zap.Error(err),
		)
	}

	s.logger.Debug("Cache miss for rule, loaded from database", zap.String("rule_id", ruleID))
	return rule, nil
}

// InvalidateProjectRules 清除项目规则缓存
func (s *CacheRuleService) InvalidateProjectRules(ctx context.Context, projectID string) error {
	// 清除项目级别的所有缓存
	patterns := []string{
		fmt.Sprintf("rules:project:%s", projectID),
		fmt.Sprintf("rules:enabled:%s:", projectID), // 前缀匹配
	}

	for _, pattern := range patterns {
		if err := s.cacheService.GetManager().Delete(ctx, pattern); err != nil {
			s.logger.Warn("Failed to invalidate cache pattern",
				zap.String("pattern", pattern),
				zap.Error(err),
			)
		}
	}

	s.logger.Debug("Project rules cache invalidated", zap.String("project_id", projectID))
	return nil
}

// InvalidateEnvironmentRules 清除环境规则缓存
func (s *CacheRuleService) InvalidateEnvironmentRules(ctx context.Context, projectID, environmentID string) error {
	pattern := fmt.Sprintf("rules:enabled:%s:%s", projectID, environmentID)

	if err := s.cacheService.GetManager().Delete(ctx, pattern); err != nil {
		s.logger.Warn("Failed to invalidate environment rules cache",
			zap.String("project_id", projectID),
			zap.String("environment_id", environmentID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Debug("Environment rules cache invalidated",
		zap.String("project_id", projectID),
		zap.String("environment_id", environmentID),
	)
	return nil
}

// InvalidateRule 清除单个规则缓存
func (s *CacheRuleService) InvalidateRule(ctx context.Context, ruleID string) error {
	pattern := fmt.Sprintf("rule:id:%s", ruleID)

	if err := s.cacheService.GetManager().Delete(ctx, pattern); err != nil {
		s.logger.Warn("Failed to invalidate rule cache",
			zap.String("rule_id", ruleID),
			zap.Error(err),
		)
		return err
	}

	s.logger.Debug("Rule cache invalidated", zap.String("rule_id", ruleID))
	return nil
}

// CacheStatsService 缓存统计服务
type CacheStatsService struct {
	cacheService *CacheService
	logger       *zap.Logger
}

// NewCacheStatsService 创建缓存统计服务
func NewCacheStatsService(
	cacheService *CacheService,
	logger *zap.Logger,
) *CacheStatsService {
	return &CacheStatsService{
		cacheService: cacheService,
		logger:       logger.Named("cache_stats_service"),
	}
}

// GetCacheStats 获取缓存统计信息
func (s *CacheStatsService) GetCacheStats(ctx context.Context) (*cache.CacheStats, error) {
	return s.cacheService.GetManager().GetStats(ctx)
}

// GetCacheStrategy 获取当前缓存策略
func (s *CacheStatsService) GetCacheStrategy() *cache.CacheStrategy {
	return s.cacheService.GetManager().GetStrategy()
}

// UpdateCacheStrategy 更新缓存策略
func (s *CacheStatsService) UpdateCacheStrategy(strategy *cache.CacheStrategy) error {
	err := s.cacheService.GetManager().UpdateStrategy(strategy)
	if err != nil {
		return fmt.Errorf("failed to update cache strategy: %w", err)
	}

	s.logger.Info("Cache strategy updated successfully")
	return nil
}

// ClearCache 清空指定级别的缓存
func (s *CacheStatsService) ClearCache(ctx context.Context, level cache.CacheLevel) error {
	err := s.cacheService.GetManager().Clear(ctx, level)
	if err != nil {
		return fmt.Errorf("failed to clear cache level %v: %w", level, err)
	}

	s.logger.Info("Cache cleared successfully", zap.String("level", cacheLevelString(level)))
	return nil
}

// cacheLevelString 将缓存级别转换为字符串
func cacheLevelString(level cache.CacheLevel) string {
	switch level {
	case cache.L1_HOT:
		return "L1_HOT"
	case cache.L2_WARM:
		return "L2_WARM"
	case cache.L3_COLD:
		return "L3_COLD"
	default:
		return "UNKNOWN"
	}
}

// CacheHealthService 缓存健康检查服务
type CacheHealthService struct {
	cacheService *CacheService
	l2Cache      cache.L2Cache
	logger       *zap.Logger
}

// NewCacheHealthService 创建缓存健康检查服务
func NewCacheHealthService(
	cacheService *CacheService,
	l2Cache cache.L2Cache,
	logger *zap.Logger,
) *CacheHealthService {
	return &CacheHealthService{
		cacheService: cacheService,
		l2Cache:      l2Cache,
		logger:       logger.Named("cache_health_service"),
	}
}

// HealthCheck 执行缓存健康检查
func (s *CacheHealthService) HealthCheck(ctx context.Context) error {
	// 检查L2 Redis连接
	if err := s.l2Cache.Ping(ctx); err != nil {
		s.logger.Error("L2 cache health check failed", zap.Error(err))
		return fmt.Errorf("L2 cache unhealthy: %w", err)
	}

	// 测试缓存读写
	testKey := "health_check_test"
	testValue := fmt.Sprintf("test_value_%d", time.Now().Unix())

	// 写入测试
	if err := s.cacheService.GetManager().Set(ctx, testKey, testValue, 1*time.Minute); err != nil {
		s.logger.Error("Cache write test failed", zap.Error(err))
		return fmt.Errorf("cache write failed: %w", err)
	}

	// 读取测试
	value, err := s.cacheService.GetManager().Get(ctx, testKey)
	if err != nil {
		s.logger.Error("Cache read test failed", zap.Error(err))
		return fmt.Errorf("cache read failed: %w", err)
	}

	if value != testValue {
		err := fmt.Errorf("cache value mismatch: expected %v, got %v", testValue, value)
		s.logger.Error("Cache value verification failed", zap.Error(err))
		return err
	}

	// 清理测试数据
	if err := s.cacheService.GetManager().Delete(ctx, testKey); err != nil {
		s.logger.Warn("Failed to cleanup health check data", zap.Error(err))
	}

	s.logger.Debug("Cache health check passed")
	return nil
}
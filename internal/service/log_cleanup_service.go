package service

import (
	"context"
	"time"

	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// LogCleanupService 日志清理服务
type LogCleanupService struct {
	repo           repository.RequestLogRepository
	retentionDays  int
	cleanupEnabled bool
	stopChan       chan struct{}
}

// NewLogCleanupService 创建日志清理服务
func NewLogCleanupService(repo repository.RequestLogRepository, retentionDays int) *LogCleanupService {
	if retentionDays <= 0 {
		retentionDays = 7 // 默认保留7天
	}

	return &LogCleanupService{
		repo:           repo,
		retentionDays:  retentionDays,
		cleanupEnabled: true,
		stopChan:       make(chan struct{}),
	}
}

// Start 启动定时清理任务
func (s *LogCleanupService) Start() {
	logger.Info("log cleanup service started", zap.Int("retention_days", s.retentionDays))

	// 立即执行一次清理
	s.cleanup()

	// 每天凌晨2点执行清理
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			// 每天凌晨2点执行
			if now.Hour() == 2 {
				s.cleanup()
			}
		case <-s.stopChan:
			logger.Info("log cleanup service stopped")
			return
		}
	}
}

// Stop 停止清理服务
func (s *LogCleanupService) Stop() {
	close(s.stopChan)
}

// cleanup 执行清理
func (s *LogCleanupService) cleanup() {
	if !s.cleanupEnabled {
		return
	}

	logger.Info("starting log cleanup", zap.Int("retention_days", s.retentionDays))

	before := time.Now().AddDate(0, 0, -s.retentionDays)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	count, err := s.repo.DeleteBefore(ctx, before)
	if err != nil {
		logger.Error("failed to cleanup logs", zap.Error(err))
		return
	}

	logger.Info("log cleanup completed",
		zap.Int64("deleted_count", count),
		zap.Time("before", before))
}

// SetEnabled 设置是否启用自动清理
func (s *LogCleanupService) SetEnabled(enabled bool) {
	s.cleanupEnabled = enabled
}

// SetRetentionDays 设置保留天数
func (s *LogCleanupService) SetRetentionDays(days int) {
	if days > 0 {
		s.retentionDays = days
		logger.Info("log retention days updated", zap.Int("retention_days", days))
	}
}

// ManualCleanup 手动触发清理
func (s *LogCleanupService) ManualCleanup(ctx context.Context, beforeDays int) (int64, error) {
	if beforeDays <= 0 {
		beforeDays = s.retentionDays
	}

	before := time.Now().AddDate(0, 0, -beforeDays)
	count, err := s.repo.DeleteBefore(ctx, before)
	if err != nil {
		logger.Error("manual cleanup failed", zap.Error(err))
		return 0, err
	}

	logger.Info("manual cleanup completed",
		zap.Int64("deleted_count", count),
		zap.Int("before_days", beforeDays))

	return count, nil
}

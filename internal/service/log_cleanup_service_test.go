package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRequestLogRepository mock实现
type MockRequestLogRepositoryForCleanup struct {
	mock.Mock
}

func (m *MockRequestLogRepositoryForCleanup) Create(ctx context.Context, log *models.RequestLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockRequestLogRepositoryForCleanup) FindByID(ctx context.Context, id string) (*models.RequestLog, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RequestLog), args.Error(1)
}

func (m *MockRequestLogRepositoryForCleanup) List(ctx context.Context, filter repository.RequestLogFilter) ([]*models.RequestLog, int64, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*models.RequestLog), args.Get(1).(int64), args.Error(2)
}

func (m *MockRequestLogRepositoryForCleanup) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	args := m.Called(ctx, before)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequestLogRepositoryForCleanup) DeleteByProjectID(ctx context.Context, projectID string) error {
	args := m.Called(ctx, projectID)
	return args.Error(0)
}

func (m *MockRequestLogRepositoryForCleanup) CountByProjectID(ctx context.Context, projectID string, startTime, endTime time.Time) (int64, error) {
	args := m.Called(ctx, projectID, startTime, endTime)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRequestLogRepositoryForCleanup) GetStatistics(ctx context.Context, projectID, environmentID string, startTime, endTime time.Time) (*repository.RequestLogStatistics, error) {
	args := m.Called(ctx, projectID, environmentID, startTime, endTime)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.RequestLogStatistics), args.Error(1)
}

func TestNewLogCleanupService(t *testing.T) {
	mockRepo := new(MockRequestLogRepositoryForCleanup)

	t.Run("With valid retention days", func(t *testing.T) {
		service := NewLogCleanupService(mockRepo, 10)
		assert.NotNil(t, service)
		assert.Equal(t, 10, service.retentionDays)
		assert.True(t, service.cleanupEnabled)
		assert.NotNil(t, service.stopChan)
	})

	t.Run("With zero retention days", func(t *testing.T) {
		service := NewLogCleanupService(mockRepo, 0)
		assert.NotNil(t, service)
		assert.Equal(t, 7, service.retentionDays) // 默认7天
	})

	t.Run("With negative retention days", func(t *testing.T) {
		service := NewLogCleanupService(mockRepo, -5)
		assert.NotNil(t, service)
		assert.Equal(t, 7, service.retentionDays) // 默认7天
	})
}

func TestLogCleanupService_SetEnabled(t *testing.T) {
	mockRepo := new(MockRequestLogRepositoryForCleanup)
	service := NewLogCleanupService(mockRepo, 7)

	// 测试启用
	service.SetEnabled(true)
	assert.True(t, service.cleanupEnabled)

	// 测试禁用
	service.SetEnabled(false)
	assert.False(t, service.cleanupEnabled)
}

func TestLogCleanupService_SetRetentionDays(t *testing.T) {
	mockRepo := new(MockRequestLogRepositoryForCleanup)
	service := NewLogCleanupService(mockRepo, 7)

	t.Run("Set valid retention days", func(t *testing.T) {
		service.SetRetentionDays(15)
		assert.Equal(t, 15, service.retentionDays)
	})

	t.Run("Set zero retention days", func(t *testing.T) {
		originalDays := service.retentionDays
		service.SetRetentionDays(0)
		assert.Equal(t, originalDays, service.retentionDays) // 不应该改变
	})

	t.Run("Set negative retention days", func(t *testing.T) {
		originalDays := service.retentionDays
		service.SetRetentionDays(-5)
		assert.Equal(t, originalDays, service.retentionDays) // 不应该改变
	})
}

func TestLogCleanupService_ManualCleanup(t *testing.T) {
	mockRepo := new(MockRequestLogRepositoryForCleanup)
	service := NewLogCleanupService(mockRepo, 7)

	t.Run("Successful cleanup", func(t *testing.T) {
		ctx := context.Background()
		mockRepo.On("DeleteBefore", ctx, mock.AnythingOfType("time.Time")).
			Return(int64(100), nil).Once()

		count, err := service.ManualCleanup(ctx, 10)

		assert.NoError(t, err)
		assert.Equal(t, int64(100), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cleanup with error", func(t *testing.T) {
		ctx := context.Background()
		mockRepo.On("DeleteBefore", ctx, mock.AnythingOfType("time.Time")).
			Return(int64(0), errors.New("database error")).Once()

		count, err := service.ManualCleanup(ctx, 10)

		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		assert.EqualError(t, err, "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cleanup with default retention days", func(t *testing.T) {
		ctx := context.Background()
		mockRepo.On("DeleteBefore", ctx, mock.AnythingOfType("time.Time")).
			Return(int64(50), nil).Once()

		// 传入0或负数应该使用默认的retentionDays
		count, err := service.ManualCleanup(ctx, 0)

		assert.NoError(t, err)
		assert.Equal(t, int64(50), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cleanup with negative days", func(t *testing.T) {
		ctx := context.Background()
		mockRepo.On("DeleteBefore", ctx, mock.AnythingOfType("time.Time")).
			Return(int64(25), nil).Once()

		count, err := service.ManualCleanup(ctx, -5)

		assert.NoError(t, err)
		assert.Equal(t, int64(25), count)
		mockRepo.AssertExpectations(t)
	})
}

func TestLogCleanupService_Cleanup(t *testing.T) {
	t.Run("Cleanup when enabled", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepositoryForCleanup)
		service := NewLogCleanupService(mockRepo, 7)

		mockRepo.On("DeleteBefore", mock.Anything, mock.AnythingOfType("time.Time")).
			Return(int64(100), nil).Once()

		service.cleanup()

		mockRepo.AssertExpectations(t)
	})

	t.Run("Cleanup when disabled", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepositoryForCleanup)
		service := NewLogCleanupService(mockRepo, 7)
		service.SetEnabled(false)

		// 禁用时不应该调用DeleteBefore
		service.cleanup()

		// 验证没有调用
		mockRepo.AssertNotCalled(t, "DeleteBefore")
	})

	t.Run("Cleanup with error", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepositoryForCleanup)
		service := NewLogCleanupService(mockRepo, 7)

		mockRepo.On("DeleteBefore", mock.Anything, mock.AnythingOfType("time.Time")).
			Return(int64(0), errors.New("cleanup failed")).Once()

		// 应该不会panic，只是记录错误
		assert.NotPanics(t, func() {
			service.cleanup()
		})

		mockRepo.AssertExpectations(t)
	})
}

func TestLogCleanupService_Stop(t *testing.T) {
	mockRepo := new(MockRequestLogRepositoryForCleanup)
	service := NewLogCleanupService(mockRepo, 7)

	// 测试Stop不会panic
	assert.NotPanics(t, func() {
		service.Stop()
	})

	// 验证stopChan已关闭
	select {
	case <-service.stopChan:
		// 通道已关闭，测试通过
	case <-time.After(100 * time.Millisecond):
		t.Error("stopChan should be closed")
	}
}

// TestLogCleanupService_Start 测试Start方法
func TestLogCleanupService_Start(t *testing.T) {
	t.Run("Start and stop service", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepositoryForCleanup)
		service := NewLogCleanupService(mockRepo, 7)

		// Mock立即执行的清理
		mockRepo.On("DeleteBefore", mock.Anything, mock.AnythingOfType("time.Time")).
			Return(int64(10), nil).Once()

		// 在goroutine中启动服务
		go func() {
			service.Start()
		}()

		// 等待一小段时间确保启动
		time.Sleep(200 * time.Millisecond)

		// 停止服务
		service.Stop()

		// 等待服务完全停止
		time.Sleep(100 * time.Millisecond)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Start with cleanup disabled", func(t *testing.T) {
		mockRepo := new(MockRequestLogRepositoryForCleanup)
		service := NewLogCleanupService(mockRepo, 7)
		service.SetEnabled(false)

		// 禁用清理时不应该调用DeleteBefore
		// 在goroutine中启动服务
		go func() {
			service.Start()
		}()

		// 等待一小段时间
		time.Sleep(200 * time.Millisecond)

		// 停止服务
		service.Stop()

		// 验证没有调用DeleteBefore
		mockRepo.AssertNotCalled(t, "DeleteBefore")
	})
}

func TestLogCleanupService_Integration(t *testing.T) {
	mockRepo := new(MockRequestLogRepositoryForCleanup)
	service := NewLogCleanupService(mockRepo, 10)

	t.Run("Complete workflow", func(t *testing.T) {
		// 1. 设置保留天数
		service.SetRetentionDays(5)
		assert.Equal(t, 5, service.retentionDays)

		// 2. 禁用自动清理
		service.SetEnabled(false)
		assert.False(t, service.cleanupEnabled)

		// 3. 执行手动清理
		ctx := context.Background()
		mockRepo.On("DeleteBefore", ctx, mock.AnythingOfType("time.Time")).
			Return(int64(200), nil).Once()

		count, err := service.ManualCleanup(ctx, 3)
		assert.NoError(t, err)
		assert.Equal(t, int64(200), count)

		// 4. 重新启用
		service.SetEnabled(true)
		assert.True(t, service.cleanupEnabled)

		mockRepo.AssertExpectations(t)
	})
}

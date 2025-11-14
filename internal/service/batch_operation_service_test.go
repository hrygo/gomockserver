package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockRuleRepository Mock 规则仓库
type MockBatchRuleRepository struct {
	mock.Mock
}

func (m *MockBatchRuleRepository) Create(ctx context.Context, rule *models.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockBatchRuleRepository) Update(ctx context.Context, rule *models.Rule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockBatchRuleRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBatchRuleRepository) FindByID(ctx context.Context, id string) (*models.Rule, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Rule), args.Error(1)
}

func (m *MockBatchRuleRepository) FindByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error) {
	args := m.Called(ctx, projectID, environmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Rule), args.Error(1)
}

func (m *MockBatchRuleRepository) FindEnabledByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error) {
	args := m.Called(ctx, projectID, environmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Rule), args.Error(1)
}

func (m *MockBatchRuleRepository) List(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]*models.Rule, int64, error) {
	args := m.Called(ctx, filter, skip, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Rule), args.Get(1).(int64), args.Error(2)
}

// 测试辅助函数
func setupBatchService() (*batchOperationService, *MockBatchRuleRepository, *zap.Logger) {
	mockRepo := new(MockBatchRuleRepository)
	logger := zap.NewNop()
	service := &batchOperationService{
		ruleRepo: mockRepo,
		logger:   logger,
	}
	return service, mockRepo, logger
}

func createTestRule(id string, enabled bool) *models.Rule {
	return &models.Rule{
		ID:            id,
		Name:          "Test Rule " + id,
		ProjectID:     "test-project",
		EnvironmentID: "test-env",
		Enabled:       enabled,
		Priority:      100,
		Tags:          []string{"test"},
	}
}

// TestBatchEnable 测试批量启用
func TestBatchEnable(t *testing.T) {
	service, mockRepo, _ := setupBatchService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		ruleIDs := []string{"rule1", "rule2"}
		
		for _, id := range ruleIDs {
			rule := createTestRule(id, false)
			mockRepo.On("FindByID", ctx, id).Return(rule, nil).Once()
			mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
				return r.ID == id && r.Enabled == true
			})).Return(nil).Once()
		}

		result, err := service.BatchEnable(ctx, ruleIDs)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.TotalCount)
		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 0, result.FailedCount)
		assert.True(t, result.Success)
		assert.Empty(t, result.FailedIDs)
		assert.Empty(t, result.Errors)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Partial Success", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1", "rule2", "rule3"}
		
		// rule1 成功
		rule1 := createTestRule("rule1", false)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule1, nil).Once()
		mockRepo.On("Update", ctx, mock.Anything).Return(nil).Once()
		
		// rule2 查找失败
		mockRepo.On("FindByID", ctx, "rule2").Return(nil, errors.New("database error")).Once()
		
		// rule3 更新失败
		rule3 := createTestRule("rule3", false)
		mockRepo.On("FindByID", ctx, "rule3").Return(rule3, nil).Once()
		mockRepo.On("Update", ctx, mock.Anything).Return(errors.New("update failed")).Once()

		result, err := service.BatchEnable(ctx, ruleIDs)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.TotalCount)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 2, result.FailedCount)
		assert.False(t, result.Success)
		assert.Equal(t, []string{"rule2", "rule3"}, result.FailedIDs)
		assert.Len(t, result.Errors, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Rule Not Found", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"nonexistent"}
		mockRepo.On("FindByID", ctx, "nonexistent").Return(nil, nil).Once()

		result, err := service.BatchEnable(ctx, ruleIDs)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.TotalCount)
		assert.Equal(t, 0, result.SuccessCount)
		assert.Equal(t, 1, result.FailedCount)
		assert.False(t, result.Success)
		assert.Contains(t, result.FailedIDs, "nonexistent")
		mockRepo.AssertExpectations(t)
	})
}

// TestBatchDisable 测试批量禁用
func TestBatchDisable(t *testing.T) {
	service, mockRepo, _ := setupBatchService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		ruleIDs := []string{"rule1", "rule2"}
		
		for _, id := range ruleIDs {
			rule := createTestRule(id, true)
			mockRepo.On("FindByID", ctx, id).Return(rule, nil).Once()
			mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
				return r.ID == id && r.Enabled == false
			})).Return(nil).Once()
		}

		result, err := service.BatchDisable(ctx, ruleIDs)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.TotalCount)
		assert.Equal(t, 2, result.SuccessCount)
		assert.Equal(t, 0, result.FailedCount)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})
}

// TestBatchDelete 测试批量删除
func TestBatchDelete(t *testing.T) {
	service, mockRepo, _ := setupBatchService()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		ruleIDs := []string{"rule1", "rule2", "rule3"}
		
		for _, id := range ruleIDs {
			mockRepo.On("Delete", ctx, id).Return(nil).Once()
		}

		result, err := service.BatchDelete(ctx, ruleIDs)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.TotalCount)
		assert.Equal(t, 3, result.SuccessCount)
		assert.Equal(t, 0, result.FailedCount)
		assert.True(t, result.Success)
		assert.Empty(t, result.FailedIDs)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Partial Failure", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1", "rule2"}
		
		mockRepo.On("Delete", ctx, "rule1").Return(nil).Once()
		mockRepo.On("Delete", ctx, "rule2").Return(errors.New("delete failed")).Once()

		result, err := service.BatchDelete(ctx, ruleIDs)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.TotalCount)
		assert.Equal(t, 1, result.SuccessCount)
		assert.Equal(t, 1, result.FailedCount)
		assert.False(t, result.Success)
		assert.Contains(t, result.FailedIDs, "rule2")
		mockRepo.AssertExpectations(t)
	})
}

// TestBatchUpdate 测试批量更新
func TestBatchUpdate(t *testing.T) {
	service, mockRepo, _ := setupBatchService()
	ctx := context.Background()

	t.Run("Update Priority - Success", func(t *testing.T) {
		ruleIDs := []string{"rule1", "rule2"}
		updates := map[string]interface{}{
			"priority": 200,
		}
		
		for _, id := range ruleIDs {
			rule := createTestRule(id, true)
			mockRepo.On("FindByID", ctx, id).Return(rule, nil).Once()
			mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
				return r.ID == id && r.Priority == 200
			})).Return(nil).Once()
		}

		result, err := service.BatchUpdate(ctx, ruleIDs, updates)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.TotalCount)
		assert.Equal(t, 2, result.SuccessCount)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Priority - Float64", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1"}
		updates := map[string]interface{}{
			"priority": float64(150),
		}
		
		rule := createTestRule("rule1", true)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return r.Priority == 150
		})).Return(nil).Once()

		result, err := service.BatchUpdate(ctx, ruleIDs, updates)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Tags - Success", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1"}
		updates := map[string]interface{}{
			"tags": []string{"prod", "important"},
		}
		
		rule := createTestRule("rule1", true)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return len(r.Tags) == 2 && r.Tags[0] == "prod" && r.Tags[1] == "important"
		})).Return(nil).Once()

		result, err := service.BatchUpdate(ctx, ruleIDs, updates)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Tags - Interface Array", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1"}
		updates := map[string]interface{}{
			"tags": []interface{}{"tag1", "tag2"},
		}
		
		rule := createTestRule("rule1", true)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return len(r.Tags) == 2
		})).Return(nil).Once()

		result, err := service.BatchUpdate(ctx, ruleIDs, updates)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Enabled - Success", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1"}
		updates := map[string]interface{}{
			"enabled": false,
		}
		
		rule := createTestRule("rule1", true)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return r.Enabled == false
		})).Return(nil).Once()

		result, err := service.BatchUpdate(ctx, ruleIDs, updates)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Field", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1"}
		updates := map[string]interface{}{
			"invalid_field": "value",
		}

		result, err := service.BatchUpdate(ctx, ruleIDs, updates)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("Multiple Fields Update", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		ruleIDs := []string{"rule1"}
		updates := map[string]interface{}{
			"priority": 300,
			"enabled":  false,
			"tags":     []string{"updated"},
		}
		
		rule := createTestRule("rule1", true)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return r.Priority == 300 && r.Enabled == false && len(r.Tags) == 1
		})).Return(nil).Once()

		result, err := service.BatchUpdate(ctx, ruleIDs, updates)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})
}

// TestExecuteBatchOperation 测试统一入口
func TestExecuteBatchOperation(t *testing.T) {
	service, mockRepo, _ := setupBatchService()
	ctx := context.Background()

	t.Run("Empty RuleIDs", func(t *testing.T) {
		req := &models.BatchOperationRequest{
			Operation: "enable",
			RuleIDs:   []string{},
		}

		result, err := ExecuteBatchOperation(ctx, service, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Batch operation input is empty")
	})

	t.Run("Enable Operation", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		req := &models.BatchOperationRequest{
			Operation: "enable",
			RuleIDs:   []string{"rule1"},
		}
		
		rule := createTestRule("rule1", false)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

		result, err := ExecuteBatchOperation(ctx, service, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Disable Operation", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		req := &models.BatchOperationRequest{
			Operation: "disable",
			RuleIDs:   []string{"rule1"},
		}
		
		rule := createTestRule("rule1", true)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

		result, err := ExecuteBatchOperation(ctx, service, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Delete Operation", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		req := &models.BatchOperationRequest{
			Operation: "delete",
			RuleIDs:   []string{"rule1"},
		}
		
		mockRepo.On("Delete", ctx, "rule1").Return(nil).Once()

		result, err := ExecuteBatchOperation(ctx, service, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Operation", func(t *testing.T) {
		mockRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRepo
		
		req := &models.BatchOperationRequest{
			Operation: "update",
			RuleIDs:   []string{"rule1"},
			Updates:   map[string]interface{}{"priority": 100},
		}
		
		rule := createTestRule("rule1", true)
		mockRepo.On("FindByID", ctx, "rule1").Return(rule, nil).Once()
		mockRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

		result, err := ExecuteBatchOperation(ctx, service, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update Without Updates", func(t *testing.T) {
		req := &models.BatchOperationRequest{
			Operation: "update",
			RuleIDs:   []string{"rule1"},
			Updates:   nil,
		}

		result, err := ExecuteBatchOperation(ctx, service, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "updates cannot be empty")
	})

	t.Run("Unknown Operation", func(t *testing.T) {
		req := &models.BatchOperationRequest{
			Operation: "unknown",
			RuleIDs:   []string{"rule1"},
		}

		result, err := ExecuteBatchOperation(ctx, service, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unknown operation")
	})
}

// TestHandleError 测试错误处理
func TestHandleError(t *testing.T) {
	service, _, _ := setupBatchService()

	result := &models.BatchOperationResult{
		TotalCount:   3,
		SuccessCount: 0,
		FailedCount:  0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	service.handleError(result, "rule1", "test error 1")
	service.handleError(result, "rule2", "test error 2")

	assert.Equal(t, 2, result.FailedCount)
	assert.Equal(t, []string{"rule1", "rule2"}, result.FailedIDs)
	assert.Equal(t, []string{"test error 1", "test error 2"}, result.Errors)
}

// TestNewBatchOperationService 测试服务创建
func TestNewBatchOperationService(t *testing.T) {
	mockRepo := new(MockBatchRuleRepository)
	logger := zap.NewNop()

	service := NewBatchOperationService(mockRepo, logger)

	assert.NotNil(t, service)
	assert.Implements(t, (*BatchOperationService)(nil), service)
}

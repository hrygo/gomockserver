package service

import (
	"context"
	"fmt"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"go.uber.org/zap"
)

// BatchOperationService 批量操作服务接口
type BatchOperationService interface {
	// BatchEnable 批量启用规则
	BatchEnable(ctx context.Context, ruleIDs []string) (*models.BatchOperationResult, error)
	// BatchDisable 批量禁用规则
	BatchDisable(ctx context.Context, ruleIDs []string) (*models.BatchOperationResult, error)
	// BatchDelete 批量删除规则
	BatchDelete(ctx context.Context, ruleIDs []string) (*models.BatchOperationResult, error)
	// BatchUpdate 批量更新规则
	BatchUpdate(ctx context.Context, ruleIDs []string, updates map[string]interface{}) (*models.BatchOperationResult, error)
}

type batchOperationService struct {
	ruleRepo repository.RuleRepository
	logger   *zap.Logger
}

// NewBatchOperationService 创建批量操作服务
func NewBatchOperationService(ruleRepo repository.RuleRepository, logger *zap.Logger) BatchOperationService {
	return &batchOperationService{
		ruleRepo: ruleRepo,
		logger:   logger,
	}
}

// BatchEnable 批量启用规则
func (s *batchOperationService) BatchEnable(ctx context.Context, ruleIDs []string) (*models.BatchOperationResult, error) {
	return s.batchUpdateEnabled(ctx, ruleIDs, true)
}

// BatchDisable 批量禁用规则
func (s *batchOperationService) BatchDisable(ctx context.Context, ruleIDs []string) (*models.BatchOperationResult, error) {
	return s.batchUpdateEnabled(ctx, ruleIDs, false)
}

// batchUpdateEnabled 批量更新启用状态
func (s *batchOperationService) batchUpdateEnabled(ctx context.Context, ruleIDs []string, enabled bool) (*models.BatchOperationResult, error) {
	result := &models.BatchOperationResult{
		TotalCount:   len(ruleIDs),
		SuccessCount: 0,
		FailedCount:  0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, ruleID := range ruleIDs {
		rule, err := s.ruleRepo.FindByID(ctx, ruleID)
		if err != nil {
			s.handleError(result, ruleID, models.ErrBatchOperationFailed.Message)
			continue
		}
		if rule == nil {
			s.handleError(result, ruleID, models.ErrRuleNotFound.Message)
			continue
		}

		rule.Enabled = enabled
		if err := s.ruleRepo.Update(ctx, rule); err != nil {
			s.handleError(result, ruleID, models.ErrBatchOperationFailed.Message)
			continue
		}

		result.SuccessCount++
		s.logger.Info("Updated rule enabled status",
			zap.String("rule_id", ruleID),
			zap.Bool("enabled", enabled),
		)
	}

	result.Success = result.FailedCount == 0
	return result, nil
}

// BatchDelete 批量删除规则
func (s *batchOperationService) BatchDelete(ctx context.Context, ruleIDs []string) (*models.BatchOperationResult, error) {
	result := &models.BatchOperationResult{
		TotalCount:   len(ruleIDs),
		SuccessCount: 0,
		FailedCount:  0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

	for _, ruleID := range ruleIDs {
		if err := s.ruleRepo.Delete(ctx, ruleID); err != nil {
			s.handleError(result, ruleID, models.ErrBatchOperationFailed.Message)
			continue
		}

		result.SuccessCount++
		s.logger.Info("Deleted rule", zap.String("rule_id", ruleID))
	}

	result.Success = result.FailedCount == 0
	return result, nil
}

// BatchUpdate 批量更新规则
func (s *batchOperationService) BatchUpdate(ctx context.Context, ruleIDs []string, updates map[string]interface{}) (*models.BatchOperationResult, error) {
	result := &models.BatchOperationResult{
		TotalCount:   len(ruleIDs),
		SuccessCount: 0,
		FailedCount:  0,
		FailedIDs:    []string{},
		Errors:       []string{},
	}

		// 验证更新字段
		allowedFields := map[string]bool{
			"priority": true,
			"tags":     true,
			"enabled":  true,
		}

		for field := range updates {
			if !allowedFields[field] {
				return nil, fmt.Errorf("%s: field '%s' is not allowed for batch update", models.ErrBatchInvalidInput.Message, field)
			}
		}

	for _, ruleID := range ruleIDs {
		rule, err := s.ruleRepo.FindByID(ctx, ruleID)
		if err != nil {
			s.handleError(result, ruleID, models.ErrBatchOperationFailed.Message)
			continue
		}
		if rule == nil {
			s.handleError(result, ruleID, models.ErrRuleNotFound.Message)
			continue
		}

		// 应用更新
		if priority, ok := updates["priority"].(int); ok {
			rule.Priority = priority
		}
		if priority, ok := updates["priority"].(float64); ok {
			rule.Priority = int(priority)
		}
		if tags, ok := updates["tags"].([]string); ok {
			rule.Tags = tags
		}
		if tagsInterface, ok := updates["tags"].([]interface{}); ok {
			tags := make([]string, len(tagsInterface))
			for i, t := range tagsInterface {
				if str, ok := t.(string); ok {
					tags[i] = str
				}
			}
			rule.Tags = tags
		}
		if enabled, ok := updates["enabled"].(bool); ok {
			rule.Enabled = enabled
		}

		if err := s.ruleRepo.Update(ctx, rule); err != nil {
			s.handleError(result, ruleID, models.ErrBatchOperationFailed.Message)
			continue
		}

		result.SuccessCount++
		s.logger.Info("Batch updated rule", zap.String("rule_id", ruleID))
	}

	result.Success = result.FailedCount == 0
	return result, nil
}

// handleError 处理错误
func (s *batchOperationService) handleError(result *models.BatchOperationResult, ruleID string, errorMsg string) {
	result.FailedCount++
	result.FailedIDs = append(result.FailedIDs, ruleID)
	result.Errors = append(result.Errors, errorMsg)
	s.logger.Error("Batch operation failed", zap.String("rule_id", ruleID), zap.String("error", errorMsg))
}

// ExecuteBatchOperation 执行批量操作（统一入口）
func ExecuteBatchOperation(
	ctx context.Context,
	service BatchOperationService,
	req *models.BatchOperationRequest,
) (*models.BatchOperationResult, error) {
	if len(req.RuleIDs) == 0 {
		return nil, models.ErrBatchEmptyInput
	}

	switch req.Operation {
	case "enable":
		return service.BatchEnable(ctx, req.RuleIDs)
	case "disable":
		return service.BatchDisable(ctx, req.RuleIDs)
	case "delete":
		return service.BatchDelete(ctx, req.RuleIDs)
		case "update":
			if req.Updates == nil || len(req.Updates) == 0 {
				return nil, fmt.Errorf("%s: updates cannot be empty for update operation", models.ErrBatchInvalidInput.Message)
			}
			return service.BatchUpdate(ctx, req.RuleIDs, req.Updates)
	default:
		return nil, fmt.Errorf("%s: unknown operation: %s", models.ErrBatchInvalidInput.Message, req.Operation)
	}
}

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/internal/repository"
	"go.uber.org/zap"
)

// ImportExportService 导入导出服务接口
type ImportExportService interface {
	// ExportRules 导出规则
	ExportRules(ctx context.Context, req *models.ExportRequest) (*models.ExportData, error)
	// ExportProject 导出整个项目
	ExportProject(ctx context.Context, projectID string, includeMetadata bool) (*models.ExportData, error)
	// ImportData 导入数据
	ImportData(ctx context.Context, req *models.ImportRequest) (*models.ImportResult, error)
	// ValidateImportData 验证导入数据
	ValidateImportData(ctx context.Context, data *models.ExportData) error
}

type importExportService struct {
	ruleRepo    repository.RuleRepository
	projectRepo repository.ProjectRepository
	envRepo     repository.EnvironmentRepository
	logger      *zap.Logger
}

// NewImportExportService 创建导入导出服务
func NewImportExportService(
	ruleRepo repository.RuleRepository,
	projectRepo repository.ProjectRepository,
	envRepo repository.EnvironmentRepository,
	logger *zap.Logger,
) ImportExportService {
	return &importExportService{
		ruleRepo:    ruleRepo,
		projectRepo: projectRepo,
		envRepo:     envRepo,
		logger:      logger,
	}
}

// ExportRules 导出规则
func (s *importExportService) ExportRules(ctx context.Context, req *models.ExportRequest) (*models.ExportData, error) {
	exportData := &models.ExportData{
		Version:    "1.0",
		ExportTime: time.Now(),
		Data: models.ExportDataContent{
			Rules: []models.RuleExportData{},
		},
	}

	// 根据请求参数确定导出类型
	if req.IncludeProject {
		exportData.ExportType = models.ExportTypeProject
	} else if req.IncludeEnvs {
		exportData.ExportType = models.ExportTypeEnvironment
	} else {
		exportData.ExportType = models.ExportTypeRules
	}

	// 导出项目信息
	if req.IncludeProject && req.ProjectID != "" {
		project, err := s.projectRepo.FindByID(ctx, req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to find project: %w", err)
		}
		if project != nil {
			exportData.Data.Project = &models.ProjectExportData{
				Name:        project.Name,
				WorkspaceID: project.WorkspaceID,
				Description: project.Description,
			}
		}
	}

	// 导出环境信息
	envMap := make(map[string]string) // envID -> envName
	if req.IncludeEnvs && req.ProjectID != "" {
		envs, err := s.envRepo.FindByProject(ctx, req.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to find environments: %w", err)
		}
		for _, env := range envs {
			envMap[env.ID] = env.Name
			exportData.Data.Environments = append(exportData.Data.Environments, models.EnvironmentExportData{
				Name:      env.Name,
				BaseURL:   env.BaseURL,
				Variables: env.Variables,
			})
		}
	}

	// 导出规则
	var rules []*models.Rule
	var err error

	if len(req.RuleIDs) > 0 {
		// 按规则ID列表导出
		for _, ruleID := range req.RuleIDs {
			rule, err := s.ruleRepo.FindByID(ctx, ruleID)
			if err != nil {
				s.logger.Error("Failed to find rule", zap.String("rule_id", ruleID), zap.Error(err))
				continue
			}
			if rule != nil {
				rules = append(rules, rule)
			}
		}
	} else if req.EnvironmentID != "" && req.ProjectID != "" {
		// 按环境导出
		rules, err = s.ruleRepo.FindByEnvironment(ctx, req.ProjectID, req.EnvironmentID)
		if err != nil {
			return nil, fmt.Errorf("failed to find rules by environment: %w", err)
		}
	} else if req.ProjectID != "" {
		// 按项目导出所有规则
		filter := map[string]interface{}{"project_id": req.ProjectID}
		rules, _, err = s.ruleRepo.List(ctx, filter, 0, 10000) // 最多导出10000条
		if err != nil {
			return nil, fmt.Errorf("failed to find rules by project: %w", err)
		}
	}

	// 转换规则为导出格式
	for _, rule := range rules {
		ruleData := models.RuleExportData{
			Name:           rule.Name,
			EnvironmentID:  rule.EnvironmentID,
			Protocol:       rule.Protocol,
			MatchType:      rule.MatchType,
			Priority:       rule.Priority,
			Enabled:        rule.Enabled,
			MatchCondition: rule.MatchCondition,
			Response:       rule.Response,
			Tags:           rule.Tags,
		}

		// 添加环境名称
		if envName, ok := envMap[rule.EnvironmentID]; ok {
			ruleData.EnvironmentName = envName
		}

		exportData.Data.Rules = append(exportData.Data.Rules, ruleData)
	}

	// 添加元数据
	if req.IncludeMetadata {
		exportData.Metadata = &models.ExportMetadata{
			Comment: fmt.Sprintf("Exported %d rules", len(exportData.Data.Rules)),
		}
	}

	return exportData, nil
}

// ExportProject 导出整个项目
func (s *importExportService) ExportProject(ctx context.Context, projectID string, includeMetadata bool) (*models.ExportData, error) {
	req := &models.ExportRequest{
		ProjectID:       projectID,
		IncludeProject:  true,
		IncludeEnvs:     true,
		IncludeMetadata: includeMetadata,
	}
	return s.ExportRules(ctx, req)
}

// ValidateImportData 验证导入数据
func (s *importExportService) ValidateImportData(ctx context.Context, data *models.ExportData) error {
	// 检查版本
	if data.Version != "1.0" {
		return fmt.Errorf("unsupported data version: %s", data.Version)
	}

	// 检查导出类型
	validTypes := map[models.ExportType]bool{
		models.ExportTypeRules:       true,
		models.ExportTypeEnvironment: true,
		models.ExportTypeProject:     true,
	}
	if !validTypes[data.ExportType] {
		return fmt.Errorf("invalid export type: %s", data.ExportType)
	}

	// 检查规则数据
	if len(data.Data.Rules) == 0 {
		return errors.New("no rules to import")
	}

	// 验证每条规则的必填字段
	for i, rule := range data.Data.Rules {
		if rule.Name == "" {
			return fmt.Errorf("rule %d: name is required", i)
		}
		if rule.Protocol == "" {
			return fmt.Errorf("rule %d (%s): protocol is required", i, rule.Name)
		}
		if rule.MatchType == "" {
			return fmt.Errorf("rule %d (%s): match_type is required", i, rule.Name)
		}
		if rule.MatchCondition == nil {
			return fmt.Errorf("rule %d (%s): match_condition is required", i, rule.Name)
		}
	}

	return nil
}

// ImportData 导入数据
func (s *importExportService) ImportData(ctx context.Context, req *models.ImportRequest) (*models.ImportResult, error) {
	result := &models.ImportResult{
		Success:        true,
		EnvironmentIDs: make(map[string]string),
		Errors:         []models.ImportError{},
	}

	// 验证数据
	if err := s.ValidateImportData(ctx, &req.Data); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 处理项目
	projectID := req.TargetProjectID
	if req.CreateProject && req.Data.Data.Project != nil {
		// 创建新项目
		project := &models.Project{
			Name:        req.Data.Data.Project.Name,
			WorkspaceID: req.Data.Data.Project.WorkspaceID,
			Description: req.Data.Data.Project.Description,
		}
		if err := s.projectRepo.Create(ctx, project); err != nil {
			return nil, fmt.Errorf("failed to create project: %w", err)
		}
		projectID = project.ID
		result.ProjectID = projectID
		s.logger.Info("Created project", zap.String("project_id", projectID), zap.String("name", project.Name))
	} else if projectID == "" && req.Data.Data.Project == nil {
		return nil, errors.New("project_id is required or create_project must be true")
	}

	// 处理环境
	envNameToID := make(map[string]string)
	if req.CreateEnvs && len(req.Data.Data.Environments) > 0 {
		for _, envData := range req.Data.Data.Environments {
			env := &models.Environment{
				Name:      envData.Name,
				ProjectID: projectID,
				BaseURL:   envData.BaseURL,
				Variables: envData.Variables,
			}
			if err := s.envRepo.Create(ctx, env); err != nil {
				s.logger.Error("Failed to create environment", zap.String("name", envData.Name), zap.Error(err))
				continue
			}
			envNameToID[envData.Name] = env.ID
			result.EnvironmentIDs[envData.Name] = env.ID
			s.logger.Info("Created environment", zap.String("env_id", env.ID), zap.String("name", env.Name))
		}
	} else if req.TargetEnvID != "" {
		// 使用指定的目标环境
		env, err := s.envRepo.FindByID(ctx, req.TargetEnvID)
		if err != nil || env == nil {
			return nil, fmt.Errorf("target environment not found: %s", req.TargetEnvID)
		}
		envNameToID["default"] = req.TargetEnvID
	}

	// 导入规则
	for _, ruleData := range req.Data.Data.Rules {
		// 确定目标环境ID
		targetEnvID := req.TargetEnvID
		if targetEnvID == "" {
			// 根据环境名称查找
			if ruleData.EnvironmentName != "" {
				if envID, ok := envNameToID[ruleData.EnvironmentName]; ok {
					targetEnvID = envID
				}
			}
			if targetEnvID == "" && len(envNameToID) > 0 {
				// 使用第一个环境
				for _, envID := range envNameToID {
					targetEnvID = envID
					break
				}
			}
		}

		if targetEnvID == "" {
			result.Errors = append(result.Errors, models.ImportError{
				RuleName: ruleData.Name,
				Error:    "no target environment found",
			})
			result.Success = false
			continue
		}

		// 检查规则是否已存在
		existingRules, _ := s.ruleRepo.FindByEnvironment(ctx, projectID, targetEnvID)
		var existingRule *models.Rule
		for _, r := range existingRules {
			if r.Name == ruleData.Name {
				existingRule = r
				break
			}
		}

		// 根据策略处理
		switch req.Strategy {
		case models.ImportStrategySkip:
			if existingRule != nil {
				result.Skipped++
				s.logger.Info("Skipped existing rule", zap.String("name", ruleData.Name))
				continue
			}
		case models.ImportStrategyOverwrite:
			if existingRule != nil {
				// 更新现有规则
				existingRule.MatchType = ruleData.MatchType
				existingRule.Priority = ruleData.Priority
				existingRule.Enabled = ruleData.Enabled
				existingRule.MatchCondition = ruleData.MatchCondition
				existingRule.Response = ruleData.Response
				existingRule.Tags = ruleData.Tags

				if err := s.ruleRepo.Update(ctx, existingRule); err != nil {
					result.Errors = append(result.Errors, models.ImportError{
						RuleName: ruleData.Name,
						Error:    err.Error(),
					})
					result.Success = false
					continue
				}
				result.Updated++
				result.RuleIDs = append(result.RuleIDs, existingRule.ID)
				s.logger.Info("Updated rule", zap.String("rule_id", existingRule.ID), zap.String("name", ruleData.Name))
				continue
			}
		case models.ImportStrategyAppend:
			if existingRule != nil {
				// 自动重命名
				ruleData.Name = s.generateUniqueName(ctx, projectID, targetEnvID, ruleData.Name)
			}
		}

		// 创建新规则
		rule := &models.Rule{
			Name:           ruleData.Name,
			ProjectID:      projectID,
			EnvironmentID:  targetEnvID,
			Protocol:       ruleData.Protocol,
			MatchType:      ruleData.MatchType,
			Priority:       ruleData.Priority,
			Enabled:        ruleData.Enabled,
			MatchCondition: ruleData.MatchCondition,
			Response:       ruleData.Response,
			Tags:           ruleData.Tags,
		}

		if err := s.ruleRepo.Create(ctx, rule); err != nil {
			result.Errors = append(result.Errors, models.ImportError{
				RuleName: ruleData.Name,
				Error:    err.Error(),
			})
			result.Success = false
			continue
		}

		result.Created++
		result.RuleIDs = append(result.RuleIDs, rule.ID)
		s.logger.Info("Created rule", zap.String("rule_id", rule.ID), zap.String("name", rule.Name))
	}

	// 如果有任何错误，标记为不完全成功
	if len(result.Errors) > 0 {
		result.Success = false
	}

	return result, nil
}

// generateUniqueName 生成唯一名称
func (s *importExportService) generateUniqueName(ctx context.Context, projectID, envID, baseName string) string {
	existingRules, _ := s.ruleRepo.FindByEnvironment(ctx, projectID, envID)
	nameMap := make(map[string]bool)
	for _, r := range existingRules {
		nameMap[r.Name] = true
	}

	// 尝试添加后缀
	for i := 1; i <= 100; i++ {
		newName := fmt.Sprintf("%s_copy_%d", baseName, i)
		if !nameMap[newName] {
			return newName
		}
	}

	// 使用时间戳
	return fmt.Sprintf("%s_copy_%d", baseName, time.Now().Unix())
}

// CloneRuleService 规则克隆服务接口
type CloneRuleService interface {
	// CloneRule 克隆规则
	CloneRule(ctx context.Context, ruleID string, req *models.CloneRuleRequest) (*models.Rule, error)
}

type cloneRuleService struct {
	ruleRepo repository.RuleRepository
	logger   *zap.Logger
}

// NewCloneRuleService 创建规则克隆服务
func NewCloneRuleService(ruleRepo repository.RuleRepository, logger *zap.Logger) CloneRuleService {
	return &cloneRuleService{
		ruleRepo: ruleRepo,
		logger:   logger,
	}
}

// CloneRule 克隆规则
func (s *cloneRuleService) CloneRule(ctx context.Context, ruleID string, req *models.CloneRuleRequest) (*models.Rule, error) {
	// 查找源规则
	sourceRule, err := s.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find source rule: %w", err)
	}
	if sourceRule == nil {
		return nil, errors.New("source rule not found")
	}

	// 创建新规则（复制所有字段）
	newRule := &models.Rule{
		Name:           sourceRule.Name,
		ProjectID:      sourceRule.ProjectID,
		EnvironmentID:  req.TargetEnvironmentID,
		Protocol:       sourceRule.Protocol,
		MatchType:      sourceRule.MatchType,
		Priority:       sourceRule.Priority,
		Enabled:        sourceRule.Enabled,
		MatchCondition: sourceRule.MatchCondition,
		Response:       sourceRule.Response,
		Tags:           sourceRule.Tags,
		Creator:        sourceRule.Creator,
	}

	// 应用目标项目（如果指定）
	if req.TargetProjectID != "" {
		newRule.ProjectID = req.TargetProjectID
	}

	// 应用新名称
	if req.NewName != "" {
		newRule.Name = req.NewName
	} else {
		// 自动添加 "_copy" 后缀
		newRule.Name = s.generateCopyName(sourceRule.Name)
	}

	// 应用新优先级
	if req.NewPriority != nil {
		newRule.Priority = *req.NewPriority
	}

	// 创建规则
	if err := s.ruleRepo.Create(ctx, newRule); err != nil {
		return nil, fmt.Errorf("failed to create cloned rule: %w", err)
	}

	s.logger.Info("Cloned rule",
		zap.String("source_id", ruleID),
		zap.String("new_id", newRule.ID),
		zap.String("new_name", newRule.Name),
	)

	return newRule, nil
}

// generateCopyName 生成复制名称
func (s *cloneRuleService) generateCopyName(baseName string) string {
	if !strings.HasSuffix(baseName, "_copy") {
		return baseName + "_copy"
	}
	return baseName
}

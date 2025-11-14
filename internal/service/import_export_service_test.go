package service

import (
	"context"
	"testing"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockProjectRepository for import/export tests
type MockImportProjectRepository struct {
	mock.Mock
}

func (m *MockImportProjectRepository) Create(ctx context.Context, project *models.Project) error {
	args := m.Called(ctx, project)
	if args.Get(0) == nil {
		project.ID = "test-project-id"
		return nil
	}
	return args.Error(0)
}

func (m *MockImportProjectRepository) Update(ctx context.Context, project *models.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockImportProjectRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImportProjectRepository) FindByID(ctx context.Context, id string) (*models.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Project), args.Error(1)
}

func (m *MockImportProjectRepository) FindByWorkspace(ctx context.Context, workspaceID string) ([]*models.Project, error) {
	args := m.Called(ctx, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Project), args.Error(1)
}

func (m *MockImportProjectRepository) List(ctx context.Context, skip, limit int64) ([]*models.Project, int64, error) {
	args := m.Called(ctx, skip, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Project), args.Get(1).(int64), args.Error(2)
}

// MockEnvironmentRepository for import/export tests
type MockImportEnvironmentRepository struct {
	mock.Mock
}

func (m *MockImportEnvironmentRepository) Create(ctx context.Context, env *models.Environment) error {
	args := m.Called(ctx, env)
	if args.Get(0) == nil {
		if env.ID == "" {
			env.ID = "test-env-id"
		}
		return nil
	}
	return args.Error(0)
}

func (m *MockImportEnvironmentRepository) Update(ctx context.Context, env *models.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockImportEnvironmentRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockImportEnvironmentRepository) FindByID(ctx context.Context, id string) (*models.Environment, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Environment), args.Error(1)
}

func (m *MockImportEnvironmentRepository) FindByProject(ctx context.Context, projectID string) ([]*models.Environment, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Environment), args.Error(1)
}

// 测试辅助函数
func setupImportExportService() (*importExportService, *MockBatchRuleRepository, *MockImportProjectRepository, *MockImportEnvironmentRepository) {
	mockRuleRepo := new(MockBatchRuleRepository)
	mockProjectRepo := new(MockImportProjectRepository)
	mockEnvRepo := new(MockImportEnvironmentRepository)
	logger := zap.NewNop()

	service := &importExportService{
		ruleRepo:    mockRuleRepo,
		projectRepo: mockProjectRepo,
		envRepo:     mockEnvRepo,
		logger:      logger,
	}

	return service, mockRuleRepo, mockProjectRepo, mockEnvRepo
}

// TestExportRules 测试导出规则
func TestExportRules(t *testing.T) {
	service, mockRuleRepo, mockProjectRepo, mockEnvRepo := setupImportExportService()
	ctx := context.Background()

	t.Run("Export by RuleIDs", func(t *testing.T) {
		req := &models.ExportRequest{
			RuleIDs:         []string{"rule1", "rule2"},
			IncludeProject:  false,
			IncludeEnvs:     false,
			IncludeMetadata: false,
		}

		rule1 := createTestRule("rule1", true)
		rule2 := createTestRule("rule2", false)

		mockRuleRepo.On("FindByID", ctx, "rule1").Return(rule1, nil).Once()
		mockRuleRepo.On("FindByID", ctx, "rule2").Return(rule2, nil).Once()

		result, err := service.ExportRules(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "1.0", result.Version)
		assert.Equal(t, models.ExportTypeRules, result.ExportType)
		assert.Len(t, result.Data.Rules, 2)
		assert.Nil(t, result.Metadata)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Export by Environment", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		req := &models.ExportRequest{
			ProjectID:       "project1",
			EnvironmentID:   "env1",
			IncludeProject:  false,
			IncludeEnvs:     false,
			IncludeMetadata: false,
		}

		rules := []*models.Rule{createTestRule("rule1", true), createTestRule("rule2", false)}
		mockRuleRepo.On("FindByEnvironment", ctx, "project1", "env1").Return(rules, nil).Once()

		result, err := service.ExportRules(ctx, req)

		assert.NoError(t, err)
		assert.Len(t, result.Data.Rules, 2)
		assert.Equal(t, models.ExportTypeRules, result.ExportType)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Export with Project Info", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		mockProjectRepo = new(MockImportProjectRepository)
		service.ruleRepo = mockRuleRepo
		service.projectRepo = mockProjectRepo

		req := &models.ExportRequest{
			ProjectID:      "project1",
			IncludeProject: true,
			IncludeEnvs:    false,
		}

		project := &models.Project{
			ID:          "project1",
			Name:        "Test Project",
			WorkspaceID: "workspace1",
			Description: "Test Description",
		}

		filter := map[string]interface{}{"project_id": "project1"}
		rules := []*models.Rule{createTestRule("rule1", true)}

		mockProjectRepo.On("FindByID", ctx, "project1").Return(project, nil).Once()
		mockRuleRepo.On("List", ctx, filter, int64(0), int64(10000)).Return(rules, int64(1), nil).Once()

		result, err := service.ExportRules(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result.Data.Project)
		assert.Equal(t, "Test Project", result.Data.Project.Name)
		assert.Equal(t, models.ExportTypeProject, result.ExportType)
		mockProjectRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Export with Environments", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		mockEnvRepo = new(MockImportEnvironmentRepository)
		service.ruleRepo = mockRuleRepo
		service.envRepo = mockEnvRepo

		req := &models.ExportRequest{
			ProjectID:   "project1",
			IncludeEnvs: true,
		}

		envs := []*models.Environment{
			{ID: "env1", Name: "Test Env", ProjectID: "project1", BaseURL: "http://test.com"},
		}

		filter := map[string]interface{}{"project_id": "project1"}
		rule := createTestRule("rule1", true)
		rule.EnvironmentID = "env1"
		rules := []*models.Rule{rule}

		mockEnvRepo.On("FindByProject", ctx, "project1").Return(envs, nil).Once()
		mockRuleRepo.On("List", ctx, filter, int64(0), int64(10000)).Return(rules, int64(1), nil).Once()

		result, err := service.ExportRules(ctx, req)

		assert.NoError(t, err)
		assert.Len(t, result.Data.Environments, 1)
		assert.Equal(t, "Test Env", result.Data.Environments[0].Name)
		assert.Equal(t, "Test Env", result.Data.Rules[0].EnvironmentName)
		assert.Equal(t, models.ExportTypeEnvironment, result.ExportType)
		mockEnvRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Export with Metadata", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		req := &models.ExportRequest{
			RuleIDs:         []string{"rule1"},
			IncludeMetadata: true,
		}

		rule1 := createTestRule("rule1", true)
		mockRuleRepo.On("FindByID", ctx, "rule1").Return(rule1, nil).Once()

		result, err := service.ExportRules(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result.Metadata)
		assert.Contains(t, result.Metadata.Comment, "Exported 1 rules")
		mockRuleRepo.AssertExpectations(t)
	})
}

// TestExportProject 测试导出整个项目
func TestExportProject(t *testing.T) {
	service, mockRuleRepo, mockProjectRepo, mockEnvRepo := setupImportExportService()
	ctx := context.Background()

	project := &models.Project{
		ID:          "project1",
		Name:        "Test Project",
		WorkspaceID: "workspace1",
	}

	envs := []*models.Environment{
		{ID: "env1", Name: "Dev", ProjectID: "project1"},
	}

	filter := map[string]interface{}{"project_id": "project1"}
	rules := []*models.Rule{createTestRule("rule1", true)}

	mockProjectRepo.On("FindByID", ctx, "project1").Return(project, nil).Once()
	mockEnvRepo.On("FindByProject", ctx, "project1").Return(envs, nil).Once()
	mockRuleRepo.On("List", ctx, filter, int64(0), int64(10000)).Return(rules, int64(1), nil).Once()

	result, err := service.ExportProject(ctx, "project1", true)

	assert.NoError(t, err)
	assert.NotNil(t, result.Data.Project)
	assert.Len(t, result.Data.Environments, 1)
	assert.NotNil(t, result.Metadata)
	mockProjectRepo.AssertExpectations(t)
	mockEnvRepo.AssertExpectations(t)
	mockRuleRepo.AssertExpectations(t)
}

// TestValidateImportData 测试验证导入数据
func TestValidateImportData(t *testing.T) {
	service, _, _, _ := setupImportExportService()
	ctx := context.Background()

	t.Run("Valid Data", func(t *testing.T) {
		data := &models.ExportData{
			Version:    "1.0",
			ExportType: models.ExportTypeRules,
			Data: models.ExportDataContent{
				Rules: []models.RuleExportData{
					{
						Name:           "Test Rule",
						Protocol:       models.ProtocolHTTP,
						MatchType:      models.MatchTypeSimple,
						MatchCondition: map[string]interface{}{"path": "/test"},
					},
				},
			},
		}

		err := service.ValidateImportData(ctx, data)
		assert.NoError(t, err)
	})

	t.Run("Unsupported Version", func(t *testing.T) {
		data := &models.ExportData{
			Version:    "2.0",
			ExportType: models.ExportTypeRules,
			Data: models.ExportDataContent{
				Rules: []models.RuleExportData{{Name: "Test"}},
			},
		}

		err := service.ValidateImportData(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported data version")
	})

	t.Run("Invalid Export Type", func(t *testing.T) {
		data := &models.ExportData{
			Version:    "1.0",
			ExportType: "invalid",
			Data: models.ExportDataContent{
				Rules: []models.RuleExportData{{Name: "Test"}},
			},
		}

		err := service.ValidateImportData(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid export type")
	})

	t.Run("No Rules", func(t *testing.T) {
		data := &models.ExportData{
			Version:    "1.0",
			ExportType: models.ExportTypeRules,
			Data:       models.ExportDataContent{Rules: []models.RuleExportData{}},
		}

		err := service.ValidateImportData(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no rules")
	})

	t.Run("Missing Name", func(t *testing.T) {
		data := &models.ExportData{
			Version:    "1.0",
			ExportType: models.ExportTypeRules,
			Data: models.ExportDataContent{
				Rules: []models.RuleExportData{
					{Protocol: models.ProtocolHTTP},
				},
			},
		}

		err := service.ValidateImportData(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("Missing Protocol", func(t *testing.T) {
		data := &models.ExportData{
			Version:    "1.0",
			ExportType: models.ExportTypeRules,
			Data: models.ExportDataContent{
				Rules: []models.RuleExportData{
					{Name: "Test", MatchType: models.MatchTypeSimple},
				},
			},
		}

		err := service.ValidateImportData(ctx, data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "protocol is required")
	})
}

// TestImportData 测试导入数据
func TestImportData(t *testing.T) {
	service, mockRuleRepo, mockProjectRepo, mockEnvRepo := setupImportExportService()
	ctx := context.Background()

	t.Run("Create New Project and Rules", func(t *testing.T) {
		req := &models.ImportRequest{
			Data: models.ExportData{
				Version:    "1.0",
				ExportType: models.ExportTypeProject,
				Data: models.ExportDataContent{
					Project: &models.ProjectExportData{
						Name:        "New Project",
						WorkspaceID: "workspace1",
					},
					Environments: []models.EnvironmentExportData{
						{Name: "Dev", BaseURL: "http://dev.test.com"},
					},
					Rules: []models.RuleExportData{
						{
							Name:            "Test Rule",
							EnvironmentName: "Dev",
							Protocol:        models.ProtocolHTTP,
							MatchType:       models.MatchTypeSimple,
							MatchCondition:  map[string]interface{}{"path": "/test"},
							Response:        models.Response{Type: models.ResponseTypeStatic},
							Priority:        100,
							Enabled:         true,
						},
					},
				},
			},
			Strategy:      models.ImportStrategySkip,
			CreateProject: true,
			CreateEnvs:    true,
		}

		mockProjectRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
		mockEnvRepo.On("Create", ctx, mock.Anything).Return(nil).Once()
		mockRuleRepo.On("FindByEnvironment", ctx, mock.Anything, mock.Anything).Return([]*models.Rule{}, nil).Once()
		mockRuleRepo.On("Create", ctx, mock.Anything).Return(nil).Once()

		result, err := service.ImportData(ctx, req)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 1, result.Created)
		assert.Equal(t, 0, result.Updated)
		assert.Equal(t, 0, result.Skipped)
		assert.NotEmpty(t, result.ProjectID)
		assert.Len(t, result.EnvironmentIDs, 1)
		mockProjectRepo.AssertExpectations(t)
		mockEnvRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Import to Existing Project - Skip Strategy", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		req := &models.ImportRequest{
			Data: models.ExportData{
				Version:    "1.0",
				ExportType: models.ExportTypeRules,
				Data: models.ExportDataContent{
					Rules: []models.RuleExportData{
						{
							Name:           "Existing Rule",
							Protocol:       models.ProtocolHTTP,
							MatchType:      models.MatchTypeSimple,
							MatchCondition: map[string]interface{}{"path": "/test"},
							Response:       models.Response{Type: models.ResponseTypeStatic},
						},
					},
				},
			},
			TargetProjectID: "project1",
			TargetEnvID:     "env1",
			Strategy:        models.ImportStrategySkip,
		}

		existingRule := createTestRule("rule1", true)
		existingRule.Name = "Existing Rule"

		// Mock env find for TargetEnvID validation
		mockEnvRepo.On("FindByID", ctx, "env1").Return(&models.Environment{ID: "env1", Name: "Test Env"}, nil).Once()
		mockRuleRepo.On("FindByEnvironment", ctx, "project1", "env1").Return([]*models.Rule{existingRule}, nil).Once()

		result, err := service.ImportData(ctx, req)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 0, result.Created)
		assert.Equal(t, 1, result.Skipped)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Import to Existing Project - Overwrite Strategy", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		req := &models.ImportRequest{
			Data: models.ExportData{
				Version:    "1.0",
				ExportType: models.ExportTypeRules,
				Data: models.ExportDataContent{
					Rules: []models.RuleExportData{
						{
							Name:           "Existing Rule",
							Protocol:       models.ProtocolHTTP,
							MatchType:      models.MatchTypeSimple,
							MatchCondition: map[string]interface{}{"path": "/updated"},
							Response:       models.Response{Type: models.ResponseTypeStatic},
							Priority:       200,
						},
					},
				},
			},
			TargetProjectID: "project1",
			TargetEnvID:     "env1",
			Strategy:        models.ImportStrategyOverwrite,
		}

		existingRule := createTestRule("rule1", true)
		existingRule.Name = "Existing Rule"

		// Mock env find for TargetEnvID validation
		mockEnvRepo.On("FindByID", ctx, "env1").Return(&models.Environment{ID: "env1", Name: "Test Env"}, nil).Once()
		mockRuleRepo.On("FindByEnvironment", ctx, "project1", "env1").Return([]*models.Rule{existingRule}, nil).Once()
		mockRuleRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

		result, err := service.ImportData(ctx, req)

		assert.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 0, result.Created)
		assert.Equal(t, 1, result.Updated)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Import with Validation Error", func(t *testing.T) {
		req := &models.ImportRequest{
			Data: models.ExportData{
				Version:    "1.0",
				ExportType: models.ExportTypeRules,
				Data: models.ExportDataContent{
					Rules: []models.RuleExportData{}, // Empty rules
				},
			},
			TargetProjectID: "project1",
			TargetEnvID:     "env1",
			Strategy:        models.ImportStrategySkip,
		}

		result, err := service.ImportData(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Missing Target Project", func(t *testing.T) {
		req := &models.ImportRequest{
			Data: models.ExportData{
				Version:    "1.0",
				ExportType: models.ExportTypeRules,
				Data: models.ExportDataContent{
					Rules: []models.RuleExportData{
						{
							Name:           "Test",
							Protocol:       models.ProtocolHTTP,
							MatchType:      models.MatchTypeSimple,
							MatchCondition: map[string]interface{}{"path": "/test"},
						},
					},
				},
			},
			Strategy:      models.ImportStrategySkip,
			CreateProject: false,
		}

		result, err := service.ImportData(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "project_id is required")
	})
}

// TestCloneRule 测试克隆规则
func TestCloneRule(t *testing.T) {
	mockRuleRepo := new(MockBatchRuleRepository)
	logger := zap.NewNop()
	service := &cloneRuleService{
		ruleRepo: mockRuleRepo,
		logger:   logger,
	}
	ctx := context.Background()

	t.Run("Clone with Auto Name", func(t *testing.T) {
		sourceRule := createTestRule("source-rule", true)
		sourceRule.Name = "Original Rule"

		mockRuleRepo.On("FindByID", ctx, "source-rule").Return(sourceRule, nil).Once()
		mockRuleRepo.On("Create", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return r.Name == "Original Rule_copy" && r.EnvironmentID == "target-env"
		})).Return(nil).Once()

		req := &models.CloneRuleRequest{
			TargetEnvironmentID: "target-env",
		}

		result, err := service.CloneRule(ctx, "source-rule", req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Original Rule_copy", result.Name)
		assert.Equal(t, "target-env", result.EnvironmentID)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Clone with Custom Name", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		sourceRule := createTestRule("source-rule", true)

		mockRuleRepo.On("FindByID", ctx, "source-rule").Return(sourceRule, nil).Once()
		mockRuleRepo.On("Create", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return r.Name == "Custom Name"
		})).Return(nil).Once()

		req := &models.CloneRuleRequest{
			TargetEnvironmentID: "target-env",
			NewName:             "Custom Name",
		}

		result, err := service.CloneRule(ctx, "source-rule", req)

		assert.NoError(t, err)
		assert.Equal(t, "Custom Name", result.Name)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Clone with New Priority", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		sourceRule := createTestRule("source-rule", true)
		newPriority := 200

		mockRuleRepo.On("FindByID", ctx, "source-rule").Return(sourceRule, nil).Once()
		mockRuleRepo.On("Create", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return r.Priority == 200
		})).Return(nil).Once()

		req := &models.CloneRuleRequest{
			TargetEnvironmentID: "target-env",
			NewPriority:         &newPriority,
		}

		result, err := service.CloneRule(ctx, "source-rule", req)

		assert.NoError(t, err)
		assert.Equal(t, 200, result.Priority)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Clone to Different Project", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		sourceRule := createTestRule("source-rule", true)

		mockRuleRepo.On("FindByID", ctx, "source-rule").Return(sourceRule, nil).Once()
		mockRuleRepo.On("Create", ctx, mock.MatchedBy(func(r *models.Rule) bool {
			return r.ProjectID == "target-project"
		})).Return(nil).Once()

		req := &models.CloneRuleRequest{
			TargetProjectID:     "target-project",
			TargetEnvironmentID: "target-env",
		}

		result, err := service.CloneRule(ctx, "source-rule", req)

		assert.NoError(t, err)
		assert.Equal(t, "target-project", result.ProjectID)
		mockRuleRepo.AssertExpectations(t)
	})

	t.Run("Source Rule Not Found", func(t *testing.T) {
		mockRuleRepo = new(MockBatchRuleRepository)
		service.ruleRepo = mockRuleRepo

		mockRuleRepo.On("FindByID", ctx, "nonexistent").Return(nil, nil).Once()

		req := &models.CloneRuleRequest{
			TargetEnvironmentID: "target-env",
		}

		result, err := service.CloneRule(ctx, "nonexistent", req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRuleRepo.AssertExpectations(t)
	})
}

// TestGenerateCopyName 测试生成复制名称
func TestGenerateCopyName(t *testing.T) {
	mockRuleRepo := new(MockBatchRuleRepository)
	service := &cloneRuleService{
		ruleRepo: mockRuleRepo,
		logger:   zap.NewNop(),
	}

	t.Run("Name Without Copy Suffix", func(t *testing.T) {
		name := service.generateCopyName("Original Name")
		assert.Equal(t, "Original Name_copy", name)
	})

	t.Run("Name With Copy Suffix", func(t *testing.T) {
		name := service.generateCopyName("Original Name_copy")
		assert.Equal(t, "Original Name_copy", name)
	})
}

// TestGenerateUniqueName 测试生成唯一名称
func TestGenerateUniqueName(t *testing.T) {
	service, mockRuleRepo, _, _ := setupImportExportService()
	ctx := context.Background()

	t.Run("Generate Unique Name", func(t *testing.T) {
		existingRules := []*models.Rule{
			{Name: "Test Rule"},
			{Name: "Test Rule_copy_1"},
		}

		mockRuleRepo.On("FindByEnvironment", ctx, "project1", "env1").Return(existingRules, nil).Once()

		uniqueName := service.generateUniqueName(ctx, "project1", "env1", "Test Rule")

		assert.Equal(t, "Test Rule_copy_2", uniqueName)
		mockRuleRepo.AssertExpectations(t)
	})
}

// TestNewImportExportService 测试服务创建
func TestNewImportExportService(t *testing.T) {
	mockRuleRepo := new(MockBatchRuleRepository)
	mockProjectRepo := new(MockImportProjectRepository)
	mockEnvRepo := new(MockImportEnvironmentRepository)
	logger := zap.NewNop()

	service := NewImportExportService(mockRuleRepo, mockProjectRepo, mockEnvRepo, logger)

	assert.NotNil(t, service)
	assert.Implements(t, (*ImportExportService)(nil), service)
}

// TestNewCloneRuleService 测试克隆服务创建
func TestNewCloneRuleService(t *testing.T) {
	mockRuleRepo := new(MockBatchRuleRepository)
	logger := zap.NewNop()

	service := NewCloneRuleService(mockRuleRepo, logger)

	assert.NotNil(t, service)
	assert.Implements(t, (*CloneRuleService)(nil), service)
}

//go:build integration
// +build integration

package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupRuleTestDB 设置规则测试数据库
func setupRuleTestDB(t *testing.T) (*mongo.Client, *mongo.Database, RuleRepository) {
	ctx := context.Background()

	uri := "mongodb://localhost:27017"
	if envURI := os.Getenv("MONGODB_URI"); envURI != "" {
		uri = envURI
	}

	clientOpts := options.Client().ApplyURI(uri).SetServerSelectionTimeout(2 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		t.Skipf("Skipping integration test: MongoDB not available - %v", err)
		return nil, nil, nil
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		client.Disconnect(ctx)
		t.Skipf("Skipping integration test: cannot ping MongoDB - %v", err)
		return nil, nil, nil
	}

	db := client.Database("mockserver_test_rule_" + primitive.NewObjectID().Hex())
	t.Logf("Using test database: %s", db.Name())

	collection := db.Collection("rules")
	repo := &ruleRepository{collection: collection}

	return client, db, repo
}

// teardownRuleTestDB 清理规则测试数据库
func teardownRuleTestDB(t *testing.T, client *mongo.Client, db *mongo.Database) {
	if db != nil {
		ctx := context.Background()
		err := db.Drop(ctx)
		if err != nil {
			t.Logf("Warning: failed to drop test database: %v", err)
		}
	}

	if client != nil {
		ctx := context.Background()
		err := client.Disconnect(ctx)
		if err != nil {
			t.Logf("Warning: failed to disconnect: %v", err)
		}
	}
}

// TestRuleRepository_Create_MinimalFields 测试最小字段创建
func TestRuleRepository_Create_MinimalFields(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	rule := &models.Rule{
		Name:          "最小字段规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		MatchType:     models.MatchTypeSimple,
		Priority:      100,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": "GET",
			"path":   "/api/test",
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code": 200,
			},
		},
	}

	err := repo.Create(ctx, rule)
	require.NoError(t, err)
	assert.NotEmpty(t, rule.ID)
	assert.False(t, rule.CreatedAt.IsZero())
	assert.False(t, rule.UpdatedAt.IsZero())
}

// TestRuleRepository_Create_CompleteFields 测试完整字段创建
func TestRuleRepository_Create_CompleteFields(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	rule := &models.Rule{
		Name:          "完整字段规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		MatchType:     models.MatchTypeSimple,
		Priority:      500,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": []string{"GET", "POST"},
			"path":   "/api/users/:id",
			"query": map[string]interface{}{
				"type": "admin",
			},
			"headers": map[string]interface{}{
				"Authorization": "Bearer token",
			},
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Delay: &models.DelayConfig{
				Type:  "fixed",
				Fixed: 100,
			},
			Content: map[string]interface{}{
				"status_code":  200,
				"content_type": "JSON",
				"headers": map[string]interface{}{
					"X-Custom-Header": "value",
				},
				"body": map[string]interface{}{
					"code":    0,
					"message": "success",
					"data": map[string]interface{}{
						"id":   1,
						"name": "Test User",
					},
				},
			},
		},
		Tags:    []string{"test", "user"},
		Creator: "admin",
	}

	err := repo.Create(ctx, rule)
	require.NoError(t, err)
	assert.NotEmpty(t, rule.ID)

	// 验证数据完整性
	found, err := repo.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, rule.Name, found.Name)
	assert.Equal(t, rule.Priority, found.Priority)
	assert.NotNil(t, found.Response.Delay)
	assert.Equal(t, 100, found.Response.Delay.Fixed)
	assert.Len(t, found.Tags, 2)
}

// TestRuleRepository_Update_PriorityChange 测试优先级变更
func TestRuleRepository_Update_PriorityChange(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	// 创建规则
	rule := &models.Rule{
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		Priority:      100,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": "GET",
			"path":   "/test",
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code": 200,
			},
		},
	}

	err := repo.Create(ctx, rule)
	require.NoError(t, err)

	// 更新优先级
	rule.Priority = 500
	time.Sleep(10 * time.Millisecond)
	err = repo.Update(ctx, rule)
	require.NoError(t, err)

	// 验证
	found, err := repo.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, 500, found.Priority)
}

// TestRuleRepository_Update_MatchConditionChange 测试匹配条件变更
func TestRuleRepository_Update_MatchConditionChange(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	rule := &models.Rule{
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		Priority:      100,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": "GET",
			"path":   "/api/test",
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code": 200,
			},
		},
	}

	err := repo.Create(ctx, rule)
	require.NoError(t, err)

	// 更新匹配条件
	rule.MatchCondition = map[string]interface{}{
		"method": []string{"GET", "POST"},
		"path":   "/api/updated",
		"query": map[string]interface{}{
			"version": "v2",
		},
	}

	err = repo.Update(ctx, rule)
	require.NoError(t, err)

	// 验证
	found, err := repo.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.MatchCondition)
	assert.Contains(t, found.MatchCondition, "query")
}

// TestRuleRepository_Update_ResponseChange 测试响应配置变更
func TestRuleRepository_Update_ResponseChange(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	rule := &models.Rule{
		Name:          "测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		Priority:      100,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": "GET",
			"path":   "/test",
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code": 200,
				"body":        "old response",
			},
		},
	}

	err := repo.Create(ctx, rule)
	require.NoError(t, err)

	// 更新响应
	rule.Response = models.Response{
		Type: models.ResponseTypeStatic,
		Delay: &models.DelayConfig{
			Type: "random",
			Min:  50,
			Max:  150,
		},
		Content: map[string]interface{}{
			"status_code": 201,
			"body": map[string]interface{}{
				"status":  "updated",
				"message": "new response",
			},
		},
	}

	err = repo.Update(ctx, rule)
	require.NoError(t, err)

	// 验证
	found, err := repo.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.NotNil(t, found.Response.Delay)
	statusCode := found.Response.Content["status_code"].(int32) // MongoDB may return int32
	assert.Equal(t, int32(201), statusCode)
}

// TestRuleRepository_FindByEnvironment_Priority 测试优先级排序
func TestRuleRepository_FindByEnvironment_Priority(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	// 创建不同优先级的规则
	rules := []*models.Rule{
		{Name: "低优先级", ProjectID: "p1", EnvironmentID: "e1", Protocol: models.ProtocolHTTP, Priority: 10, Enabled: true,
			MatchCondition: map[string]interface{}{"method": "GET"}, Response: models.Response{Type: models.ResponseTypeStatic, Content: map[string]interface{}{"status_code": 200}}},
		{Name: "高优先级", ProjectID: "p1", EnvironmentID: "e1", Protocol: models.ProtocolHTTP, Priority: 1000, Enabled: true,
			MatchCondition: map[string]interface{}{"method": "GET"}, Response: models.Response{Type: models.ResponseTypeStatic, Content: map[string]interface{}{"status_code": 200}}},
		{Name: "中优先级", ProjectID: "p1", EnvironmentID: "e1", Protocol: models.ProtocolHTTP, Priority: 500, Enabled: true,
			MatchCondition: map[string]interface{}{"method": "GET"}, Response: models.Response{Type: models.ResponseTypeStatic, Content: map[string]interface{}{"status_code": 200}}},
	}

	for _, rule := range rules {
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}

	// 查询并验证排序
	found, err := repo.FindByEnvironment(ctx, "p1", "e1")
	require.NoError(t, err)
	assert.Len(t, found, 3)

	// 应该按优先级降序
	assert.Equal(t, "高优先级", found[0].Name)
	assert.Equal(t, "中优先级", found[1].Name)
	assert.Equal(t, "低优先级", found[2].Name)
}

// TestRuleRepository_FindEnabledByEnvironment_OnlyEnabled 测试只返回启用的规则
func TestRuleRepository_FindEnabledByEnvironment_OnlyEnabled(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	rules := []*models.Rule{
		{Name: "启用1", ProjectID: "p1", EnvironmentID: "e1", Protocol: models.ProtocolHTTP, Priority: 300, Enabled: true,
			MatchCondition: map[string]interface{}{"method": "GET"}, Response: models.Response{Type: models.ResponseTypeStatic, Content: map[string]interface{}{"status_code": 200}}},
		{Name: "禁用1", ProjectID: "p1", EnvironmentID: "e1", Protocol: models.ProtocolHTTP, Priority: 200, Enabled: false,
			MatchCondition: map[string]interface{}{"method": "GET"}, Response: models.Response{Type: models.ResponseTypeStatic, Content: map[string]interface{}{"status_code": 200}}},
		{Name: "启用2", ProjectID: "p1", EnvironmentID: "e1", Protocol: models.ProtocolHTTP, Priority: 100, Enabled: true,
			MatchCondition: map[string]interface{}{"method": "GET"}, Response: models.Response{Type: models.ResponseTypeStatic, Content: map[string]interface{}{"status_code": 200}}},
		{Name: "禁用2", ProjectID: "p1", EnvironmentID: "e1", Protocol: models.ProtocolHTTP, Priority: 50, Enabled: false,
			MatchCondition: map[string]interface{}{"method": "GET"}, Response: models.Response{Type: models.ResponseTypeStatic, Content: map[string]interface{}{"status_code": 200}}},
	}

	for _, rule := range rules {
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}

	// 查询启用的规则
	found, err := repo.FindEnabledByEnvironment(ctx, "p1", "e1")
	require.NoError(t, err)
	assert.Len(t, found, 2)

	// 验证都是启用状态
	for _, rule := range found {
		assert.True(t, rule.Enabled)
	}

	// 验证排序
	assert.Equal(t, "启用1", found[0].Name)
	assert.Equal(t, "启用2", found[1].Name)
}

// TestRuleRepository_List_FilterByEnabled 测试按启用状态过滤
func TestRuleRepository_List_FilterByEnabled(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	// 创建混合状态的规则
	for i := 0; i < 10; i++ {
		rule := &models.Rule{
			Name:          "规则" + string(rune('A'+i)),
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      i * 10,
			Enabled:       i%2 == 0, // 偶数启用，奇数禁用
			MatchCondition: map[string]interface{}{
				"method": "GET",
			},
			Response: models.Response{
				Type: models.ResponseTypeStatic,
				Content: map[string]interface{}{
					"status_code": 200,
				},
			},
		}
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}

	// 查询启用的规则
	filter := map[string]interface{}{
		"project_id":     "project-001",
		"environment_id": "env-001",
		"enabled":        true,
	}
	enabled, total, err := repo.List(ctx, filter, 0, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, enabled, 5)

	for _, rule := range enabled {
		assert.True(t, rule.Enabled)
	}

	// 查询禁用的规则
	filter["enabled"] = false
	disabled, total, err := repo.List(ctx, filter, 0, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, disabled, 5)

	for _, rule := range disabled {
		assert.False(t, rule.Enabled)
	}
}

// TestRuleRepository_List_EmptyFilter 测试空过滤条件
func TestRuleRepository_List_EmptyFilter(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	// 创建规则
	for i := 0; i < 5; i++ {
		rule := &models.Rule{
			Name:          "规则" + string(rune('A'+i)),
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      i * 10,
			Enabled:       true,
			MatchCondition: map[string]interface{}{
				"method": "GET",
			},
			Response: models.Response{
				Type: models.ResponseTypeStatic,
				Content: map[string]interface{}{
					"status_code": 200,
				},
			},
		}
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
	}

	// 使用空过滤条件查询（应该返回所有规则）
	filter := map[string]interface{}{}
	rules, total, err := repo.List(ctx, filter, 0, 20)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, rules, 5)
}

// TestRuleRepository_Delete_MultipleRules 测试删除多个规则
func TestRuleRepository_Delete_MultipleRules(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	// 创建多个规则
	var ruleIDs []string
	for i := 0; i < 5; i++ {
		rule := &models.Rule{
			Name:          "规则" + string(rune('A'+i)),
			ProjectID:     "project-001",
			EnvironmentID: "env-001",
			Protocol:      models.ProtocolHTTP,
			Priority:      i * 10,
			Enabled:       true,
			MatchCondition: map[string]interface{}{
				"method": "GET",
			},
			Response: models.Response{
				Type: models.ResponseTypeStatic,
				Content: map[string]interface{}{
					"status_code": 200,
				},
			},
		}
		err := repo.Create(ctx, rule)
		require.NoError(t, err)
		ruleIDs = append(ruleIDs, rule.ID)
	}

	// 删除其中3个
	for i := 0; i < 3; i++ {
		err := repo.Delete(ctx, ruleIDs[i])
		require.NoError(t, err)
	}

	// 验证删除结果
	for i, id := range ruleIDs {
		found, err := repo.FindByID(ctx, id)
		require.NoError(t, err)
		if i < 3 {
			assert.Nil(t, found, "前3个规则应该被删除")
		} else {
			assert.NotNil(t, found, "后2个规则应该还存在")
		}
	}
}

// TestRuleRepository_FindByEnvironment_EmptyResult 测试空结果集
func TestRuleRepository_FindByEnvironment_EmptyResult(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	// 不创建任何规则，直接查询
	rules, err := repo.FindByEnvironment(ctx, "non-existent-project", "non-existent-env")
	require.NoError(t, err)
	assert.NotNil(t, rules)
	assert.Len(t, rules, 0)
}

// TestRuleRepository_Update_InvalidID 测试无效ID更新
func TestRuleRepository_Update_InvalidID(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	tests := []struct {
		name string
		id   string
	}{
		{"无效格式", "invalid-id"},
		{"空ID", ""},
		{"短ID", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &models.Rule{
				ID:            tt.id,
				Name:          "测试",
				ProjectID:     "p1",
				EnvironmentID: "e1",
				Protocol:      models.ProtocolHTTP,
				Priority:      100,
				Enabled:       true,
				MatchCondition: map[string]interface{}{
					"method": "GET",
				},
				Response: models.Response{
					Type: models.ResponseTypeStatic,
					Content: map[string]interface{}{
						"status_code": 200,
					},
				},
			}

			err := repo.Update(ctx, rule)
			assert.Error(t, err)
		})
	}
}

// TestRuleRepository_ConcurrentUpdate 测试并发更新
func TestRuleRepository_ConcurrentUpdate(t *testing.T) {
	client, db, repo := setupRuleTestDB(t)
	if client == nil {
		return
	}
	defer teardownRuleTestDB(t, client, db)

	ctx := context.Background()

	// 创建规则
	rule := &models.Rule{
		Name:          "并发测试规则",
		ProjectID:     "project-001",
		EnvironmentID: "env-001",
		Protocol:      models.ProtocolHTTP,
		Priority:      100,
		Enabled:       true,
		MatchCondition: map[string]interface{}{
			"method": "GET",
		},
		Response: models.Response{
			Type: models.ResponseTypeStatic,
			Content: map[string]interface{}{
				"status_code": 200,
			},
		},
	}

	err := repo.Create(ctx, rule)
	require.NoError(t, err)

	// 并发更新
	concurrency := 5
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			updateRule := *rule
			updateRule.Priority = 100 + index
			updateRule.Name = "更新" + string(rune('A'+index))
			err := repo.Update(ctx, &updateRule)
			if err != nil {
				t.Logf("Update error: %v", err)
			}
			done <- true
		}(i)
	}

	// 等待完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// 验证规则仍然存在
	found, err := repo.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
}

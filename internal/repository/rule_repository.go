package repository

import (
	"context"
	"time"

	"github.com/gomockserver/mockserver/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RuleRepository 规则仓库接口
type RuleRepository interface {
	Create(ctx context.Context, rule *models.Rule) error
	Update(ctx context.Context, rule *models.Rule) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.Rule, error)
	FindByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error)
	FindEnabledByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error)
	List(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]*models.Rule, int64, error)
}

type ruleRepository struct {
	collection *mongo.Collection
}

// NewRuleRepository 创建规则仓库
func NewRuleRepository() RuleRepository {
	return &ruleRepository{
		collection: GetCollection("rules"),
	}
}

// Create 创建规则
func (r *ruleRepository) Create(ctx context.Context, rule *models.Rule) error {
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, rule)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		rule.ID = oid.Hex()
	}

	return nil
}

// Update 更新规则
func (r *ruleRepository) Update(ctx context.Context, rule *models.Rule) error {
	rule.UpdatedAt = time.Now()

	objectID, err := primitive.ObjectIDFromHex(rule.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": rule}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete 删除规则
func (r *ruleRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	_, err = r.collection.DeleteOne(ctx, filter)
	return err
}

// FindByID 根据ID查找规则
func (r *ruleRepository) FindByID(ctx context.Context, id string) (*models.Rule, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}

	var rule models.Rule
	err = r.collection.FindOne(ctx, filter).Decode(&rule)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &rule, nil
}

// FindByEnvironment 查找环境下的所有规则
func (r *ruleRepository) FindByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error) {
	filter := bson.M{
		"project_id":     projectID,
		"environment_id": environmentID,
	}

	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []*models.Rule
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// FindEnabledByEnvironment 查找环境下所有启用的规则
func (r *ruleRepository) FindEnabledByEnvironment(ctx context.Context, projectID, environmentID string) ([]*models.Rule, error) {
	filter := bson.M{
		"project_id":     projectID,
		"environment_id": environmentID,
		"enabled":        true,
	}

	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var rules []*models.Rule
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

// List 列出规则（支持分页和过滤）
func (r *ruleRepository) List(ctx context.Context, filter map[string]interface{}, skip, limit int64) ([]*models.Rule, int64, error) {
	// 转换过滤条件
	mongoFilter := bson.M{}
	for k, v := range filter {
		mongoFilter[k] = v
	}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, mongoFilter)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var rules []*models.Rule
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

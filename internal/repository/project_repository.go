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

// ProjectRepository 项目仓库接口
type ProjectRepository interface {
	Create(ctx context.Context, project *models.Project) error
	Update(ctx context.Context, project *models.Project) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.Project, error)
	FindByWorkspace(ctx context.Context, workspaceID string) ([]*models.Project, error)
	List(ctx context.Context, skip, limit int64) ([]*models.Project, int64, error)
}

type projectRepository struct {
	collection *mongo.Collection
}

// NewProjectRepository 创建项目仓库
func NewProjectRepository() ProjectRepository {
	return &projectRepository{
		collection: GetCollection("projects"),
	}
}

// Create 创建项目
func (r *projectRepository) Create(ctx context.Context, project *models.Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, project)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		project.ID = oid.Hex()
	}

	return nil
}

// Update 更新项目
func (r *projectRepository) Update(ctx context.Context, project *models.Project) error {
	project.UpdatedAt = time.Now()

	objectID, err := primitive.ObjectIDFromHex(project.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	// 排除 _id 字段，避免更新不可变字段
	update := bson.M{"$set": bson.M{
		"name":         project.Name,
		"workspace_id": project.WorkspaceID,
		"description":  project.Description,
		"updated_at":   project.UpdatedAt,
	}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete 删除项目
func (r *projectRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	_, err = r.collection.DeleteOne(ctx, filter)
	return err
}

// FindByID 根据ID查找项目
func (r *projectRepository) FindByID(ctx context.Context, id string) (*models.Project, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}

	var project models.Project
	err = r.collection.FindOne(ctx, filter).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}

// FindByWorkspace 查找工作空间下的所有项目
func (r *projectRepository) FindByWorkspace(ctx context.Context, workspaceID string) ([]*models.Project, error) {
	filter := bson.M{"workspace_id": workspaceID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var projects []*models.Project
	if err = cursor.All(ctx, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// List 列出项目（支持分页）
func (r *projectRepository) List(ctx context.Context, skip, limit int64) ([]*models.Project, int64, error) {
	filter := bson.M{}

	// 获取总数
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// 查询数据
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{primitive.E{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var projects []*models.Project
	if err = cursor.All(ctx, &projects); err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

// EnvironmentRepository 环境仓库接口
type EnvironmentRepository interface {
	Create(ctx context.Context, environment *models.Environment) error
	Update(ctx context.Context, environment *models.Environment) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.Environment, error)
	FindByProject(ctx context.Context, projectID string) ([]*models.Environment, error)
}

type environmentRepository struct {
	collection *mongo.Collection
}

// NewEnvironmentRepository 创建环境仓库
func NewEnvironmentRepository() EnvironmentRepository {
	return &environmentRepository{
		collection: GetCollection("environments"),
	}
}

// Create 创建环境
func (r *environmentRepository) Create(ctx context.Context, environment *models.Environment) error {
	environment.CreatedAt = time.Now()
	environment.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, environment)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		environment.ID = oid.Hex()
	}

	return nil
}

// Update 更新环境
func (r *environmentRepository) Update(ctx context.Context, environment *models.Environment) error {
	environment.UpdatedAt = time.Now()

	objectID, err := primitive.ObjectIDFromHex(environment.ID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	// 排除 _id 字段，避免更新不可变字段
	update := bson.M{"$set": bson.M{
		"name":       environment.Name,
		"project_id": environment.ProjectID,
		"base_url":   environment.BaseURL,
		"variables":  environment.Variables,
		"updated_at": environment.UpdatedAt,
	}}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}

// Delete 删除环境
func (r *environmentRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	_, err = r.collection.DeleteOne(ctx, filter)
	return err
}

// FindByID 根据ID查找环境
func (r *environmentRepository) FindByID(ctx context.Context, id string) (*models.Environment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": objectID}

	var environment models.Environment
	err = r.collection.FindOne(ctx, filter).Decode(&environment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &environment, nil
}

// FindByProject 查找项目下的所有环境
func (r *environmentRepository) FindByProject(ctx context.Context, projectID string) ([]*models.Environment, error) {
	filter := bson.M{"project_id": projectID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var environments []*models.Environment
	if err = cursor.All(ctx, &environments); err != nil {
		return nil, err
	}

	return environments, nil
}

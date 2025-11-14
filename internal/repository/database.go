package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/gomockserver/mockserver/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client   *mongo.Client
	database *mongo.Database
)

// Init 初始化数据库连接
func Init(cfg *config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.MongoDB.Timeout)
	defer cancel()

	// 创建客户端选项
	clientOptions := options.Client().
		ApplyURI(cfg.Database.MongoDB.URI).
		SetMinPoolSize(uint64(cfg.Database.MongoDB.Pool.Min)).
		SetMaxPoolSize(uint64(cfg.Database.MongoDB.Pool.Max))

	// 连接数据库
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// 验证连接
	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// 获取数据库实例
	database = client.Database(cfg.Database.MongoDB.Database)

	// 创建索引
	if err := createIndexes(ctx); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// createIndexes 创建索引
func createIndexes(ctx context.Context) error {
	// Rules 集合索引
	rulesCollection := database.Collection("rules")
	rulesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "project_id", Value: 1},
				{Key: "environment_id", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "protocol", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "enabled", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "priority", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}
	if _, err := rulesCollection.Indexes().CreateMany(ctx, rulesIndexes); err != nil {
		return err
	}

	// Projects 集合索引
	projectsCollection := database.Collection("projects")
	projectsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspace_id", Value: 1}},
		},
	}
	if _, err := projectsCollection.Indexes().CreateMany(ctx, projectsIndexes); err != nil {
		return err
	}

	// Environments 集合索引
	environmentsCollection := database.Collection("environments")
	environmentsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "project_id", Value: 1}},
		},
	}
	if _, err := environmentsCollection.Indexes().CreateMany(ctx, environmentsIndexes); err != nil {
		return err
	}

	// Logs 集合索引（带 TTL）
	logsCollection := database.Collection("logs")
	logsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "request_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "rule_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "protocol", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(7 * 24 * 60 * 60), // 7天过期
		},
		{
			Keys: bson.D{
				{Key: "project_id", Value: 1},
				{Key: "environment_id", Value: 1},
			},
		},
	}
	if _, err := logsCollection.Indexes().CreateMany(ctx, logsIndexes); err != nil {
		return err
	}

	// Versions 集合索引
	versionsCollection := database.Collection("versions")
	versionsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "rule_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}
	if _, err := versionsCollection.Indexes().CreateMany(ctx, versionsIndexes); err != nil {
		return err
	}

	// Users 集合索引
	usersCollection := database.Collection("users")
	usersIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := usersCollection.Indexes().CreateMany(ctx, usersIndexes); err != nil {
		return err
	}

	return nil
}

// GetDatabase 获取数据库实例
func GetDatabase() *mongo.Database {
	return database
}

// GetCollection 获取集合
func GetCollection(name string) *mongo.Collection {
	return database.Collection(name)
}

// Close 关闭数据库连接
func Close() error {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return client.Disconnect(ctx)
	}
	return nil
}

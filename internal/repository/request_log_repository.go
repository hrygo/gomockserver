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

// RequestLogRepository 请求日志仓库接口
type RequestLogRepository interface {
	Create(ctx context.Context, log *models.RequestLog) error
	FindByID(ctx context.Context, id string) (*models.RequestLog, error)
	List(ctx context.Context, filter RequestLogFilter) ([]*models.RequestLog, int64, error)
	DeleteBefore(ctx context.Context, before time.Time) (int64, error)
	DeleteByProjectID(ctx context.Context, projectID string) error
	CountByProjectID(ctx context.Context, projectID string, startTime, endTime time.Time) (int64, error)
	GetStatistics(ctx context.Context, projectID, environmentID string, startTime, endTime time.Time) (*RequestLogStatistics, error)
}

// RequestLogFilter 请求日志查询过滤器
type RequestLogFilter struct {
	ProjectID     string
	EnvironmentID string
	RuleID        string
	Protocol      models.ProtocolType
	Method        string
	Path          string
	StatusCode    int
	SourceIP      string
	StartTime     time.Time
	EndTime       time.Time
	Page          int
	PageSize      int
	SortBy        string
	SortOrder     int // 1: asc, -1: desc
}

// RequestLogStatistics 请求日志统计信息
type RequestLogStatistics struct {
	TotalRequests   int64              `json:"total_requests"`
	SuccessRequests int64              `json:"success_requests"`
	ErrorRequests   int64              `json:"error_requests"`
	AvgDuration     float64            `json:"avg_duration"`
	MaxDuration     int64              `json:"max_duration"`
	MinDuration     int64              `json:"min_duration"`
	ProtocolStats   map[string]int64   `json:"protocol_stats"`
	StatusCodeStats map[string]int64   `json:"status_code_stats"`
	TopPaths        []PathStat         `json:"top_paths"`
	TopRules        []RuleStat         `json:"top_rules"`
	HourlyStats     []HourlyStat       `json:"hourly_stats"`
}

// PathStat 路径统计
type PathStat struct {
	Path  string `json:"path"`
	Count int64  `json:"count"`
}

// RuleStat 规则统计
type RuleStat struct {
	RuleID string `json:"rule_id"`
	Count  int64  `json:"count"`
}

// HourlyStat 小时统计
type HourlyStat struct {
	Hour  string `json:"hour"`
	Count int64  `json:"count"`
}

type mongoRequestLogRepository struct {
	collection *mongo.Collection
}

// NewMongoRequestLogRepository 创建 MongoDB 请求日志仓库
func NewMongoRequestLogRepository(db *mongo.Database) RequestLogRepository {
	return &mongoRequestLogRepository{
		collection: db.Collection("request_logs"),
	}
}

// EnsureIndexes 创建索引
func (r *mongoRequestLogRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "timestamp", Value: -1},
				{Key: "project_id", Value: 1},
				{Key: "environment_id", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "request_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "rule_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status_code", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "source_ip", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "protocol", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(7 * 24 * 60 * 60), // 7天自动过期
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// Create 创建请求日志
func (r *mongoRequestLogRepository) Create(ctx context.Context, log *models.RequestLog) error {
	if log.ID == "" {
		log.ID = primitive.NewObjectID().Hex()
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

// FindByID 根据 ID 查询请求日志
func (r *mongoRequestLogRepository) FindByID(ctx context.Context, id string) (*models.RequestLog, error) {
	var log models.RequestLog
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&log)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &log, nil
}

// List 列表查询请求日志
func (r *mongoRequestLogRepository) List(ctx context.Context, filter RequestLogFilter) ([]*models.RequestLog, int64, error) {
	// 构建查询条件
	query := bson.M{}
	
	if filter.ProjectID != "" {
		query["project_id"] = filter.ProjectID
	}
	if filter.EnvironmentID != "" {
		query["environment_id"] = filter.EnvironmentID
	}
	if filter.RuleID != "" {
		query["rule_id"] = filter.RuleID
	}
	if filter.Protocol != "" {
		query["protocol"] = filter.Protocol
	}
	if filter.Method != "" {
		query["method"] = filter.Method
	}
	if filter.Path != "" {
		query["path"] = bson.M{"$regex": filter.Path}
	}
	if filter.StatusCode > 0 {
		query["status_code"] = filter.StatusCode
	}
	if filter.SourceIP != "" {
		query["source_ip"] = filter.SourceIP
	}
	if !filter.StartTime.IsZero() || !filter.EndTime.IsZero() {
		timeQuery := bson.M{}
		if !filter.StartTime.IsZero() {
			timeQuery["$gte"] = filter.StartTime
		}
		if !filter.EndTime.IsZero() {
			timeQuery["$lte"] = filter.EndTime
		}
		query["timestamp"] = timeQuery
	}

	// 统计总数
	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// 分页和排序
	opts := options.Find()
	if filter.Page > 0 && filter.PageSize > 0 {
		skip := int64((filter.Page - 1) * filter.PageSize)
		limit := int64(filter.PageSize)
		opts.SetSkip(skip).SetLimit(limit)
	}

	sortBy := "timestamp"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := -1
	if filter.SortOrder != 0 {
		sortOrder = filter.SortOrder
	}
	opts.SetSort(bson.D{{Key: sortBy, Value: sortOrder}})

	// 查询
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []*models.RequestLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// DeleteBefore 删除指定时间之前的日志
func (r *mongoRequestLogRepository) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, bson.M{
		"timestamp": bson.M{"$lt": before},
	})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// DeleteByProjectID 删除指定项目的所有日志
func (r *mongoRequestLogRepository) DeleteByProjectID(ctx context.Context, projectID string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"project_id": projectID})
	return err
}

// CountByProjectID 统计指定项目的日志数量
func (r *mongoRequestLogRepository) CountByProjectID(ctx context.Context, projectID string, startTime, endTime time.Time) (int64, error) {
	query := bson.M{"project_id": projectID}
	
	if !startTime.IsZero() || !endTime.IsZero() {
		timeQuery := bson.M{}
		if !startTime.IsZero() {
			timeQuery["$gte"] = startTime
		}
		if !endTime.IsZero() {
			timeQuery["$lte"] = endTime
		}
		query["timestamp"] = timeQuery
	}

	return r.collection.CountDocuments(ctx, query)
}

// GetStatistics 获取统计信息
func (r *mongoRequestLogRepository) GetStatistics(ctx context.Context, projectID, environmentID string, startTime, endTime time.Time) (*RequestLogStatistics, error) {
	matchStage := bson.M{}
	if projectID != "" {
		matchStage["project_id"] = projectID
	}
	if environmentID != "" {
		matchStage["environment_id"] = environmentID
	}
	if !startTime.IsZero() || !endTime.IsZero() {
		timeQuery := bson.M{}
		if !startTime.IsZero() {
			timeQuery["$gte"] = startTime
		}
		if !endTime.IsZero() {
			timeQuery["$lte"] = endTime
		}
		matchStage["timestamp"] = timeQuery
	}

	stats := &RequestLogStatistics{
		ProtocolStats:   make(map[string]int64),
		StatusCodeStats: make(map[string]int64),
		TopPaths:        []PathStat{},
		TopRules:        []RuleStat{},
		HourlyStats:     []HourlyStat{},
	}

	// 基础统计
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id":             nil,
			"total_requests":  bson.M{"$sum": 1},
			"success_count":   bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$and": []interface{}{bson.M{"$gte": []interface{}{"$status_code", 200}}, bson.M{"$lt": []interface{}{"$status_code", 400}}}}, 1, 0}}},
			"error_count":     bson.M{"$sum": bson.M{"$cond": []interface{}{bson.M{"$gte": []interface{}{"$status_code", 400}}, 1, 0}}},
			"avg_duration":    bson.M{"$avg": "$duration"},
			"max_duration":    bson.M{"$max": "$duration"},
			"min_duration":    bson.M{"$min": "$duration"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			TotalRequests  int64   `bson:"total_requests"`
			SuccessCount   int64   `bson:"success_count"`
			ErrorCount     int64   `bson:"error_count"`
			AvgDuration    float64 `bson:"avg_duration"`
			MaxDuration    int64   `bson:"max_duration"`
			MinDuration    int64   `bson:"min_duration"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		stats.TotalRequests = result.TotalRequests
		stats.SuccessRequests = result.SuccessCount
		stats.ErrorRequests = result.ErrorCount
		stats.AvgDuration = result.AvgDuration
		stats.MaxDuration = result.MaxDuration
		stats.MinDuration = result.MinDuration
	}

	// 协议统计
	protocolPipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$protocol",
			"count": bson.M{"$sum": 1},
		}}},
	}
	protocolCursor, err := r.collection.Aggregate(ctx, protocolPipeline)
	if err == nil {
		defer protocolCursor.Close(ctx)
		for protocolCursor.Next(ctx) {
			var result struct {
				Protocol string `bson:"_id"`
				Count    int64  `bson:"count"`
			}
			if err := protocolCursor.Decode(&result); err == nil {
				stats.ProtocolStats[result.Protocol] = result.Count
			}
		}
	}

	// 状态码统计
	statusPipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$status_code",
			"count": bson.M{"$sum": 1},
		}}},
	}
	statusCursor, err := r.collection.Aggregate(ctx, statusPipeline)
	if err == nil {
		defer statusCursor.Close(ctx)
		for statusCursor.Next(ctx) {
			var result struct {
				StatusCode int   `bson:"_id"`
				Count      int64 `bson:"count"`
			}
			if err := statusCursor.Decode(&result); err == nil {
				stats.StatusCodeStats[string(rune(result.StatusCode+'0'))] = result.Count
			}
		}
	}

	return stats, nil
}

package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gomockserver/mockserver/internal/adapter"
	"github.com/gomockserver/mockserver/internal/middleware"
	"github.com/gomockserver/mockserver/internal/models"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

// MatchEngineInterface 匹配引擎接口
type MatchEngineInterface interface {
	Match(ctx context.Context, request *adapter.Request, projectID, environmentID string) (*models.Rule, error)
}

// MockExecutorInterface Mock 执行器接口
type MockExecutorInterface interface {
	Execute(request *adapter.Request, rule *models.Rule) (*adapter.Response, error)
	GetDefaultResponse() *adapter.Response
}

// MockService Mock 服务
type MockService struct {
	httpAdapter  *adapter.HTTPAdapter
	matchEngine  MatchEngineInterface
	mockExecutor MockExecutorInterface
}

// NewMockService 创建 Mock 服务
func NewMockService(matchEngine MatchEngineInterface, mockExecutor MockExecutorInterface) *MockService {
	return &MockService{
		httpAdapter:  adapter.NewHTTPAdapter(),
		matchEngine:  matchEngine,
		mockExecutor: mockExecutor,
	}
}

// HandleMockRequest 处理 Mock 请求
func (s *MockService) HandleMockRequest(c *gin.Context) {
	// 从路径中提取项目ID和环境ID
	// 请求格式：/:projectID/:environmentID/*path
	projectID := c.Param("projectID")
	environmentID := c.Param("environmentID")

	if projectID == "" || environmentID == "" {
		c.JSON(400, gin.H{
			"error": "projectID and environmentID are required",
		})
		return
	}

	// 解析请求为统一模型
	request, err := s.httpAdapter.Parse(c)
	if err != nil {
		logger.Error("failed to parse request", zap.Error(err))
		c.JSON(500, gin.H{
			"error": "Failed to parse request",
		})
		return
	}

	// 匹配规则
	ctx := context.Background()
	rule, err := s.matchEngine.Match(ctx, request, projectID, environmentID)
	if err != nil {
		logger.Error("failed to match rule", zap.Error(err))
		c.JSON(500, gin.H{
			"error": "Failed to match rule",
		})
		return
	}

	var response *adapter.Response

	// 如果没有匹配的规则，返回默认响应
	if rule == nil {
		logger.Info("no rule matched, using default response",
			zap.String("path", request.Path),
			zap.String("project_id", projectID),
			zap.String("environment_id", environmentID))
		response = s.mockExecutor.GetDefaultResponse()
	} else {
		// 执行 Mock 响应生成
		response, err = s.mockExecutor.Execute(request, rule)
		if err != nil {
			logger.Error("failed to execute mock", zap.Error(err))
			c.JSON(500, gin.H{
				"error": "Failed to execute mock",
			})
			return
		}
	}

	// 写入响应
	s.httpAdapter.WriteResponse(c, response)
}

// StartMockServer 启动 Mock 服务器
func StartMockServer(addr string, service *MockService) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	// 添加 CORS 支持，允许前端直接调用 Mock 服务
	r.Use(middleware.CORS())

	// Mock 请求处理路由
	// 格式：/:projectID/:environmentID/*path
	r.Any("/:projectID/:environmentID/*path", service.HandleMockRequest)

	logger.Info("starting mock server", zap.String("address", addr))
	return r.Run(addr)
}

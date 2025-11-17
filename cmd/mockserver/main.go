package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/gomockserver/mockserver/internal/api"
	"github.com/gomockserver/mockserver/internal/config"
	"github.com/gomockserver/mockserver/internal/engine"
	"github.com/gomockserver/mockserver/internal/executor"
	"github.com/gomockserver/mockserver/internal/repository"
	"github.com/gomockserver/mockserver/internal/service"
	"github.com/gomockserver/mockserver/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	// 初始化日志
	if err := logger.Init(
		cfg.Logging.Level,
		cfg.Logging.Format,
		cfg.Logging.Output,
		cfg.Logging.File.Path,
		cfg.Logging.File.MaxSize,
		cfg.Logging.File.MaxBackups,
		cfg.Logging.File.MaxAge,
	); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("mockserver admin starting...")

	// 初始化数据库
	if err := repository.Init(cfg); err != nil {
		logger.Fatal("failed to init database", zap.Error(err))
	}
	defer repository.Close()

	logger.Info("database connected successfully")

	// 创建仓库
	ruleRepo := repository.NewRuleRepository()
	projectRepo := repository.NewProjectRepository()
	environmentRepo := repository.NewEnvironmentRepository()

	// 创建处理器
	ruleHandler := api.NewRuleHandler(ruleRepo, projectRepo, environmentRepo)
	projectHandler := api.NewProjectHandler(projectRepo, environmentRepo)
	statisticsHandler := api.NewStatisticsHandler(repository.NewMongoRequestLogRepository(repository.GetDatabase()), repository.GetDatabase())

	// 创建导入导出服务
	importExportService := service.NewImportExportService(ruleRepo, projectRepo, environmentRepo, logger.Get())

	// 创建服务
	adminService := service.NewAdminService(ruleHandler, projectHandler, statisticsHandler, importExportService)

	// 同时启动 Mock 服务器
	matchEngine := engine.NewMatchEngine(ruleRepo)
	mockExecutor := executor.NewMockExecutor()
	mockService := service.NewMockService(matchEngine, mockExecutor)

	// 启动 Mock 服务器（在 goroutine 中）
	go func() {
		logger.Info("starting mock server", zap.String("address", cfg.GetMockAddress()))
		if err := service.StartMockServer(cfg.GetMockAddress(), mockService); err != nil {
			logger.Fatal("failed to start mock server", zap.Error(err))
		}
	}()

	// 启动管理服务器（主 goroutine）
	go func() {
		logger.Info("starting admin server", zap.String("address", cfg.GetAdminAddress()))
		if err := service.StartAdminServer(cfg.GetAdminAddress(), adminService); err != nil {
			logger.Fatal("failed to start admin server", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down mockserver...")
}

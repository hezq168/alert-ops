package main

import (
	"alert-ops/internal/alert/service"
	"alert-ops/internal/api"
	"alert-ops/internal/config"
	"alert-ops/internal/model"
	"alert-ops/internal/repo"
	"alert-ops/internal/scheduler"
	logger "alert-ops/pkg"
	"alert-ops/pkg/jwt"
	"fmt"
	"os"

	"go.uber.org/zap"

	_ "alert-ops/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title alert-ops API
// @version 1.0
// @description alert-ops API 文档
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	var configFile string
	if len(os.Args) < 2 {
		configFile = "./conf/config.yaml"
	} else {
		configFile = os.Args[1]
	}

	// 加载配置文件
	if err := config.Init(configFile); err != nil {
		fmt.Printf("配置文件加载失败: %v\n", err)
		os.Exit(1)
	}
	// 初始化日志
	if err := logger.Init(config.Conf.LogConfig, config.Conf.Mode); err != nil {
		zap.L().Fatal("初始化日志失败", zap.Error(err))
		os.Exit(1)
	}
	zap.L().Debug("日志初始化成功")

	// 初始化 JWT
	jwt.InitJWTSecret()
	zap.L().Debug("JWT初始化成功")

	// 初始化数据库
	if err := repo.InitDatabase(); err != nil {
		zap.L().Fatal("数据库初始化失败", zap.Error(err))
	}
	zap.L().Info("数据库初始化成功")

	// 初始化默认数据
	repo.InitDefaultData()

	// 初始化 AI 分析服务（先用配置文件默认值）
	aiCfg := config.Conf.AlertConfig.AI
	ais := service.NewAIService(aiCfg.Provider, aiCfg.APIKey, aiCfg.BaseURL, aiCfg.Model)
	zap.L().Info("AI服务初始化成功（配置文件默认值）",
		zap.String("provider", aiCfg.Provider),
		zap.String("model", aiCfg.Model),
	)

	// 尝试从数据库加载用户配置覆盖
	var dbAICfg model.AIConfig
	if err := repo.DB.First(&dbAICfg).Error; err == nil {
		// 数据库有配置，用数据库的覆盖
		if dbAICfg.Provider != "" && dbAICfg.APIKey != "" {
			dbProvider := service.CreateProvider(dbAICfg.Provider, dbAICfg.APIKey, dbAICfg.BaseURL, dbAICfg.Model)
			if dbProvider != nil {
				ais.SetProvider(dbProvider)
				zap.L().Info("AI服务已从数据库加载用户配置",
					zap.String("provider", dbAICfg.Provider),
					zap.String("model", dbAICfg.Model),
				)
			}
		}
	} else {
		zap.L().Info("数据库无AI配置，使用配置文件默认值")
	}

	// 初始化规则引擎服务
	engineSvc := service.NewEngineService(ais)

	// 启动抑制告警定时发送任务
	scheduler.StartSuppressedSender(engineSvc)

	router := api.SetupRouter(ais)

	// Swagger 路由
	if config.Conf.Mode == "dev" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
	addr := fmt.Sprintf(":%d", config.Conf.AppInfo.Port)
	zap.L().Info("服务器启动", zap.String("addr", addr))
	if err := router.Run(addr); err != nil {
		zap.L().Fatal("服务器启动失败", zap.Error(err))
	}

}

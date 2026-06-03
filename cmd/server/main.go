package main

import (
	"alert-ops/internal/api"
	"alert-ops/internal/config"
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
	// 使用命令行参数指定配置文件
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

	// 启动抑制告警定时发送任务（不依赖 engineSvc，直接操作 DB）
	scheduler.StartSuppressedSender()

	router := api.SetupRouter()

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

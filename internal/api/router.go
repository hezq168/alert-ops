package api

import (
	alertHandler "alert-ops/internal/alert/handler"
	"alert-ops/internal/alert/service"
	"alert-ops/internal/middleware"
	userHandler "alert-ops/internal/user/handler"
	logger "alert-ops/pkg"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	r.RedirectTrailingSlash = false
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 静态文件服务 - 托管前端构建产物
	r.Static("/assets", "./web/dist/assets")
	r.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		uHandler := userHandler.NewUserHandler()
		rHandler := userHandler.NewRoleHandler()

		// ======================
		// 告警 Webhook 路由（公开，无需登录）
		// ======================
		engineSvc := service.NewEngineService()
		wHandler := alertHandler.NewWebhookHandler(engineSvc)
		webhook := v1.Group("/webhook")
		{
			// POST /api/v1/webhook/alertmanager/:slug
			webhook.POST("/alertmanager/:slug", wHandler.ReceiveAlertmanager)
			// POST /api/v1/webhook/:type/:slug （预留扩展）
			webhook.POST("/:type/:slug", wHandler.ReceiveWebhook)
		}

		// 公开路由（不需要登录）
		auth := v1.Group("/auth")
		/*
			/api/v1/auth/register - 注册（公开）
			/api/v1/auth/login - 登录（公开）
		*/
		{
			auth.POST("/register", uHandler.Register)
			auth.POST("/login", uHandler.Login)
			auth.GET("/captcha", uHandler.GetCaptcha)
		}

		// 需要认证的路由
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth())
		protected.Use(middleware.PermissionMiddleware())
		/*
			/api/v1/user/info - 获取当前用户信息（需要登录）
			/api/v1/users - 用户列表（需要登录）
			/api/v1/users/:id - 更新/删除用户（需要登录）
		*/
		{
			// ======================
			// 用户模块
			// ======================
			protected.GET("/user/info", uHandler.GetUserInfo)
			protected.GET("/user/menus", uHandler.GetUserMenus)
			protected.POST("/user/change-password", uHandler.ChangePassword)
			protected.GET("/users", uHandler.ListUsers)
			protected.PUT("/users/:id", uHandler.UpdateUser)
			protected.DELETE("/users/:id", uHandler.DeleteUser)

			// 用户状态修改
			protected.PUT("/users/:id/status", uHandler.UpdateUserStatus)
			// 分配角色
			protected.POST("/users/:id/roles", rHandler.AssignRole)
			// 删除角色
			protected.DELETE("/users/:id/roles", rHandler.RemoveUserRole)
			// 获取单个角色
			protected.GET("/users/:id/roles", rHandler.GetUserRoles)

			// ======================
			// 角色模块
			// ======================
			protected.GET("/roles", rHandler.ListRoles)
			// 创建
			protected.POST("/roles", rHandler.CreateRole)
			// 删除
			protected.DELETE("/roles/:id", rHandler.DeleteRole)
			// 获取用户权限
			protected.GET("/roles/:id/permissions", rHandler.GetRolePermissions)
			// 分配权限
			protected.POST("/roles/:id/permissions", rHandler.AssignPermissions)

			// ======================
			// 权限模块
			// ======================
			// 获取所有权限
			protected.GET("/permissions", rHandler.ListPermissions)
			// 创建权限
			protected.POST("/permissions", rHandler.CreatePermission)
			// 删创权限
			protected.DELETE("/permissions/:id", rHandler.DeletePermission)
			// 更新权限
			protected.PUT("/permissions/:id", rHandler.UpdatePermission)

			// ======================
			// 告警模块 - 告警源管理
			// ======================
			srcHandler := alertHandler.NewSourceHandler()
			protected.GET("/alert-sources", srcHandler.ListSources)
			protected.POST("/alert-sources", srcHandler.CreateSource)
			protected.GET("/alert-sources/:id", srcHandler.GetSource)
			protected.PUT("/alert-sources/:id", srcHandler.UpdateSource)
			protected.DELETE("/alert-sources/:id", srcHandler.DeleteSource)

			// ======================
			// 告警模块 - 规则管理
			// ======================
			ruleHandler := alertHandler.NewRuleHandler()
			protected.GET("/alert-rules", ruleHandler.ListRules)
			protected.POST("/alert-rules", ruleHandler.CreateRule)
			protected.GET("/alert-rules/:id", ruleHandler.GetRule)
			protected.PUT("/alert-rules/:id", ruleHandler.UpdateRule)
			protected.DELETE("/alert-rules/:id", ruleHandler.DeleteRule)

			// ======================
			// 告警模块 - 模板管理
			// ======================
			tplHandler := alertHandler.NewTemplateHandler()
			protected.GET("/alert-templates", tplHandler.ListTemplates)
			protected.POST("/alert-templates", tplHandler.CreateTemplate)
			protected.GET("/alert-templates/:id", tplHandler.GetTemplate)
			protected.PUT("/alert-templates/:id", tplHandler.UpdateTemplate)
			protected.DELETE("/alert-templates/:id", tplHandler.DeleteTemplate)

			// ======================
			// 告警模块 - 通道管理
			// ======================
			chHandler := alertHandler.NewChannelHandler()
			protected.GET("/alert-channels", chHandler.ListChannels)
			protected.POST("/alert-channels", chHandler.CreateChannel)
			protected.GET("/alert-channels/:id", chHandler.GetChannel)
			protected.PUT("/alert-channels/:id", chHandler.UpdateChannel)
			protected.DELETE("/alert-channels/:id", chHandler.DeleteChannel)
			protected.POST("/alert-channels/:id/test", chHandler.TestSend)

			// ======================
			// 告警模块 - 流水查询
			// ======================
			recHandler := alertHandler.NewRecordHandler(engineSvc)
			protected.GET("/alert-records", recHandler.ListRecords)
			protected.POST("/alert-records/:id/analyze", recHandler.AnalyzeRecord)

			// ======================
			// 告警模块 - AI 配置
			// ======================
			aiCfgHandler := alertHandler.NewAIConfigHandler()
			protected.GET("/ai-config", aiCfgHandler.GetConfig)
			protected.PUT("/ai-config", aiCfgHandler.UpdateConfig)
			protected.POST("/ai-config/test", aiCfgHandler.TestConnection)

		}
	}

	// SPA 回退：非 API 路由返回 index.html（前端路由接管）
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// API 路由返回 404 JSON
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"msg": "404",
			})
			return
		}
		// 其他路由返回 index.html（SPA）
		c.File("./web/dist/index.html")
	})

	return r
}

package middleware

import (
	"alert-ops/internal/repo"
	"alert-ops/internal/user/service"
	"go.uber.org/zap"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func PermissionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("user_id")
		// 超级管理员直接放行
		if userID == 1 {
			c.Next()
			return
		}
		// 当前请求
		method := c.Request.Method
		// 一定用 FullPath
		path := c.FullPath()
		// 去掉 /api/v1
		path = strings.TrimPrefix(path, "/api/v1")

		zap.L().Debug("请求权限", zap.String("method", method), zap.String("path", path))

		// 查当前接口权限
		var permissionCode string

		err := repo.DB.Table("permissions").Select("code").
			Where("api_method = ? AND api_path = ? AND status = 1", method, path).
			Scan(&permissionCode).Error

		zap.L().Debug("接口权限", zap.String("permissionCode", permissionCode))

		// 接口未配置权限
		if err != nil || permissionCode == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"msg": "接口未配置权限",
			})
			c.Abort()
			return
		}

		// 查询用户权限
		userService := service.NewUserService()
		permissions, err := userService.GetUserPermissions(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "权限查询失败",
			})
			c.Abort()
			return
		}

		// 判断权限
		for _, p := range permissions {
			if p.Code == permissionCode {
				c.Next()
				return
			}
		}
		zap.L().Debug("用户权限", zap.Any("permissions", permissions))
		zap.L().Debug("用户权限", zap.String("permissionCode", permissionCode))
		zap.L().Debug("用户权限", zap.String("method", method), zap.String("path", path))

		c.JSON(http.StatusForbidden, gin.H{
			"msg": "无权限",
		})
		c.Abort()
	}
}

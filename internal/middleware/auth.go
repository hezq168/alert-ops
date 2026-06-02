package middleware

import (
	"alert-ops/pkg/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 优先从 URL 参数获取 Token（WebSocket 场景）
		tokenString := c.Query("token")

		// 2. 如果 URL 没有，再从请求头获取（常规 HTTP 请求）
		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    401,
					"message": "未提供认证令牌",
				})
				c.Abort()
				return
			}

			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				tokenString = authHeader[7:]
			} else {
				tokenString = authHeader
			}
		}

		// 3. 解析 Token
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 4. 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		// 5. 继续处理请求
		c.Next()
	}
}

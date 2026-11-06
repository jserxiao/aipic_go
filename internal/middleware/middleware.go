package middleware

import (
	"os"
	"strings"

	"aipic_go/internal/config"
	"aipic_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// Auth 认证中间件
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorUnauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}

		// Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorUnauthorized(c, "认证格式错误")
			c.Abort()
			return
		}

		// 限制 token 长度，防止超长 token 攻击
		if len(parts[1]) > 2048 {
			utils.ErrorUnauthorized(c, "token 长度异常")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.ErrorUnauthorized(c, "token 无效或已过期")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// OptionalAuth 可选认证中间件（有 token 则解析，没有也通过）
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" && len(parts[1]) <= 2048 {
				if claims, err := utils.ParseToken(parts[1]); err == nil {
					c.Set("userID", claims.UserID)
				}
			}
		}
		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	allowedOrigin := os.Getenv("CORS_ORIGIN")
	devMode := config.Get().DevMode

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		if devMode {
			// 开发模式：允许所有来源
			c.Header("Access-Control-Allow-Origin", "*")
		} else if allowedOrigin != "" {
			// 生产模式：只允许配置的来源
			if origin == allowedOrigin {
				c.Header("Access-Control-Allow-Origin", origin)
			}
		} else {
			// 生产模式且未配置 CORS_ORIGIN：拒绝跨域
			c.Header("Access-Control-Allow-Origin", "null")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// GetUserID 从上下文获取 userID
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return "" // 不应该走到这里，调用方应确保在 Auth 中间件之后
	}
	if id, ok := userID.(string); ok {
		return id
	}
	return ""
}

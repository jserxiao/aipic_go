package handler

import (
	"aipic_go/internal/middleware"
	"aipic_go/internal/service"
	"aipic_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct{}

// Login 微信登录
// POST /auth/login
// Body: { "code": "wx_login_code" }
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBadRequest(c, "缺少 code 参数")
		return
	}

	token, err := service.Login(req.Code)
	if err != nil {
		utils.ErrorServerError(c, err.Error())
		return
	}

	utils.Success(c, token)
}

// Refresh 刷新 token
// POST /auth/refresh
// Body: { "refreshToken": "xxx" }
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBadRequest(c, "缺少 refreshToken 参数")
		return
	}

	token, err := service.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.ErrorUnauthorized(c, err.Error())
		return
	}

	utils.Success(c, token)
}

// Logout 退出登录
// POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 客户端清除 token 即可，服务端可选维护黑名单
	utils.SuccessEmpty(c)
}

// GetUserID 从上下文获取 userID
func GetUserID(c *gin.Context) string {
	return middleware.GetUserID(c)
}

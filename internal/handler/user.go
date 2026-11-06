package handler

import (
	"aipic_go/internal/service"
	"aipic_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct{}

// GetUserInfo 获取用户信息
// GET /user/info
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	userID := GetUserID(c)

	user, err := service.GetUserInfo(userID)
	if err != nil {
		utils.ErrorNotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// UpdateAvatar 更新用户头像
// PUT /user/avatar
// Body: { "avatar": "https://..." }
func (h *UserHandler) UpdateAvatar(c *gin.Context) {
	userID := GetUserID(c)

	var req struct {
		Avatar string `json:"avatar" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBadRequest(c, "缺少 avatar 参数")
		return
	}

	// 校验头像 URL 长度
	if len(req.Avatar) > 2048 {
		utils.ErrorBadRequest(c, "头像 URL 长度异常")
		return
	}

	user, err := service.UpdateUserAvatar(userID, req.Avatar)
	if err != nil {
		utils.ErrorServerError(c, err.Error())
		return
	}

	utils.Success(c, user)
}

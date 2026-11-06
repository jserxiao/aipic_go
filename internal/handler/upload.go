package handler

import (
	"io"
	"path/filepath"

	"aipic_go/internal/config"
	"aipic_go/internal/service"
	"aipic_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// UploadHandler 上传处理器
type UploadHandler struct{}

// UploadImage 上传图片
// POST /upload/image
// Form: file (multipart/form-data)
func (h *UploadHandler) UploadImage(c *gin.Context) {
	userID := GetUserID(c)

	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorBadRequest(c, "缺少文件")
		return
	}

	// 校验文件大小
	cfg := config.Get()
	if file.Size > cfg.MaxUploadSize {
		utils.ErrorBadRequest(c, "图片大小不能超过 5MB")
		return
	}

	// 校验文件类型（扩展名）
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".bmp": true,
	}
	if !allowedExts[ext] {
		utils.ErrorBadRequest(c, "不支持的图片格式")
		return
	}

	// 读取文件内容
	src, err := file.Open()
	if err != nil {
		utils.ErrorServerError(c, "文件读取失败")
		return
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		utils.ErrorServerError(c, "文件读取失败")
		return
	}

	result, err := service.SaveUploadFile(data, file.Filename, userID)
	if err != nil {
		utils.ErrorBadRequest(c, err.Error())
		return
	}

	utils.Success(c, result)
}

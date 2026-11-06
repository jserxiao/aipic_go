package service

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aipic_go/internal/config"
)

// UploadResult 上传结果
type UploadResult struct {
	URL string `json:"url"`
}

// 允许的文件扩展名
var allowedExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".gif": true, ".webp": true, ".bmp": true,
}

// SaveUploadFile 保存上传的文件
func SaveUploadFile(fileBytes []byte, originalFilename string, userID string) (*UploadResult, error) {
	cfg := config.Get()

	// 校验文件扩展名
	ext := strings.ToLower(filepath.Ext(originalFilename))
	if !allowedExtensions[ext] {
		return nil, fmt.Errorf("不支持的图片格式")
	}

	// 校验真实文件类型（magic bytes）
	contentType := http.DetectContentType(fileBytes)
	if !isAllowedImageType(contentType) {
		return nil, fmt.Errorf("文件内容不是有效图片")
	}

	// 确保上传目录存在
	dateDir := time.Now().Format("2006/01/02")
	uploadPath := filepath.Join(cfg.UploadDir, dateDir)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 生成唯一文件名（使用 userID + 时间戳，避免冲突）
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), userID, ext)
	filePath := filepath.Join(uploadPath, filename)

	// 写入文件
	if err := os.WriteFile(filePath, fileBytes, 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	// 返回可访问的 URL 路径
	url := fmt.Sprintf("/uploads/%s/%s", dateDir, filename)

	return &UploadResult{URL: url}, nil
}

// isAllowedContentType 检查 content type 是否为允许的图片类型
func isAllowedImageType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"image/bmp":  true,
	}
	return allowedTypes[contentType]
}

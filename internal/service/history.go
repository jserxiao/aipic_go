package service

import (
	"fmt"
	"strconv"
	"strings"

	"aipic_go/internal/model"
	"aipic_go/internal/store"
	"aipic_go/internal/utils"
)

// GetHistory 获取历史记录（分页）
func GetHistory(userID string, page, pageSize int) ([]*model.GenerateRecord, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return store.GetGenerateRecords(userID, page, pageSize)
}

// DeleteHistoryRecord 删除单条历史记录（带越权校验）
func DeleteHistoryRecord(recordID string, userID string) error {
	return store.DeleteGenerateRecord(recordID, userID)
}

// ClearHistory 清空用户所有历史记录
func ClearHistory(userID string) error {
	return store.DeleteAllGenerateRecords(userID)
}

// SaveGenerateRecord 保存生成记录
func SaveGenerateRecord(userID string, task *model.GenerateTask) error {
	// 构建分辨率字符串
	resolution := task.Resolution
	if task.Resolution == "custom" {
		resolution = fmt.Sprintf("%dx%d", task.CustomWidth, task.CustomHeight)
	}

	record := &model.GenerateRecord{
		RecordID:    utils.GenerateRecordID(),
		UserID:      userID,
		Prompt:      task.Prompt,
		Ratio:       task.Ratio,
		Resolution:  resolution,
		ImageURL:    task.ResultURL,
		Status:      "success",
		CreditsCost: task.CreditsCost,
	}

	return store.CreateGenerateRecord(record)
}

// GetTodayGenerateCount 获取今日生成次数
func GetTodayGenerateCount(userID string) (int64, error) {
	return store.GetTodayGenerateCount(userID)
}

// ParseResolution 解析分辨率字符串为宽高
func ParseResolution(resolution string) (int, int) {
	parts := strings.Split(resolution, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	w, _ := strconv.Atoi(parts[0])
	h, _ := strconv.Atoi(parts[1])
	return w, h
}

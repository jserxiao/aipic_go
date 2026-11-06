package service

import (
	"fmt"
	"time"

	"aipic_go/internal/model"
	"aipic_go/internal/store"
	"aipic_go/internal/utils"
)

// 比例对应尺寸映射
var ratioSizeMap = map[string]struct{ W, H int }{
	"1:1":  {1024, 1024},
	"3:4":  {768, 1024},
	"4:3":  {1024, 768},
	"16:9": {1280, 720},
	"9:16": {720, 1280},
}

// 状态文案
var statusTextMap = map[string]string{
	"queued":     "排队中...",
	"processing": "生成中...",
	"succeed":    "生成完成",
	"failed":     "生成失败",
}

// SubmitGenerateTask 提交生成任务
// 注意：积分扣减应在调用方完成，此处仅创建任务
func SubmitGenerateTask(userID string, config *model.GenerateTask) (string, error) {
	// 设置任务ID和用户ID
	config.TaskID = utils.GenerateTaskID()
	config.UserID = userID
	config.Status = "queued"
	config.Progress = 0

	// 保存到数据库
	if err := store.CreateGenerateTask(config); err != nil {
		return "", fmt.Errorf("创建任务失败: %w", err)
	}

	// TODO: 对接真实 AI 模型（如通义千问 qwen-image-3.0）
	// 当前为模拟异步生成，后台协程模拟进度推进
	go mockGenerateProcess(config.TaskID, config.Ratio)

	return config.TaskID, nil
}

// GetTaskProgress 查询任务进度（带越权校验）
func GetTaskProgress(taskID string, userID string) (*model.AsyncTaskProgress, error) {
	task, err := store.GetGenerateTaskByTaskID(taskID)
	if err != nil {
		return &model.AsyncTaskProgress{
			TaskID:       taskID,
			Status:       "failed",
			Progress:     0,
			StatusText:   statusTextMap["failed"],
			ErrorMessage: "任务不存在",
		}, nil
	}

	// 越权校验：只能查看自己的任务
	if task.UserID != userID {
		return &model.AsyncTaskProgress{
			TaskID:       taskID,
			Status:       "failed",
			Progress:     0,
			StatusText:   statusTextMap["failed"],
			ErrorMessage: "任务不存在",
		}, nil
	}

	progress := &model.AsyncTaskProgress{
		TaskID:     task.TaskID,
		Status:     task.Status,
		Progress:   task.Progress,
		StatusText: statusTextMap[task.Status],
	}

	if task.Status == "succeed" {
		progress.ResultURL = task.ResultURL
	}

	if task.Status == "failed" {
		progress.ErrorMessage = task.ErrorMessage
	}

	return progress, nil
}

// CancelGenerateTask 取消生成任务（带越权校验）
func CancelGenerateTask(taskID string, userID string) error {
	task, err := store.GetGenerateTaskByTaskID(taskID)
	if err != nil {
		return fmt.Errorf("任务不存在")
	}

	// 越权校验：只能取消自己的任务
	if task.UserID != userID {
		return fmt.Errorf("任务不存在")
	}

	return store.DeleteGenerateTask(taskID)
}

// mockGenerateProcess 模拟异步生成过程
func mockGenerateProcess(taskID string, ratio string) {
	// 获取比例尺寸
	size, ok := ratioSizeMap[ratio]
	if !ok {
		size = ratioSizeMap["1:1"]
	}

	// 模拟生成过程：5秒完成
	totalDuration := 5 * time.Second
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()
	resultURL := fmt.Sprintf("https://placehold.co/%dx%d/000000/FFFFFF?text=AI+Generated", size.W, size.H)

	for {
		elapsed := time.Since(startTime)
		if elapsed >= totalDuration {
			// 完成
			store.UpdateGenerateTaskStatus(taskID, "succeed", 100, resultURL, "")
			// 保存生成记录
			saveRecordFromTask(taskID)
			break
		}

		progress := int((float64(elapsed) / float64(totalDuration)) * 100)
		if progress < 5 {
			progress = 5
		}

		store.UpdateGenerateTaskStatus(taskID, "processing", progress, "", "")

		select {
		case <-ticker.C:
			continue
		case <-time.After(totalDuration - elapsed):
			store.UpdateGenerateTaskStatus(taskID, "succeed", 100, resultURL, "")
			saveRecordFromTask(taskID)
			return
		}
	}
}

// saveRecordFromTask 从任务创建生成记录
func saveRecordFromTask(taskID string) {
	task, err := store.GetGenerateTaskByTaskID(taskID)
	if err != nil || task == nil {
		return
	}
	if task.Status != "succeed" {
		return
	}
	_ = SaveGenerateRecord(task.UserID, task)
}

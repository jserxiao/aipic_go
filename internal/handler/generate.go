package handler

import (
	"aipic_go/internal/model"
	"aipic_go/internal/service"
	"aipic_go/internal/store"
	"aipic_go/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GenerateHandler 生成处理器
type GenerateHandler struct{}

// SubmitTask 提交生图任务
// POST /generate/submit
func (h *GenerateHandler) SubmitTask(c *gin.Context) {
	userID := GetUserID(c)

	var req struct {
		Prompt         string `json:"prompt" binding:"required"`
		Ratio          string `json:"ratio" binding:"required"`
		Resolution     string `json:"resolution" binding:"required"`
		Model          string `json:"model" binding:"required"`
		CustomWidth    int    `json:"customWidth"`
		CustomHeight   int    `json:"customHeight"`
		ReferenceImage string `json:"referenceImage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 校验文字描述长度
	if len(req.Prompt) < 2 {
		utils.ErrorBadRequest(c, "请输入至少 2 个字的描述")
		return
	}
	if len(req.Prompt) > 500 {
		utils.ErrorBadRequest(c, "描述文字不能超过 500 字")
		return
	}

	// 校验 ratio 和 model 合法性（白名单）
	if !isValidRatio(req.Ratio) {
		utils.ErrorBadRequest(c, "无效的图片比例")
		return
	}
	if !isValidModel(req.Model) {
		utils.ErrorBadRequest(c, "无效的 AI 模型")
		return
	}

	// 获取模型消耗积分
	creditsCost := getModelCreditsCost(req.Model)

	// 校验积分
	user, err := service.GetUserInfo(userID)
	if err != nil {
		utils.ErrorNotFound(c, "用户不存在")
		return
	}
	if user.Credits < int64(creditsCost) {
		utils.Error(c, utils.CodeBadRequest, "积分不足，当前剩余 "+strconv.FormatInt(user.Credits, 10)+"，需要 "+strconv.Itoa(creditsCost))
		return
	}

	// 校验今日生成次数
	todayCount, err := service.GetTodayGenerateCount(userID)
	if err == nil && todayCount >= 30 {
		utils.ErrorBadRequest(c, "今日生成次数已达上限（30张），明天再试吧")
		return
	}

	// 先扣除积分（事务安全）
	newCredits, err := store.DeductUserCredits(userID, int64(creditsCost))
	if err != nil {
		utils.Error(c, utils.CodeBadRequest, err.Error())
		return
	}

	// 再创建生成任务
	task := &model.GenerateTask{
		Prompt:         req.Prompt,
		Ratio:          req.Ratio,
		Resolution:     req.Resolution,
		Model:          req.Model,
		CustomWidth:    req.CustomWidth,
		CustomHeight:   req.CustomHeight,
		ReferenceImage: req.ReferenceImage,
		CreditsCost:    creditsCost,
	}

	taskID, err := service.SubmitGenerateTask(userID, task)
	if err != nil {
		// 任务创建失败，退还积分
		store.AddUserCredits(userID, int64(creditsCost))
		utils.ErrorServerError(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"taskId":  taskID,
		"credits": newCredits,
	})
}

// GetTaskProgress 查询任务状态
// GET /generate/task/:taskId
func (h *GenerateHandler) GetTaskProgress(c *gin.Context) {
	userID := GetUserID(c)
	taskID := c.Param("taskId")
	if taskID == "" {
		utils.ErrorBadRequest(c, "缺少 taskId")
		return
	}

	progress, err := service.GetTaskProgress(taskID, userID)
	if err != nil {
		utils.ErrorServerError(c, err.Error())
		return
	}

	utils.Success(c, progress)
}

// CancelTask 取消生成任务
// POST /generate/task/:taskId/cancel
func (h *GenerateHandler) CancelTask(c *gin.Context) {
	userID := GetUserID(c)
	taskID := c.Param("taskId")
	if taskID == "" {
		utils.ErrorBadRequest(c, "缺少 taskId")
		return
	}

	if err := service.CancelGenerateTask(taskID, userID); err != nil {
		utils.ErrorBadRequest(c, err.Error())
		return
	}

	utils.SuccessEmpty(c)
}

// GetHistory 获取历史记录
// GET /generate/history?page=1&pageSize=20
func (h *GenerateHandler) GetHistory(c *gin.Context) {
	userID := GetUserID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	records, total, err := service.GetHistory(userID, page, pageSize)
	if err != nil {
		utils.ErrorServerError(c, err.Error())
		return
	}

	utils.PaginatedSuccess(c, records, total, page, pageSize)
}

// DeleteHistoryRecord 删除单条历史记录
// DELETE /generate/history/:recordId
func (h *GenerateHandler) DeleteHistoryRecord(c *gin.Context) {
	userID := GetUserID(c)
	recordID := c.Param("recordId")
	if recordID == "" {
		utils.ErrorBadRequest(c, "缺少 recordId")
		return
	}

	if err := service.DeleteHistoryRecord(recordID, userID); err != nil {
		utils.ErrorBadRequest(c, err.Error())
		return
	}

	utils.SuccessEmpty(c)
}

// ClearHistory 清空历史记录
// DELETE /generate/history
func (h *GenerateHandler) ClearHistory(c *gin.Context) {
	userID := GetUserID(c)

	if err := service.ClearHistory(userID); err != nil {
		utils.ErrorServerError(c, err.Error())
		return
	}

	utils.SuccessEmpty(c)
}

// isValidRatio 校验比例是否合法
func isValidRatio(ratio string) bool {
	validRatios := map[string]bool{
		"1:1": true, "3:4": true, "4:3": true,
		"16:9": true, "9:16": true,
	}
	return validRatios[ratio]
}

// isValidModel 校验模型是否合法
func isValidModel(model string) bool {
	validModels := map[string]bool{
		"qwen-image-3.0": true,
	}
	return validModels[model]
}

// getModelCreditsCost 根据模型获取消耗积分
func getModelCreditsCost(model string) int {
	modelCreditsMap := map[string]int{
		"qwen-image-3.0": 18,
	}
	if cost, ok := modelCreditsMap[model]; ok {
		return cost
	}
	return 10 // 默认
}

package handler

import (
	"aipic_go/internal/config"
	"aipic_go/internal/service"
	"aipic_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// PaymentHandler 支付处理器
type PaymentHandler struct{}

// CreateRechargeOrder 创建积分充值订单
// POST /payment/recharge/create
// Body: { "amount": 10 }  // amount 单位为元
func (h *PaymentHandler) CreateRechargeOrder(c *gin.Context) {
	userID := GetUserID(c)

	var req struct {
		Amount int64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBadRequest(c, "缺少 amount 参数")
		return
	}

	if req.Amount <= 0 {
		utils.ErrorBadRequest(c, "充值金额必须大于 0")
		return
	}

	order, err := service.CreateRechargeOrder(userID, req.Amount)
	if err != nil {
		utils.ErrorBadRequest(c, err.Error())
		return
	}

	utils.Success(c, order)
}

// CreateVipOrder 创建 VIP 订阅订单
// POST /payment/vip/create
// Body: { "planType": "monthly" }
func (h *PaymentHandler) CreateVipOrder(c *gin.Context) {
	userID := GetUserID(c)

	var req struct {
		PlanType string `json:"planType" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBadRequest(c, "缺少 planType 参数")
		return
	}

	order, err := service.CreateVipOrder(userID, req.PlanType)
	if err != nil {
		utils.ErrorBadRequest(c, err.Error())
		return
	}

	utils.Success(c, order)
}

// QueryOrderStatus 查询订单支付状态
// GET /payment/order/:orderId
// 只能查询自己的订单
func (h *PaymentHandler) QueryOrderStatus(c *gin.Context) {
	userID := GetUserID(c)
	orderID := c.Param("orderId")
	if orderID == "" {
		utils.ErrorBadRequest(c, "缺少 orderId")
		return
	}

	result, err := service.QueryOrderStatus(orderID, userID)
	if err != nil {
		utils.ErrorServerError(c, err.Error())
		return
	}

	utils.Success(c, result)
}

// MockPayConfirm 模拟支付确认
// POST /payment/mock-confirm
// Body: { "orderId": "xxx" }
// 在开发模式或支付 Mock 模式下可用，模拟微信支付回调确认
func (h *PaymentHandler) MockPayConfirm(c *gin.Context) {
	cfg := config.Get()
	if !cfg.DevMode && !cfg.PayMockMode {
		utils.Error(c, utils.CodeForbidden, "该接口仅在开发/Mock支付模式可用")
		return
	}

	userID := GetUserID(c)

	var req struct {
		OrderID string `json:"orderId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorBadRequest(c, "缺少 orderId 参数")
		return
	}

	// 越权校验：确认订单属于当前用户
	result, err := service.QueryOrderStatus(req.OrderID, userID)
	if err != nil || result.Status == "closed" {
		utils.ErrorBadRequest(c, "订单不存在")
		return
	}

	if result.Status == "paid" {
		utils.Success(c, result)
		return
	}

	// 模拟支付确认
	if err := service.ConfirmPayment(req.OrderID); err != nil {
		utils.ErrorServerError(c, err.Error())
		return
	}

	// 返回最新状态
	updated, _ := service.QueryOrderStatus(req.OrderID, userID)
	utils.Success(c, updated)
}

// WxPayNotify 微信支付回调通知
// POST /payment/notify
// 生产环境：由微信支付服务器调用，需要验签
// 开发环境：模拟回调
func (h *PaymentHandler) WxPayNotify(c *gin.Context) {
	cfg := config.Get()

	if cfg.DevMode || cfg.PayMockMode {
		// 开发模式：直接从 URL 参数获取订单 ID 并确认
		orderID := c.Query("order_id")
		if orderID == "" {
			// 尝试从 body 读取
			var req struct {
				OrderID string `json:"order_id"`
			}
			if err := c.ShouldBindJSON(&req); err == nil && req.OrderID != "" {
				orderID = req.OrderID
			}
		}
		if orderID == "" {
			utils.ErrorBadRequest(c, "缺少 order_id")
			return
		}

		if err := service.ConfirmPayment(orderID); err != nil {
			utils.ErrorServerError(c, err.Error())
			return
		}

		utils.SuccessEmpty(c)
		return
	}

	// TODO: 生产环境实现微信支付回调验签
	// 1. 读取请求 body（微信发送的 XML/JSON 数据）
	// 2. 验证签名（使用微信支付平台证书）
	// 3. 解密回调数据，获取订单号和交易状态
	// 4. 确认支付成功后调用 service.ConfirmPayment(orderID)
	// 5. 返回符合微信要求的响应

	// body, err := io.ReadAll(c.Request.Body)
	// if err != nil {
	//     utils.ErrorBadRequest(c, "读取请求失败")
	//     return
	// }
	// // 验签 & 解密...
	// // 确认支付
	// if err := service.ConfirmPayment(orderID); err != nil {
	//     utils.ErrorServerError(c, err.Error())
	//     return
	// }

	utils.ErrorBadRequest(c, "支付回调验签未实现，请配置微信支付证书")
}

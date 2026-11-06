package service

import (
	"fmt"
	"time"

	"aipic_go/internal/model"
	"aipic_go/internal/store"
	"aipic_go/internal/utils"
)

// VIP 套餐配置
type VipPlan struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	Price        int64  `json:"price"`        // 分
	BonusCredits int64  `json:"bonusCredits"`
	Description  string `json:"description"`
	Duration     int    `json:"duration"` // 月数
}

// 预设 VIP 套餐
var vipPlans = map[string]*VipPlan{
	"monthly": {
		Type:         "monthly",
		Name:         "月卡",
		Price:        4900, // 49 元 = 4900 分
		BonusCredits: 1000,
		Description:  "每月自动续费",
		Duration:     1,
	},
	"quarterly": {
		Type:         "quarterly",
		Name:         "季卡",
		Price:        13900, // 139 元 = 13900 分
		BonusCredits: 3000,
		Description:  "每季自动续费",
		Duration:     3,
	},
	"yearly": {
		Type:         "yearly",
		Name:         "年卡",
		Price:        39900, // 399 元 = 39900 分
		BonusCredits: 10000,
		Description:  "每年自动续费",
		Duration:     12,
	},
}

// 充值金额限制
const (
	MinRechargeAmount int64 = 1     // 最低 1 元
	MaxRechargeAmount int64 = 10000 // 最高 10000 元
)

// CreateRechargeOrder 创建积分充值订单
// amount: 充值金额（元）
func CreateRechargeOrder(userID string, amount int64) (*model.PaymentOrderResponse, error) {
	if amount < MinRechargeAmount {
		return nil, fmt.Errorf("充值金额不能少于 %d 元", MinRechargeAmount)
	}
	if amount > MaxRechargeAmount {
		return nil, fmt.Errorf("充值金额不能超过 %d 元", MaxRechargeAmount)
	}

	// 1元 = 10积分
	credits := amount * 10

	order := &model.PaymentOrder{
		OrderID:   utils.GenerateOrderID("order"),
		UserID:    userID,
		OrderType: "recharge",
		Amount:    amount * 100, // 转为分
		Credits:   credits,
		Status:    "pending",
	}

	if err := store.CreatePaymentOrder(order); err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// TODO: 对接微信支付，生成真实支付参数
	// 当前返回 mock 支付参数（开发模式下使用）
	return buildOrderResponse(order), nil
}

// CreateVipOrder 创建 VIP 订阅订单
func CreateVipOrder(userID string, planType string) (*model.PaymentOrderResponse, error) {
	plan, ok := vipPlans[planType]
	if !ok {
		return nil, fmt.Errorf("无效的套餐类型: %s", planType)
	}

	order := &model.PaymentOrder{
		OrderID:   utils.GenerateOrderID("vip_order"),
		UserID:    userID,
		OrderType: "vip",
		Amount:    plan.Price,
		Credits:   plan.BonusCredits,
		Status:    "pending",
		PlanType:  planType,
	}

	if err := store.CreatePaymentOrder(order); err != nil {
		return nil, fmt.Errorf("创建订单失败: %w", err)
	}

	// TODO: 对接微信支付，生成真实支付参数
	return buildOrderResponse(order), nil
}

// QueryOrderStatus 查询订单支付状态（带越权校验）
func QueryOrderStatus(orderID string, userID string) (*model.PaymentResult, error) {
	order, err := store.GetPaymentOrderByOrderIDAndUser(orderID, userID)
	if err != nil {
		return &model.PaymentResult{
			OrderID: orderID,
			Status:  "closed",
		}, nil
	}

	return &model.PaymentResult{
		OrderID: orderID,
		Status:  order.Status,
		Credits: order.Credits,
	}, nil
}

// ConfirmPayment 确认支付成功
// 生产环境：由微信支付服务器回调调用，需要验签
// 开发环境：模拟回调
func ConfirmPayment(orderID string) error {
	// 查询订单
	order, err := store.GetPaymentOrderByOrderID(orderID)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}

	// 计算 VIP 到期时间（如有）
	var vipPlanType string
	var vipExpireDate string
	if order.OrderType == "vip" {
		vipPlanType = order.PlanType
		vipExpireDate = calculateVipExpireDate(order.UserID, order.PlanType)
	}

	// 事务内完成：更新订单状态 + 增加积分 + 更新 VIP 状态
	if err := store.ConfirmPaymentInTransaction(orderID, vipPlanType, vipExpireDate); err != nil {
		return err
	}

	return nil
}

// calculateVipExpireDate 计算 VIP 到期时间
// 如果用户已有 VIP 且未过期，则在当前到期时间基础上续费
func calculateVipExpireDate(userID string, planType string) string {
	plan, ok := vipPlans[planType]
	if !ok {
		// 默认 1 个月
		now := time.Now().AddDate(0, 1, 0)
		return now.Format("2006-01-02T15:04:05.000Z")
	}

	// 检查用户当前 VIP 到期时间
	currentExpire, err := store.GetUserVipExpireAt(userID)
	if err == nil && currentExpire != nil && currentExpire.After(time.Now()) {
		// 当前 VIP 未过期，在现有基础上续费
		newExpire := currentExpire.AddDate(0, plan.Duration, 0)
		return newExpire.Format("2006-01-02T15:04:05.000Z")
	}

	// 当前无 VIP 或已过期，从现在开始计算
	now := time.Now().AddDate(0, plan.Duration, 0)
	return now.Format("2006-01-02T15:04:05.000Z")
}

// buildOrderResponse 构建订单响应（mock 支付参数）
func buildOrderResponse(order *model.PaymentOrder) *model.PaymentOrderResponse {
	return &model.PaymentOrderResponse{
		OrderID: order.OrderID,
		PayParams: model.PayParams{
			TimeStamp: fmt.Sprintf("%d", time.Now().Unix()),
			NonceStr:  utils.GenerateRandomID(),
			Package:   "prepay_id=mock_prepay",
			SignType:  "MD5",
			PaySign:   "mock_sign",
		},
		Amount:  order.Amount,
		Credits: order.Credits,
	}
}

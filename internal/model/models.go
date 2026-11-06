package model

import "time"

// User 用户模型
type User struct {
	ID          int64     `json:"-" db:"id"`
	UserID      string    `json:"userId" db:"user_id"`       // 对外展示的用户ID
	RandomID    string    `json:"randomId" db:"random_id"`   // 随机展示ID，如 AP_8xK2m9
	Avatar      string    `json:"avatar" db:"avatar"`         // 头像URL
	IsVip       bool      `json:"isVip" db:"is_vip"`          // 是否为VIP
	VipLevel    string    `json:"vipLevel" db:"vip_level"`    // 会员等级: normal / vip
	VipExpireAt *string   `json:"vipExpireAt" db:"vip_expire_at"` // VIP到期时间（ISO字符串）
	Credits     int64     `json:"credits" db:"credits"`       // 剩余积分
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

// GenerateTask 生成任务模型
type GenerateTask struct {
	ID             int64     `json:"-" db:"id"`
	TaskID         string    `json:"taskId" db:"task_id"`         // 任务唯一标识
	UserID         string    `json:"-" db:"user_id"`              // 用户ID
	Prompt         string    `json:"prompt" db:"prompt"`          // 文字描述
	Ratio          string    `json:"ratio" db:"ratio"`            // 图片比例
	Resolution     string    `json:"resolution" db:"resolution"`  // 分辨率
	Model          string    `json:"model" db:"model"`            // AI模型
	CustomWidth    int       `json:"customWidth" db:"custom_width"` // 自定义宽
	CustomHeight   int       `json:"customHeight" db:"custom_height"` // 自定义高
	ReferenceImage string    `json:"referenceImage" db:"reference_image"` // 参考图URL
	Status         string    `json:"status" db:"status"`          // 任务状态: queued / processing / succeed / failed
	Progress       int       `json:"progress" db:"progress"`      // 进度百分比
	ResultURL      string    `json:"resultUrl" db:"result_url"`   // 结果图片URL
	ErrorMessage   string    `json:"errorMessage" db:"error_message"` // 失败原因
	CreditsCost    int       `json:"creditsCost" db:"credits_cost"` // 消耗积分
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"-" db:"updated_at"`
}

// GenerateRecord 生成记录（历史记录）
type GenerateRecord struct {
	ID          int64     `json:"-" db:"id"`
	RecordID    string    `json:"id" db:"record_id"`       // 记录唯一标识
	UserID      string    `json:"-" db:"user_id"`          // 用户ID
	Prompt      string    `json:"prompt" db:"prompt"`      // 描述文字
	Ratio       string    `json:"ratio" db:"ratio"`        // 比例
	Resolution  string    `json:"resolution" db:"resolution"` // 分辨率
	ImageURL    string    `json:"imageUrl" db:"image_url"` // 生成图片URL
	Status      string    `json:"status" db:"status"`      // 生成状态: pending / success / failed
	CreditsCost int       `json:"creditsCost" db:"credits_cost"` // 消耗积分
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

// PaymentOrder 支付订单
type PaymentOrder struct {
	ID         int64     `json:"-" db:"id"`
	OrderID    string    `json:"orderId" db:"order_id"`     // 订单ID
	UserID     string    `json:"-" db:"user_id"`            // 用户ID
	OrderType  string    `json:"-" db:"order_type"`         // 订单类型: recharge / vip
	Amount     int64     `json:"amount" db:"amount"`        // 支付金额（分）
	Credits    int64     `json:"credits" db:"credits"`      // 获得积分
	Status     string    `json:"status" db:"status"`        // 支付状态: pending / paid / failed / closed
	PlanType   string    `json:"-" db:"plan_type"`          // 套餐类型（VIP订单）
	CreatedAt  time.Time `json:"-" db:"created_at"`
	UpdatedAt  time.Time `json:"-" db:"updated_at"`
}

// AuthToken 认证令牌
type AuthToken struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         *User  `json:"user"`
}

// AsyncTaskProgress 异步任务进度
type AsyncTaskProgress struct {
	TaskID       string `json:"taskId"`
	Status       string `json:"status"`
	Progress     int    `json:"progress"`
	StatusText   string `json:"statusText"`
	ResultURL    string `json:"resultUrl,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// PaymentResult 支付结果
type PaymentResult struct {
	OrderID string `json:"orderId"`
	Status  string `json:"status"`
	Credits int64  `json:"credits,omitempty"`
}

// PayParams 微信支付参数
type PayParams struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

// PaymentOrderResponse 支付订单响应
type PaymentOrderResponse struct {
	OrderID   string    `json:"orderId"`
	PayParams PayParams `json:"payParams"`
	Amount    int64     `json:"amount"`
	Credits   int64     `json:"credits"`
}

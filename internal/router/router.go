package router

import (
	"aipic_go/internal/config"
	"aipic_go/internal/handler"
	"aipic_go/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Setup 初始化路由
func Setup() *gin.Engine {
	// 生产环境设为 release 模式
	cfg := config.Get()
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())

	// 静态文件服务（上传的图片）
	r.Static("/uploads", cfg.UploadDir)

	// 健康检查（不限流）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "ok",
			"service":     "aipic-go",
			"version":     "1.0.0",
			"devMode":     cfg.DevMode,
			"payMockMode": cfg.PayMockMode,
		})
	})

	// API 路由组
	api := r.Group("/api")
	api.Use(middleware.NormalRateLimit()) // 全局限流：每分钟 120 次

	// 认证相关（不需要鉴权，但需要严格限流）
	authHandler := &handler.AuthHandler{}
	api.POST("/auth/login", middleware.StrictRateLimit(), authHandler.Login)
	api.POST("/auth/refresh", middleware.StrictRateLimit(), authHandler.Refresh)
	api.POST("/auth/logout", authHandler.Logout)

	// 以下接口需要鉴权
	authed := api.Group("")
	authed.Use(middleware.Auth())

	// 用户相关
	userHandler := &handler.UserHandler{}
	authed.GET("/user/info", userHandler.GetUserInfo)
	authed.PUT("/user/avatar", userHandler.UpdateAvatar)

	// 图片生成相关
	generateHandler := &handler.GenerateHandler{}
	authed.POST("/generate/submit", generateHandler.SubmitTask)
	authed.GET("/generate/task/:taskId", generateHandler.GetTaskProgress)
	authed.POST("/generate/task/:taskId/cancel", generateHandler.CancelTask)
	authed.GET("/generate/history", generateHandler.GetHistory)
	authed.DELETE("/generate/history", generateHandler.ClearHistory)                    // 清空所有
	authed.DELETE("/generate/history/:recordId", generateHandler.DeleteHistoryRecord)   // 删除单条

	// 图片上传
	uploadHandler := &handler.UploadHandler{}
	authed.POST("/upload/image", uploadHandler.UploadImage)

	// 支付相关
	paymentHandler := &handler.PaymentHandler{}
	authed.POST("/payment/recharge/create", paymentHandler.CreateRechargeOrder)
	authed.POST("/payment/vip/create", paymentHandler.CreateVipOrder)
	authed.GET("/payment/order/:orderId", paymentHandler.QueryOrderStatus)

	// Mock 支付确认接口（开发模式或支付 Mock 模式可用）
	if cfg.DevMode || cfg.PayMockMode {
		authed.POST("/payment/mock-confirm", paymentHandler.MockPayConfirm)
	}

	// 微信支付回调（不需要鉴权，由微信服务器调用）
	// 生产环境需要验签，在 handler 内部处理
	api.POST("/payment/notify", paymentHandler.WxPayNotify)

	return r
}

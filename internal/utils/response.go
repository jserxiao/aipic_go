package utils

import "github.com/gin-gonic/gin"

// 业务状态码
const (
	CodeSuccess      = 0
	CodeBadRequest    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeRateLimit     = 429
	CodeServerError   = 500
)

// ApiResponse 统一响应体
type ApiResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// PaginatedData 分页数据
type PaginatedData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	HasMore  bool        `json:"hasMore"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, ApiResponse{
		Code:    CodeSuccess,
		Data:    data,
		Message: "success",
	})
}

// SuccessWithMessage 成功响应（带消息）
func SuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(200, ApiResponse{
		Code:    CodeSuccess,
		Data:    data,
		Message: message,
	})
}

// SuccessEmpty 成功响应（无数据）
func SuccessEmpty(c *gin.Context) {
	c.JSON(200, ApiResponse{
		Code:    CodeSuccess,
		Data:    nil,
		Message: "success",
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(200, ApiResponse{
		Code:    code,
		Data:    nil,
		Message: message,
	})
}

// ErrorBadRequest 参数错误
func ErrorBadRequest(c *gin.Context, message string) {
	Error(c, CodeBadRequest, message)
}

// ErrorUnauthorized 未登录
func ErrorUnauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

// ErrorNotFound 资源不存在
func ErrorNotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

// ErrorServerError 服务器错误
func ErrorServerError(c *gin.Context, message string) {
	Error(c, CodeServerError, message)
}

// PaginatedSuccess 分页成功响应
func PaginatedSuccess(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	hasMore := int64(page*pageSize) < total
	Success(c, PaginatedData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	})
}

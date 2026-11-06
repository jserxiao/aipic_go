package middleware

import (
	"net/http"
	"sync"
	"time"

	"aipic_go/internal/utils"

	"github.com/gin-gonic/gin"
)

// visitor 访问者记录
type visitor struct {
	count    int
	lastTime time.Time
}

// RateLimiter 令牌桶限流器
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int           // 时间窗口内最大请求数
	window   time.Duration // 时间窗口
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	// 定期清理过期记录
	go rl.cleanup()
	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	now := time.Now()

	if !exists || now.Sub(v.lastTime) > rl.window {
		rl.visitors[key] = &visitor{count: 1, lastTime: now}
		return true
	}

	if v.count >= rl.limit {
		return false
	}

	v.count++
	v.lastTime = now
	return true
}

// cleanup 定期清理过期访问者记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			if now.Sub(v.lastTime) > rl.window {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit 限流中间件
// limit: 时间窗口内最大请求数
// window: 时间窗口
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusOK, utils.ApiResponse{
				Code:    utils.CodeRateLimit,
				Data:    nil,
				Message: "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}

// StrictRateLimit 严格限流（针对敏感接口如登录）
func StrictRateLimit() gin.HandlerFunc {
	return RateLimit(10, time.Minute) // 每分钟 10 次
}

// NormalRateLimit 普通限流（针对一般接口）
func NormalRateLimit() gin.HandlerFunc {
	return RateLimit(120, time.Minute) // 每分钟 120 次
}

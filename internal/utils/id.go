package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateRandomID 生成随机展示ID，如 AP_8xK2m9
func GenerateRandomID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return fmt.Sprintf("AP_%s", string(b))
}

// GenerateTaskID 生成任务ID
func GenerateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixMilli())
}

// GenerateRecordID 生成记录ID
func GenerateRecordID() string {
	return fmt.Sprintf("gr_%d", time.Now().UnixMilli())
}

// GenerateOrderID 生成订单ID
func GenerateOrderID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixMilli())
}

package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// bodyWriter 拦截 ResponseWriter 的输出，用于记录响应体
type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

var logBorder = strings.Repeat("-", 72)

// RequestLogger 请求日志中间件
// 打印每个请求的方法、路径、请求体、响应体、耗时
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成请求 ID
		requestID := uuid.NewString()
		c.Set("requestID", requestID)

		// 设置响应头
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		query := c.Request.URL.RawQuery

		// 读取请求体（用于打印）
		var reqBody string
		if c.Request.Body != nil && c.Request.ContentLength != 0 {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			// 恢复 body 供后续 handler 读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			reqBody = string(bodyBytes)
			// 截断过长的请求体
			if len(reqBody) > 2000 {
				reqBody = reqBody[:2000] + "...(truncated)"
			}
		}

		// 拦截响应体
		bw := &bodyWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = bw

		// 打印请求日志
		fmt.Printf("\n+-- %s\n", logBorder)
		fmt.Printf("| [%s] >> REQUEST\n", requestID[:8])
		fmt.Printf("| %s %s", method, path)
		if query != "" {
			fmt.Printf("?%s", query)
		}
		fmt.Println()
		fmt.Printf("| Time: %s\n", start.Format("2006-01-02 15:04:05.000"))
		if reqBody != "" {
			if pretty := prettyJSON(reqBody); pretty != "" {
				fmt.Printf("| Body:\n%s\n", indent(pretty, "|   "))
			} else {
				fmt.Printf("| Body: %s\n", reqBody)
			}
		} else {
			fmt.Printf("| Body: (empty)\n")
		}
		fmt.Printf("+-- %s\n", logBorder)

		c.Next()

		// 计算耗时
		latency := time.Since(start)
		status := c.Writer.Status()

		// 读取响应体
		respBody := bw.body.String()
		if len(respBody) > 3000 {
			respBody = respBody[:3000] + "...(truncated)"
		}

		// 状态标记
		icon := "[OK]"
		if status >= 400 {
			icon = "[ERR]"
		} else if status >= 300 {
			icon = "[WARN]"
		}

		fmt.Printf("\n+-- %s\n", logBorder)
		fmt.Printf("| [%s] << RESPONSE %s %d (%v)\n", requestID[:8], icon, status, latency)
		if respBody != "" {
			if pretty := prettyJSON(respBody); pretty != "" {
				fmt.Printf("| Body:\n%s\n", indent(pretty, "|   "))
			} else {
				fmt.Printf("| Body: %s\n", respBody)
			}
		} else {
			fmt.Printf("| Body: (empty)\n")
		}
		fmt.Printf("+-- %s\n", logBorder)
	}
}

// prettyJSON 尝试美化 JSON 字符串，如果不是 JSON 则返回空
func prettyJSON(s string) string {
	var v interface{}
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return ""
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}
	return string(out)
}

// indent 给每一行加前缀
func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

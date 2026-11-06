package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"aipic_go/internal/config"
	"aipic_go/internal/router"
	"aipic_go/internal/store"
	"aipic_go/internal/utils"
)

// loadEnvFile 从 .env 文件加载环境变量（不覆盖已有的环境变量）
func loadEnvFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return // 文件不存在时静默跳过
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 解析 KEY=VALUE
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		// 去除可能的引号
		val = strings.Trim(val, `"'`)
		// 不覆盖已有的环境变量
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func main() {
	// 0. 加载 .env 文件
	loadEnvFile(".env")

	// 1. 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("❌ 配置初始化失败: %v", err)
	}
	cfg := config.Get()

	// 2. 设置 JWT 密钥
	utils.SetJWTSecret(cfg.JWTSecret)

	// 3. 确保上传目录存在
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Fatalf("❌ 创建上传目录失败: %v", err)
	}

	// 4. 初始化数据库
	if err := store.Init(cfg.DBPath); err != nil {
		log.Fatalf("❌ 初始化数据库失败: %v", err)
	}
	defer store.Close()

	// 5. 执行数据库迁移
	if err := store.Migrate(); err != nil {
		log.Fatalf("❌ 数据库迁移失败: %v", err)
	}

	// 6. 开发模式：确保测试用户存在
	if cfg.DevMode {
		store.EnsureTestUser()
		log.Println("[DEV] Mock 登录 + Mock 支付")
	} else if cfg.PayMockMode {
		log.Println("[REAL-LOGIN] 真实微信登录 + Mock 支付")
	} else {
		log.Println("[PROD] 真实微信登录 + 真实微信支付")
	}

	// 7. 初始化路由
	r := router.Setup()

	// 8. 启动服务
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 aipic-go 服务启动: http://localhost%s", addr)
	log.Printf("📁 数据库路径: %s", filepath.Clean(cfg.DBPath))
	log.Printf("📁 上传目录: %s", filepath.Clean(cfg.UploadDir))
	log.Printf("🔑 JWT 过期: %d小时, Refresh 过期: %d小时", cfg.JWTExpire, cfg.RefreshExpire)

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}

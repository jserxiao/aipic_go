package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 全局配置
type Config struct {
	Port          string // 服务端口
	DBPath        string // SQLite 数据库路径
	JWTSecret     string // JWT 签名密钥
	JWTExpire     int    // JWT 过期时间（小时）
	RefreshExpire int    // RefreshToken 过期时间（小时）
	UploadDir     string // 上传文件目录
	MaxUploadSize int64  // 最大上传大小（字节）
	WxAppID       string // 微信小程序 AppID
	WxAppSecret   string // 微信小程序 AppSecret
	MchID         string // 微信支付商户号
	MchKey        string // 微信支付 API 密钥
	MchCertSerial string // 微信支付证书序列号
	MchPrivateKey string // 微信支付商户私钥路径
	WxNotifyURL   string // 微信支付回调通知地址
	DevMode       bool   // 是否开发模式（无微信登录配置）
	PayMockMode   bool   // 是否支付 Mock 模式（有微信登录但无支付配置）
}

var cfg *Config

// Init 初始化配置
func Init() error {
	cfg = &Config{
		Port:          getEnv("PORT", "8080"),
		DBPath:        getEnv("DB_PATH", "./data/aipic.db"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
		JWTExpire:     getEnvInt("JWT_EXPIRE_HOURS", 2),       // access token 2小时
		RefreshExpire: getEnvInt("REFRESH_EXPIRE_HOURS", 168), // refresh token 7天
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSize: getEnvInt64("MAX_UPLOAD_SIZE", 5*1024*1024), // 5MB
		WxAppID:       getEnv("WX_APP_ID", ""),
		WxAppSecret:   getEnv("WX_APP_SECRET", ""),
		MchID:         getEnv("WX_MCH_ID", ""),
		MchKey:        getEnv("WX_MCH_KEY", ""),
		MchCertSerial: getEnv("WX_MCH_CERT_SERIAL", ""),
		MchPrivateKey: getEnv("WX_MCH_PRIVATE_KEY_PATH", ""),
		WxNotifyURL:   getEnv("WX_NOTIFY_URL", ""),
	}

	// 判断模式
	cfg.DevMode = cfg.WxAppID == "" || cfg.WxAppSecret == ""
	cfg.PayMockMode = !cfg.DevMode && (cfg.MchID == "" || cfg.MchKey == "")

	// 生产环境安全检查
	if !cfg.DevMode {
		if cfg.JWTSecret == "" || cfg.JWTSecret == "aipic_default_secret_key" {
			return fmt.Errorf("配置了微信登录后必须设置 JWT_SECRET 环境变量")
		}
		if len(cfg.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET 长度不能少于 32 字符")
		}
		// 支付配置检查：只有在非 Mock 支付模式下才强制要求
		if !cfg.PayMockMode {
			if cfg.WxNotifyURL == "" {
				return fmt.Errorf("生产环境必须配置微信支付回调地址 (WX_NOTIFY_URL)")
			}
		}
	} else {
		// 开发模式使用默认密钥
		if cfg.JWTSecret == "" {
			cfg.JWTSecret = "aipic_dev_secret_key_DO_NOT_USE_IN_PRODUCTION"
		}
	}

	return nil
}

// Get 获取配置实例
func Get() *Config {
	if cfg == nil {
		_ = Init()
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultVal
}

func getEnvInt64(key string, defaultVal int64) int64 {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.ParseInt(val, 10, 64); err == nil {
			return n
		}
	}
	return defaultVal
}

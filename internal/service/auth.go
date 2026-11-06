package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"aipic_go/internal/config"
	"aipic_go/internal/model"
	"aipic_go/internal/store"
	"aipic_go/internal/utils"

	"golang.org/x/sync/singleflight"
)

// WxLoginResponse 微信登录响应
type WxLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

var loginSF singleflight.Group

// Login 微信登录
// 前端通过 Taro.login() 获取 code，传给后端
// 后端用 code 换取 openid，然后生成 token
func Login(code string) (*model.AuthToken, error) {
	// code 长度校验：微信 code 通常为 32~128 字符
	if code == "" {
		return nil, fmt.Errorf("code 不能为空")
	}
	if len(code) > 256 {
		return nil, fmt.Errorf("code 长度异常")
	}

	cfg := config.Get()

	// 开发模式：没有配置微信 AppID 时，使用 code 作为用户标识
	if cfg.WxAppID == "" || cfg.WxAppSecret == "" {
		return mockLogin(code)
	}

	// 真实环境：调用微信接口
	openid, err := getWxOpenID(cfg.WxAppID, cfg.WxAppSecret, code)
	if err != nil {
		return nil, fmt.Errorf("微信登录失败: %w", err)
	}

	// 使用 singleflight 防止重复登录
	result, err, _ := loginSF.Do(openid, func() (interface{}, error) {
		return loginOrCreateUser(openid)
	})
	if err != nil {
		return nil, err
	}

	return result.(*model.AuthToken), nil
}

// RefreshToken 刷新 token
func RefreshToken(refreshToken string) (*model.AuthToken, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refreshToken 不能为空")
	}
	if len(refreshToken) > 2048 {
		return nil, fmt.Errorf("refreshToken 长度异常")
	}

	// 使用 ParseRefreshToken 确保只接受 refresh 类型
	claims, err := utils.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("refreshToken 无效或已过期")
	}

	user, err := store.GetUserByUserID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	cfg := config.Get()
	token, err := utils.GenerateToken(user.UserID, cfg.JWTExpire)
	if err != nil {
		return nil, err
	}
	newRefreshToken, err := utils.GenerateRefreshToken(user.UserID, cfg.RefreshExpire)
	if err != nil {
		return nil, err
	}

	return &model.AuthToken{
		Token:        token,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

// getWxOpenID 调用微信接口获取 openid
func getWxOpenID(appID, appSecret, code string) (string, error) {
	// URL 参数编码，防止特殊字符注入
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 限制读取大小，防止超大响应
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return "", err
	}

	var wxResp WxLoginResponse
	if err := json.Unmarshal(body, &wxResp); err != nil {
		return "", err
	}

	if wxResp.ErrCode != 0 {
		return "", fmt.Errorf("微信接口错误: %d %s", wxResp.ErrCode, wxResp.ErrMsg)
	}

	if wxResp.OpenID == "" {
		return "", fmt.Errorf("微信返回 openid 为空")
	}

	return wxResp.OpenID, nil
}

// loginOrCreateUser 登录或创建用户
func loginOrCreateUser(openid string) (*model.AuthToken, error) {
	// 使用 openid 作为 userID
	userID := "wx_" + openid

	user, err := store.GetUserByUserID(userID)
	if err != nil {
		// 用户不存在，创建新用户
		user = &model.User{
			UserID:      userID,
			RandomID:    utils.GenerateRandomID(),
			Avatar:      "",
			IsVip:       false,
			VipLevel:    "normal",
			VipExpireAt: nil,
			Credits:     100, // 新用户赠送 100 积分
		}
		if err := store.CreateUser(user); err != nil {
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}
	}

	cfg := config.Get()
	token, err := utils.GenerateToken(user.UserID, cfg.JWTExpire)
	if err != nil {
		return nil, err
	}
	refreshToken, err := utils.GenerateRefreshToken(user.UserID, cfg.RefreshExpire)
	if err != nil {
		return nil, err
	}

	return &model.AuthToken{
		Token:        token,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// mockLogin 开发模式登录
func mockLogin(code string) (*model.AuthToken, error) {
	// 开发模式：使用固定用户 u_001
	userID := "u_001"

	result, err, _ := loginSF.Do(userID, func() (interface{}, error) {
		user, err := store.GetUserByUserID(userID)
		if err != nil {
			// 创建测试用户
			user = &model.User{
				UserID:      userID,
				RandomID:    utils.GenerateRandomID(),
				Avatar:      "",
				IsVip:       true,
				VipLevel:    "vip",
				VipExpireAt: strPtr("2027-01-01T00:00:00.000Z"),
				Credits:     1000,
			}
			if err := store.CreateUser(user); err != nil {
				return nil, fmt.Errorf("创建用户失败: %w", err)
			}
		}

		cfg := config.Get()
		token, err := utils.GenerateToken(user.UserID, cfg.JWTExpire)
		if err != nil {
			return nil, err
		}
		refreshToken, err := utils.GenerateRefreshToken(user.UserID, cfg.RefreshExpire)
		if err != nil {
			return nil, err
		}

		return &model.AuthToken{
			Token:        token,
			RefreshToken: refreshToken,
			User:         user,
		}, nil
	})

	if err != nil {
		return nil, err
	}
	return result.(*model.AuthToken), nil
}

func strPtr(s string) *string {
	return &s
}

// IsDevMode 判断是否为开发模式
func IsDevMode() bool {
	cfg := config.Get()
	return cfg.WxAppID == "" || cfg.WxAppSecret == ""
}

// SuppressUnused 抑制未使用导入
var _ = strings.TrimSpace

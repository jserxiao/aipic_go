package service

import (
	"aipic_go/internal/model"
	"aipic_go/internal/store"
)

// GetUserInfo 获取用户信息
func GetUserInfo(userID string) (*model.User, error) {
	return store.GetUserByUserID(userID)
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(userID string, avatarURL string) (*model.User, error) {
	return store.UpdateUserAvatar(userID, avatarURL)
}

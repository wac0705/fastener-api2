package utils

import (
	"fmt"
	"go.uber.org/zap"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 對密碼進行 Bcrypt 雜湊
func HashPassword(password string) (string, error) {
	// bcrypt.DefaultCost 是一個合理的默認成本參數，可以根據需要調整
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("Utils: Failed to hash password", zap.Error(err))
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash 比較明文密碼與雜湊密碼是否匹配
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// 如果 err 不為 nil，表示不匹配或雜湊值無效
	if err != nil {
		// zap.L().Debug("Utils: Password hash comparison failed", zap.Error(err)) // 在調試時可以啟用
		return false
	}
	return true
}

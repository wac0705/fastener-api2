package jwt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models" // 導入 Account 模型
	"github.com/wac0705/fastener-api/utils"  // 導入工具 (包含自定義錯誤)
)

// AccessClaims 定義 Access Token 的 JWT Claim 結構
type AccessClaims struct {
	AccountID int    `json:"account_id"`
	Username  string `json:"username"`
	RoleID    int    `json:"role_id"` // 角色 ID
	jwt.RegisteredClaims
}

// RefreshClaims 定義 Refresh Token 的 JWT Claim 結構
type RefreshClaims struct {
	AccountID int `json:"account_id"`
	jwt.RegisteredClaims
}

// GenerateAuthTokens 創建 Access Token 和 Refresh Token
func GenerateAuthTokens(account models.Account, secret string, accessExpiresHours, refreshExpiresHours int) (accessToken string, refreshToken string, err error) {
	// Access Token
	accessClaims := &AccessClaims{
		AccountID: account.ID,
		Username:  account.Username,
		RoleID:    account.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(accessExpiresHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fastener-api", // Token 發行者
			Subject:   fmt.Sprintf("%d", account.ID),
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secret))
	if err != nil {
		zap.L().Error("Failed to generate access token", zap.Error(err), zap.Int("account_id", account.ID))
		return "", "", utils.ErrInternalServer.SetDetails("Failed to generate access token")
	}

	// Refresh Token
	refreshClaims := &RefreshClaims{
		AccountID: account.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(refreshExpiresHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "fastener-api",
			Subject:   fmt.Sprintf("%d", account.ID),
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret))
	if err != nil {
		zap.L().Error("Failed to generate refresh token", zap.Error(err), zap.Int("account_id", account.ID))
		return "", "", utils.ErrInternalServer.SetDetails("Failed to generate refresh token")
	}

	return accessToken, refreshToken, nil
}

// JwtAccessConfig 返回 Echo 的 JWT 中介軟體配置，用於 Access Token 驗證
func JwtAccessConfig(secret string) echojwt.Config {
	return echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(AccessClaims) // 使用 AccessClaims 結構
		},
		SigningKey:  []byte(secret),
		TokenLookup: "header:" + echo.HeaderAuthorization, // 從 Authorization 頭部查找 Token
		AuthScheme:  "Bearer",                             // Token 方案
		ErrorHandler: func(c echo.Context, err error) error {
			zap.L().Info("Access Token validation failed", zap.Error(err), zap.String("path", c.Path()))
			return c.JSON(http.StatusUnauthorized, utils.ErrUnauthorized.SetDetails("Invalid or expired access token"))
		},
	}
}

// VerifyRefreshToken 驗證 Refresh Token 並返回 Claims
// 這個函數會在 RefreshToken API 處理器中被調用
func VerifyRefreshToken(tokenString string, secret string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		zap.L().Info("Refresh Token parsing failed", zap.Error(err))
		return nil, utils.ErrUnauthorized.SetDetails("Invalid refresh token")
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		zap.L().Info("Refresh Token validation failed: invalid claims or token", zap.Any("claims", claims), zap.Bool("valid", token.Valid))
		return nil, utils.ErrUnauthorized.SetDetails("Invalid refresh token")
	}
	return claims, nil
}

// NewJwtVerifier 創建 JWT 驗證器，可在需要時手動驗證 Token (Access 或 Refresh)
// 這是通用驗證器，可以根據 needsAccess 參數決定驗證 AccessClaims 或 RefreshClaims
type JwtVerifier struct {
	Secret string
}

func NewJwtVerifier(secret string) *JwtVerifier {
	return &JwtVerifier{Secret: secret}
}

// VerifyToken 通用驗證器，根據上下文判斷驗證哪種 Token
func (jv *JwtVerifier) VerifyToken(tokenString string, needsRefresh bool) (interface{}, error) {
	if needsRefresh {
		return VerifyRefreshToken(tokenString, jv.Secret)
	}
	// 預設為 Access Token
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jv.Secret), nil
	})

	if err != nil {
		zap.L().Info("Token parsing failed", zap.Error(err))
		return nil, utils.ErrUnauthorized.SetDetails("Invalid token")
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		zap.L().Info("Token validation failed: invalid claims or token", zap.Any("claims", claims), zap.Bool("valid", token.Valid))
		return nil, utils.ErrUnauthorized.SetDetails("Invalid token")
	}
	return claims, nil
}

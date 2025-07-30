package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/middleware/jwt" // 導入 JWT 相關模型
	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/service" // 導入 Service 層
	"github.com/wac0705/fastener-api/utils"  // 導入工具 (包含自定義錯誤)
)

// AuthHandler 定義身份驗證處理器結構，包含 AuthService 的依賴
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler 創建 AuthHandler 實例
func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{authService: s}
}

// Login 處理用戶登入請求
func (h *AuthHandler) Login(c echo.Context) error {
	req := new(models.LoginRequest)

	// 綁定請求體
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 驗證請求數據
	if err := c.Validate(req); err != nil {
		return err // 驗證錯誤會被全局錯誤處理器捕獲和格式化
	}

	// 調用 Service 層進行登入
	accessToken, refreshToken, account, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Login failed due to internal error", zap.String("username", req.Username), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	// 成功登入，返回 Access Token 和 Refresh Token 以及用戶基本信息
	resp := struct {
		AccessToken  string         `json:"access_token"`
		RefreshToken string         `json:"refresh_token"`
		Account      *models.Account `json:"account"`
	}{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Account:      account,
	}
	resp.Account.Password = "" // 清除密碼敏感信息
	return c.JSON(http.StatusOK, resp)
}

// Register 處理用戶註冊請求
func (h *AuthHandler) Register(c echo.Context) error {
	req := new(models.RegisterRequest)

	// 綁定請求體
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 驗證請求數據
	if err := c.Validate(req); err != nil {
		return err // 驗證錯誤
	}

	// 調用 Service 層進行註冊
	account, err := h.authService.Register(req.Username, req.Password, req.RoleID)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Registration failed due to internal error", zap.String("username", req.Username), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	account.Password = "" // 清除密碼敏感信息
	return c.JSON(http.StatusCreated, account)
}

// RefreshToken 處理 Token 刷新請求
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	req := new(models.RefreshTokenRequest)

	// 綁定請求體 (只需 Refresh Token)
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 調用 Service 層刷新 Token
	newAccessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to refresh token", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token": newAccessToken,
	})
}

// GetMyProfile 獲取當前用戶的資料 (受保護路由)
// 這是新增的範例，用於演示如何從 Context 中獲取 Claims
func (h *AuthHandler) GetMyProfile(c echo.Context) error {
    claims, ok := c.Get("claims").(*jwt.AccessClaims)
    if !ok || claims == nil {
        // 這條路徑通常不會被觸發，因為有 JWT 中介軟體保護
        zap.L().Warn("Claims not found in context for GetMyProfile")
        return c.JSON(http.StatusUnauthorized, utils.ErrUnauthorized)
    }

    // 這裡可以呼叫 service 層根據 claims.AccountID 獲取更詳細的用戶資訊
    // 例如：account, err := h.authService.GetAccountProfile(claims.AccountID)
    // 為了簡化，直接返回 claims 中的部分資訊
    
    // 從資料庫獲取完整帳戶信息，包括角色名
    account, err := h.authService.GetAccountByID(claims.AccountID)
    if err != nil {
        if customErr, ok := err.(*utils.CustomError); ok {
            return c.JSON(customErr.Code, customErr)
        }
        zap.L().Error("Failed to get account profile", zap.Int("account_id", claims.AccountID), zap.Error(err))
        return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
    }
    if account == nil {
        return c.JSON(http.StatusNotFound, utils.ErrNotFound)
    }

    account.Password = "" // 不返回密碼

    return c.JSON(http.StatusOK, account)
}

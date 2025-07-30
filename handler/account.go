package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap" // 使用 zap 進行日誌記錄

	"github.com/wac0705/fastener-api/middleware/jwt" // 導入 JWT Claims
	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/service" // 導入 Service 層
	"github.com/wac0705/fastener-api/utils"  // 導入工具 (包含自定義錯誤)
)

// AccountHandler 定義帳戶處理器結構，包含 AccountService 的依賴
type AccountHandler struct {
	accountService service.AccountService
}

// NewAccountHandler 創建 AccountHandler 實例
func NewAccountHandler(s service.AccountService) *AccountHandler {
	return &AccountHandler{accountService: s}
}

// CreateAccount 創建新帳戶
func (h *AccountHandler) CreateAccount(c echo.Context) error {
	account := new(models.Account)

	// 綁定請求體到結構體
	if err := c.Bind(account); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 驗證請求數據
	if err := c.Validate(account); err != nil {
		// Echo 的 Validate 會觸發我們在 main.go 中設定的錯誤處理器
		return err // 驗證錯誤會被全局錯誤處理器捕獲和格式化
	}

	// 調用 Service 層創建帳戶
	if err := h.accountService.CreateAccount(account); err != nil {
		// 如果是自定義錯誤，直接返回
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		// 其他未知錯誤，記錄並返回內部錯誤
		zap.L().Error("Failed to create account", zap.Error(err), zap.Any("account", account))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	// 成功創建後，不返回密碼等敏感信息
	account.Password = "" // 清除密碼字段
	return c.JSON(http.StatusCreated, account)
}

// GetAccounts 獲取所有帳戶
func (h *AccountHandler) GetAccounts(c echo.Context) error {
	accounts, err := h.accountService.GetAllAccounts()
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get accounts", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, accounts)
}

// GetAccountById 根據 ID 獲取帳戶
func (h *AccountHandler) GetAccountById(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	account, err := h.accountService.GetAccountByID(id)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get account by ID", zap.Int("account_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	if account == nil { // Service 層返回 nil, nil 表示未找到
		return c.JSON(http.StatusNotFound, utils.ErrNotFound)
	}

	account.Password = "" // 清除密碼字段
	return c.JSON(http.StatusOK, account)
}

// UpdateAccount 更新帳戶信息
func (h *AccountHandler) UpdateAccount(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	account := new(models.Account)
	if err := c.Bind(account); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 確保更新的是正確的帳戶 ID
	account.ID = id

	// 驗證請求數據
	// 注意：對於部分更新，如果驗證器要求所有字段都存在，這裡可能需要特殊處理
	if err := c.Validate(account); err != nil {
		return err
	}

	// 調用 Service 層更新帳戶
	if err := h.accountService.UpdateAccount(account); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to update account", zap.Int("account_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	account.Password = "" // 清除密碼字段
	return c.JSON(http.StatusOK, account)
}

// DeleteAccount 刪除帳戶
func (h *AccountHandler) DeleteAccount(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 調用 Service 層刪除帳戶
	if err := h.accountService.DeleteAccount(id); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to delete account", zap.Int("account_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.NoContent(http.StatusNoContent) // 成功刪除，返回 204 No Content
}

// UpdateAccountPassword 更新帳戶密碼
func (h *AccountHandler) UpdateAccountPassword(c echo.Context) error {
    id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取目標帳戶 ID
    if err != nil {
        return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
    }

    // 獲取當前請求用戶的 claims，用於檢查是否是自己修改密碼或有權限的管理員修改
    claims := c.Get("claims").(*jwt.AccessClaims)
    if claims == nil {
        return c.JSON(http.StatusUnauthorized, utils.ErrUnauthorized) // 未經授權，理論上不會發生因為有 JWT 中介軟體
    }

    req := new(models.UpdatePasswordRequest)
    if err := c.Bind(req); err != nil {
        return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
    }

    if err := c.Validate(req); err != nil {
        return err // 驗證錯誤
    }

    // 調用 Service 層更新密碼
    if err := h.accountService.UpdatePassword(id, req.OldPassword, req.NewPassword, claims.AccountID, claims.RoleID); err != nil {
        if customErr, ok := err.(*utils.CustomError); ok {
            return c.JSON(customErr.Code, customErr)
        }
        zap.L().Error("Failed to update account password", zap.Int("account_id", id), zap.Error(err))
        return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
    }

    return c.NoContent(http.StatusNoContent) // 成功更新，返回 204 No Content
}

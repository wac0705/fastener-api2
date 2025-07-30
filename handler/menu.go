package handler

import (
	"database/sql" // 導入 sql 包，用於檢查 ErrNoRows
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/service"
	"github.com/wac0705/fastener-api/utils"
)

// MenuHandler 定義選單處理器結構，包含 MenuService 的依賴
type MenuHandler struct {
	menuService service.MenuService
}

// NewMenuHandler 創建 MenuHandler 實例
func NewMenuHandler(s service.MenuService) *MenuHandler {
	return &MenuHandler{menuService: s}
}

// CreateMenu 創建新選單
func (h *MenuHandler) CreateMenu(c echo.Context) error {
	menu := new(models.Menu)

	if err := c.Bind(menu); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := c.Validate(menu); err != nil {
		return err // 驗證錯誤
	}

	if err := h.menuService.CreateMenu(menu); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to create menu", zap.Error(err), zap.String("menu_name", menu.Name))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusCreated, menu)
}

// GetMenus 獲取所有選單
func (h *MenuHandler) GetMenus(c echo.Context) error {
	menus, err := h.menuService.GetAllMenus()
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get menus", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, menus)
}

// GetMenuById 根據 ID 獲取選單
func (h *MenuHandler) GetMenuById(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	menu, err := h.menuService.GetMenuByID(id)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get menu by ID", zap.Int("menu_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	if menu == nil { // Service 層返回 nil, nil 表示未找到
		return c.JSON(http.StatusNotFound, utils.ErrNotFound)
	}

	return c.JSON(http.StatusOK, menu)
}

// UpdateMenu 更新選單信息
func (h *MenuHandler) UpdateMenu(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	menu := new(models.Menu)
	if err := c.Bind(menu); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 確保更新的是正確的選單 ID
	menu.ID = id

	if err := c.Validate(menu); err != nil {
		return err // 驗證錯誤
	}

	if err := h.menuService.UpdateMenu(menu); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to update menu", zap.Int("menu_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusOK, menu)
}

// DeleteMenu 刪除選單
func (h *MenuHandler) DeleteMenu(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := h.menuService.DeleteMenu(id); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to delete menu", zap.Int("menu_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.NoContent(http.StatusNoContent) // 成功刪除，返回 204 No Content
}

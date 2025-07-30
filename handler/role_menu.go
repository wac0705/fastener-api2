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

// RoleMenuHandler 定義角色選單處理器結構，包含 RoleMenuService 的依賴
type RoleMenuHandler struct {
	roleMenuService service.RoleMenuService
}

// NewRoleMenuHandler 創建 RoleMenuHandler 實例
func NewRoleMenuHandler(s service.RoleMenuService) *RoleMenuHandler {
	return &RoleMenuHandler{roleMenuService: s}
}

// CreateRoleMenu 創建新的角色選單關聯
func (h *RoleMenuHandler) CreateRoleMenu(c echo.Context) error {
	roleMenu := new(models.RoleMenu)

	if err := c.Bind(roleMenu); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := c.Validate(roleMenu); err != nil {
		return err // 驗證錯誤
	}

	if err := h.roleMenuService.CreateRoleMenu(roleMenu); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to create role menu", zap.Error(err), zap.Int("role_id", roleMenu.RoleID), zap.Int("menu_id", roleMenu.MenuID))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusCreated, roleMenu)
}

// GetRoleMenus 獲取所有角色選單關聯 (或根據查詢參數過濾)
func (h *RoleMenuHandler) GetRoleMenus(c echo.Context) error {
	roleIDStr := c.QueryParam("role_id")
	menuIDStr := c.QueryParam("menu_id")

	var roleID *int
	if roleIDStr != "" {
		id, err := strconv.Atoi(roleIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrBadRequest.SetDetails("Invalid role_id"))
		}
		roleID = &id
	}

	var menuID *int
	if menuIDStr != "" {
		id, err := strconv.Atoi(menuIDStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrBadRequest.SetDetails("Invalid menu_id"))
		}
		menuID = &id
	}

	roleMenus, err := h.roleMenuService.GetAllRoleMenus(roleID, menuID)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get role menus", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, roleMenus)
}

// DeleteRoleMenu 刪除角色選單關聯
func (h *RoleMenuHandler) DeleteRoleMenu(c echo.Context) error {
	roleID, err := strconv.Atoi(c.Param("id1")) // 假設 URL 參數是 /role_menus/:role_id/:menu_id
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest.SetDetails("Invalid role_id in path"))
	}
	menuID, err := strconv.Atoi(c.Param("id2")) // 假設 URL 參數是 /role_menus/:role_id/:menu_id
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest.SetDetails("Invalid menu_id in path"))
	}

	if err := h.roleMenuService.DeleteRoleMenu(roleID, menuID); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to delete role menu", zap.Error(err), zap.Int("role_id", roleID), zap.Int("menu_id", menuID))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.NoContent(http.StatusNoContent) // 成功刪除，返回 204 No Content
}

// UpdateRoleMenu 由於是複合主鍵，更新操作通常是先刪除再創建，或者直接更新（如果僅更新非主鍵字段）。
// 這裡假設是根據 role_id 和 menu_id 查詢後更新其他可能的屬性，但通常複合主鍵表不會有太多可更新的屬性。
// 如果實際需求是修改關聯，通常是通過 delete + create 來實現。
// 但為了提供一個範例，我們假設可以更新一個新的菜單 ID
func (h *RoleMenuHandler) UpdateRoleMenu(c echo.Context) error {
	oldRoleID, err := strconv.Atoi(c.Param("id1"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest.SetDetails("Invalid old role_id in path"))
	}
	oldMenuID, err := strconv.Atoi(c.Param("id2"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest.SetDetails("Invalid old menu_id in path"))
	}

	req := new(models.RoleMenu) // 新的關聯數據，可能包含新的 menu_id 或 role_id
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	// 這裡假定更新是從 oldRoleID, oldMenuID 更改為 req.RoleID, req.MenuID
	// 實際操作中，如果是更新複合主鍵，一般是先刪後插
	// 這裡我們直接調用 Service 層的 Update 方法來處理邏輯
	if err := h.roleMenuService.UpdateRoleMenu(oldRoleID, oldMenuID, req.RoleID, req.MenuID); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to update role menu", zap.Error(err),
			zap.Int("old_role_id", oldRoleID), zap.Int("old_menu_id", oldMenuID),
			zap.Int("new_role_id", req.RoleID), zap.Int("new_menu_id", req.MenuID))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, req)
}

package service

import (
	"fmt"
	"net/http" // 用於檢查錯誤類型

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository"
	"github.com/wac0705/fastener-api/utils"
)

// MenuService 定義選單服務介面
type MenuService interface {
	GetAllMenus() ([]models.Menu, error)
	GetMenuByID(id int) (*models.Menu, error)
	CreateMenu(menu *models.Menu) error
	UpdateMenu(menu *models.Menu) error
	DeleteMenu(id int) error
	GetMenusByRoleID(roleID int) ([]models.Menu, error) // 新增：根據角色 ID 獲取選單
}

// menuServiceImpl 實現 MenuService 介面
type menuServiceImpl struct {
	menuRepo repository.MenuRepository
	roleMenuRepo repository.RoleMenuRepository // 導入 RoleMenuRepository
}

// NewMenuService 創建 MenuService 實例
func NewMenuService(menuRepo repository.MenuRepository, roleMenuRepo repository.RoleMenuRepository) MenuService {
	return &menuServiceImpl{menuRepo: menuRepo, roleMenuRepo: roleMenuRepo}
}

// CreateMenu 創建新選單
func (s *menuServiceImpl) CreateMenu(menu *models.Menu) error {
	// 檢查 Path 是否重複
	existingMenu, err := s.menuRepo.FindByPath(menu.Path) // 假設 Repository 有 FindByPath
	if err != nil {
		zap.L().Error("Service: Error checking existing menu by path during creation", zap.Error(err), zap.String("path", menu.Path))
		return utils.ErrInternalServer
	}
	if existingMenu != nil {
		return utils.ErrBadRequest.SetDetails("Menu with this path already exists.")
	}

	// 如果有 ParentID，檢查父選單是否存在
	if menu.ParentID != nil {
		parentMenu, err := s.menuRepo.FindByID(*menu.ParentID)
		if err != nil {
			zap.L().Error("Service: Error checking parent menu ID for new menu", zap.Error(err), zap.Int("parent_id", *menu.ParentID))
			return utils.ErrInternalServer
		}
		if parentMenu == nil {
			return utils.ErrBadRequest.SetDetails("Provided Parent Menu ID does not exist.")
		}
	}

	if err := s.menuRepo.Create(menu); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusBadRequest {
			return customErr // 假設 Repository 返回的錯誤已包含詳細信息
		}
		zap.L().Error("Service: Failed to create menu in repository", zap.Error(err), zap.String("name", menu.Name))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to create menu: %v", err))
	}
	return nil
}

// GetAllMenus 獲取所有選單
func (s *menuServiceImpl) GetAllMenus() ([]models.Menu, error) {
	menus, err := s.menuRepo.FindAll()
	if err != nil {
		zap.L().Error("Service: Failed to get all menus", zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	return menus, nil
}

// GetMenuByID 根據 ID 獲取選單
func (s *menuServiceImpl) GetMenuByID(id int) (*models.Menu, error) {
	menu, err := s.menuRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Failed to get menu by ID", zap.Int("id", id), zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	if menu == nil {
		return nil, nil // Repository 返回 nil, nil 表示未找到
	}
	return menu, nil
}

// UpdateMenu 更新選單信息
func (s *menuServiceImpl) UpdateMenu(menu *models.Menu) error {
	// 檢查選單是否存在
	existingMenu, err := s.menuRepo.FindByID(menu.ID)
	if err != nil {
		zap.L().Error("Service: Error checking existing menu for update", zap.Error(err), zap.Int("menu_id", menu.ID))
		return utils.ErrInternalServer
	}
	if existingMenu == nil {
		return utils.ErrNotFound
	}

	// 如果 Path 有更改，檢查是否重複
	if existingMenu.Path != menu.Path {
		otherMenu, err := s.menuRepo.FindByPath(menu.Path) // 假設 Repository 有 FindByPath
		if err != nil {
			zap.L().Error("Service: Error checking menu path for update conflict", zap.Error(err), zap.String("new_path", menu.Path))
			return utils.ErrInternalServer
		}
		if otherMenu != nil && otherMenu.ID != menu.ID {
			return utils.ErrBadRequest.SetDetails("Menu path already exists for another menu")
		}
	}

	// 如果有 ParentID，檢查父選單是否存在
	if menu.ParentID != nil {
		parentMenu, err := s.menuRepo.FindByID(*menu.ParentID)
		if err != nil {
			zap.L().Error("Service: Error checking parent menu ID for menu update", zap.Error(err), zap.Int("parent_id", *menu.ParentID))
			return utils.ErrInternalServer
		}
		if parentMenu == nil {
			return utils.ErrBadRequest.SetDetails("Provided Parent Menu ID for update does not exist.")
		}
	}

	if err := s.menuRepo.Update(menu); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusBadRequest {
			return customErr
		}
		zap.L().Error("Service: Failed to update menu in repository", zap.Error(err), zap.Int("menu_id", menu.ID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to update menu: %v", err))
	}
	return nil
}

// DeleteMenu 刪除選單
func (s *menuServiceImpl) DeleteMenu(id int) error {
	// 檢查選單是否存在
	existingMenu, err := s.menuRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Error checking existing menu for delete", zap.Error(err), zap.Int("menu_id", id))
		return utils.ErrInternalServer
	}
	if existingMenu == nil {
		return utils.ErrNotFound
	}

	// 這裡可以添加額外業務邏輯，例如檢查是否有子選單或角色關聯到此選單
	// 如果資料庫外鍵設置為 RESTRICT，則會自動阻止刪除
	// 如果有多個子選單，也可以考慮先將子選單的 parent_id 設為 NULL

	if err := s.menuRepo.Delete(id); err != nil {
		zap.L().Error("Service: Failed to delete menu in repository", zap.Error(err), zap.Int("menu_id", id))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to delete menu: %v", err))
	}
	return nil
}

// GetMenusByRoleID 根據角色 ID 獲取選單 (供前端使用)
func (s *menuServiceImpl) GetMenusByRoleID(roleID int) ([]models.Menu, error) {
	// 檢查角色是否存在
	// 這是為了防止查詢一個不存在的角色ID
	// role, err := s.roleRepo.FindByID(roleID) // 如果有 roleRepo 依賴，可以在這裡檢查
	// if err != nil || role == nil { return nil, utils.ErrBadRequest.SetDetails("Invalid Role ID") }

	menus, err := s.roleMenuRepo.FindMenusByRoleID(roleID) // 呼叫 RoleMenuRepository
	if err != nil {
		zap.L().Error("Service: Failed to get menus by role ID from repository", zap.Error(err), zap.Int("role_id", roleID))
		return nil, utils.ErrInternalServer
	}
	return menus, nil
}

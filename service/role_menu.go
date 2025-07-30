package service

import (
	"fmt"
	"net/http" // 用於錯誤檢查

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository"
	"github.com/wac0705/fastener-api/utils"
)

// RoleMenuService 定義角色選單服務介面
type RoleMenuService interface {
	CreateRoleMenu(roleMenu *models.RoleMenu) error
	GetAllRoleMenus(roleID, menuID *int) ([]models.RoleMenuDetail, error)
	DeleteRoleMenu(roleID, menuID int) error
	UpdateRoleMenu(oldRoleID, oldMenuID, newRoleID, newMenuID int) error
}

// roleMenuServiceImpl 實現 RoleMenuService 介面
type roleMenuServiceImpl struct {
	roleMenuRepo repository.RoleMenuRepository
	roleRepo     repository.RoleRepository // 依賴 RoleRepository 檢查角色是否存在
	menuRepo     repository.MenuRepository // 依賴 MenuRepository 檢查選單是否存在
}

// NewRoleMenuService 創建 RoleMenuService 實例
func NewRoleMenuService(roleMenuRepo repository.RoleMenuRepository, roleRepo repository.RoleRepository, menuRepo repository.MenuRepository) RoleMenuService {
	return &roleMenuServiceImpl{roleMenuRepo: roleMenuRepo, roleRepo: roleRepo, menuRepo: menuRepo}
}

// CreateRoleMenu 創建新的角色選單關聯
func (s *roleMenuServiceImpl) CreateRoleMenu(roleMenu *models.RoleMenu) error {
	// 業務驗證：檢查 roleID 和 menuID 是否真實存在
	role, err := s.roleRepo.FindByID(roleMenu.RoleID)
	if err != nil {
		zap.L().Error("Service: Error checking role for role menu creation", zap.Error(err), zap.Int("role_id", roleMenu.RoleID))
		return utils.ErrInternalServer
	}
	if role == nil {
		return utils.ErrBadRequest.SetDetails("Invalid Role ID")
	}

	menu, err := s.menuRepo.FindByID(roleMenu.MenuID)
	if err != nil {
		zap.L().Error("Service: Error checking menu for role menu creation", zap.Error(err), zap.Int("menu_id", roleMenu.MenuID))
		return utils.ErrInternalServer
	}
	if menu == nil {
		return utils.ErrBadRequest.SetDetails("Invalid Menu ID")
	}

	// 檢查是否已存在相同的關聯 (Repository 的 ON CONFLICT DO NOTHING 會處理，但這裡可以提前返回錯誤)
	existingRelations, err := s.roleMenuRepo.FindAll(&roleMenu.RoleID, &roleMenu.MenuID)
	if err != nil {
		zap.L().Error("Service: Error checking existing role menu relationship", zap.Error(err))
		return utils.ErrInternalServer
	}
	if len(existingRelations) > 0 {
		return utils.ErrBadRequest.SetDetails("Role-menu relationship already exists.")
	}


	if err := s.roleMenuRepo.Create(roleMenu); err != nil {
		zap.L().Error("Service: Failed to create role menu in repository", zap.Error(err), zap.Int("role_id", roleMenu.RoleID), zap.Int("menu_id", roleMenu.MenuID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to create role menu: %v", err))
	}
	return nil
}

// GetAllRoleMenus 獲取所有角色選單關聯
func (s *roleMenuServiceImpl) GetAllRoleMenus(roleID, menuID *int) ([]models.RoleMenuDetail, error) {
	roleMenus, err := s.roleMenuRepo.FindAll(roleID, menuID)
	if err != nil {
		zap.L().Error("Service: Failed to get all role menus", zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	return roleMenus, nil
}

// DeleteRoleMenu 刪除角色選單關聯
func (s *roleMenuServiceImpl) DeleteRoleMenu(roleID, menuID int) error {
	// 業務驗證：檢查關聯是否存在
	existingRelations, err := s.roleMenuRepo.FindAll(&roleID, &menuID)
	if err != nil {
		zap.L().Error("Service: Error checking existing role menu relationship for delete", zap.Error(err))
		return utils.ErrInternalServer
	}
	if len(existingRelations) == 0 {
		return utils.ErrNotFound.SetDetails(fmt.Sprintf("Role-menu relationship (role_id: %d, menu_id: %d) not found.", roleID, menuID))
	}

	if err := s.roleMenuRepo.Delete(roleID, menuID); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusNotFound {
			return customErr // 如果 Repository 返回的是未找到錯誤
		}
		zap.L().Error("Service: Failed to delete role menu in repository", zap.Error(err), zap.Int("role_id", roleID), zap.Int("menu_id", menuID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to delete role menu: %v", err))
	}
	return nil
}

// UpdateRoleMenu 更新角色選單關聯
func (s *roleMenuServiceImpl) UpdateRoleMenu(oldRoleID, oldMenuID, newRoleID, newMenuID int) error {
	// 業務驗證：檢查新的 roleID 和 menuID 是否存在
	role, err := s.roleRepo.FindByID(newRoleID)
	if err != nil {
		zap.L().Error("Service: Error checking new role for role menu update", zap.Error(err), zap.Int("role_id", newRoleID))
		return utils.ErrInternalServer
	}
	if role == nil {
		return utils.ErrBadRequest.SetDetails("Invalid New Role ID")
	}

	menu, err := s.menuRepo.FindByID(newMenuID)
	if err != nil {
		zap.L().Error("Service: Error checking new menu for role menu update", zap.Error(err), zap.Int("menu_id", newMenuID))
		return utils.ErrInternalServer
	}
	if menu == nil {
		return utils.ErrBadRequest.SetDetails("Invalid New Menu ID")
	}

	// 檢查新關聯是否已存在 (如果新舊ID相同，且新關聯已存在，則視為成功)
	if oldRoleID != newRoleID || oldMenuID != newMenuID {
		existingNewRelations, err := s.roleMenuRepo.FindAll(&newRoleID, &newMenuID)
		if err != nil {
			zap.L().Error("Service: Error checking existing new role menu relationship for update", zap.Error(err))
			return utils.ErrInternalServer
		}
		if len(existingNewRelations) > 0 {
			return utils.ErrBadRequest.SetDetails("New role-menu relationship already exists.")
		}
	}

	if err := s.roleMenuRepo.Update(oldRoleID, oldMenuID, newRoleID, newMenuID); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusNotFound {
			return customErr // 如果 Repository 返回的是未找到錯誤
		}
		zap.L().Error("Service: Failed to update role menu in repository", zap.Error(err),
			zap.Int("old_role_id", oldRoleID), zap.Int("old_menu_id", oldMenuID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to update role menu: %v", err))
	}
	return nil
}

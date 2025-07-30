package service

import (
	"fmt"
	"net/http" // 用於檢查錯誤類型

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository"
	"github.com/wac0705/fastener-api/utils"
)

// RoleService 定義角色服務介面
type RoleService interface {
	GetAllRoles() ([]models.Role, error)
	GetRoleByID(id int) (*models.Role, error)
	CreateRole(role *models.Role) error
	UpdateRole(role *models.Role) error
	DeleteRole(id int) error
}

// roleServiceImpl 實現 RoleService 介面
type roleServiceImpl struct {
	roleRepo repository.RoleRepository
}

// NewRoleService 創建 RoleService 實例
func NewRoleService(repo repository.RoleRepository) RoleService {
	return &roleServiceImpl{roleRepo: repo}
}

// CreateRole 創建新角色
func (s *roleServiceImpl) CreateRole(role *models.Role) error {
	// 檢查角色名稱是否已存在
	existingRole, err := s.roleRepo.FindByName(role.Name)
	if err != nil {
		zap.L().Error("Service: Error checking existing role by name during creation", zap.Error(err), zap.String("name", role.Name))
		return utils.ErrInternalServer
	}
	if existingRole != nil {
		return utils.ErrBadRequest.SetDetails("Role with this name already exists.")
	}

	if err := s.roleRepo.Create(role); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusBadRequest {
			return customErr // 假設 Repository 返回的錯誤已包含詳細信息
		}
		zap.L().Error("Service: Failed to create role in repository", zap.Error(err), zap.String("name", role.Name))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to create role: %v", err))
	}
	return nil
}

// GetAllRoles 獲取所有角色
func (s *roleServiceImpl) GetAllRoles() ([]models.Role, error) {
	roles, err := s.roleRepo.FindAll()
	if err != nil {
		zap.L().Error("Service: Failed to get all roles", zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	return roles, nil
}

// GetRoleByID 根據 ID 獲取角色
func (s *roleServiceImpl) GetRoleByID(id int) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Failed to get role by ID", zap.Int("id", id), zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	if role == nil {
		return nil, nil // Repository 返回 nil, nil 表示未找到
	}
	return role, nil
}

// UpdateRole 更新角色信息
func (s *roleServiceImpl) UpdateRole(role *models.Role) error {
	// 檢查角色是否存在
	existingRole, err := s.roleRepo.FindByID(role.ID)
	if err != nil {
		zap.L().Error("Service: Error checking existing role for update", zap.Error(err), zap.Int("role_id", role.ID))
		return utils.ErrInternalServer
	}
	if existingRole == nil {
		return utils.ErrNotFound
	}

	// 檢查新名稱是否被其他角色占用 (如果名稱有更改)
	if existingRole.Name != role.Name {
		otherRole, err := s.roleRepo.FindByName(role.Name)
		if err != nil {
			zap.L().Error("Service: Error checking role name for update conflict", zap.Error(err), zap.String("new_name", role.Name))
			return utils.ErrInternalServer
		}
		if otherRole != nil && otherRole.ID != role.ID {
			return utils.ErrBadRequest.SetDetails("Role name already exists for another role")
		}
	}

	if err := s.roleRepo.Update(role); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusBadRequest {
			return customErr
		}
		zap.L().Error("Service: Failed to update role in repository", zap.Error(err), zap.Int("role_id", role.ID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to update role: %v", err))
	}
	return nil
}

// DeleteRole 刪除角色
func (s *roleServiceImpl) DeleteRole(id int) error {
	// 檢查角色是否存在
	existingRole, err := s.roleRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Error checking existing role for delete", zap.Error(err), zap.Int("role_id", id))
		return utils.ErrInternalServer
	}
	if existingRole == nil {
		return utils.ErrNotFound
	}

	// 業務邏輯：檢查是否有用戶或選單關聯到此角色，如果資料庫外鍵是 RESTRICT 會阻止刪除
	// 也可以在這裡主動檢查，並返回更友好的錯誤訊息
	// 例如：userCount, _ := s.accountRepo.CountByRoleID(id)
	// if userCount > 0 { return utils.ErrBadRequest.SetDetails("Cannot delete role with associated accounts") }

	if err := s.roleRepo.Delete(id); err != nil {
		zap.L().Error("Service: Failed to delete role in repository", zap.Error(err), zap.Int("role_id", id))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to delete role: %v", err))
	}
	return nil
}

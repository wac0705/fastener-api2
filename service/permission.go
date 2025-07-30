package service

import (
	"fmt"
	"sync" // 用於緩存的併發安全

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository"
	"github.com/wac0705/fastener-api/utils"
)

// PermissionService 定義權限服務介面
type PermissionService interface {
	HasPermission(roleID int, permission string) (bool, error)
	// 可以新增其他權限管理方法，例如：
	// GetRolePermissions(roleID int) ([]models.Permission, error)
	// AssignPermissionToRole(roleID, permissionID int) error
	// RevokePermissionFromRole(roleID, permissionID int) error
}

// permissionServiceImpl 實現 PermissionService 介面
type permissionServiceImpl struct {
	permissionRepo repository.PermissionRepository
	roleRepo       repository.RoleRepository // 依賴 RoleRepository 以獲取角色信息

	// 考慮新增一個緩存機制來儲存角色-權限映射，避免每次都查詢資料庫
	rolePermissionsCache map[int]map[string]bool // map[roleID]map[permissionName]true
	cacheMutex           sync.RWMutex            // 讀寫鎖保護緩存
}

// NewPermissionService 創建 PermissionService 實例
func NewPermissionService(permissionRepo repository.PermissionRepository, roleRepo repository.RoleRepository) PermissionService {
	s := &permissionServiceImpl{
		permissionRepo:       permissionRepo,
		roleRepo:             roleRepo,
		rolePermissionsCache: make(map[int]map[string]bool),
	}
	// 在服務啟動時預載入一些核心權限到緩存 (可選)
	// s.loadInitialPermissions()
	return s
}

// loadPermissionsForRole 從資料庫載入特定角色的所有權限到緩存
func (s *permissionServiceImpl) loadPermissionsForRole(roleID int) error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	permissions, err := s.permissionRepo.FindPermissionsByRoleID(roleID)
	if err != nil {
		zap.L().Error("Service: Failed to load permissions for role from repository", zap.Error(err), zap.Int("role_id", roleID))
		return fmt.Errorf("failed to load permissions for role %d: %w", roleID, err)
	}

	permissionMap := make(map[string]bool)
	for _, p := range permissions {
		permissionMap[p.Name] = true
	}
	s.rolePermissionsCache[roleID] = permissionMap
	zap.L().Info("Service: Loaded permissions into cache for role", zap.Int("role_id", roleID), zap.Int("count", len(permissionMap)))
	return nil
}

// HasPermission 檢查指定角色是否擁有特定權限
func (s *permissionServiceImpl) HasPermission(roleID int, permission string) (bool, error) {
	// 優先從緩存中讀取
	s.cacheMutex.RLock()
	rolePerms, ok := s.rolePermissionsCache[roleID]
	s.cacheMutex.RUnlock()

	if ok {
		// 緩存命中
		_, has := rolePerms[permission]
		return has, nil
	}

	// 緩存未命中，從資料庫載入
	err := s.loadPermissionsForRole(roleID)
	if err != nil {
		zap.L().Error("Service: Failed to load permissions to cache for role", zap.Error(err), zap.Int("role_id", roleID))
		return false, utils.ErrInternalServer.SetDetails("Failed to retrieve permissions")
	}

	// 再次從緩存中檢查 (因為現在已經載入)
	s.cacheMutex.RLock()
	rolePerms, ok = s.rolePermissionsCache[roleID]
	s.cacheMutex.RUnlock()

	if ok {
		_, has := rolePerms[permission]
		return has, nil
	}

	// 理論上不應該到達這裡，除非 loadPermissionsForRole 失敗但沒有返回錯誤
	zap.L().Error("Service: Permissions not found in cache after load attempt", zap.Int("role_id", roleID), zap.String("permission", permission))
	return false, utils.ErrInternalServer.SetDetails("Could not verify permission")
}

// 以下為範例，如果需要通過 Service 層管理權限賦予/撤銷，可以實現：
/*
func (s *permissionServiceImpl) AssignPermissionToRole(roleID, permissionID int) error {
    // 檢查角色和權限是否存在
    if _, err := s.roleRepo.FindByID(roleID); err != nil || role == nil {
        return utils.ErrBadRequest.SetDetails("Invalid Role ID")
    }
    if _, err := s.permissionRepo.FindByID(permissionID); err != nil || perm == nil {
        return utils.ErrBadRequest.SetDetails("Invalid Permission ID")
    }

    err := s.permissionRepo.AssignPermissionToRole(roleID, permissionID)
    if err != nil {
        zap.L().Error("Service: Failed to assign permission to role in repository", zap.Error(err), zap.Int("role_id", roleID), zap.Int("permission_id", permissionID))
        return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to assign permission: %v", err))
    }
    s.invalidateCache(roleID) // 權限變更後使緩存失效
    return nil
}

func (s *permissionServiceImpl) RevokePermissionFromRole(roleID, permissionID int) error {
    err := s.permissionRepo.RevokePermissionFromRole(roleID, permissionID)
    if err != nil {
        zap.L().Error("Service: Failed to revoke permission from role in repository", zap.Error(err), zap.Int("role_id", roleID), zap.Int("permission_id", permissionID))
        return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to revoke permission: %v", err))
    }
    s.invalidateCache(roleID) // 權限變更後使緩存失效
    return nil
}

// invalidateCache 權限變更後使特定角色的緩存失效
func (s *permissionServiceImpl) invalidateCache(roleID int) {
    s.cacheMutex.Lock()
    defer s.cacheMutex.Unlock()
    delete(s.rolePermissionsCache, roleID)
    zap.L().Info("Service: Invalidated permission cache for role", zap.Int("role_id", roleID))
}
*/

package repository

import (
	"database/sql"
	"fmt"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
)

// PermissionRepository 定義權限資料庫操作介面
type PermissionRepository interface {
	FindByID(id int) (*models.Permission, error)
	FindByName(name string) (*models.Permission, error)
	FindPermissionsByRoleID(roleID int) ([]models.Permission, error) // 獲取某個角色擁有的所有權限
	AssignPermissionToRole(roleID, permissionID int) error
	RevokePermissionFromRole(roleID, permissionID int) error
}

// permissionRepositoryImpl 實現 PermissionRepository 介面
type permissionRepositoryImpl struct {
	db *sql.DB
}

// NewPermissionRepository 創建 PermissionRepository 實例
func NewPermissionRepository(db *sql.DB) PermissionRepository {
	return &permissionRepositoryImpl{db: db}
}

// FindByID 根據 ID 獲取權限
func (r *permissionRepositoryImpl) FindByID(id int) (*models.Permission, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM permissions WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var permission models.Permission
	if err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		zap.L().Error("Repository: Failed to get permission by ID", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get permission by ID %d: %w", id, err)
	}
	return &permission, nil
}

// FindByName 根據名稱獲取權限
func (r *permissionRepositoryImpl) FindByName(name string) (*models.Permission, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM permissions WHERE name = $1`
	row := r.db.QueryRow(query, name)
	var permission models.Permission
	if err := row.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		zap.L().Error("Repository: Failed to get permission by name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to get permission by name %s: %w", name, err)
	}
	return &permission, nil
}

// FindPermissionsByRoleID 獲取某個角色擁有的所有權限
func (r *permissionRepositoryImpl) FindPermissionsByRoleID(roleID int) ([]models.Permission, error) {
	query := `SELECT p.id, p.name, p.description, p.created_at, p.updated_at
              FROM permissions p
              JOIN role_permissions rp ON p.id = rp.permission_id
              WHERE rp.role_id = $1`
	rows, err := r.db.Query(query, roleID)
	if err != nil {
		zap.L().Error("Repository: Failed to get permissions by role ID", zap.Int("role_id", roleID), zap.Error(err))
		return nil, fmt.Errorf("failed to get permissions for role %d: %w", roleID, err)
	}
	defer rows.Close()

	permissions := []models.Permission{}
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			zap.L().Error("Repository: Failed to scan permission data for role", zap.Int("role_id", roleID), zap.Error(err))
			return nil, fmt.Errorf("failed to scan permission data for role %d: %w", roleID, err)
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

// AssignPermissionToRole 將權限賦予角色
func (r *permissionRepositoryImpl) AssignPermissionToRole(roleID, permissionID int) error {
	query := `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT (role_id, permission_id) DO NOTHING`
	_, err := r.db.Exec(query, roleID, permissionID)
	if err != nil {
		zap.L().Error("Repository: Failed to assign permission to role", zap.Error(err), zap.Int("role_id", roleID), zap.Int("permission_id", permissionID))
		return fmt.Errorf("failed to assign permission %d to role %d: %w", permissionID, roleID, err)
	}
	return nil
}

// RevokePermissionFromRole 從角色撤銷權限
func (r *permissionRepositoryImpl) RevokePermissionFromRole(roleID, permissionID int) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`
	res, err := r.db.Exec(query, roleID, permissionID)
	if err != nil {
		zap.L().Error("Repository: Failed to revoke permission from role", zap.Error(err), zap.Int("role_id", roleID), zap.Int("permission_id", permissionID))
		return fmt.Errorf("failed to revoke permission %d from role %d: %w", permissionID, roleID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after revoke", zap.Error(err), zap.Int("role_id", roleID), zap.Int("permission_id", permissionID))
		return fmt.Errorf("failed to check rows affected for revoke %d from %d: %w", permissionID, roleID, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("permission %d not found for role %d", permissionID, roleID) // 沒有找到要刪除的關聯
	}
	return nil
}

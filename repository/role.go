package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/utils"
)

// RoleRepository 定義角色資料庫操作介面
type RoleRepository interface {
	Create(role *models.Role) error
	FindAll() ([]models.Role, error)
	FindByID(id int) (*models.Role, error)
	FindByName(name string) (*models.Role, error) // 根據名稱查找角色
	Update(role *models.Role) error
	Delete(id int) error
}

// roleRepositoryImpl 實現 RoleRepository 介面
type roleRepositoryImpl struct {
	db *sql.DB
}

// NewRoleRepository 創建 RoleRepository 實例
func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepositoryImpl{db: db}
}

// Create 創建新角色
func (r *roleRepositoryImpl) Create(role *models.Role) error {
	query := `INSERT INTO roles (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(query, role.Name).
		Scan(&role.ID, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		zap.L().Error("Repository: Failed to create role", zap.Error(err), zap.String("name", role.Name))
		// 檢查是否是唯一約束衝突錯誤
		if err.Error() == `pq: duplicate key value violates unique constraint "roles_name_key"` {
			return utils.ErrBadRequest.SetDetails("Role name already exists")
		}
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

// FindAll 獲取所有角色
func (r *roleRepositoryImpl) FindAll() ([]models.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles`
	rows, err := r.db.Query(query)
	if err != nil {
		zap.L().Error("Repository: Failed to get all roles", zap.Error(err))
		return nil, fmt.Errorf("failed to get all roles: %w", err)
	}
	defer rows.Close()

	roles := []models.Role{}
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			zap.L().Error("Repository: Failed to scan role data", zap.Error(err))
			return nil, fmt.Errorf("failed to scan role data: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// FindByID 根據 ID 獲取角色
func (r *roleRepositoryImpl) FindByID(id int) (*models.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var role models.Role
	if err := row.Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到
		}
		zap.L().Error("Repository: Failed to get role by ID", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get role by ID %d: %w", id, err)
	}
	return &role, nil
}

// FindByName 根據名稱獲取角色
func (r *roleRepositoryImpl) FindByName(name string) (*models.Role, error) {
	query := `SELECT id, name, created_at, updated_at FROM roles WHERE name = $1`
	row := r.db.QueryRow(query, name)
	var role models.Role
	if err := row.Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到
		}
		zap.L().Error("Repository: Failed to get role by name", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to get role by name %s: %w", name, err)
	}
	return &role, nil
}

// Update 更新角色信息
func (r *roleRepositoryImpl) Update(role *models.Role) error {
	query := `UPDATE roles SET name = $1, updated_at = NOW() WHERE id = $2 RETURNING updated_at`
	err := r.db.QueryRow(query, role.Name, role.ID).Scan(&role.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrNotFound // 未找到要更新的記錄
		}
		zap.L().Error("Repository: Failed to update role", zap.Error(err), zap.Int("id", role.ID))
		// 檢查是否是唯一約束衝突錯誤
		if err.Error() == `pq: duplicate key value violates unique constraint "roles_name_key"` {
			return utils.ErrBadRequest.SetDetails("Role name already exists")
		}
		return fmt.Errorf("failed to update role %d: %w", role.ID, err)
	}
	return nil
}

// Delete 刪除角色
func (r *roleRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM roles WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		zap.L().Error("Repository: Failed to delete role", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("failed to delete role %d: %w", id, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after delete", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("failed to check delete rows affected %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return utils.ErrNotFound // 未找到要刪除的記錄
	}
	return nil
}

package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/utils"
)

// RoleMenuRepository 定義角色選單資料庫操作介面
type RoleMenuRepository interface {
	Create(roleMenu *models.RoleMenu) error
	FindAll(roleID, menuID *int) ([]models.RoleMenuDetail, error) // 允許按角色或選單ID過濾
	Delete(roleID, menuID int) error
	Update(oldRoleID, oldMenuID, newRoleID, newMenuID int) error // 由於複合主鍵，更新是特殊操作
	FindMenusByRoleID(roleID int) ([]models.Menu, error) // 新增：根據角色ID獲取所有選單
}

// roleMenuRepositoryImpl 實現 RoleMenuRepository 介面
type roleMenuRepositoryImpl struct {
	db *sql.DB
}

// NewRoleMenuRepository 創建 RoleMenuRepository 實例
func NewRoleMenuRepository(db *sql.DB) RoleMenuRepository {
	return &roleMenuRepositoryImpl{db: db}
}

// Create 創建新的角色選單關聯
func (r *roleMenuRepositoryImpl) Create(roleMenu *models.RoleMenu) error {
	query := `INSERT INTO role_menus (role_id, menu_id) VALUES ($1, $2) ON CONFLICT (role_id, menu_id) DO NOTHING`
	_, err := r.db.Exec(query, roleMenu.RoleID, roleMenu.MenuID)
	if err != nil {
		zap.L().Error("Repository: Failed to create role menu", zap.Error(err), zap.Int("role_id", roleMenu.RoleID), zap.Int("menu_id", roleMenu.MenuID))
		return fmt.Errorf("failed to create role menu: %w", err)
	}
	return nil
}

// FindAll 獲取所有角色選單關聯，並帶上詳細資訊
func (r *roleMenuRepositoryImpl) FindAll(roleIDFilter, menuIDFilter *int) ([]models.RoleMenuDetail, error) {
	query := `SELECT rm.role_id, r.name AS role_name, rm.menu_id, m.name AS menu_name, m.path AS menu_path
              FROM role_menus rm
              JOIN roles r ON rm.role_id = r.id
              JOIN menus m ON rm.menu_id = m.id
              WHERE TRUE` // TRUE 允許動態添加 WHERE 條件

	args := []interface{}{}
	argCounter := 1

	if roleIDFilter != nil {
		query += fmt.Sprintf(" AND rm.role_id = $%d", argCounter)
		args = append(args, *roleIDFilter)
		argCounter++
	}
	if menuIDFilter != nil {
		query += fmt.Sprintf(" AND rm.menu_id = $%d", argCounter)
		args = append(args, *menuIDFilter)
		argCounter++
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		zap.L().Error("Repository: Failed to get all role menus", zap.Error(err))
		return nil, fmt.Errorf("failed to get all role menus: %w", err)
	}
	defer rows.Close()

	roleMenus := []models.RoleMenuDetail{}
	for rows.Next() {
		var rm models.RoleMenuDetail
		if err := rows.Scan(&rm.RoleID, &rm.RoleName, &rm.MenuID, &rm.MenuName, &rm.MenuPath); err != nil {
			zap.L().Error("Repository: Failed to scan role menu data", zap.Error(err))
			return nil, fmt.Errorf("failed to scan role menu data: %w", err)
		}
		roleMenus = append(roleMenus, rm)
	}
	return roleMenus, nil
}

// Delete 刪除角色選單關聯
func (r *roleMenuRepositoryImpl) Delete(roleID, menuID int) error {
	query := `DELETE FROM role_menus WHERE role_id = $1 AND menu_id = $2`
	res, err := r.db.Exec(query, roleID, menuID)
	if err != nil {
		zap.L().Error("Repository: Failed to delete role menu", zap.Error(err), zap.Int("role_id", roleID), zap.Int("menu_id", menuID))
		return fmt.Errorf("failed to delete role menu %d-%d: %w", roleID, menuID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after delete", zap.Error(err), zap.Int("role_id", roleID), zap.Int("menu_id", menuID))
		return fmt.Errorf("failed to check delete rows affected %d-%d: %w", roleID, menuID, err)
	}
	if rowsAffected == 0 {
		return utils.ErrNotFound.SetDetails(fmt.Sprintf("Role menu relationship role_id %d, menu_id %d not found", roleID, menuID))
	}
	return nil
}

// Update 更新角色選單關聯
// 由於複合主鍵，這實際上是先刪除舊關聯，再創建新關聯。
func (r *roleMenuRepositoryImpl) Update(oldRoleID, oldMenuID, newRoleID, newMenuID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		zap.L().Error("Repository: Failed to begin transaction for role menu update", zap.Error(err))
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() // 確保在函數返回前回滾，除非明確提交

	// 1. 刪除舊的關聯
	deleteQuery := `DELETE FROM role_menus WHERE role_id = $1 AND menu_id = $2`
	res, err := tx.Exec(deleteQuery, oldRoleID, oldMenuID)
	if err != nil {
		zap.L().Error("Repository: Failed to delete old role menu for update", zap.Error(err),
			zap.Int("old_role_id", oldRoleID), zap.Int("old_menu_id", oldMenuID))
		return fmt.Errorf("failed to delete old role menu: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after delete for update", zap.Error(err))
		return fmt.Errorf("failed to check deleted rows: %w", err)
	}
	if rowsAffected == 0 {
		return utils.ErrNotFound.SetDetails(fmt.Sprintf("Old role menu relationship %d-%d not found for update", oldRoleID, oldMenuID))
	}

	// 2. 創建新的關聯
	createQuery := `INSERT INTO role_menus (role_id, menu_id) VALUES ($1, $2) ON CONFLICT (role_id, menu_id) DO NOTHING`
	_, err = tx.Exec(createQuery, newRoleID, newMenuID)
	if err != nil {
		zap.L().Error("Repository: Failed to create new role menu for update", zap.Error(err),
			zap.Int("new_role_id", newRoleID), zap.Int("new_menu_id", newMenuID))
		return fmt.Errorf("failed to create new role menu: %w", err)
	}

	return tx.Commit() // 提交事務
}

// FindMenusByRoleID 根據角色 ID 獲取該角色能訪問的所有選單
func (r *roleMenuRepositoryImpl) FindMenusByRoleID(roleID int) ([]models.Menu, error) {
	query := `SELECT m.id, m.name, m.path, m.icon, m.parent_id, m.display_order, m.created_at, m.updated_at
              FROM menus m
              JOIN role_menus rm ON m.id = rm.menu_id
              WHERE rm.role_id = $1
              ORDER BY m.display_order ASC`
	rows, err := r.db.Query(query, roleID)
	if err != nil {
		zap.L().Error("Repository: Failed to get menus by role ID", zap.Int("role_id", roleID), zap.Error(err))
		return nil, fmt.Errorf("failed to get menus for role %d: %w", roleID, err)
	}
	defer rows.Close()

	menus := []models.Menu{}
	for rows.Next() {
		var menu models.Menu
		var parentID sql.NullInt64
		if err := rows.Scan(
			&menu.ID,
			&menu.Name,
			&menu.Path,
			&menu.Icon,
			&parentID,
			&menu.DisplayOrder,
			&menu.CreatedAt,
			&menu.UpdatedAt,
		); err != nil {
			zap.L().Error("Repository: Failed to scan menu data for role", zap.Int("role_id", roleID), zap.Error(err))
			return nil, fmt.Errorf("failed to scan menu data for role %d: %w", roleID, err)
		}
		if parentID.Valid {
			menu.ParentID = new(int)
			*menu.ParentID = int(parentID.Int64)
		} else {
			menu.ParentID = nil
		}
		menus = append(menus, menu)
	}
	return menus, nil
}

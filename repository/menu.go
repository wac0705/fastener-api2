package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/utils"
)

// MenuRepository 定義選單資料庫操作介面
type MenuRepository interface {
	Create(menu *models.Menu) error
	FindAll() ([]models.Menu, error)
	FindByID(id int) (*models.Menu, error)
	Update(menu *models.Menu) error
	Delete(id int) error
}

// menuRepositoryImpl 實現 MenuRepository 介面
type menuRepositoryImpl struct {
	db *sql.DB
}

// NewMenuRepository 創建 MenuRepository 實例
func NewMenuRepository(db *sql.DB) MenuRepository {
	return &menuRepositoryImpl{db: db}
}

// Create 創建新選單
func (r *menuRepositoryImpl) Create(menu *models.Menu) error {
	query := `INSERT INTO menus (name, path, icon, parent_id, display_order) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	var parentID sql.NullInt64
	if menu.ParentID != nil {
		parentID = sql.NullInt64{Int64: int64(*menu.ParentID), Valid: true}
	} else {
		parentID = sql.NullInt64{Valid: false}
	}

	err := r.db.QueryRow(query, menu.Name, menu.Path, menu.Icon, parentID, menu.DisplayOrder).
		Scan(&menu.ID, &menu.CreatedAt, &menu.UpdatedAt)
	if err != nil {
		zap.L().Error("Repository: Failed to create menu", zap.Error(err), zap.String("name", menu.Name))
		// 檢查是否是唯一約束衝突錯誤 (例如，path 已存在)
		if err.Error() == `pq: duplicate key value violates unique constraint "menus_path_key"` {
			return utils.ErrBadRequest.SetDetails("Menu path already exists")
		}
		return fmt.Errorf("failed to create menu: %w", err)
	}
	return nil
}

// FindAll 獲取所有選單
func (r *menuRepositoryImpl) FindAll() ([]models.Menu, error) {
	query := `SELECT id, name, path, icon, parent_id, display_order, created_at, updated_at FROM menus ORDER BY display_order ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		zap.L().Error("Repository: Failed to get all menus", zap.Error(err))
		return nil, fmt.Errorf("failed to get all menus: %w", err)
	}
	defer rows.Close()

	menus := []models.Menu{}
	for rows.Next() {
		var menu models.Menu
		var parentID sql.NullInt64 // 用於處理 NULLABLE 的 parent_id
		if err := rows.Scan(
			&menu.ID,
			&menu.Name,
			&menu.Path,
			&menu.Icon,
			&parentID, // Scan 到 sql.NullInt64
			&menu.DisplayOrder,
			&menu.CreatedAt,
			&menu.UpdatedAt,
		); err != nil {
			zap.L().Error("Repository: Failed to scan menu data", zap.Error(err))
			return nil, fmt.Errorf("failed to scan menu data: %w", err)
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

// FindByID 根據 ID 獲取選單
func (r *menuRepositoryImpl) FindByID(id int) (*models.Menu, error) {
	query := `SELECT id, name, path, icon, parent_id, display_order, created_at, updated_at FROM menus WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var menu models.Menu
	var parentID sql.NullInt64
	if err := row.Scan(
		&menu.ID,
		&menu.Name,
		&menu.Path,
		&menu.Icon,
		&parentID,
		&menu.DisplayOrder,
		&menu.CreatedAt,
		&menu.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到
		}
		zap.L().Error("Repository: Failed to get menu by ID", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get menu by ID %d: %w", id, err)
	}
	if parentID.Valid {
		menu.ParentID = new(int)
		*menu.ParentID = int(parentID.Int64)
	} else {
		menu.ParentID = nil
	}
	return &menu, nil
}

// Update 更新選單信息
func (r *menuRepositoryImpl) Update(menu *models.Menu) error {
	query := `UPDATE menus SET name = $1, path = $2, icon = $3, parent_id = $4, display_order = $5, updated_at = NOW() WHERE id = $6 RETURNING updated_at`
	var parentID sql.NullInt64
	if menu.ParentID != nil {
		parentID = sql.NullInt64{Int64: int64(*menu.ParentID), Valid: true}
	} else {
		parentID = sql.NullInt64{Valid: false}
	}

	res, err := r.db.Exec(query,
		menu.Name,
		menu.Path,
		menu.Icon,
		parentID,
		menu.DisplayOrder,
		menu.ID,
	)
	if err != nil {
		zap.L().Error("Repository: Failed to update menu", zap.Error(err), zap.Int("id", menu.ID))
		// 檢查是否是唯一約束衝突錯誤
		if err.Error() == `pq: duplicate key value violates unique constraint "menus_path_key"` {
			return utils.ErrBadRequest.SetDetails("Menu path already exists")
		}
		return fmt.Errorf("failed to update menu %d: %w", menu.ID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after update", zap.Error(err), zap.Int("id", menu.ID))
		return fmt.Errorf("failed to check update rows affected %d: %w", menu.ID, err)
	}
	if rowsAffected == 0 {
		return utils.ErrNotFound // 未找到要更新的記錄
	}
	// 重新讀取 updated_at
	row := r.db.QueryRow(`SELECT updated_at FROM menus WHERE id = $1`, menu.ID)
	if err := row.Scan(&menu.UpdatedAt); err != nil {
		zap.L().Error("Repository: Failed to scan updated_at after update", zap.Error(err), zap.Int("id", menu.ID))
		return fmt.Errorf("failed to scan updated_at for menu %d: %w", menu.ID, err)
	}
	return nil
}

// Delete 刪除選單
func (r *menuRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM menus WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		zap.L().Error("Repository: Failed to delete menu", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("failed to delete menu %d: %w", id, err)
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

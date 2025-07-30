package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/utils"
)

// AccountRepository 定義帳戶資料庫操作介面
type AccountRepository interface {
	Create(account *models.Account) error
	FindAll() ([]models.Account, error)
	FindByID(id int) (*models.Account, error)
	FindByUsername(username string) (*models.Account, error)
	Update(account *models.Account) error
	Delete(id int) error
	UpdatePassword(accountID int, hashedPassword string) error
	UpdateAdminPassword(username, hashedPassword string) error // 專門為 resetadmin 工具提供的方法
}

// accountRepositoryImpl 實現 AccountRepository 介面
type accountRepositoryImpl struct {
	db *sql.DB
}

// NewAccountRepository 創建 AccountRepository 實例
func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepositoryImpl{db: db}
}

// Create 創建新帳戶
func (r *accountRepositoryImpl) Create(account *models.Account) error {
	query := `INSERT INTO accounts (username, password, role_id) VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(query, account.Username, account.Password, account.RoleID).
		Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		zap.L().Error("Repository: Failed to create account", zap.Error(err), zap.String("username", account.Username))
		return fmt.Errorf("failed to create account: %w", err) // 包裝原始錯誤
	}
	return nil
}

// FindAll 獲取所有帳戶，並帶上角色名稱
func (r *accountRepositoryImpl) FindAll() ([]models.Account, error) {
	query := `SELECT a.id, a.username, a.role_id, r.name AS role_name, a.created_at, a.updated_at
              FROM accounts a
              JOIN roles r ON a.role_id = r.id`
	rows, err := r.db.Query(query)
	if err != nil {
		zap.L().Error("Repository: Failed to get all accounts", zap.Error(err))
		return nil, fmt.Errorf("failed to get all accounts: %w", err)
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var account models.Account
		if err := rows.Scan(&account.ID, &account.Username, &account.RoleID, &account.RoleName, &account.CreatedAt, &account.UpdatedAt); err != nil {
			zap.L().Error("Repository: Failed to scan account data", zap.Error(err))
			return nil, fmt.Errorf("failed to scan account data: %w", err)
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// FindByID 根據 ID 獲取帳戶，並帶上角色名稱
func (r *accountRepositoryImpl) FindByID(id int) (*models.Account, error) {
	query := `SELECT a.id, a.username, a.role_id, r.name AS role_name, a.created_at, a.updated_at
              FROM accounts a
              JOIN roles r ON a.role_id = r.id
              WHERE a.id = $1`
	row := r.db.QueryRow(query, id)
	var account models.Account
	if err := row.Scan(&account.ID, &account.Username, &account.RoleID, &account.RoleName, &account.CreatedAt, &account.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到
		}
		zap.L().Error("Repository: Failed to get account by ID", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get account by ID %d: %w", id, err)
	}
	return &account, nil
}

// FindByUsername 根據用戶名獲取帳戶
func (r *accountRepositoryImpl) FindByUsername(username string) (*models.Account, error) {
	query := `SELECT a.id, a.username, a.password, a.role_id, r.name AS role_name, a.created_at, a.updated_at
              FROM accounts a
              JOIN roles r ON a.role_id = r.id
              WHERE a.username = $1`
	row := r.db.QueryRow(query, username)
	var account models.Account
	if err := row.Scan(&account.ID, &account.Username, &account.Password, &account.RoleID, &account.RoleName, &account.CreatedAt, &account.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到
		}
		zap.L().Error("Repository: Failed to get account by username", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("failed to get account by username %s: %w", username, err)
	}
	return &account, nil
}

// Update 更新帳戶信息
func (r *accountRepositoryImpl) Update(account *models.Account) error {
	query := `UPDATE accounts SET username = $1, role_id = $2, updated_at = NOW() WHERE id = $3 RETURNING updated_at`
	err := r.db.QueryRow(query, account.Username, account.RoleID, account.ID).Scan(&account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrNotFound // 未找到要更新的記錄
		}
		zap.L().Error("Repository: Failed to update account", zap.Error(err), zap.Int("id", account.ID))
		return fmt.Errorf("failed to update account %d: %w", account.ID, err)
	}
	return nil
}

// Delete 刪除帳戶
func (r *accountRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM accounts WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		zap.L().Error("Repository: Failed to delete account", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("failed to delete account %d: %w", id, err)
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

// UpdatePassword 更新帳戶密碼
func (r *accountRepositoryImpl) UpdatePassword(accountID int, hashedPassword string) error {
	query := `UPDATE accounts SET password = $1, updated_at = NOW() WHERE id = $2 RETURNING updated_at`
	res, err := r.db.Exec(query, hashedPassword, accountID)
	if err != nil {
		zap.L().Error("Repository: Failed to update password", zap.Error(err), zap.Int("account_id", accountID))
		return fmt.Errorf("failed to update password for account %d: %w", accountID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after password update", zap.Error(err), zap.Int("account_id", accountID))
		return fmt.Errorf("failed to check rows affected for password update %d: %w", accountID, err)
	}
	if rowsAffected == 0 {
		return utils.ErrNotFound // 未找到要更新的記錄
	}
	return nil
}

// UpdateAdminPassword 專門用於重設管理員密碼的工具
func (r *accountRepositoryImpl) UpdateAdminPassword(username, hashedPassword string) error {
	query := `UPDATE accounts SET password = $1, updated_at = NOW() WHERE username = $2 AND role_id = (SELECT id FROM roles WHERE name = 'admin')`
	res, err := r.db.Exec(query, hashedPassword, username)
	if err != nil {
		zap.L().Error("Repository: Failed to update admin password", zap.Error(err), zap.String("username", username))
		return fmt.Errorf("failed to update admin password for '%s': %w", username, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after admin password update", zap.Error(err), zap.String("username", username))
		return fmt.Errorf("failed to check rows affected for admin password update '%s': %w", username, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("admin account '%s' not found or not an admin role", username)
	}
	return nil
}

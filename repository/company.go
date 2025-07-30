package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/utils"
)

// CompanyRepository 定義公司資料庫操作介面
type CompanyRepository interface {
	Create(company *models.Company) error
	FindAll() ([]models.Company, error)
	FindByID(id int) (*models.Company, error)
	Update(company *models.Company) error
	Delete(id int) error
}

// companyRepositoryImpl 實現 CompanyRepository 介面
type companyRepositoryImpl struct {
	db *sql.DB
}

// NewCompanyRepository 創建 CompanyRepository 實例
func NewCompanyRepository(db *sql.DB) CompanyRepository {
	return &companyRepositoryImpl{db: db}
}

// Create 創建新公司
func (r *companyRepositoryImpl) Create(company *models.Company) error {
	query := `INSERT INTO companies (name) VALUES ($1) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(query, company.Name).
		Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)
	if err != nil {
		zap.L().Error("Repository: Failed to create company", zap.Error(err), zap.String("name", company.Name))
		// 檢查是否是唯一約束衝突錯誤 (例如，公司名稱已存在)
		if err.Error() == `pq: duplicate key value violates unique constraint "companies_name_key"` { // 這是 PostgreSQL 特有的錯誤訊息
			return utils.ErrBadRequest.SetDetails("Company name already exists")
		}
		return fmt.Errorf("failed to create company: %w", err)
	}
	return nil
}

// FindAll 獲取所有公司
func (r *companyRepositoryImpl) FindAll() ([]models.Company, error) {
	query := `SELECT id, name, created_at, updated_at FROM companies`
	rows, err := r.db.Query(query)
	if err != nil {
		zap.L().Error("Repository: Failed to get all companies", zap.Error(err))
		return nil, fmt.Errorf("failed to get all companies: %w", err)
	}
	defer rows.Close()

	companies := []models.Company{}
	for rows.Next() {
		var company models.Company
		if err := rows.Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt); err != nil {
			zap.L().Error("Repository: Failed to scan company data", zap.Error(err))
			return nil, fmt.Errorf("failed to scan company data: %w", err)
		}
		companies = append(companies, company)
	}
	return companies, nil
}

// FindByID 根據 ID 獲取公司
func (r *companyRepositoryImpl) FindByID(id int) (*models.Company, error) {
	query := `SELECT id, name, created_at, updated_at FROM companies WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var company models.Company
	if err := row.Scan(&company.ID, &company.Name, &company.CreatedAt, &company.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到
		}
		zap.L().Error("Repository: Failed to get company by ID", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get company by ID %d: %w", id, err)
	}
	return &company, nil
}

// Update 更新公司信息
func (r *companyRepositoryImpl) Update(company *models.Company) error {
	query := `UPDATE companies SET name = $1, updated_at = NOW() WHERE id = $2 RETURNING updated_at`
	err := r.db.QueryRow(query, company.Name, company.ID).Scan(&company.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return utils.ErrNotFound // 未找到要更新的記錄
		}
		zap.L().Error("Repository: Failed to update company", zap.Error(err), zap.Int("id", company.ID))
		// 檢查是否是唯一約束衝突錯誤
		if err.Error() == `pq: duplicate key value violates unique constraint "companies_name_key"` {
			return utils.ErrBadRequest.SetDetails("Company name already exists")
		}
		return fmt.Errorf("failed to update company %d: %w", company.ID, err)
	}
	return nil
}

// Delete 刪除公司
func (r *companyRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM companies WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		zap.L().Error("Repository: Failed to delete company", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("failed to delete company %d: %w", id, err)
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

package repository

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/utils"
)

// CustomerRepository 定義客戶資料庫操作介面
type CustomerRepository interface {
	Create(customer *models.Customer) error
	FindAll() ([]models.Customer, error)
	FindByID(id int) (*models.Customer, error)
	Update(customer *models.Customer) error
	Delete(id int) error
}

// customerRepositoryImpl 實現 CustomerRepository 介面
type customerRepositoryImpl struct {
	db *sql.DB
}

// NewCustomerRepository 創建 CustomerRepository 實例
func NewCustomerRepository(db *sql.DB) CustomerRepository {
	return &customerRepositoryImpl{db: db}
}

// Create 創建新客戶
func (r *customerRepositoryImpl) Create(customer *models.Customer) error {
	query := `INSERT INTO customers (name, contact_person, email, phone, company_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(query,
		customer.Name,
		customer.ContactPerson,
		customer.Email,
		customer.Phone,
		customer.CompanyID,
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)
	if err != nil {
		zap.L().Error("Repository: Failed to create customer", zap.Error(err), zap.String("name", customer.Name))
		return fmt.Errorf("failed to create customer: %w", err)
	}
	return nil
}

// FindAll 獲取所有客戶
func (r *customerRepositoryImpl) FindAll() ([]models.Customer, error) {
	query := `SELECT id, name, contact_person, email, phone, company_id, created_at, updated_at FROM customers`
	rows, err := r.db.Query(query)
	if err != nil {
		zap.L().Error("Repository: Failed to get all customers", zap.Error(err))
		return nil, fmt.Errorf("failed to get all customers: %w", err)
	}
	defer rows.Close()

	customers := []models.Customer{}
	for rows.Next() {
		var customer models.Customer
		// 注意這裡對 company_id 的處理，因為它是 NULLABLE
		var companyID sql.NullInt64
		if err := rows.Scan(
			&customer.ID,
			&customer.Name,
			&customer.ContactPerson,
			&customer.Email,
			&customer.Phone,
			&companyID, // Scan 到 sql.NullInt64
			&customer.CreatedAt,
			&customer.UpdatedAt,
		); err != nil {
			zap.L().Error("Repository: Failed to scan customer data", zap.Error(err))
			return nil, fmt.Errorf("failed to scan customer data: %w", err)
		}
		if companyID.Valid {
			customer.CompanyID = new(int)
			*customer.CompanyID = int(companyID.Int64)
		} else {
			customer.CompanyID = nil
		}
		customers = append(customers, customer)
	}
	return customers, nil
}

// FindByID 根據 ID 獲取客戶
func (r *customerRepositoryImpl) FindByID(id int) (*models.Customer, error) {
	query := `SELECT id, name, contact_person, email, phone, company_id, created_at, updated_at FROM customers WHERE id = $1`
	row := r.db.QueryRow(query, id)
	var customer models.Customer
	var companyID sql.NullInt64 // 用於處理 NULLABLE 的 company_id
	if err := row.Scan(
		&customer.ID,
		&customer.Name,
		&customer.ContactPerson,
		&customer.Email,
		&customer.Phone,
		&companyID,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 未找到
		}
		zap.L().Error("Repository: Failed to get customer by ID", zap.Int("id", id), zap.Error(err))
		return nil, fmt.Errorf("failed to get customer by ID %d: %w", id, err)
	}
	if companyID.Valid {
		customer.CompanyID = new(int)
		*customer.CompanyID = int(companyID.Int64)
	} else {
		customer.CompanyID = nil
	}
	return &customer, nil
}

// Update 更新客戶信息
func (r *customerRepositoryImpl) Update(customer *models.Customer) error {
	query := `UPDATE customers SET name = $1, contact_person = $2, email = $3, phone = $4, company_id = $5, updated_at = NOW() WHERE id = $6 RETURNING updated_at`
	res, err := r.db.Exec(query,
		customer.Name,
		customer.ContactPerson,
		customer.Email,
		customer.Phone,
		customer.CompanyID,
		customer.ID,
	)
	if err != nil {
		zap.L().Error("Repository: Failed to update customer", zap.Error(err), zap.Int("id", customer.ID))
		return fmt.Errorf("failed to update customer %d: %w", customer.ID, err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		zap.L().Error("Repository: Failed to get rows affected after update", zap.Error(err), zap.Int("id", customer.ID))
		return fmt.Errorf("failed to check update rows affected %d: %w", customer.ID, err)
	}
	if rowsAffected == 0 {
		return utils.ErrNotFound // 未找到要更新的記錄
	}
	// 重新讀取 updated_at
	row := r.db.QueryRow(`SELECT updated_at FROM customers WHERE id = $1`, customer.ID)
	if err := row.Scan(&customer.UpdatedAt); err != nil {
		zap.L().Error("Repository: Failed to scan updated_at after update", zap.Error(err), zap.Int("id", customer.ID))
		return fmt.Errorf("failed to scan updated_at for customer %d: %w", customer.ID, err)
	}
	return nil
}

// Delete 刪除客戶
func (r *customerRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM customers WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		zap.L().Error("Repository: Failed to delete customer", zap.Error(err), zap.Int("id", id))
		return fmt.Errorf("failed to delete customer %d: %w", id, err)
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

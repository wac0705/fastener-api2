package service

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository"
	"github.com/wac0705/fastener-api/utils"
)

// CustomerService 定義客戶服務介面
type CustomerService interface {
	GetAllCustomers() ([]models.Customer, error)
	GetCustomerByID(id int) (*models.Customer, error)
	CreateCustomer(customer *models.Customer) error
	UpdateCustomer(customer *models.Customer) error
	DeleteCustomer(id int) error
}

// customerServiceImpl 實現 CustomerService 介面
type customerServiceImpl struct {
	customerRepo repository.CustomerRepository
	companyRepo  repository.CompanyRepository // 依賴 CompanyRepository 檢查公司是否存在
}

// NewCustomerService 創建 CustomerService 實例
func NewCustomerService(customerRepo repository.CustomerRepository, companyRepo repository.CompanyRepository) CustomerService {
	return &customerServiceImpl{customerRepo: customerRepo, companyRepo: companyRepo}
}

// CreateCustomer 創建新客戶
func (s *customerServiceImpl) CreateCustomer(customer *models.Customer) error {
	// 如果提供了 company_id，檢查公司是否存在
	if customer.CompanyID != nil {
		company, err := s.companyRepo.FindByID(*customer.CompanyID)
		if err != nil {
			zap.L().Error("Service: Error checking company ID for new customer", zap.Error(err), zap.Int("company_id", *customer.CompanyID))
			return utils.ErrInternalServer
		}
		if company == nil {
			return utils.ErrBadRequest.SetDetails("Provided Company ID does not exist.")
		}
	}

	if err := s.customerRepo.Create(customer); err != nil {
		zap.L().Error("Service: Failed to create customer in repository", zap.Error(err), zap.String("name", customer.Name))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to create customer: %v", err))
	}
	return nil
}

// GetAllCustomers 獲取所有客戶
func (s *customerServiceImpl) GetAllCustomers() ([]models.Customer, error) {
	customers, err := s.customerRepo.FindAll()
	if err != nil {
		zap.L().Error("Service: Failed to get all customers", zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	return customers, nil
}

// GetCustomerByID 根據 ID 獲取客戶
func (s *customerServiceImpl) GetCustomerByID(id int) (*models.Customer, error) {
	customer, err := s.customerRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Failed to get customer by ID", zap.Int("id", id), zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	if customer == nil {
		return nil, nil // Repository 返回 nil, nil 表示未找到
	}
	return customer, nil
}

// UpdateCustomer 更新客戶信息
func (s *customerServiceImpl) UpdateCustomer(customer *models.Customer) error {
	// 檢查客戶是否存在
	existingCustomer, err := s.customerRepo.FindByID(customer.ID)
	if err != nil {
		zap.L().Error("Service: Error checking existing customer for update", zap.Error(err), zap.Int("customer_id", customer.ID))
		return utils.ErrInternalServer
	}
	if existingCustomer == nil {
		return utils.ErrNotFound
	}

	// 如果提供了新的 company_id，檢查公司是否存在
	if customer.CompanyID != nil {
		company, err := s.companyRepo.FindByID(*customer.CompanyID)
		if err != nil {
			zap.L().Error("Service: Error checking company ID for customer update", zap.Error(err), zap.Int("company_id", *customer.CompanyID))
			return utils.ErrInternalServer
		}
		if company == nil {
			return utils.ErrBadRequest.SetDetails("Provided Company ID for update does not exist.")
		}
	}

	if err := s.customerRepo.Update(customer); err != nil {
		zap.L().Error("Service: Failed to update customer in repository", zap.Error(err), zap.Int("customer_id", customer.ID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to update customer: %v", err))
	}
	return nil
}

// DeleteCustomer 刪除客戶
func (s *customerServiceImpl) DeleteCustomer(id int) error {
	// 檢查客戶是否存在
	existingCustomer, err := s.customerRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Error checking existing customer for delete", zap.Error(err), zap.Int("customer_id", id))
		return utils.ErrInternalServer
	}
	if existingCustomer == nil {
		return utils.ErrNotFound
	}

	if err := s.customerRepo.Delete(id); err != nil {
		zap.L().Error("Service: Failed to delete customer in repository", zap.Error(err), zap.Int("customer_id", id))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to delete customer: %v", err))
	}
	return nil
}

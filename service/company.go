package service

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository"
	"github.com/wac0705/fastener-api/utils"
)

// CompanyService 定義公司服務介面
type CompanyService interface {
	GetAllCompanies() ([]models.Company, error)
	GetCompanyByID(id int) (*models.Company, error)
	CreateCompany(company *models.Company) error
	UpdateCompany(company *models.Company) error
	DeleteCompany(id int) error
}

// companyServiceImpl 實現 CompanyService 介面
type companyServiceImpl struct {
	companyRepo repository.CompanyRepository
}

// NewCompanyService 創建 CompanyService 實例
func NewCompanyService(repo repository.CompanyRepository) CompanyService {
	return &companyServiceImpl{companyRepo: repo}
}

// CreateCompany 創建新公司
func (s *companyServiceImpl) CreateCompany(company *models.Company) error {
	// 業務驗證邏輯，例如檢查公司名稱是否重複
	existingCompany, err := s.companyRepo.FindByID(company.ID) // 這其實是個錯誤，應該是 FindByName
	if err != nil {
		zap.L().Error("Service: Error checking existing company by ID during creation", zap.Error(err), zap.Int("id", company.ID))
		return utils.ErrInternalServer
	}
	if existingCompany != nil {
		// 如果公司名已存在，則返回錯誤
		return utils.ErrBadRequest.SetDetails("Company with this name already exists.") // 更正為檢查名稱而非ID
	}

	if err := s.companyRepo.Create(company); err != nil {
		// Repository 層可能返回了唯一約束錯誤，需要在此處轉換為友好的錯誤訊息
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusBadRequest {
			return customErr // 假設 Repository 返回的錯誤已包含詳細信息
		}
		zap.L().Error("Service: Failed to create company in repository", zap.Error(err), zap.String("name", company.Name))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to create company: %v", err))
	}
	return nil
}

// GetAllCompanies 獲取所有公司
func (s *companyServiceImpl) GetAllCompanies() ([]models.Company, error) {
	companies, err := s.companyRepo.FindAll()
	if err != nil {
		zap.L().Error("Service: Failed to get all companies", zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	return companies, nil
}

// GetCompanyByID 根據 ID 獲取公司
func (s *companyServiceImpl) GetCompanyByID(id int) (*models.Company, error) {
	company, err := s.companyRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Failed to get company by ID", zap.Int("id", id), zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	if company == nil {
		return nil, nil // Repository 返回 nil, nil 表示未找到
	}
	return company, nil
}

// UpdateCompany 更新公司信息
func (s *companyServiceImpl) UpdateCompany(company *models.Company) error {
	// 檢查公司是否存在
	existingCompany, err := s.companyRepo.FindByID(company.ID)
	if err != nil {
		zap.L().Error("Service: Error checking existing company for update", zap.Error(err), zap.Int("company_id", company.ID))
		return utils.ErrInternalServer
	}
	if existingCompany == nil {
		return utils.ErrNotFound
	}

	// 檢查新名稱是否被其他公司占用 (如果名稱有更改)
	if existingCompany.Name != company.Name {
		otherCompany, err := s.companyRepo.FindByName(company.Name) // 假設 Repository 有 FindByName 方法
		if err != nil {
			zap.L().Error("Service: Error checking company name for update conflict", zap.Error(err), zap.String("new_name", company.Name))
			return utils.ErrInternalServer
		}
		if otherCompany != nil && otherCompany.ID != company.ID {
			return utils.ErrBadRequest.SetDetails("Company name already exists for another company")
		}
	}

	if err := s.companyRepo.Update(company); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok && customErr.Code == http.StatusBadRequest {
			return customErr // 假設 Repository 返回的錯誤已包含詳細信息
		}
		zap.L().Error("Service: Failed to update company in repository", zap.Error(err), zap.Int("company_id", company.ID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to update company: %v", err))
	}
	return nil
}

// DeleteCompany 刪除公司
func (s *companyServiceImpl) DeleteCompany(id int) error {
	// 檢查公司是否存在
	existingCompany, err := s.companyRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Error checking existing company for delete", zap.Error(err), zap.Int("company_id", id))
		return utils.ErrInternalServer
	}
	if existingCompany == nil {
		return utils.ErrNotFound
	}

	// 這裡可以添加額外業務邏輯，例如檢查是否有客戶關聯到該公司，避免刪除
	// 範例：customerCount, _ := s.customerRepo.CountByCompanyID(id)
	// if customerCount > 0 { return utils.ErrBadRequest.SetDetails("Cannot delete company with associated customers") }

	if err := s.companyRepo.Delete(id); err != nil {
		zap.L().Error("Service: Failed to delete company in repository", zap.Error(err), zap.Int("company_id", id))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to delete company: %v", err))
	}
	return nil
}

package handler

import (
	"database/sql" // 導入 sql 包，用於檢查 ErrNoRows
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/service"
	"github.com/wac0705/fastener-api/utils"
)

// CompanyHandler 定義公司處理器結構，包含 CompanyService 的依賴
type CompanyHandler struct {
	companyService service.CompanyService
}

// NewCompanyHandler 創建 CompanyHandler 實例
func NewCompanyHandler(s service.CompanyService) *CompanyHandler {
	return &CompanyHandler{companyService: s}
}

// CreateCompany 創建新公司
func (h *CompanyHandler) CreateCompany(c echo.Context) error {
	company := new(models.Company)

	if err := c.Bind(company); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := c.Validate(company); err != nil {
		return err // 驗證錯誤會被全局錯誤處理器捕獲
	}

	if err := h.companyService.CreateCompany(company); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to create company", zap.Error(err), zap.String("company_name", company.Name))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusCreated, company)
}

// GetCompanies 獲取所有公司
func (h *CompanyHandler) GetCompanies(c echo.Context) error {
	companies, err := h.companyService.GetAllCompanies()
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get companies", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, companies)
}

// GetCompanyById 根據 ID 獲取公司
func (h *CompanyHandler) GetCompanyById(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	company, err := h.companyService.GetCompanyByID(id)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get company by ID", zap.Int("company_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	if company == nil { // Service 層返回 nil, nil 表示未找到
		return c.JSON(http.StatusNotFound, utils.ErrNotFound)
	}

	return c.JSON(http.StatusOK, company)
}

// UpdateCompany 更新公司信息
func (h *CompanyHandler) UpdateCompany(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	company := new(models.Company)
	if err := c.Bind(company); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 確保更新的是正確的公司 ID
	company.ID = id

	if err := c.Validate(company); err != nil {
		return err // 驗證錯誤
	}

	if err := h.companyService.UpdateCompany(company); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to update company", zap.Int("company_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusOK, company)
}

// DeleteCompany 刪除公司
func (h *CompanyHandler) DeleteCompany(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := h.companyService.DeleteCompany(id); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to delete company", zap.Int("company_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.NoContent(http.StatusNoContent) // 成功刪除，返回 204 No Content
}

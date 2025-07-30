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

// CustomerHandler 定義客戶處理器結構，包含 CustomerService 的依賴
type CustomerHandler struct {
	customerService service.CustomerService
}

// NewCustomerHandler 創建 CustomerHandler 實例
func NewCustomerHandler(s service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: s}
}

// CreateCustomer 創建新客戶
func (h *CustomerHandler) CreateCustomer(c echo.Context) error {
	customer := new(models.Customer)

	if err := c.Bind(customer); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := c.Validate(customer); err != nil {
		return err // 驗證錯誤
	}

	if err := h.customerService.CreateCustomer(customer); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to create customer", zap.Error(err), zap.String("customer_name", customer.Name))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusCreated, customer)
}

// GetCustomers 獲取所有客戶
func (h *CustomerHandler) GetCustomers(c echo.Context) error {
	customers, err := h.customerService.GetAllCustomers()
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get customers", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, customers)
}

// GetCustomerById 根據 ID 獲取客戶
func (h *CustomerHandler) GetCustomerById(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	customer, err := h.customerService.GetCustomerByID(id)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get customer by ID", zap.Int("customer_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	if customer == nil { // Service 層返回 nil, nil 表示未找到
		return c.JSON(http.StatusNotFound, utils.ErrNotFound)
	}

	return c.JSON(http.StatusOK, customer)
}

// UpdateCustomer 更新客戶信息
func (h *CustomerHandler) UpdateCustomer(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	customer := new(models.Customer)
	if err := c.Bind(customer); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 確保更新的是正確的客戶 ID
	customer.ID = id

	if err := c.Validate(customer); err != nil {
		return err // 驗證錯誤
	}

	if err := h.customerService.UpdateCustomer(customer); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to update customer", zap.Int("customer_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusOK, customer)
}

// DeleteCustomer 刪除客戶
func (h *CustomerHandler) DeleteCustomer(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := h.customerService.DeleteCustomer(id); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to delete customer", zap.Int("customer_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.NoContent(http.StatusNoContent) // 成功刪除，返回 204 No Content
}

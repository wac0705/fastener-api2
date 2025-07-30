package handler

import (
	"database/sql" // 導入 sql 包，用於檢查 ErrNoRows
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com->wac0705/fastener-api/service"
	"github.com/wac0705/fastener-api/utils"
)

// ProductDefinitionHandler 定義產品定義處理器結構，包含 ProductDefinitionService 的依賴
type ProductDefinitionHandler struct {
	productDefinitionService service.ProductDefinitionService
}

// NewProductDefinitionHandler 創建 ProductDefinitionHandler 實例
func NewProductDefinitionHandler(s service.ProductDefinitionService) *ProductDefinitionHandler {
	return &ProductDefinitionHandler{productDefinitionService: s}
}

// CreateProductCategory 創建新產品類別
func (h *ProductDefinitionHandler) CreateProductCategory(c echo.Context) error {
	category := new(models.ProductCategory)

	if err := c.Bind(category); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := c.Validate(category); err != nil {
		return err // 驗證錯誤
	}

	if err := h.productDefinitionService.CreateProductCategory(category); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to create product category", zap.Error(err), zap.String("category_name", category.Name))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusCreated, category)
}

// GetProductCategories 獲取所有產品類別
func (h *ProductDefinitionHandler) GetProductCategories(c echo.Context) error {
	categories, err := h.productDefinitionService.GetAllProductCategories()
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get product categories", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, categories)
}

// GetProductCategoryById 根據 ID 獲取產品類別
func (h *ProductDefinitionHandler) GetProductCategoryById(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	category, err := h.productDefinitionService.GetProductCategoryByID(id)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get product category by ID", zap.Int("category_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	if category == nil { // Service 層返回 nil, nil 表示未找到
		return c.JSON(http.StatusNotFound, utils.ErrNotFound)
	}

	return c.JSON(http.StatusOK, category)
}

// UpdateProductCategory 更新產品類別信息
func (h *ProductDefinitionHandler) UpdateProductCategory(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	category := new(models.ProductCategory)
	if err := c.Bind(category); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 確保更新的是正確的類別 ID
	category.ID = id

	if err := c.Validate(category); err != nil {
		return err // 驗證錯誤
	}

	if err := h.productDefinitionService.UpdateProductCategory(category); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to update product category", zap.Int("category_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusOK, category)
}

// DeleteProductCategory 刪除產品類別
func (h *ProductDefinitionHandler) DeleteProductCategory(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := h.productDefinitionService.DeleteProductCategory(id); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to delete product category", zap.Int("category_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.NoContent(http.StatusNoContent) // 成功刪除，返回 204 No Content
}

// CreateProductDefinition 創建新產品定義
func (h *ProductDefinitionHandler) CreateProductDefinition(c echo.Context) error {
	definition := new(models.ProductDefinition)

	if err := c.Bind(definition); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := c.Validate(definition); err != nil {
		return err // 驗證錯誤
	}

	if err := h.productDefinitionService.CreateProductDefinition(definition); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to create product definition", zap.Error(err), zap.String("definition_name", definition.Name))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusCreated, definition)
}

// GetProductDefinitions 獲取所有產品定義
func (h *ProductDefinitionHandler) GetProductDefinitions(c echo.Context) error {
	definitions, err := h.productDefinitionService.GetAllProductDefinitions()
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get product definitions", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	return c.JSON(http.StatusOK, definitions)
}

// GetProductDefinitionById 根據 ID 獲取產品定義
func (h *ProductDefinitionHandler) GetProductDefinitionById(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	definition, err := h.productDefinitionService.GetProductDefinitionByID(id)
	if err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to get product definition by ID", zap.Int("definition_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}
	if definition == nil { // Service 層返回 nil, nil 表示未找到
		return c.JSON(http.StatusNotFound, utils.ErrNotFound)
	}

	return c.JSON(http.StatusOK, definition)
}

// UpdateProductDefinition 更新產品定義信息
func (h *ProductDefinitionHandler) UpdateProductDefinition(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	definition := new(models.ProductDefinition)
	if err := c.Bind(definition); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	// 確保更新的是正確的定義 ID
	definition.ID = id

	if err := c.Validate(definition); err != nil {
		return err // 驗證錯誤
	}

	if err := h.productDefinitionService.UpdateProductDefinition(definition); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to update product definition", zap.Int("definition_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.JSON(http.StatusOK, definition)
}

// DeleteProductDefinition 刪除產品定義
func (h *ProductDefinitionHandler) DeleteProductDefinition(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id")) // 從 URL 參數獲取 ID
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrBadRequest)
	}

	if err := h.productDefinitionService.DeleteProductDefinition(id); err != nil {
		if customErr, ok := err.(*utils.CustomError); ok {
			return c.JSON(customErr.Code, customErr)
		}
		zap.L().Error("Failed to delete product definition", zap.Int("definition_id", id), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	return c.NoContent(http.StatusNoContent) // 成功刪除，返回 204 No Content
}

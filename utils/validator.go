package utils

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator 結構體，包裝 go-playground/validator 實例
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator 創建一個新的 CustomValidator 實例
func NewCustomValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

// Validate 實現 Echo 的 Validator 介面
// 當 Echo 接收到請求並嘗試綁定數據到結構體時，如果該結構體定義了 `validate` 標籤，
// Echo 會自動調用這個 Validate 方法。
func (cv *CustomValidator) Validate(i interface{}) error {
	// 使用 validator 庫對結構體進行驗證
	if err := cv.validator.Struct(i); err != nil {
		// 返回原始的驗證錯誤，Echo 的 HTTPErrorHandler 將會處理它
		return err
	}
	return nil
}

// 你可以在這裡添加自定義的驗證規則，例如：
/*
func (cv *CustomValidator) RegisterCustomValidations() {
    // 註冊一個自定義的日期格式驗證器
    cv.validator.RegisterValidation("date_format", func(fl validator.FieldLevel) bool {
        dateStr := fl.Field().String()
        _, err := time.Parse("2006-01-02", dateStr) // 例如，驗證 "YYYY-MM-DD" 格式
        return err == nil
    })
}
*/

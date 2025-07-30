package utils

import (
	"fmt"
	"net/http"
)

// CustomError 自定義錯誤結構，用於統一 API 響應格式
type CustomError struct {
	Code    int         `json:"code"`    // HTTP 狀態碼
	Message string      `json:"message"` // 錯誤訊息
	Details interface{} `json:"details,omitempty"` // 錯誤細節 (例如驗證錯誤列表、原始錯誤等)
}

// Error 實現 error 介面，讓 CustomError 可以作為 Go 的錯誤類型使用
func (e *CustomError) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("Error %d: %s (Details: %v)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

// SetDetails 設置錯誤的詳細信息，並返回 CustomError 實例
func (e *CustomError) SetDetails(details interface{}) *CustomError {
	e.Details = details
	return e
}

// 常用錯誤實例
// 這些都是預定義的錯誤，可以在應用程式的任何地方直接使用
var (
	ErrBadRequest     = &CustomError{Code: http.StatusBadRequest, Message: "Bad Request"}
	ErrUnauthorized   = &CustomError{Code: http.StatusUnauthorized, Message: "Unauthorized"}
	ErrForbidden      = &CustomError{Code: http.StatusForbidden, Message: "Forbidden"}
	ErrNotFound       = &CustomError{Code: http.StatusNotFound, Message: "Resource not found"}
	ErrInternalServer = &CustomError{Code: http.StatusInternalServerError, Message: "Internal server error"}
)

// NewValidationError 創建一個特定用於驗證失敗的錯誤實例
func NewValidationError(details interface{}) *CustomError {
	return &CustomError{Code: http.StatusBadRequest, Message: "Validation failed", Details: details}
}

// NewCustomError 創建一個新的 CustomError 實例
func NewCustomError(code int, message string, details interface{}) *CustomError {
	return &CustomError{Code: code, Message: message, Details: details}
}

package models

import "time"

// Customer 客戶模型
type Customer struct {
	ID           int       `json:"id"`
	Name         string    `json:"name" validate:"required,min=2,max=255"`
	ContactPerson string    `json:"contact_person"`
	Email        string    `json:"email" validate:"omitempty,email"` // omitempty 表示可選，email 驗證格式
	Phone        string    `json:"phone" validate:"omitempty,min=7,max=20"`
	CompanyID    *int      `json:"company_id,omitempty"` // 指針類型允許為 NULL
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

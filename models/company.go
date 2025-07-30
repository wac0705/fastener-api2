package models

import "time"

// Company 公司模型
type Company struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" validate:"required,min=2,max=255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

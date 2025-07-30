package models

import "time"

// ProductCategory 產品類別模型
type ProductCategory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" validate:"required,min=2,max=255"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductDefinition 產品定義模型
type ProductDefinition struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" validate:"required,min=2,max=255"`
	Description string    `json:"description,omitempty"`
	CategoryID  int       `json:"category_id" validate:"required,min=1"`
	Unit        string    `json:"unit,omitempty"`
	Price       float64   `json:"price" validate:"required,min=0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

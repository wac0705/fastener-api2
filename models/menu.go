package models

import "time"

// Menu 選單模型
type Menu struct {
	ID           int       `json:"id"`
	Name         string    `json:"name" validate:"required,min=2,max=100"`
	Path         string    `json:"path" validate:"required,min=1,max=255"` // 前端路由路徑
	Icon         string    `json:"icon,omitempty"`                         // 選單圖標
	ParentID     *int      `json:"parent_id,omitempty"`                    // 父選單 ID，允許為 NULL
	DisplayOrder int       `json:"display_order"`                          // 顯示順序
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

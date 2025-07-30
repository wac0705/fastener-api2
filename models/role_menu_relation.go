package models

import "time"

// RoleMenu 角色與選單的關聯模型
type RoleMenu struct {
	RoleID    int       `json:"role_id" validate:"required,min=1"`
	MenuID    int       `json:"menu_id" validate:"required,min=1"`
	CreatedAt time.Time `json:"created_at"` // 在關聯創建時自動設置
	UpdatedAt time.Time `json:"updated_at"` // 在關聯更新時自動設置 (如果需要)
}

// 這個模型可能用於返回給前端，包含更多詳細資訊
type RoleMenuDetail struct {
	RoleID   int    `json:"role_id"`
	RoleName string `json:"role_name"`
	MenuID   int    `json:"menu_id"`
	MenuName string `json:"menu_name"`
	MenuPath string `json:"menu_path"`
}

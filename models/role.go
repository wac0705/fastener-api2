package models

import "time"

// Role 角色模型
type Role struct {
	ID        int       `json:"id"`
	Name      string    `json:"name" validate:"required,min=2,max=50,alphanum"` // 例如: "admin", "finance", "user"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Permission 權限模型
type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name" validate:"required,min=3,max=100,alphanum"` // 例如: "company:read", "account:create"
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RolePermission 角色與權限的關聯模型 (用於多對多關係)
type RolePermission struct {
	RoleID      int `json:"role_id" validate:"required,min=1"`
	PermissionID int `json:"permission_id" validate:"required,min=1"`
}

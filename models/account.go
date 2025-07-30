package models

import "time"

// Account 帳戶模型，用於應用程式用戶
type Account struct {
	ID        int       `json:"id"`
	Username  string    `json:"username" validate:"required,min=3,max=50"`
	Password  string    `json:"password,omitempty" validate:"required,min=6"` // `omitempty` 在 JSON 序列化時忽略空值
	RoleID    int       `json:"role_id"`
	RoleName  string    `json:"role_at_read,omitempty"` // 角色名稱，通常在讀取時通過 JOIN 填充
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest 用於登入請求的結構
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest 用於註冊請求的結構
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	RoleID   int    `json:"role_id" validate:"required,min=1"` // 註冊時必須指定角色
}

// UpdatePasswordRequest 用於更新密碼請求
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// RefreshTokenRequest 用於刷新 Token 請求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

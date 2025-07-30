package service

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/middleware/jwt" // 導入 JWT 相關函式
	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository" // 導入 Repository 層
	"github.com/wac0705/fastener-api/utils"      // 導入工具 (包含自定義錯誤)
)

// AuthService 定義身份驗證服務介面
type AuthService interface {
	Login(username, password string) (accessToken, refreshToken string, account *models.Account, err error)
	Register(username, password string, roleID int) (*models.Account, error)
	RefreshToken(refreshToken string) (newAccessToken string, err error)
    GetAccountByID(accountID int) (*models.Account, error) // 用於獲取我的資料
}

// authServiceImpl 實現 AuthService 介面
type authServiceImpl struct {
	accountRepo        repository.AccountRepository
	roleRepo           repository.RoleRepository
	jwtSecret          string
	jwtAccessExpires   int
	jwtRefreshExpires  int
}

// NewAuthService 創建 AuthService 實例
func NewAuthService(
	accountRepo repository.AccountRepository,
	roleRepo repository.RoleRepository,
	jwtSecret string,
	jwtAccessExpires, jwtRefreshExpires int,
) AuthService {
	return &authServiceImpl{
		accountRepo:       accountRepo,
		roleRepo:          roleRepo,
		jwtSecret:         jwtSecret,
		jwtAccessExpires:  jwtAccessExpires,
		jwtRefreshExpires: jwtRefreshExpires,
	}
}

// Login 處理用戶登入邏輯
func (s *authServiceImpl) Login(username, password string) (string, string, *models.Account, error) {
	account, err := s.accountRepo.FindByUsername(username)
	if err != nil {
		zap.L().Error("AuthService: Error finding account by username during login", zap.Error(err), zap.String("username", username))
		return "", "", nil, utils.ErrInternalServer
	}
	if account == nil {
		return "", "", nil, utils.ErrUnauthorized.SetDetails("Invalid credentials") // 用戶不存在或密碼錯誤都返回通用錯誤
	}

	// 驗證密碼
	if !utils.CheckPasswordHash(password, account.Password) {
		return "", "", nil, utils.ErrUnauthorized.SetDetails("Invalid credentials")
	}

	// 獲取角色名稱 (用於返回給前端顯示)
	role, err := s.roleRepo.FindByID(account.RoleID)
	if err != nil {
		zap.L().Error("AuthService: Error finding role for account", zap.Error(err), zap.Int("account_id", account.ID))
		return "", "", nil, utils.ErrInternalServer
	}
	if role == nil {
		// 這種情況不應該發生，表示數據不一致
		zap.L().Error("AuthService: Role not found for account", zap.Int("account_id", account.ID), zap.Int("role_id", account.RoleID))
		return "", "", nil, utils.ErrInternalServer.SetDetails("Account role not configured correctly")
	}
	account.RoleName = role.Name

	// 生成 Access Token 和 Refresh Token
	accessToken, refreshToken, err := jwt.GenerateAuthTokens(*account, s.jwtSecret, s.jwtAccessExpires, s.jwtRefreshExpires)
	if err != nil {
		zap.L().Error("AuthService: Failed to generate tokens during login", zap.Error(err), zap.Int("account_id", account.ID))
		return "", "", nil, utils.ErrInternalServer
	}

	return accessToken, refreshToken, account, nil
}

// Register 處理用戶註冊邏輯
func (s *authServiceImpl) Register(username, password string, roleID int) (*models.Account, error) {
	// 檢查用戶名是否已存在
	existingAccount, err := s.accountRepo.FindByUsername(username)
	if err != nil {
		zap.L().Error("AuthService: Error checking existing account by username during registration", zap.Error(err), zap.String("username", username))
		return nil, utils.ErrInternalServer
	}
	if existingAccount != nil {
		return nil, utils.ErrBadRequest.SetDetails("Username already exists")
	}

	// 檢查角色 ID 是否有效
	role, err := s.roleRepo.FindByID(roleID)
	if err != nil {
		zap.L().Error("AuthService: Error checking role ID during registration", zap.Error(err), zap.Int("role_id", roleID))
		return nil, utils.ErrInternalServer
	}
	if role == nil {
		return nil, utils.ErrBadRequest.SetDetails("Invalid Role ID")
	}

	// 雜湊密碼
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		zap.L().Error("AuthService: Failed to hash password during registration", zap.Error(err))
		return nil, utils.ErrInternalServer
	}

	// 創建帳戶模型
	newAccount := &models.Account{
		Username: username,
		Password: hashedPassword,
		RoleID:   roleID,
	}

	// 調用 Repository 創建帳戶
	if err := s.accountRepo.Create(newAccount); err != nil {
		zap.L().Error("AuthService: Failed to create account in repository during registration", zap.Error(err), zap.String("username", username))
		return nil, utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to register account: %v", err))
	}
	newAccount.RoleName = role.Name // 填充角色名稱
	return newAccount, nil
}

// RefreshToken 處理 Refresh Token 刷新 Access Token 的邏輯
func (s *authServiceImpl) RefreshToken(refreshToken string) (string, error) {
	// 驗證 Refresh Token
	claims, err := jwt.VerifyRefreshToken(refreshToken, s.jwtSecret)
	if err != nil {
		// VerifyRefreshToken 已在內部記錄錯誤
		return "", utils.ErrUnauthorized.SetDetails("Invalid or expired refresh token")
	}

	// 查找對應的帳戶
	account, err := s.accountRepo.FindByID(claims.AccountID)
	if err != nil {
		zap.L().Error("AuthService: Error finding account for refresh token", zap.Error(err), zap.Int("account_id", claims.AccountID))
		return "", utils.ErrInternalServer
	}
	if account == nil {
		zap.L().Info("AuthService: Account not found for refresh token", zap.Int("account_id", claims.AccountID))
		return "", utils.ErrUnauthorized.SetDetails("Invalid refresh token: Account not found")
	}

	// 這裡可以選擇性地實現 Refresh Token 的黑名單機制，
	// 確保 Refresh Token 只能使用一次或在特定情況下失效
	// ... (例如，在資料庫或 Redis 中標記 Refresh Token 為已使用)

	// 生成新的 Access Token
	newAccessToken, _, err := jwt.GenerateAuthTokens(*account, s.jwtSecret, s.jwtAccessExpires, s.jwtRefreshExpires) // 只返回 Access Token
	if err != nil {
		zap.L().Error("AuthService: Failed to generate new access token during refresh", zap.Error(err), zap.Int("account_id", account.ID))
		return "", utils.ErrInternalServer
	}

	return newAccessToken, nil
}

// GetAccountByID 獲取帳戶資料，用於我的資料
func (s *authServiceImpl) GetAccountByID(accountID int) (*models.Account, error) {
    account, err := s.accountRepo.FindByID(accountID)
    if err != nil {
        zap.L().Error("AuthService: Failed to get account by ID", zap.Int("account_id", accountID), zap.Error(err))
        return nil, utils.ErrInternalServer
    }
    if account == nil {
        return nil, nil // 未找到
    }
    account.Password = "" // 清除密碼

    return account, nil
}

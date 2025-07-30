package service

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/models"
	"github.com/wac0705/fastener-api/repository" // 導入 Repository 層
	"github.com/wac0705/fastener-api/utils"      // 導入工具 (包含自定義錯誤)
)

// AccountService 定義帳戶服務介面
type AccountService interface {
	CreateAccount(account *models.Account) error
	GetAllAccounts() ([]models.Account, error)
	GetAccountByID(id int) (*models.Account, error)
	UpdateAccount(account *models.Account) error
	DeleteAccount(id int) error
	UpdatePassword(accountID int, oldPassword, newPassword string, requesterAccountID int, requesterRoleID int) error
}

// accountServiceImpl 實現 AccountService 介面
type accountServiceImpl struct {
	accountRepo repository.AccountRepository
	roleRepo    repository.RoleRepository // 依賴 RoleRepository 以獲取角色信息
}

// NewAccountService 創建 AccountService 實例
func NewAccountService(accountRepo repository.AccountRepository, roleRepo repository.RoleRepository) AccountService {
	return &accountServiceImpl{accountRepo: accountRepo, roleRepo: roleRepo}
}

// CreateAccount 創建新帳戶
func (s *accountServiceImpl) CreateAccount(account *models.Account) error {
	// 檢查用戶名是否已存在
	existingAccount, err := s.accountRepo.FindByUsername(account.Username)
	if err != nil {
		zap.L().Error("Service: Error checking existing account by username", zap.Error(err), zap.String("username", account.Username))
		return utils.ErrInternalServer
	}
	if existingAccount != nil {
		return utils.ErrBadRequest.SetDetails("Username already exists")
	}

	// 檢查角色 ID 是否有效
	role, err := s.roleRepo.FindByID(account.RoleID)
	if err != nil {
		zap.L().Error("Service: Error checking role ID", zap.Error(err), zap.Int("role_id", account.RoleID))
		return utils.ErrInternalServer
	}
	if role == nil {
		return utils.ErrBadRequest.SetDetails("Invalid Role ID")
	}

	// 雜湊密碼
	hashedPassword, err := utils.HashPassword(account.Password)
	if err != nil {
		zap.L().Error("Service: Failed to hash password for new account", zap.Error(err))
		return utils.ErrInternalServer
	}
	account.Password = hashedPassword

	// 調用 Repository 創建帳戶
	if err := s.accountRepo.Create(account); err != nil {
		// Repository 可能已經處理了一些重複鍵錯誤，但這裡可以再次確保
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to create account: %v", err))
	}
	return nil
}

// GetAllAccounts 獲取所有帳戶
func (s *accountServiceImpl) GetAllAccounts() ([]models.Account, error) {
	accounts, err := s.accountRepo.FindAll()
	if err != nil {
		zap.L().Error("Service: Failed to get all accounts", zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	// 在返回之前清除敏感資訊
	for i := range accounts {
		accounts[i].Password = ""
	}
	return accounts, nil
}

// GetAccountByID 根據 ID 獲取帳戶
func (s *accountServiceImpl) GetAccountByID(id int) (*models.Account, error) {
	account, err := s.accountRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Failed to get account by ID", zap.Int("id", id), zap.Error(err))
		return nil, utils.ErrInternalServer
	}
	if account == nil {
		return nil, nil // Repository 返回 nil, nil 表示未找到
	}
	account.Password = "" // 清除敏感資訊
	return account, nil
}

// UpdateAccount 更新帳戶信息
func (s *accountServiceImpl) UpdateAccount(account *models.Account) error {
	// 檢查帳戶是否存在
	existingAccount, err := s.accountRepo.FindByID(account.ID)
	if err != nil {
		zap.L().Error("Service: Error checking existing account for update", zap.Error(err), zap.Int("account_id", account.ID))
		return utils.ErrInternalServer
	}
	if existingAccount == nil {
		return utils.ErrNotFound
	}

	// 檢查新的用戶名是否被其他帳戶占用 (如果用戶名有更改)
	if existingAccount.Username != account.Username {
		otherAccount, err := s.accountRepo.FindByUsername(account.Username)
		if err != nil {
			zap.L().Error("Service: Error checking username for update conflict", zap.Error(err), zap.String("new_username", account.Username))
			return utils.ErrInternalServer
		}
		if otherAccount != nil && otherAccount.ID != account.ID {
			return utils.ErrBadRequest.SetDetails("Username already taken by another account")
		}
	}

	// 檢查新的角色 ID 是否有效
	role, err := s.roleRepo.FindByID(account.RoleID)
	if err != nil {
		zap.L().Error("Service: Error checking role ID for update", zap.Error(err), zap.Int("role_id", account.RoleID))
		return utils.ErrInternalServer
	}
	if role == nil {
		return utils.ErrBadRequest.SetDetails("Invalid Role ID")
	}

	// 調用 Repository 更新帳戶
	if err := s.accountRepo.Update(account); err != nil {
		zap.L().Error("Service: Failed to update account in repository", zap.Error(err), zap.Int("account_id", account.ID))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to update account: %v", err))
	}
	return nil
}

// DeleteAccount 刪除帳戶
func (s *accountServiceImpl) DeleteAccount(id int) error {
	// 檢查帳戶是否存在
	existingAccount, err := s.accountRepo.FindByID(id)
	if err != nil {
		zap.L().Error("Service: Error checking existing account for delete", zap.Error(err), zap.Int("account_id", id))
		return utils.ErrInternalServer
	}
	if existingAccount == nil {
		return utils.ErrNotFound
	}

	// 可以添加業務邏輯，例如不允許刪除管理員帳戶
	// if existingAccount.RoleID == adminRoleID { ... }

	if err := s.accountRepo.Delete(id); err != nil {
		zap.L().Error("Service: Failed to delete account in repository", zap.Error(err), zap.Int("account_id", id))
		return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to delete account: %v", err))
	}
	return nil
}

// UpdatePassword 更新帳戶密碼
// requesterAccountID 是發起密碼修改的用戶ID，用於權限判斷（是否是自己或有權限的管理員）
func (s *accountServiceImpl) UpdatePassword(accountID int, oldPassword, newPassword string, requesterAccountID int, requesterRoleID int) error {
    // 獲取目標帳戶信息
    targetAccount, err := s.accountRepo.FindByID(accountID)
    if err != nil {
        zap.L().Error("Service: Error getting target account for password update", zap.Error(err), zap.Int("account_id", accountID))
        return utils.ErrInternalServer
    }
    if targetAccount == nil {
        return utils.ErrNotFound
    }

    // 檢查請求者是否有權修改密碼：
    // 1. 如果是修改自己的密碼
    // 2. 如果請求者是管理員 (假設 RoleID=1 是 admin) 且有權限修改他人密碼
    isAdminRoleID, err := s.roleRepo.FindByName("admin")
    if err != nil {
        zap.L().Error("Service: Failed to get admin role ID", zap.Error(err))
        return utils.ErrInternalServer
    }
    if isAdminRoleID == nil {
        zap.L().Error("Service: Admin role not found in database, check initial setup.")
        return utils.ErrInternalServer.SetDetails("Admin role not configured.")
    }

    if requesterAccountID != accountID && requesterRoleID != isAdminRoleID.ID {
        return utils.ErrForbidden.SetDetails("You do not have permission to change this account's password.")
    }

    // 如果是修改自己的密碼，需要驗證舊密碼
    if requesterAccountID == accountID {
        currentAccount, err := s.accountRepo.FindByID(accountID)
        if err != nil {
            zap.L().Error("Service: Error retrieving current account for password verification", zap.Error(err), zap.Int("account_id", accountID))
            return utils.ErrInternalServer
        }
        if currentAccount == nil { // 應當不會發生，因為前面已經檢查過 targetAccount
            return utils.ErrNotFound
        }
        if !utils.CheckPasswordHash(oldPassword, currentAccount.Password) {
            return utils.ErrUnauthorized.SetDetails("Old password is incorrect")
        }
    } else {
        // 如果是管理員修改他人的密碼，不需要舊密碼，但要確保 newPassword 不為空
        if newPassword == "" {
             return utils.ErrBadRequest.SetDetails("New password cannot be empty for admin password reset.")
        }
    }

    // 雜湊新密碼
    hashedNewPassword, err := utils.HashPassword(newPassword)
    if err != nil {
        zap.L().Error("Service: Failed to hash new password", zap.Error(err))
        return utils.ErrInternalServer
    }

    if err := s.accountRepo.UpdatePassword(accountID, hashedNewPassword); err != nil {
        if err == utils.ErrNotFound { // Repository 返回的未找到錯誤
            return utils.ErrNotFound // 帳戶可能被刪除
        }
        zap.L().Error("Service: Failed to update password in repository", zap.Error(err), zap.Int("account_id", accountID))
        return utils.ErrInternalServer.SetDetails(fmt.Sprintf("Failed to update password: %v", err))
    }

    return nil
}

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/wac0705/fastener-api/config" // 導入配置模組
	"github.com/wac0705/fastener-api/db"     // 導入資料庫模組
	"github.com/wac0705/fastener-api/repository" // 導入 Repository 層
	"github.com/wac0705/fastener-api/utils"  // 導入工具模組
)

func main() {
	// 載入應用程式配置
	config.LoadConfig()

	// 初始化資料庫連接
	db.InitDB(config.Cfg.DatabaseURL)
	defer func() {
		sqlDB, err := db.DB.DB()
		if err != nil {
			log.Printf("Error getting underlying SQL DB for resetadmin: %v\n", err)
		} else if sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing database for resetadmin: %v\n", err)
			}
		}
	}()

	// 從配置中獲取管理員帳戶和新密碼
	adminUsername := config.Cfg.AdminUsername
	adminPassword := config.Cfg.AdminPassword

	if adminUsername == "" || adminPassword == "" {
		log.Fatal("ADMIN_USERNAME and ADMIN_PASSWORD environment variables must be set in .env or environment for resetadmin.")
	}

	// 創建 Account Repository 實例
	accountRepo := repository.NewAccountRepository(db.DB)

	// 雜湊新密碼
	hashedPassword, err := utils.HashPassword(adminPassword)
	if err != nil {
		log.Fatalf("Error hashing password: %v", err)
	}

	// 更新資料庫中的管理員密碼
	// 假設有一個方法可以直接更新指定用戶名的密碼，且只針對 'admin' 角色
	err = accountRepo.UpdateAdminPassword(adminUsername, hashedPassword)
	if err != nil {
		log.Fatalf("Error updating admin password for '%s': %v", adminUsername, err)
	}

	fmt.Printf("Admin account '%s' password reset successfully.\n", adminUsername)
}

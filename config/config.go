package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// AppConfig 應用程式的配置結構
type AppConfig struct {
	Port                string
	DatabaseURL         string
	JwtSecret           string
	JwtAccessExpiresHours  int
	JwtRefreshExpiresHours int
	CorsAllowOrigin     string
	AdminUsername       string
	AdminPassword       string
	AppEnv              string
	LogLevel            string
}

var Cfg *AppConfig // 全局配置實例

// LoadConfig 載入應用程式配置
func LoadConfig() {
	// 載入 .env 檔案，生產環境可能沒有，所以錯誤不Fatal
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, assuming environment variables are set or using default.")
	}

	// 從環境變數讀取配置，並提供預設值
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required.")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required.")
	}

	jwtAccessExpiresHoursStr := os.Getenv("JWT_ACCESS_EXPIRES_HOURS")
	jwtAccessExpiresHours, err := strconv.Atoi(jwtAccessExpiresHoursStr)
	if err != nil || jwtAccessExpiresHours == 0 {
		jwtAccessExpiresHours = 1 // 預設 Access Token 有效期為 1 小時
		log.Printf("JWT_ACCESS_EXPIRES_HOURS not set or invalid, using default %d hours.\n", jwtAccessExpiresHours)
	}

	jwtRefreshExpiresHoursStr := os.Getenv("JWT_REFRESH_EXPIRES_HOURS")
	jwtRefreshExpiresHours, err := strconv.Atoi(jwtRefreshExpiresHoursStr)
	if err != nil || jwtRefreshExpiresHours == 0 {
		jwtRefreshExpiresHours = 720 // 預設 Refresh Token 有效期為 720 小時 (30 天)
		log.Printf("JWT_REFRESH_EXPIRES_HOURS not set or invalid, using default %d hours.\n", jwtRefreshExpiresHours)
	}

	corsAllowOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if corsAllowOrigin == "" {
		corsAllowOrigin = "*" // 預設允許所有來源 (開發環境可接受，生產環境應限制)
		log.Println("CORS_ALLOW_ORIGIN not set, defaulting to '*'.")
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD") // 注意：此密碼僅用於初始化或重設工具，不應長期存在

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	Cfg = &AppConfig{
		Port:                port,
		DatabaseURL:         dbURL,
		JwtSecret:           jwtSecret,
		JwtAccessExpiresHours:  jwtAccessExpiresHours,
		JwtRefreshExpiresHours: jwtRefreshExpiresHours,
		CorsAllowOrigin:     corsAllowOrigin,
		AdminUsername:       adminUsername,
		AdminPassword:       adminPassword,
		AppEnv:              appEnv,
		LogLevel:            logLevel,
	}

	// 敏感資訊的警告 (僅在開發環境輸出)
	if Cfg.AppEnv == "development" {
		log.Println("--- WARNING: Using .env file for sensitive configurations. ---")
		log.Println("--- For production, use secure secrets management (e.g., Kubernetes Secrets, Vault, AWS Secrets Manager). ---")
	}
}

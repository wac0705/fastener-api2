package main

import (
	"errors" // 用於錯誤類型斷言
	"fmt"
	"net/http"
	"os"
	"time" // 用於 CORS MaxAge

	"github.com/go-playground/validator/v10" // 驗證器
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"           // 結構化日誌庫
	"go.uber.org/zap/zapcore"    // zap 的核心組件

	"github.com/wac0705/fastener-api/config"        // 應用程式配置
	"github.com/wac0705/fastener-api/db"            // 資料庫初始化
	"github.com/wac0705/fastener-api/handler"       // 處理器
	"github.com/wac0705/fastener-api/middleware/authz" // 授權中介軟體
	"github.com/wac0705/fastener-api/middleware/jwt" // JWT 中介軟體
	"github.com/wac0705/fastener-api/repository"    // Repository 層
	"github.com/wac0705/fastener-api/routes"        // 路由定義
	"github.com/wac0705/fastener-api/service"       // Service 層
	"github.com/wac0705/fastener-api/utils"         // 工具函式 (包含自定義錯誤)
)

var logger *zap.Logger // 全局日誌器

// init 函數會在 main 函數之前執行，用於初始化日誌器
func init() {
	var cfg zap.Config
	var err error

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development" // 預設為開發環境
	}

	if appEnv == "production" {
		cfg = zap.NewProductionConfig() // 生產環境：JSON 格式，更利於機器解析
	} else {
		cfg = zap.NewDevelopmentConfig() // 開發環境：彩色，更利於人類閱讀
	}

	// 設定日誌級別
	logLevelStr := os.Getenv("LOG_LEVEL")
	if logLevelStr == "" {
		logLevelStr = "info" // 預設日誌級別
	}
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevelStr)); err != nil {
		fmt.Printf("Invalid LOG_LEVEL '%s', defaulting to info: %v\n", logLevelStr, err)
		level = zapcore.InfoLevel
	}
	cfg.Level.SetLevel(level)

	logger, err = cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	// zap.ReplaceGlobals(logger) // 設定為全局 Zap logger，以便其他包直接使用 zap.L() 或 zap.S()
}

func main() {
	defer func() {
		// 確保所有緩衝日誌都被寫入。對於某些輸出（如 /dev/stderr），sync 可能會返回錯誤，需要忽略。
		if err := logger.Sync(); err != nil && err.Error() != "sync /dev/stderr: invalid argument" {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	// 載入應用程式配置
	config.LoadConfig()

	// 初始化資料庫
	db.InitDB(config.Cfg.DatabaseURL)
	defer func() {
		sqlDB, err := db.DB.DB()
		if err != nil {
			logger.Error("Error getting underlying SQL DB", zap.Error(err))
		} else if sqlDB != nil {
			if err := sqlDB.Close(); err != nil {
				logger.Error("Error closing database", zap.Error(err))
			}
		}
	}()

	e := echo.New() // 創建 Echo 實例

	// 設定自定義錯誤處理器
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		var he *echo.HTTPError
		if errors.As(err, &he) { // 如果是 Echo 內部錯誤
			// 如果內部錯誤是我們自定義的錯誤，則直接使用
			if he.Internal != nil {
				if customErr, ok := he.Internal.(*utils.CustomError); ok {
					c.JSON(customErr.Code, customErr)
					return
				}
			}
			// 否則，將 Echo HTTP 錯誤轉換為自定義錯誤格式
			c.JSON(he.Code, &utils.CustomError{Code: he.Code, Message: he.Message.(string)})
			return
		}

		// 如果錯誤是我們自定義的錯誤
		if customErr, ok := err.(*utils.CustomError); ok {
			c.JSON(customErr.Code, customErr)
			return
		}

		// 如果是驗證錯誤 (來自 go-playground/validator)
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			details := make(map[string]string)
			for _, fieldErr := range validationErrors {
				details[fieldErr.Field()] = fieldErr.Tag() // 簡化處理，實際應用中可轉換為更友好的訊息
			}
			customErr := utils.NewValidationError(details)
			c.JSON(customErr.Code, customErr)
			return
		}

		// 其他未處理的錯誤，記錄到日誌並返回通用的內部伺服器錯誤
		logger.Error("Unhandled internal server error", zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Request().Method),
			zap.Any("error_type", fmt.Sprintf("%T", err)), // 記錄錯誤類型
		)
		c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
	}

	// Echo 全局中介軟體
	e.Use(middleware.Recover()) // 錯誤恢復
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{ // CORS 設定
		AllowOrigins:     []string{config.Cfg.CorsAllowOrigin},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodPatch},
		AllowCredentials: true,
		MaxAge:           int(12 * time.Hour / time.Second), // CORS 預檢請求緩存時間
	}))

	// 設定 RequestLogger 以使用 zap
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogMethod:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("method", v.Method),
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
				zap.String("remote_ip", v.RemoteIP),
				// 可以在這裡加入更多上下文，例如如果已經經過 JWT 驗證，可以加入用戶 ID
			)
			return nil
		},
	}))

	// 設定 Echo 的日誌輸出到 Zap
	e.Logger.SetOutput(zap.NewStdLog(logger).Writer())
	e.Logger.SetLevel(echo.Lvl(config.Cfg.LogLevel)) // 設定 Echo 日誌級別

	// 將 JWT 驗證器實例綁定到 Echo 上下文 (用於處理器內部手動驗證，如果需要)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("jwtVerifier", jwt.NewJwtVerifier(config.Cfg.JwtSecret))
			return next(c)
		}
	})

	// 設置靜態檔案伺服 (如果需要，可創建 public 目錄)
	// e.Static("/", "public")

	// --- 依賴注入和服務啟動 ---
	// 實例化 Repository 層
	accountRepo := repository.NewAccountRepository(db.DB)
	companyRepo := repository.NewCompanyRepository(db.DB)
	customerRepo := repository.NewCustomerRepository(db.DB)
	menuRepo := repository.NewMenuRepository(db.DB)
	productDefinitionRepo := repository.NewProductDefinitionRepository(db.DB)
	roleRepo := repository.NewRoleRepository(db.DB)             // 新增 Role Repository
	roleMenuRepo := repository.NewRoleMenuRepository(db.DB)     // 新增 RoleMenu Repository
	permissionRepo := repository.NewPermissionRepository(db.DB) // 新增 Permission Repository

	// 實例化 Service 層，並注入 Repository 依賴
	accountService := service.NewAccountService(accountRepo, roleRepo) // AccountService 依賴 AccountRepo 和 RoleRepo
	authService := service.NewAuthService(accountRepo, roleRepo, config.Cfg.JwtSecret, config.Cfg.JwtAccessExpiresHours, config.Cfg.JwtRefreshExpiresHours) // AuthService 依賴 AccountRepo, RoleRepo, JWT配置
	companyService := service.NewCompanyService(companyRepo)
	customerService := service.NewCustomerService(customerRepo)
	menuService := service.NewMenuService(menuRepo)
	productDefinitionService := service.NewProductDefinitionService(productDefinitionRepo)
	roleService := service.NewRoleService(roleRepo)             // 新增 RoleService
	roleMenuService := service.NewRoleMenuService(roleMenuRepo) // 新增 RoleMenuService
	permissionService := service.NewPermissionService(permissionRepo, roleRepo) // 新增 PermissionService 依賴 PermissionRepo 和 RoleRepo

	// 實例化 Handler 層，並注入 Service 依賴
	accountHandler := handler.NewAccountHandler(accountService)
	authHandler := handler.NewAuthHandler(authService)
	companyHandler := handler.NewCompanyHandler(companyService)
	customerHandler := handler.NewCustomerHandler(customerService)
	menuHandler := handler.NewMenuHandler(menuService)
	productDefinitionHandler := handler.NewProductDefinitionHandler(productDefinitionService)
	roleMenuHandler := handler.NewRoleMenuHandler(roleMenuService)

	// --- API 路由定義 ---
	// 使用 routes 包來集中定義所有路由
	routes.RegisterAPIRoutes(e,
		authHandler,
		accountHandler,
		companyHandler,
		customerHandler,
		menuHandler,
		productDefinitionHandler,
		roleMenuHandler,
		permissionService, // 將權限服務傳入以便在路由中介軟體中使用
		config.Cfg.JwtSecret, // JWT Secret 也傳入
	)

	// 啟動伺服器
	port := config.Cfg.Port
	if port == "" {
		port = "8080" // 預設端口
	}
	logger.Fatal("Server failed to start", zap.Error(e.Start(":"+port))) // 使用 zap 記錄 Fatal 錯誤
}

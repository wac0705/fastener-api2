package routes

import (
	"net/http" // 導入 http 包，用於定義方法常數

	"github.com/labstack/echo/v4"

	"github.com/wac0705/fastener-api/handler"
	"github.com/wac0705/fastener-api/middleware/authz"
	"github.com/wac0705/fastener-api/middleware/jwt"
	"github.com/wac0705/fastener-api/service" // 導入 service 包以傳遞 PermissionService
)

// RegisterAPIRoutes 註冊所有 API 路由
func RegisterAPIRoutes(e *echo.Echo,
	authHandler *handler.AuthHandler,
	accountHandler *handler.AccountHandler,
	companyHandler *handler.CompanyHandler,
	customerHandler *handler.CustomerHandler,
	menuHandler *handler.MenuHandler,
	productDefinitionHandler *handler.ProductDefinitionHandler,
	roleMenuHandler *handler.RoleMenuHandler,
	permissionService service.PermissionService, // 注入權限服務
	jwtSecret string, // 注入 JWT Secret
) {
	apiGroup := e.Group("/api")

	// --- 公開路由 (無需身份驗證) ---
	apiGroup.POST("/login", authHandler.Login)
	apiGroup.POST("/register", authHandler.Register)
	apiGroup.POST("/refresh-token", authHandler.RefreshToken)

	// --- 受保護路由 (需要 JWT Access Token 驗證和細粒度授權) ---
	authGroup := apiGroup.Group("") // 創建一個新的分組，應用 JWT 中介軟體
	authGroup.Use(jwt.JwtAccessConfig(jwtSecret)) // 應用 JWT Access Token 驗證

	// 額外中介軟體：將 Access Token Claims 存入 Echo Context
	// 這樣後續的 authz 中介軟體和 handler 就可以方便地訪問用戶資訊
	authGroup.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := c.Get("user").(*jwt.Token) // Echo JWT 將解析後的 token 存為 "user"
			claims, ok := token.Claims.(*jwt.AccessClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, "Invalid token claims type")
			}
			c.Set("claims", claims) // 將自定義的 AccessClaims 存入上下文
			return next(c)
		}
	})

	// --- 應用細粒度授權中介軟體 (authz.Authorize) ---
	// 傳入每個 API 端點所需的特定權限字串
	// 格式通常是 "資源:操作"，例如 "company:read", "account:create"

	// 帳戶管理路由
	authGroup.GET("/accounts", accountHandler.GetAccounts, authz.Authorize("account:read", permissionService))
	authGroup.GET("/accounts/:id", accountHandler.GetAccountById, authz.Authorize("account:read", permissionService))
	authGroup.POST("/accounts", accountHandler.CreateAccount, authz.Authorize("account:create", permissionService))
	authGroup.PUT("/accounts/:id", accountHandler.UpdateAccount, authz.Authorize("account:update", permissionService))
	authGroup.DELETE("/accounts/:id", accountHandler.DeleteAccount, authz.Authorize("account:delete", permissionService))
	authGroup.POST("/accounts/:id/password", accountHandler.UpdateAccountPassword, authz.Authorize("account:update_password", permissionService))
	authGroup.GET("/my-profile", authHandler.GetMyProfile, authz.Authorize("account:read_own_profile", permissionService)) // 用戶查看自己資料

	// 公司管理路由
	authGroup.GET("/companies", companyHandler.GetCompanies, authz.Authorize("company:read", permissionService))
	authGroup.GET("/companies/:id", companyHandler.GetCompanyById, authz.Authorize("company:read", permissionService))
	authGroup.POST("/companies", companyHandler.CreateCompany, authz.Authorize("company:create", permissionService))
	authGroup.PUT("/companies/:id", companyHandler.UpdateCompany, authz.Authorize("company:update", permissionService))
	authGroup.DELETE("/companies/:id", companyHandler.DeleteCompany, authz.Authorize("company:delete", permissionService))

	// 客戶管理路由
	authGroup.GET("/customers", customerHandler.GetCustomers, authz.Authorize("customer:read", permissionService))
	authGroup.GET("/customers/:id", customerHandler.GetCustomerById, authz.Authorize("customer:read", permissionService))
	authGroup.POST("/customers", customerHandler.CreateCustomer, authz.Authorize("customer:create", permissionService))
	authGroup.PUT("/customers/:id", customerHandler.UpdateCustomer, authz.Authorize("customer:update", permissionService))
	authGroup.DELETE("/customers/:id", customerHandler.DeleteCustomer, authz.Authorize("customer:delete", permissionService))

	// 選單管理路由
	authGroup.GET("/menus", menuHandler.GetMenus, authz.Authorize("menu:read", permissionService))
	authGroup.GET("/menus/:id", menuHandler.GetMenuById, authz.Authorize("menu:read", permissionService))
	authGroup.POST("/menus", menuHandler.CreateMenu, authz.Authorize("menu:create", permissionService))
	authGroup.PUT("/menus/:id", menuHandler.UpdateMenu, authz.Authorize("menu:update", permissionService))
	authGroup.DELETE("/menus/:id", menuHandler.DeleteMenu, authz.Authorize("menu:delete", permissionService))

	// 產品類別和產品定義管理路由
	authGroup.GET("/product_categories", productDefinitionHandler.GetProductCategories, authz.Authorize("product_category:read", permissionService))
	authGroup.POST("/product_categories", productDefinitionHandler.CreateProductCategory, authz.Authorize("product_category:create", permissionService))
	authGroup.PUT("/product_categories/:id", productDefinitionHandler.UpdateProductCategory, authz.Authorize("product_category:update", permissionService))
	authGroup.DELETE("/product_categories/:id", productDefinitionHandler.DeleteProductCategory, authz.Authorize("product_category:delete", permissionService))

	authGroup.GET("/product_definitions", productDefinitionHandler.GetProductDefinitions, authz.Authorize("product_definition:read", permissionService))
	authGroup.GET("/product_definitions/:id", productDefinitionHandler.GetProductDefinitionById, authz.Authorize("product_definition:read", permissionService))
	authGroup.POST("/product_definitions", productDefinitionHandler.CreateProductDefinition, authz.Authorize("product_definition:create", permissionService))
	authGroup.PUT("/product_definitions/:id", productDefinitionHandler.UpdateProductDefinition, authz.Authorize("product_definition:update", permissionService))
	authGroup.DELETE("/product_definitions/:id", productDefinitionHandler.DeleteProductDefinition, authz.Authorize("product_definition:delete", permissionService))

	// 角色選單關聯管理路由
	authGroup.GET("/role_menus", roleMenuHandler.GetRoleMenus, authz.Authorize("role_menu:read", permissionService))
	authGroup.POST("/role_menus", roleMenuHandler.CreateRoleMenu, authz.Authorize("role_menu:create", permissionService))
	authGroup.DELETE("/role_menus/:id1/:id2", roleMenuHandler.DeleteRoleMenu, authz.Authorize("role_menu:delete", permissionService)) // 複合主鍵刪除
	authGroup.PUT("/role_menus/:id1/:id2", roleMenuHandler.UpdateRoleMenu, authz.Authorize("role_menu:update", permissionService)) // 複合主鍵更新

	// (範例) 獲取特定角色可訪問的選單 - 這個路由可以直接從前端使用來獲取動態選單
	// 由於這個是專門為前端獲取選單數據而設計，其權限檢查可能略有不同，
	// 例如只檢查是否登入，而不是是否有特定選單管理權限。
	// 或者，只允許「admin」角色呼叫這個 API。
	authGroup.GET("/roles/:roleID/menus", menuHandler.GetMenusByRoleID, authz.Authorize("role:read_menus", permissionService)) // 新增權限字串
}

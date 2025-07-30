package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/wac0705/fastener-api/middleware/jwt" // 導入 JWT Claims
	"github.com/wac0705/fastener-api/service"       // 導入權限服務
	"github.com/wac0705/fastener-api/utils"         // 導入自定義錯誤
)

// Authorize 授權中介軟體，根據用戶角色檢查是否具備指定權限
// permission 參數是這個 API 端點所需的權限字串，例如 "company:read"
func Authorize(permission string, permissionService service.PermissionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 從上下文中獲取 JWT claims (假設 JWT 中介軟體已將 claims 設置為 "claims")
			claims, ok := c.Get("claims").(*jwt.AccessClaims)
			if !ok || claims == nil {
				// 這通常表示 JWT 中介軟體沒有正確執行，或者 Token 解析失敗
				zap.L().Warn("Authorization failed: JWT claims not found or invalid in context",
					zap.String("path", c.Path()), zap.String("method", c.Request().Method))
				return c.JSON(http.StatusUnauthorized, utils.ErrUnauthorized.SetDetails("Invalid or missing authentication credentials"))
			}

			// 如果是超級管理員角色 (假設 RoleID=1 是 admin)，則直接放行所有權限
			// 這是快速路徑，實際 RoleID 需要和你的資料庫設定一致
			if claims.RoleID == 1 { // 假設 1 是 admin 角色 ID
				return next(c)
			}

			// 檢查用戶角色是否具備所需權限
			hasPermission, err := permissionService.HasPermission(claims.RoleID, permission)
			if err != nil {
				zap.L().Error("Error checking permission for user",
					zap.Int("account_id", claims.AccountID),
					zap.Int("role_id", claims.RoleID),
					zap.String("required_permission", permission),
					zap.Error(err),
					zap.String("path", c.Path()), zap.String("method", c.Request().Method))
				return c.JSON(http.StatusInternalServerError, utils.ErrInternalServer)
			}

			if !hasPermission {
				zap.L().Warn("User forbidden from accessing resource due to insufficient permissions",
					zap.Int("account_id", claims.AccountID),
					zap.Int("role_id", claims.RoleID),
					zap.String("required_permission", permission),
					zap.String("path", c.Path()), zap.String("method", c.Request().Method))
				return c.JSON(http.StatusForbidden, utils.ErrForbidden.SetDetails("Insufficient permissions to perform this action"))
			}

			return next(c) // 繼續處理請求
		}
	}
}

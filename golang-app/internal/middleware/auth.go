package middleware

import (
	"golang-app/internal/config"
	"golang-app/pkg/apperror"
	"golang-app/pkg/response"
	"golang-app/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		authHeader := c.GetHeader("Authorization")

		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}

		// If header is missing or invalid, check query parameter (for WebSockets)
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			response.Error(c, apperror.NewUnauthorized(nil, "Missing Authorization Header or Token Query Param"))
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenString, cfg.App.JWTSecret)
		if err != nil {
			response.Error(c, apperror.NewUnauthorized(err, "Invalid or Expired Token"))
			c.Abort()
			return
		}

		// Set claims to context for later use
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists {
			response.Error(c, apperror.NewUnauthorized(nil, "User role not found in context"))
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.Error(c, apperror.NewInternal(nil)) // Nên log lỗi "kiểu dữ liệu role không hợp lệ"
			c.Abort()
			return
		}

		isAllowed := false
		for _, r := range allowedRoles {
			if userRole == r {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response.Error(c, apperror.NewForbidden(nil, "Bạn không có quyền truy cập chức năng này (Forbidden)")) // Hoặc new 403 Forbidden Error
			c.Abort()
			return
		}

		c.Next()
	}
}

package middleware

import (
	"net/http"
	"user-service/internal/domain"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoleVal, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Yêu cầu không được xác thực danh tính"})
			c.Abort()
			return 
		}

		userRole := userRoleVal.(string)

		if userRole == domain.RoleAdmin {
			c.Next()
			return 
		}

		isAllowed := false 
		for _, role := range allowedRoles {
			if userRole == role {
				isAllowed = true 
				break 
			}
		}

		if !isAllowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Quyền truy cập bị từ chối! bạn không có thẩm quyền thực hiện hành động này.",
			})
			c.Abort()
			return 
		}
		c.Next()
	}
}
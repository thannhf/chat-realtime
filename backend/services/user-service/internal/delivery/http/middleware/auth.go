package middleware

import (
	"net/http"
	"strings"

	"common/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc{
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error":"thiếu token xác thực"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer"){
			c.JSON(http.StatusUnauthorized, gin.H{"error":"định dạng token không hợp lệ"})
			c.Abort()
			return 
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error":"Token hết hạn hoặc không hợp lệ"})
			c.Abort()
			return 
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}
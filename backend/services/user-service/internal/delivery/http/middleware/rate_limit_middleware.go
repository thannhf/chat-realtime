package middleware

import (
	"net/http"
	"time"
	"user-service/internal/domain"

	"github.com/gin-gonic/gin"
)

func RateLimiter(cache domain.UserCache, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		key := clientIP + ":" + c.FullPath()

		isLimited, err := cache.IsRateLimited(c.Request.Context(), key, limit, window)
		if err != nil {
			c.Next()
			return 
		}

		if isLimited {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":"Bạn đang thao tác quá nhanh, vui lòng thử lại sau ít phút",
			})
			c.Abort()
			return 
		}
		c.Next()
	}
}
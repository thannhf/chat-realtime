package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	limiters = make(map[string]*clientLimiter)
	mu       sync.Mutex
)

func init() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			mu.Lock()
			for ip, client := range limiters {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(limiters, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

func RateLimiter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		
		if _, exists := limiters[ip]; !exists {
			limiters[ip] = &clientLimiter{
				limiter: rate.NewLimiter(rate.Every(time.Second/10), 20),
			}
		}
		limiters[ip].lastSeen = time.Now()

		if !limiters[ip].limiter.Allow() {
			mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Hệ thống bận: Bạn đang gửi quá nhiều yêu cầu, vui lòng thử lại sau!",
			})
			c.Abort()
			return
		}
		mu.Unlock()
		c.Next()
	}
}
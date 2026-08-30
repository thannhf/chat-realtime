package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc{
	return func(c *gin.Context) {
		start := time.Now()
		reqID, _ := c.Get("requestID")
		c.Next()
		log.Printf("[GATEWAY] ID:%v | %s | %s | %d | %v", 
			reqID, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
	}
}
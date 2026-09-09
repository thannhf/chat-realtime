package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db			*gorm.DB 
	redisClient	*redis.Client
}

func NewHealthHandler(r *gin.Engine, db *gorm.DB, redisClient *redis.Client) {
	h := &HealthHandler{db: db, redisClient: redisClient}

	r.GET("/health", h.Check)
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2 * time.Second)
	defer cancel()

	status := gin.H{
		"status": "UP",
		"timestamp": time.Now().Format(time.RFC3339),
		"details": gin.H{},
	}
	isHealthy := true 

	// check health 
	sqlDB, err := h.db.DB()
	if err != nil {
		status["details"].(gin.H)["postgres"] = "DOWN - Cannot get SQL instance"
		isHealthy = false
	} else if err := sqlDB.PingContext(ctx); err != nil {
		status["details"].(gin.H)["postgres"] = "Down - ping failed"
		isHealthy = false 
	} else {
		status["details"].(gin.H)["postgres"] = "Up"
	}

	if h.redisClient == nil {
		status["details"].(gin.H)["redis"] = "Down - Client not initialized"
		isHealthy = false 
	} else if err := h.redisClient.Ping(ctx).Err(); err != nil {
		status["details"].(gin.H)["redis"] = "Down - Ping failed"
		isHealthy = false 
	} else {
		status["details"].(gin.H)["redis"] = "Up"
	}

	if !isHealthy {
		status["status"] = "Down"
		c.JSON(http.StatusServiceUnavailable, status)
		return 
	}
	c.JSON(http.StatusOK, status)
}
package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/config"
	"api-gateway/internal/middleware" 

	"github.com/gin-gonic/gin"
)

func proxyHandler(targetURL string) gin.HandlerFunc {
	target, _ := url.Parse(targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		c.Request.Host = target.Host
		c.Request.URL.Host = target.Host
		c.Request.URL.Scheme = target.Scheme
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	log.Println("--- KHỞI ĐỘNG API GATEWAY ---")
	cfg := config.LoadConfig()

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())           
	router.Use(middleware.RequestID())  
	router.Use(middleware.Logger())   
	router.Use(middleware.CORS())      
	router.Use(middleware.RateLimiter()) 


	router.Any("/api/v1/auth/*any", proxyHandler(cfg.UserServiceURL))
	router.Any("/api/v1/users/*any", proxyHandler(cfg.UserServiceURL))
	router.Any("/api/v1/chat/*any", proxyHandler(cfg.ChatServiceURL))
	router.Any("/api/v1/media/*any", proxyHandler(cfg.MediaServiceURL))
	router.Static("/uploads", cfg.MediaUploadDir)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: router}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi nghiêm trọng: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Đang dừng Gateway...")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	srv.Shutdown(ctx)
	log.Println("API Gateway đã đóng hoàn toàn sạch sẽ.")
}
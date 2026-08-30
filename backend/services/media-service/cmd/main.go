package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"media-service/config"
	deliveryHttp "media-service/internal/delivery/http"
	"media-service/internal/delivery/http/middleware"
	"media-service/internal/repository/postgres"
	"media-service/internal/repository/storage"
	"media-service/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("--- ĐANG KHỞI ĐỘNG MEDIA SERVICE ---")

	cfg := config.LoadConfig()
	db := postgres.InitDatabase(&cfg.DB)

	if cfg.Storage.Provider == "local" {
		if err := os.MkdirAll(cfg.Storage.UploadDir, os.ModePerm); err != nil {
			log.Fatalf("Không thể khởi tạo thư mục lưu trữ file: %v", err)
		}
	}

	mediaRepo := postgres.NewMediaRepository(db)
	storageEngine := storage.NewLocalStorageEngine(cfg)
	mediaUsecase := usecase.NewMediaUsecase(mediaRepo, storageEngine, cfg)

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())

	if cfg.Storage.Provider == "local" {
		router.Static("/uploads", cfg.Storage.UploadDir)
	}

	api := router.Group("/api/v1/media")
	api.Use(middleware.AuthMiddleware()) 
	{
		deliveryHttp.NewMediaHandler(api, mediaUsecase)
	}

	internalAPI := router.Group("/api/internal")
	{
		deliveryHttp.NewInternalMediaHandler(internalAPI, mediaUsecase)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.App.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Media Service đang lắng nghe tại cổng: %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi nghiêm trọng khi chạy server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Đang tiến hành ngắt kết nối Media Service an toàn...")
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server buộc phải dừng đột ngột: %v", err)
	}

	log.Println("Media Service đã dừng hoạt động 100% an toàn.")
}
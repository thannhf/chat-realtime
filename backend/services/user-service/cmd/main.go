package main

import (
	"context"
	"log"
	"time"
	"os/signal"
	"syscall"
	"os"

	"common/logger"
	"user-service/config"
	"user-service/internal/bootstrap"
	"user-service/internal/delivery/http"
	"user-service/internal/domain"
	"user-service/internal/repository"
	"user-service/internal/usecase"

	"common/redis"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	netHttp "net/http"
)

func main() {
	logger.InitLogger()
	err := godotenv.Load()
	cfg := config.LoadConfig()
	
	if err != nil {
		logger.Log.Warn("không tìm thấy file .env")
	}
	
	// connect DB
	db, err := bootstrap.NewDatabase(&cfg.DB)
	if err != nil {
		logger.Log.Fatal("không thể kết nối database", logger.Error(err))
	}
	if err := db.Write.AutoMigrate(&domain.User{}); err != nil {
		logger.Log.Fatal("Không thể migrate database", logger.Error(err))
	}

	// connect redis cache
	rdb, err := redis.InitRedis(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		logger.Log.Warn("Hệ thống sẽ chạy không có Cache do lỗi kết nối Redis", logger.Error(err))
	} else {
		logger.Log.Info("Kết nối cơ sở dữ liệu Redis thành công!")
	}

	userRepo := repository.NewPostgresUserRepository(db.Write, db.Read)
	userCache := repository.NewUserCache(rdb) 
	userUC := usecase.NewUserUsecase(userRepo, userCache)

	r := gin.Default()
	http.NewHealthHandler(r, db.Write, rdb)
	http.NewUserHandler(r, userUC, userCache, cfg)

	port := cfg.App.Port
	if port == "" {
		port = "8001" 
	}

	server := &netHttp.Server{
		Addr:    ":" + port,
		Handler: r, 
		ReadTimeout:       5 * time.Second,   
		WriteTimeout:      10 * time.Second,  
		IdleTimeout:       120 * time.Second, 
		ReadHeaderTimeout: 2 * time.Second,   
	}

	go func() {
		logger.Log.Info("User Service bắt đầu khởi chạy an toàn...", logger.String("port", port))

		if err := server.ListenAndServe(); err != nil && err != netHttp.ErrServerClosed {
			logger.Log.Fatal("Lỗi nghiêm trọng, Server bị sập đột ngột!", logger.Error(err))
		}	
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("đang tiến hành ngắt kết nối chat service an toàn...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server buộc phải đóng đột ngột: %v", err)
	}

	logger.Log.Info("User service đã dừng hoạt động 100% an toàn. Hệ thống sạch")

	db.Close()
}
package main

import (
	"chat-service/config"
	deliveryHttp "chat-service/internal/delivery/http"
	"chat-service/internal/delivery/http/middleware"
	"chat-service/internal/delivery/websocket"
	"chat-service/internal/repository/external"
	"chat-service/internal/repository/postgres"
	redisRepository "chat-service/internal/repository/redis"
	"chat-service/internal/usecase"
	redisRepo "common/redis"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("--- ĐANG KHỞI ĐỘNG CHAT SERVICE ---")

	cfg := config.LoadConfig()
	db := postgres.InitDatabase(&cfg.DB)

	rdb, err := redisRepo.InitRedis(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Không thể kết nối đến Redis Cache: %v", err)
	}
	log.Println("Kết nối đến Redis thành công với cấu hình Pool tối ưu!")
	
	// Repositories
	msgRepo := postgres.NewMessageRepository(db)
	convRepo := postgres.NewConversationRepository(db)
	userGateway := external.NewUserGateway(cfg)
	presenceRepo := redisRepository.NewPresenceRepository(rdb)

	// usecase
	presenceUsecase := usecase.NewPresenceUsecase(presenceRepo)
	chatUsecase := usecase.NewChatUsecase(msgRepo, convRepo, userGateway, nil)

	// Delivery 
	hub := websocket.NewHub(chatUsecase, presenceUsecase)

	chatUsecase.SetReadReceiptPublisher(hub)
	go hub.Run() 
	log.Println("WebSocket Hub định danh UUID đã sẵn sàng vận hành!")

	gin.SetMode(gin.ReleaseMode) 
	router := gin.New()

	router.Use(gin.Recovery()) 

	authMidd := middleware.AuthMiddleware()
	router.Use(authMidd)

	deliveryHttp.NewChatHandler(router, chatUsecase, presenceUsecase, hub)

	srv := &http.Server{
		Addr:    ":" + cfg.App.Port, 
		Handler: router,
	}

	go func() {
		log.Printf("Chat Service đang lắng nghe tại cổng: %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi nghiêm trọng khi khởi chạy server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Đang tiến hành ngắt kết nối Chat Service an toàn...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server buộc phải đóng đột ngột: %v", err)
	}

	log.Println("Chat Service đã dừng hoạt động 100% an toàn. Hệ thống sạch!")
}
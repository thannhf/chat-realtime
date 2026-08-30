package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string
	Env string 
	UserServiceURL string 
	ChatServiceURL string 
	MediaServiceURL string 
	MediaUploadDir string 
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Cảnh báo: không thể tìm thấy file .env của api-gateway")
	}

	gatewayPort := os.Getenv("GATEWAY_PORT")
	if gatewayPort == "" {
		gatewayPort = "8000"
	}

	return &Config{
		Port: gatewayPort,
		Env:             os.Getenv("APP_ENV"),
		UserServiceURL:  os.Getenv("USER_SERVICE_URL"),
		ChatServiceURL:  os.Getenv("CHAT_SERVICE_URL"),
		MediaServiceURL: os.Getenv("MEDIA_SERVICE_URL"),
		MediaUploadDir:  os.Getenv("MEDIA_UPLOAD_DIR"),
	}
}
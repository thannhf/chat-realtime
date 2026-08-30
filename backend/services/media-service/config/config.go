package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App AppConfig
	DB DBConfig
	Storage StorageConfig
}

type AppConfig struct {
	Port string 
	Env string 
}

type DBConfig struct {
	WriteHost string 
	ReadHost string
}

type StorageConfig struct {
	Provider string 
	UploadDir string 
	MaxFileSize int64  
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("không tìm thấy hoặc không thể nạp file .env")
	}

	appPort := os.Getenv("SERVICE_PORT")
	if appPort == "" {
		appPort = ""
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = ""
	}

	maxSize, err := strconv.ParseInt(os.Getenv("UPLOAD_MAX_SIZE"), 10, 64)
	if err != nil || maxSize <= 0 {
		maxSize = 10 * 1024 * 1024
	}

	storageProvider := os.Getenv("STORAGE_PROVIDER")
	if storageProvider == "" {
		storageProvider = ""
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = ""
	}

	return &Config{
		App: AppConfig{
			Port: appPort,
			Env: appEnv,
		},
		DB: DBConfig{
			WriteHost: os.Getenv("DB_MASTER_DSN"),
			ReadHost: os.Getenv("DB_REPLICA_DSN"),
		},
		Storage: StorageConfig{
			Provider: storageProvider,
			UploadDir: uploadDir,
			MaxFileSize: maxSize,
		},
	}
}
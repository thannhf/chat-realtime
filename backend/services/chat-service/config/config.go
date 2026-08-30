package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	App AppConfig
	DB DBConfig
	Redis RedisConfig
}

type AppConfig struct {
	Port string 
	AppURL string 
}

type DBConfig struct {
	WriteHost string 
	ReadHost string
	Port string 
	User string 
	Password string 
	Name string 
	SSLMode string 
}

type RedisConfig struct {
	Host string 
	Port string 
	Password string 
	DB int 
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("không thể load được file .env")
	}
	
	appPort := os.Getenv("SERVICE_PORT")
	if appPort == "" {
		appPort = ""
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		fmt.Printf("lỗi không tìm thấy redisDB")
	}

	dbWriteHost := os.Getenv("DB_MASTER_DSN")
	if dbWriteHost == "" {
		dbWriteHost = ""
	}

	dbReadHost := os.Getenv("DB_REPLICA_DSN")
	if dbReadHost == "" {
		dbReadHost = ""
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = ""
	}

	return &Config{
		App: AppConfig{
			Port: appPort,
		},

		DB: DBConfig{
			WriteHost: dbWriteHost,
			ReadHost: dbReadHost,
			Port: dbPort,
			User: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			Name: os.Getenv("DB_NAME"),
			SSLMode: "disable",
		},

		Redis: RedisConfig{
			Host: os.Getenv("REDIS_HOST"),
			Port: os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB: redisDB,
		},
	}
}
package config

import (
	"os"
	"strconv"
)

type Config struct {
	App    AppConfig
	DB     DBConfig
	Redis  RedisConfig
	AppURL URLConfig
	GatewayURL URLConfig
}

type AppConfig struct {
	Port string
	Env  string
}

type URLConfig struct {
	URL string
}

type DBConfig struct {
	WriteHost string
	ReadHost  string
	Port      string
	User      string
	Password  string
	Name      string
	SSLMode   string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadConfig() *Config {
	appPort := os.Getenv("SERVICE_PORT")
	if appPort == "" {
		appPort = ""
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}

	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		redisDB = 0
	}

	dbWriteHost := os.Getenv("DB_WRITE_HOST")
	if dbWriteHost == "" {
		dbWriteHost = os.Getenv("DB_HOST")
	}

	dbReadHost := os.Getenv("DB_READ_HOST")
	if dbReadHost == "" {
		dbReadHost = os.Getenv("DB_HOST")
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = ""
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = ""
	}

	gatewayURL := os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = ""
	}

	return &Config{
		App: AppConfig{
			Port: appPort,
			Env:  appEnv,
		},
		AppURL: URLConfig{
			URL: appURL,
		},
		GatewayURL: URLConfig{
			URL: gatewayURL,
		},
		DB: DBConfig{
			WriteHost: dbWriteHost,
			ReadHost:  dbReadHost,
			Port:      dbPort,
			User:      os.Getenv("DB_USER"),
			Password:  os.Getenv("DB_PASSWORD"),
			Name:      os.Getenv("DB_NAME"),
			SSLMode:   "disable",
		},
		Redis: RedisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Port:     os.Getenv("REDIS_PORT"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
	}
}

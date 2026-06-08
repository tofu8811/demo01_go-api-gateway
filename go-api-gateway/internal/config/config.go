package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort           string
	AuthServiceURL   string
	ProductServiceURL string
	OrderServiceURL   string
	GatewaySecret     string
}

func Load() Config {
	_ = godotenv.Load()

	return Config{
		AppPort:           getEnv("APP_PORT", "4000"),
		AuthServiceURL:   os.Getenv("AUTH_SERVICE_URL"),
		ProductServiceURL: os.Getenv("PRODUCT_SERVICE_URL"),
		OrderServiceURL:   os.Getenv("ORDER_SERVICE_URL"),
		GatewaySecret:     os.Getenv("GATEWAY_SECRET"),
	}
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
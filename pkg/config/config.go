package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	BackendURL      string
	RateLimit       int
	ContainerID     string
	WebhookURL      string
	ProxyTimeout    time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	config := &Config{
		BackendURL:      getEnv("BACKEND_URL", "http://localhost:8080"),
		RateLimit:       getEnvAsInt("RATE_LIMIT", 10),
		ContainerID:     getEnv("CONTAINER_ID", "unknown-container"),
		WebhookURL:      getEnv("WEBHOOK_URL", ""),
		ProxyTimeout:    time.Duration(getEnvAsInt("PROXY_TIMEOUT_SECONDS", 30)) * time.Second,
		MaxIdleConns:    getEnvAsInt("MAX_IDLE_CONNS", 100),
		IdleConnTimeout: time.Duration(getEnvAsInt("IDLE_CONN_TIMEOUT_SECONDS", 90)) * time.Second,
	}

	return config, nil
}

// getEnv reads an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}

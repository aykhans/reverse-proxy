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
		RateLimit:       getEnv("RATE_LIMIT", 10),
		ContainerID:     getEnv("CONTAINER_ID", "unknown-container"),
		WebhookURL:      getEnv("WEBHOOK_URL", ""),
		ProxyTimeout:    getEnv("PROXY_TIMEOUT", time.Second*30),
		MaxIdleConns:    getEnv("MAX_IDLE_CONNS", 100),
		IdleConnTimeout: getEnv("IDLE_CONN_TIMEOUT", time.Second*90),
	}

	return config, nil
}

// getEnv is a generic function that retrieves and parses environment variables of different types.
// It takes two parameters:
//   - s: the name of the environment variable to retrieve
//   - defaultValue: default value to return if environment variable is missing or invalid
//
// Returns parsed value of type T or the default value if parsing fails.
func getEnv[T string | int | float64 | bool | time.Duration](s string, defaultValue T) T {
	env := os.Getenv(s)
	if env == "" {
		return defaultValue
	}

	switch any(defaultValue).(type) {
	case string:
		return any(env).(T)
	case int:
		i, err := strconv.Atoi(env)
		if err != nil {
			log.Printf("Warning: Invalid integer value %s=%s, using default %v", s, env, defaultValue)
			return defaultValue
		}
		return any(i).(T)
	case float64:
		f, err := strconv.ParseFloat(env, 64)
		if err != nil {
			log.Printf("Warning: Invalid float value %s=%s, using default %v", s, env, defaultValue)
			return defaultValue
		}
		return any(f).(T)
	case bool:
		b, err := strconv.ParseBool(env)
		if err != nil {
			log.Printf("Warning: Invalid boolean value %s=%s, using default %v", s, env, defaultValue)
			return defaultValue
		}
		return any(b).(T)
	case time.Duration:
		d, err := time.ParseDuration(env)
		if err != nil {
			log.Printf("Warning: Invalid duration value %s=%s, using default %v", s, env, defaultValue)
			return defaultValue
		}
		return any(d).(T)
	default:
		log.Printf("Warning: Unsupported type %T for %s=%s, using default %v", defaultValue, s, env, defaultValue)
		return defaultValue
	}
}

// Package config loads runtime settings from environment variables.
package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds every runtime setting of the service.
type Config struct {
	ServerPort string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSchema   string

	JWTPublicKey string

	HTTPReadHeaderTimeout time.Duration
	HTTPReadTimeout       time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	HTTPShutdownTimeout   time.Duration
}

// Load reads settings from the environment and applies safe defaults.
func Load() (*Config, error) {
	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "telemed_ia"),
		DBUser:     getEnv("DB_USER", "medical_consultation_app"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBSchema:   getEnv("DB_SCHEMA", "medical_consultation"),

		JWTPublicKey: os.Getenv("JWT_PUBLIC_KEY"),

		HTTPReadHeaderTimeout: getDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
		HTTPReadTimeout:       getDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		HTTPWriteTimeout:      getDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		HTTPIdleTimeout:       getDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
		HTTPShutdownTimeout:   getDuration("HTTP_SHUTDOWN_TIMEOUT", 20*time.Second),
	}

	if cfg.JWTPublicKey == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY is required")
	}
	return cfg, nil
}

// DSN returns the PostgreSQL connection string with search_path set to the
// service schema, so every query runs inside `medical_consultation`.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSchema,
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

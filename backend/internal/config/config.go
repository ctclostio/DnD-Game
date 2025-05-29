package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	AI       AIConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        string
	Environment string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// RedisConfig holds Redis-related configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret              string
	AccessTokenDuration    time.Duration
	RefreshTokenDuration   time.Duration
	BcryptCost             int
}

// AIConfig holds AI/LLM-related configuration
type AIConfig struct {
	Provider string // "openai", "anthropic", or "mock"
	APIKey   string
	Model    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Server configuration
	cfg.Server.Port = getEnv("PORT", "8080")
	cfg.Server.Environment = getEnv("ENV", "development")

	// Database configuration
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnvAsInt("DB_PORT", 5432)
	cfg.Database.User = getEnv("DB_USER", "dndgame")
	cfg.Database.Password = getEnv("DB_PASSWORD", "dndgamepass")
	cfg.Database.DatabaseName = getEnv("DB_NAME", "dndgame")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Database.MaxOpenConns = getEnvAsInt("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getEnvAsInt("DB_MAX_IDLE_CONNS", 25)
	cfg.Database.MaxLifetime = getEnvAsDuration("DB_MAX_LIFETIME", 5*time.Minute)

	// Redis configuration
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnvAsInt("REDIS_PORT", 6379)
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)

	// Auth configuration
	cfg.Auth.JWTSecret = getEnv("JWT_SECRET", "your-secret-key-change-this-in-production")
	cfg.Auth.AccessTokenDuration = getEnvAsDuration("ACCESS_TOKEN_DURATION", 15*time.Minute)
	cfg.Auth.RefreshTokenDuration = getEnvAsDuration("REFRESH_TOKEN_DURATION", 7*24*time.Hour)
	cfg.Auth.BcryptCost = getEnvAsInt("BCRYPT_COST", 10)

	// AI configuration
	cfg.AI.Provider = getEnv("AI_PROVIDER", "mock") // Default to mock for development
	cfg.AI.APIKey = getEnv("AI_API_KEY", "")
	cfg.AI.Model = getEnv("AI_MODEL", "gpt-4-turbo-preview") // Default model

	return cfg, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// getEnvAsDuration gets an environment variable as duration with a fallback value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	strValue := getEnv(key, "")
	if strValue == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(strValue)
	if err != nil {
		return defaultValue
	}
	return duration
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.DatabaseName == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "your-secret-key-change-this-in-production" {
		return fmt.Errorf("JWT secret must be set to a secure value")
	}
	if c.Auth.AccessTokenDuration <= 0 {
		return fmt.Errorf("access token duration must be positive")
	}
	if c.Auth.RefreshTokenDuration <= 0 {
		return fmt.Errorf("refresh token duration must be positive")
	}
	if c.Auth.BcryptCost < 4 || c.Auth.BcryptCost > 31 {
		return fmt.Errorf("bcrypt cost must be between 4 and 31")
	}
	return nil
}
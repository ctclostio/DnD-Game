package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testConfigJWTSecret = "a-very-long-secret-key-that-is-at-least-32-chars"

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"PORT", "ENV",
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_MAX_LIFETIME",
		"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD", "REDIS_DB",
		"JWT_SECRET", "ACCESS_TOKEN_DURATION", "REFRESH_TOKEN_DURATION", "BCRYPT_COST",
		"AI_PROVIDER", "AI_API_KEY", "AI_MODEL",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		require.NoError(t, os.Unsetenv(key))
	}
	defer func() {
		// Restore original env vars
		for key, value := range originalEnv {
			if value != "" {
				require.NoError(t, os.Setenv(key, value))
			} else {
				require.NoError(t, os.Unsetenv(key))
			}
		}
	}()

	t.Run("loads default configuration", func(t *testing.T) {
		cfg, err := Load()
		require.NoError(t, err)

		// Check default values
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, "development", cfg.Server.Environment)

		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, "dndgame", cfg.Database.User)
		assert.Equal(t, "", cfg.Database.Password) // No default for password
		assert.Equal(t, "dndgame", cfg.Database.DatabaseName)
		assert.Equal(t, "disable", cfg.Database.SSLMode)
		assert.Equal(t, 25, cfg.Database.MaxOpenConns)
		assert.Equal(t, 25, cfg.Database.MaxIdleConns)
		assert.Equal(t, 5*time.Minute, cfg.Database.MaxLifetime)

		assert.Equal(t, "localhost", cfg.Redis.Host)
		assert.Equal(t, 6379, cfg.Redis.Port)
		assert.Equal(t, "", cfg.Redis.Password)
		assert.Equal(t, 0, cfg.Redis.DB)

		assert.Equal(t, "", cfg.Auth.JWTSecret) // No default for JWT secret
		assert.Equal(t, 15*time.Minute, cfg.Auth.AccessTokenDuration)
		assert.Equal(t, 7*24*time.Hour, cfg.Auth.RefreshTokenDuration)
		assert.Equal(t, 10, cfg.Auth.BcryptCost)

		assert.Equal(t, "mock", cfg.AI.Provider)
		assert.Equal(t, "", cfg.AI.APIKey)
		assert.Equal(t, "gpt-4-turbo-preview", cfg.AI.Model)
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		// Set test env vars
		require.NoError(t, os.Setenv("PORT", "3000"))
		require.NoError(t, os.Setenv("ENV", "production"))
		require.NoError(t, os.Setenv("DB_HOST", "test-host"))
		require.NoError(t, os.Setenv("DB_PORT", "5433"))
		require.NoError(t, os.Setenv("DB_USER", "test-user"))
		require.NoError(t, os.Setenv("DB_PASSWORD", "test-pass"))
		require.NoError(t, os.Setenv("DB_NAME", "test-db"))
		require.NoError(t, os.Setenv("DB_SSLMODE", "require"))
		require.NoError(t, os.Setenv("DB_MAX_OPEN_CONNS", "50"))
		require.NoError(t, os.Setenv("DB_MAX_IDLE_CONNS", "10"))
		require.NoError(t, os.Setenv("DB_MAX_LIFETIME", "10m"))
		require.NoError(t, os.Setenv("REDIS_HOST", "redis-host"))
		require.NoError(t, os.Setenv("REDIS_PORT", "6380"))
		require.NoError(t, os.Setenv("REDIS_PASSWORD", "redis-pass"))
		require.NoError(t, os.Setenv("REDIS_DB", "1"))
		require.NoError(t, os.Setenv("JWT_SECRET", "test-secret-key-that-is-long-enough"))
		require.NoError(t, os.Setenv("ACCESS_TOKEN_DURATION", "30m"))
		require.NoError(t, os.Setenv("REFRESH_TOKEN_DURATION", "336h"))
		require.NoError(t, os.Setenv("BCRYPT_COST", "12"))
		require.NoError(t, os.Setenv("AI_PROVIDER", "openai"))
		require.NoError(t, os.Setenv("AI_API_KEY", "test-api-key"))
		require.NoError(t, os.Setenv("AI_MODEL", "gpt-4"))

		cfg, err := Load()
		require.NoError(t, err)

		assert.Equal(t, "3000", cfg.Server.Port)
		assert.Equal(t, "production", cfg.Server.Environment)
		assert.Equal(t, "test-host", cfg.Database.Host)
		assert.Equal(t, 5433, cfg.Database.Port)
		assert.Equal(t, "test-user", cfg.Database.User)
		assert.Equal(t, "test-pass", cfg.Database.Password)
		assert.Equal(t, "test-db", cfg.Database.DatabaseName)
		assert.Equal(t, "require", cfg.Database.SSLMode)
		assert.Equal(t, 50, cfg.Database.MaxOpenConns)
		assert.Equal(t, 10, cfg.Database.MaxIdleConns)
		assert.Equal(t, 10*time.Minute, cfg.Database.MaxLifetime)
		assert.Equal(t, "redis-host", cfg.Redis.Host)
		assert.Equal(t, 6380, cfg.Redis.Port)
		assert.Equal(t, "redis-pass", cfg.Redis.Password)
		assert.Equal(t, 1, cfg.Redis.DB)
		assert.Equal(t, "test-secret-key-that-is-long-enough", cfg.Auth.JWTSecret)
		assert.Equal(t, 30*time.Minute, cfg.Auth.AccessTokenDuration)
		assert.Equal(t, 14*24*time.Hour, cfg.Auth.RefreshTokenDuration)
		assert.Equal(t, 12, cfg.Auth.BcryptCost)
		assert.Equal(t, "openai", cfg.AI.Provider)
		assert.Equal(t, "test-api-key", cfg.AI.APIKey)
		assert.Equal(t, "gpt-4", cfg.AI.Model)
	})

	t.Run("handles invalid port", func(t *testing.T) {
		require.NoError(t, os.Setenv("DB_PORT", "invalid"))

		cfg, err := Load()
		require.NoError(t, err)
		// Should fall back to default
		assert.Equal(t, 5432, cfg.Database.Port)
	})

	t.Run("handles invalid duration", func(t *testing.T) {
		require.NoError(t, os.Setenv("ACCESS_TOKEN_DURATION", "invalid"))

		cfg, err := Load()
		require.NoError(t, err)
		// Should fall back to default
		assert.Equal(t, 15*time.Minute, cfg.Auth.AccessTokenDuration)
	})
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid configuration",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
				AI: AIConfig{
					Provider: "mock", // Mock provider doesn't require API key
				},
			},
			wantErr: false,
		},
		{
			name: "missing server port",
			config: &Config{
				Server: ServerConfig{
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
			},
			wantErr: true,
			errMsg:  "server port is required",
		},
		{
			name: "missing database host",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "missing database user",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
			},
			wantErr: true,
			errMsg:  "database user is required",
		},
		{
			name: "missing database password",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
			},
			wantErr: true,
			errMsg:  "database password is required",
		},
		{
			name: "missing database name",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "user",
					Password: "pass",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
			},
			wantErr: true,
			errMsg:  "database name is required",
		},
		{
			name: "missing JWT secret",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
			},
			wantErr: true,
			errMsg:  "JWT secret is required",
		},
		{
			name: "JWT secret too short",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            "short-secret",
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
			},
			wantErr: true,
			errMsg:  "JWT secret must be at least 32 characters",
		},
		{
			name: "invalid bcrypt cost",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           3, // Too low
				},
			},
			wantErr: true,
			errMsg:  "bcrypt cost must be between 4 and 31",
		},
		{
			name: "AI provider requires API key",
			config: &Config{
				Server: ServerConfig{
					Port:        "8080",
					Environment: "development",
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret:            testConfigJWTSecret,
					AccessTokenDuration:  15 * time.Minute,
					RefreshTokenDuration: 7 * 24 * time.Hour,
					BcryptCost:           10,
				},
				AI: AIConfig{
					Provider: "openai",
					APIKey:   "", // Missing API key
				},
			},
			wantErr: true,
			errMsg:  "AI API key is required when AI provider is not 'mock'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

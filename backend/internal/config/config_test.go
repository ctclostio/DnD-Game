package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := make(map[string]string)
	envVars := []string{
		"DATABASE_HOST", "DATABASE_PORT", "DATABASE_USER",
		"DATABASE_PASSWORD", "DATABASE_NAME", "DATABASE_SSLMODE",
		"JWT_SECRET", "SERVER_PORT",
	}
	for _, key := range envVars {
		originalEnv[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		// Restore original env vars
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("loads default configuration", func(t *testing.T) {
		cfg, err := Load()
		require.NoError(t, err)
		
		// Check default values
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 5432, cfg.Database.Port)
		assert.Equal(t, "postgres", cfg.Database.User)
		assert.Equal(t, "disable", cfg.Database.SSLMode)
		assert.Equal(t, 10, cfg.Database.MaxOpenConns)
		assert.Equal(t, 5, cfg.Database.MaxIdleConns)
		assert.Equal(t, 1*time.Hour, cfg.Database.MaxLifetime)
		
		assert.Equal(t, "8080", cfg.Server.Port)
		assert.Equal(t, 15*time.Minute, cfg.Auth.AccessTokenDuration)
		assert.Equal(t, 7*24*time.Hour, cfg.Auth.RefreshTokenDuration)
		
		// JWT secret should be generated
		assert.NotEmpty(t, cfg.Auth.JWTSecret)
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		// Set test env vars
		os.Setenv("DATABASE_HOST", "test-host")
		os.Setenv("DATABASE_PORT", "5433")
		os.Setenv("DATABASE_USER", "test-user")
		os.Setenv("DATABASE_PASSWORD", "test-pass")
		os.Setenv("DATABASE_NAME", "test-db")
		os.Setenv("DATABASE_SSLMODE", "require")
		os.Setenv("JWT_SECRET", "test-secret")
		os.Setenv("SERVER_PORT", "3000")
		
		cfg, err := Load()
		require.NoError(t, err)
		
		assert.Equal(t, "test-host", cfg.Database.Host)
		assert.Equal(t, 5433, cfg.Database.Port)
		assert.Equal(t, "test-user", cfg.Database.User)
		assert.Equal(t, "test-pass", cfg.Database.Password)
		assert.Equal(t, "test-db", cfg.Database.DatabaseName)
		assert.Equal(t, "require", cfg.Database.SSLMode)
		assert.Equal(t, "test-secret", cfg.Auth.JWTSecret)
		assert.Equal(t, "3000", cfg.Server.Port)
	})

	t.Run("handles invalid port", func(t *testing.T) {
		os.Setenv("DATABASE_PORT", "invalid")
		
		cfg, err := Load()
		require.NoError(t, err)
		// Should fall back to default
		assert.Equal(t, 5432, cfg.Database.Port)
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
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret: "secret",
				},
			},
			wantErr: false,
		},
		{
			name: "missing database host",
			config: &Config{
				Database: DatabaseConfig{
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret: "secret",
				},
			},
			wantErr: true,
			errMsg:  "database host is required",
		},
		{
			name: "invalid database port",
			config: &Config{
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         0,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret: "secret",
				},
			},
			wantErr: true,
			errMsg:  "database port must be positive",
		},
		{
			name: "missing database user",
			config: &Config{
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{
					JWTSecret: "secret",
				},
			},
			wantErr: true,
			errMsg:  "database user is required",
		},
		{
			name: "missing database name",
			config: &Config{
				Database: DatabaseConfig{
					Host:     "localhost",
					Port:     5432,
					User:     "user",
					Password: "pass",
				},
				Auth: AuthConfig{
					JWTSecret: "secret",
				},
			},
			wantErr: true,
			errMsg:  "database name is required",
		},
		{
			name: "missing JWT secret",
			config: &Config{
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					User:         "user",
					Password:     "pass",
					DatabaseName: "db",
				},
				Auth: AuthConfig{},
			},
			wantErr: true,
			errMsg:  "JWT secret is required",
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
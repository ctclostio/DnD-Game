package testutil

import (
	"fmt"
	"os"
	"time"
)

// SecureTestConfig provides secure test configuration
type SecureTestConfig struct {
	DefaultPassword string
	DefaultEmail    string
	DefaultUsername string
}

// GetTestConfig returns secure test configuration
// Uses environment variables when available, otherwise generates secure defaults
func GetTestConfig() *SecureTestConfig {
	return &SecureTestConfig{
		DefaultPassword: getTestPassword(),
		DefaultEmail:    getTestEmail(),
		DefaultUsername: getTestUsername(),
	}
}

// getTestPassword returns a test password from environment or generates one
func getTestPassword() string {
	if pwd := os.Getenv("TEST_PASSWORD"); pwd != "" {
		return pwd
	}
	// Generate a unique password for this test run
	return fmt.Sprintf("Test_%d_Pass!", time.Now().Unix())
}

// getTestEmail returns a test email
func getTestEmail() string {
	if email := os.Getenv("TEST_EMAIL"); email != "" {
		return email
	}
	return fmt.Sprintf("test_%d@example.com", time.Now().Unix())
}

// getTestUsername returns a test username
func getTestUsername() string {
	if username := os.Getenv("TEST_USERNAME"); username != "" {
		return username
	}
	return fmt.Sprintf("testuser_%d", time.Now().Unix())
}

// TestUserConfig represents a test user configuration
type TestUserConfig struct {
	Username string
	Email    string
	Password string
}

// NewTestUser creates a new test user with secure defaults
func NewTestUser() *TestUserConfig {
	config := GetTestConfig()
	timestamp := time.Now().Unix()

	return &TestUserConfig{
		Username: fmt.Sprintf("%s_%d", config.DefaultUsername, timestamp),
		Email:    fmt.Sprintf("test_%d@example.com", timestamp),
		Password: config.DefaultPassword,
	}
}

// Constants for test data that don't need to be secured
const (
	TestRole        = "player"
	TestAdminRole   = "admin"
	TestCharacterID = "test-character-id"
	TestSessionID   = "test-session-id"
)

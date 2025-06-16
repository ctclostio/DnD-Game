package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateSecureID generates a cryptographically secure random ID
func GenerateSecureID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure ID: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("token length must be positive")
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	
	// Use URL-safe base64 encoding
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSecureInt generates a cryptographically secure random integer in [0, max)
func GenerateSecureInt(max int64) (int64, error) {
	if max <= 0 {
		return 0, fmt.Errorf("max must be positive")
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, fmt.Errorf("failed to generate secure int: %w", err)
	}
	
	return n.Int64(), nil
}

// GenerateSecureBytes generates cryptographically secure random bytes
func GenerateSecureBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("length must be positive")
	}
	
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate secure bytes: %w", err)
	}
	
	return bytes, nil
}

// GenerateSessionID generates a secure session ID
func GenerateSessionID() (string, error) {
	// Use 32 bytes (256 bits) for session IDs
	return GenerateSecureToken(32)
}

// GenerateNonce generates a secure nonce for cryptographic operations
func GenerateNonce(length int) ([]byte, error) {
	return GenerateSecureBytes(length)
}
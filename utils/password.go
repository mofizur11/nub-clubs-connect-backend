package utils

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost = 10
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateResetToken generates a random token for password reset
func GenerateResetToken() (string, error) {
	// In production, use a cryptographically secure random token generator
	// For now, we'll use a simple approach
	randomBytes := make([]byte, 32)
	hashedToken, err := bcrypt.GenerateFromPassword(randomBytes, bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedToken), nil
}

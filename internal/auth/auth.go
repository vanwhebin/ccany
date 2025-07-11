package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles user authentication
type AuthService struct {
	sessionTimeout time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(sessionTimeout time.Duration) *AuthService {
	return &AuthService{
		sessionTimeout: sessionTimeout,
	}
}

// HashPassword hashes a password with bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	salt, err := s.generateSalt()
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Combine password with salt
	saltedPassword := password + salt

	// Hash with bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Return salt + hash (separated by $)
	return salt + "$" + string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func (s *AuthService) VerifyPassword(password, storedHash string) bool {
	// Split salt and hash
	parts := splitHash(storedHash)
	if len(parts) != 2 {
		return false
	}

	salt := parts[0]
	hash := parts[1]

	// Combine password with salt
	saltedPassword := password + salt

	// Compare with bcrypt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}

// GenerateSessionToken generates a secure session token
func (s *AuthService) GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create hash of random bytes + timestamp
	hasher := sha256.New()
	hasher.Write(bytes)
	if _, err := fmt.Fprintf(hasher, "%d", time.Now().UnixNano()); err != nil {
		return "", fmt.Errorf("failed to write to hasher: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// IsSessionValid checks if a session token is still valid
func (s *AuthService) IsSessionValid(sessionExpires time.Time) bool {
	return time.Now().Before(sessionExpires)
}

// GetSessionExpiry returns the expiry time for a new session
func (s *AuthService) GetSessionExpiry() time.Time {
	return time.Now().Add(s.sessionTimeout)
}

// generateSalt generates a random salt
func (s *AuthService) generateSalt() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// splitHash splits a stored hash into salt and hash parts
func splitHash(storedHash string) []string {
	// Find the first $ separator
	for i, char := range storedHash {
		if char == '$' {
			return []string{storedHash[:i], storedHash[i+1:]}
		}
	}
	return []string{storedHash} // Fallback for malformed hash
}

// ValidatePassword validates password strength
func (s *AuthService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 33 && char <= 126: // Printable ASCII special characters
			if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && (char < '0' || char > '9') {
				hasSpecial = true
			}
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// ValidateUsername validates username format
func (s *AuthService) ValidateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(username) > 32 {
		return fmt.Errorf("username cannot exceed 32 characters")
	}

	// Check for valid characters (alphanumeric and underscore)
	for _, char := range username {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '_' {
			return fmt.Errorf("username can only contain letters, numbers, and underscores")
		}
	}

	return nil
}

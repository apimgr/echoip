package database

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminCredentials represents admin user credentials
type AdminCredentials struct {
	Username string
	Password string
	Token    string
}

// CreateAdminUser creates the initial admin user
func (db *DB) CreateAdminUser(username, password string) (*AdminCredentials, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate token (64 characters)
	token, err := generateToken(64)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Insert into database
	_, err = db.Exec(`
		INSERT INTO admin_users (username, password_hash, token)
		VALUES (?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET
			password_hash = excluded.password_hash,
			token = excluded.token
	`, username, string(hashedPassword), token)

	if err != nil {
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	return &AdminCredentials{
		Username: username,
		Password: password,
		Token:    token,
	}, nil
}

// ValidateCredentials validates username and password
func (db *DB) ValidateCredentials(username, password string) (bool, error) {
	var hashedPassword string
	err := db.QueryRow(`
		SELECT password_hash FROM admin_users WHERE username = ?
	`, username).Scan(&hashedPassword)

	if err != nil {
		return false, nil // User not found
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, nil // Invalid password
	}

	// Update last login
	db.Exec(`UPDATE admin_users SET last_login = CURRENT_TIMESTAMP WHERE username = ?`, username)

	return true, nil
}

// ValidateToken validates an API token
func (db *DB) ValidateToken(token string) (bool, error) {
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM admin_users WHERE token = ?)
	`, token).Scan(&exists)

	return exists, err
}

// GetTokenByUsername retrieves the token for a username
func (db *DB) GetTokenByUsername(username string) (string, error) {
	var token string
	err := db.QueryRow(`
		SELECT token FROM admin_users WHERE username = ?
	`, username).Scan(&token)

	return token, err
}

// generateToken generates a cryptographically secure random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// HashString returns SHA-256 hash of a string
func HashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

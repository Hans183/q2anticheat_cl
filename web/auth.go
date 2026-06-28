package web

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/user/anticheat_cl/database"
)

// Auth handles authentication
type Auth struct {
	db *database.DB
}

// NewAuth creates a new auth handler
func NewAuth(db *database.DB) *Auth {
	return &Auth{db: db}
}

// HashPassword hashes a password with SHA-256
func HashPassword(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}

// CheckPassword checks if a password matches the hash
func CheckPassword(hash, password string) bool {
	return hash == HashPassword(password)
}

// InitDefaultAdmin creates the default admin user if none exists
func InitDefaultAdmin(db *database.DB, username, password string) error {
	return db.EnsureDefaultAdmin(username, HashPassword(password))
}

// InitAdmins creates multiple admin users from a list, skipping existing ones
func InitAdmins(db *database.DB, admins []database.AdminInput) error {
	return db.EnsureAdmins(admins)
}

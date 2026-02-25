package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const defaultCost = bcrypt.DefaultCost

// Password hashes a plain-text password using bcrypt.
func Password(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}
	return string(hashed), nil
}

// CheckPassword compares a plain-text password against a bcrypt hash.
// Returns nil on match, an error otherwise.
func CheckPassword(plain, hashed string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

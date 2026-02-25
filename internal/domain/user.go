package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the user entity in the domain.
type User struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password_hash"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// RefreshToken represents a refresh token tied to a user and device.
type RefreshToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	DeviceID  string    `json:"device_id" db:"device_id"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RegisterRequest is the payload for registering a new user.
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

// LoginRequest is the payload for user login.
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	DeviceID string `json:"device_id" validate:"required,max=255"`
}

// AuthResponse is returned after a successful authentication.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
}

// RefreshTokenRequest is the payload for refreshing access tokens.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	DeviceID     string `json:"device_id" validate:"required"`
}

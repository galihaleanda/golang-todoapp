package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenType differentiates access and refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims extends standard JWT claims with application-specific fields.
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// Manager handles JWT creation and parsing.
type Manager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// New creates a Manager with the provided secrets and TTL values.
func New(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// GenerateAccessToken creates a signed access JWT for the given user ID.
func (m *Manager) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return m.generate(userID, AccessToken, m.accessSecret, m.accessTTL)
}

// GenerateRefreshToken creates a signed refresh JWT for the given user ID.
func (m *Manager) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return m.generate(userID, RefreshToken, m.refreshSecret, m.refreshTTL)
}

func (m *Manager) generate(userID uuid.UUID, tokenType TokenType, secret []byte, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

// ParseAccessToken validates and parses an access token string.
func (m *Manager) ParseAccessToken(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.accessSecret, AccessToken)
}

// ParseRefreshToken validates and parses a refresh token string.
func (m *Manager) ParseRefreshToken(tokenStr string) (*Claims, error) {
	return m.parse(tokenStr, m.refreshSecret, RefreshToken)
}

func (m *Manager) parse(tokenStr string, secret []byte, expectedType TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s got %s", expectedType, claims.TokenType)
	}

	return claims, nil
}

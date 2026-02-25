package service

import (
	"context"
	"fmt"
	"time"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/galihaleanda/todo-app/pkg/hash"
	pkgjwt "github.com/galihaleanda/todo-app/pkg/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// AuthService handles authentication use cases.
type AuthService struct {
	userRepo         domain.UserRepository
	refreshTokenRepo domain.RefreshTokenRepository
	jwtManager       *pkgjwt.Manager
	log              *logrus.Logger
}

// NewAuthService constructs an AuthService with its dependencies.
func NewAuthService(
	userRepo domain.UserRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	jwtManager *pkgjwt.Manager,
	log *logrus.Logger,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		log:              log,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.AuthResponse, error) {
	// Check uniqueness
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil && err != domain.ErrNotFound {
		return nil, fmt.Errorf("authService.Register FindByEmail: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrAlreadyExists
	}

	passwordHash, err := hash.Password(req.Password)
	if err != nil {
		return nil, fmt.Errorf("authService.Register hash password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		Password:  passwordHash,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("authService.Register create user: %w", err)
	}

	s.log.WithField("user_id", user.ID).Info("new user registered")
	return s.buildAuthResponse(ctx, user, "register-device")
}

// Login authenticates a user and returns tokens.
func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest, userAgent string) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("authService.Login FindByEmail: %w", err)
	}

	if err := hash.CheckPassword(req.Password, user.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.buildAuthResponse(ctx, user, req.DeviceID)
}

// RefreshTokens rotates the refresh token and issues a new access token.
func (s *AuthService) RefreshTokens(ctx context.Context, req *domain.RefreshTokenRequest) (*domain.AuthResponse, error) {
	claims, err := s.jwtManager.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	storedToken, err := s.refreshTokenRepo.FindByToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	if storedToken.ExpiresAt.Before(time.Now()) {
		_ = s.refreshTokenRepo.DeleteByToken(ctx, req.RefreshToken)
		return nil, domain.ErrTokenExpired
	}

	// Rotate — delete old, issue new
	if err := s.refreshTokenRepo.DeleteByToken(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("authService.RefreshTokens delete old: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("authService.RefreshTokens FindByID: %w", err)
	}

	return s.buildAuthResponse(ctx, user, req.DeviceID)
}

// Logout revokes refresh tokens for a specific device or all devices.
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string, allDevices bool) error {
	if allDevices {
		return s.refreshTokenRepo.DeleteByUserID(ctx, userID)
	}
	return s.refreshTokenRepo.DeleteByToken(ctx, refreshToken)
}

// buildAuthResponse generates both tokens, stores the refresh token, and returns the response.
func (s *AuthService) buildAuthResponse(ctx context.Context, user *domain.User, deviceID string) (*domain.AuthResponse, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshTokenStr, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshTokenStr,
		DeviceID:  deviceID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, rt); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		User:         user,
	}, nil
}

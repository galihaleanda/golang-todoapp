package handler

import (
	"errors"

	"github.com/galihaleanda/todo-app/internal/domain"
	"github.com/galihaleanda/todo-app/internal/middleware"
	"github.com/galihaleanda/todo-app/internal/service"
	"github.com/galihaleanda/todo-app/internal/validator"
	"github.com/galihaleanda/todo-app/pkg/response"
	"github.com/gin-gonic/gin"
)

// AuthHandler exposes authentication endpoints.
type AuthHandler struct {
	authSvc *service.AuthService
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body domain.RegisterRequest true "Registration payload"
// @Success 201 {object} response.Envelope{data=domain.AuthResponse}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if errs, err := validator.BindAndValidate(c, &req); err != nil {
		response.InternalError(c)
		return
	} else if errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	authResp, err := h.authSvc.Register(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAlreadyExists):
			response.Conflict(c, "email already registered")
		default:
			response.InternalError(c)
		}
		return
	}

	response.Created(c, authResp)
}

// Login godoc
// @Summary Authenticate a user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body domain.LoginRequest true "Login payload"
// @Success 200 {object} response.Envelope{data=domain.AuthResponse}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if errs, err := validator.BindAndValidate(c, &req); err != nil {
		response.InternalError(c)
		return
	} else if errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	authResp, err := h.authSvc.Login(c.Request.Context(), &req, c.GetHeader("User-Agent"))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			response.Unauthorized(c, "invalid email or password")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, authResp)
}

// RefreshToken godoc
// @Summary Rotate access and refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body domain.RefreshTokenRequest true "Refresh token payload"
// @Success 200 {object} response.Envelope{data=domain.AuthResponse}
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req domain.RefreshTokenRequest
	if errs, err := validator.BindAndValidate(c, &req); err != nil {
		response.InternalError(c)
		return
	} else if errs != nil {
		response.UnprocessableEntity(c, errs)
		return
	}

	authResp, err := h.authSvc.RefreshTokens(c.Request.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTokenInvalid), errors.Is(err, domain.ErrTokenExpired):
			response.Unauthorized(c, "invalid or expired refresh token")
		default:
			response.InternalError(c)
		}
		return
	}

	response.OK(c, authResp)
}

// Logout godoc
// @Summary Revoke tokens
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param all_devices query bool false "Revoke all devices"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.CurrentUserID(c)
	refreshToken := c.GetHeader("X-Refresh-Token")
	allDevices := c.Query("all_devices") == "true"

	if err := h.authSvc.Logout(c.Request.Context(), userID, refreshToken, allDevices); err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{"message": "logged out successfully"})
}

package domain

import "errors"

// Sentinel errors for the domain layer.
// Handlers map these to HTTP status codes.
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrValidation        = errors.New("validation error")
	ErrInternal          = errors.New("internal server error")
)

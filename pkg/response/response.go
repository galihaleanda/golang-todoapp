package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope is the standard API response wrapper.
type Envelope struct {
	Success bool        `json:"success"`
	Data    any         `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorBody carries structured error information.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Meta carries pagination information.
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// OK sends a 200 response with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

// Created sends a 201 response with data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

// OKPaginated sends a 200 response with data and pagination metadata.
func OKPaginated(c *gin.Context, data any, page, limit, total int) {
	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, Envelope{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	})
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, code, msg string, details any) {
	c.JSON(http.StatusBadRequest, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: msg, Details: details},
	})
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: "UNAUTHORIZED", Message: msg},
	})
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: "FORBIDDEN", Message: msg},
	})
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: "NOT_FOUND", Message: msg},
	})
}

// UnprocessableEntity sends a 422 error response (validation errors).
func UnprocessableEntity(c *gin.Context, details any) {
	c.JSON(http.StatusUnprocessableEntity, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: "VALIDATION_ERROR", Message: "request validation failed", Details: details},
	})
}

// InternalError sends a 500 error response.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: "INTERNAL_ERROR", Message: "an internal server error occurred"},
	})
}

// Conflict sends a 409 error response.
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: "CONFLICT", Message: msg},
	})
}

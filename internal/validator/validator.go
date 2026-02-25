package validator

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationError represents a single field validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// BindAndValidate decodes the JSON body into dst and runs struct validation.
// Returns (nil, nil) on success; (nil, errDetails) when there are validation errors.
func BindAndValidate(c *gin.Context, dst any) ([]ValidationError, error) {
	if err := c.ShouldBindJSON(dst); err != nil {
		return []ValidationError{{Field: "body", Message: "invalid JSON: " + err.Error()}}, nil
	}

	if err := validate.Struct(dst); err != nil {
		var errs validator.ValidationErrors
		if ok := isValidationErrors(err, &errs); ok {
			return formatErrors(errs), nil
		}
		return nil, fmt.Errorf("unexpected validation error: %w", err)
	}

	return nil, nil
}

func isValidationErrors(err error, target *validator.ValidationErrors) bool {
	if v, ok := err.(validator.ValidationErrors); ok {
		*target = v
		return true
	}
	return false
}

func formatErrors(errs validator.ValidationErrors) []ValidationError {
	out := make([]ValidationError, 0, len(errs))
	for _, e := range errs {
		out = append(out, ValidationError{
			Field:   strings.ToLower(e.Field()),
			Message: fieldMessage(e),
		})
	}
	return out
}

func fieldMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", e.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	case "hexcolor":
		return "must be a valid hex color (e.g. #3B82F6)"
	default:
		return fmt.Sprintf("failed validation: %s", e.Tag())
	}
}

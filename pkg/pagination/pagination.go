package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

// Params holds parsed pagination query parameters.
type Params struct {
	Page   int
	Limit  int
	Offset int
}

// FromContext parses ?page= and ?limit= from a gin request context.
func FromContext(c *gin.Context) Params {
	page := parseInt(c.Query("page"), DefaultPage)
	limit := parseInt(c.Query("limit"), DefaultLimit)

	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}

	return Params{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

func parseInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

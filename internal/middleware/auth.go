package middleware

import (
	"strings"

	pkgjwt "github.com/galihaleanda/todo-app/pkg/jwt"
	"github.com/galihaleanda/todo-app/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const userIDKey = "user_id"

// Auth is a Gin middleware that validates Bearer access tokens.
func Auth(jwtManager *pkgjwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseAccessToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "invalid or expired access token")
			c.Abort()
			return
		}

		c.Set(userIDKey, claims.UserID)
		c.Next()
	}
}

// CurrentUserID extracts the authenticated user's UUID from the gin context.
// Panics if called outside of an Auth-protected route — by design.
func CurrentUserID(c *gin.Context) uuid.UUID {
	return c.MustGet(userIDKey).(uuid.UUID)
}

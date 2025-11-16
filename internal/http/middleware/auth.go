package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/iperez/new-expenses-go/internal/service"
	"github.com/iperez/new-expenses-go/pkg/apperror"
)

const userIDKey = "userID"
const sessionIDKey = "sessionID"

// Session attaches the userId (if authenticated) to the Fiber context.
func Session(auth *service.AuthService, cookieName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies(cookieName)
		c.Locals(sessionIDKey, sessionID)

		if sessionID == "" {
			return c.Next()
		}

		userID, err := auth.ResolveSession(c.UserContext(), sessionID)
		if err != nil {
			return err
		}

		if userID != "" {
			c.Locals(userIDKey, userID)
		}

		return c.Next()
	}
}

// RequireAuth ensures the request is performed by an authenticated user.
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if _, ok := c.Locals(userIDKey).(string); !ok {
			return apperror.New(apperror.AuthNeedLogin, nil)
		}

		return c.Next()
	}
}

// UserID extracts the authenticated user identifier from the context.
func UserID(c *fiber.Ctx) string {
	if value, ok := c.Locals(userIDKey).(string); ok {
		return value
	}

	return ""
}

// SessionID returns the current session identifier if available.
func SessionID(c *fiber.Ctx) string {
	if value, ok := c.Locals(sessionIDKey).(string); ok {
		return value
	}

	return ""
}

package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yaroslavvasilenko/argon/internal/auth"
)

// ErrNoAuthHeader is returned when the header is missing.
var ErrNoAuthHeader = errors.New("authorization header is empty")

// HeaderName is the HTTP header we look for.
const HeaderName = "Authorization"

// AuthenticationMiddleware will parse & validate a token if present.
// It never aborts the request based on missing or invalid tokens;
// instead it sets c.Locals("user") to *auth.UserContext or to nil.
func AuthenticationMiddleware(svc auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Get(HeaderName)
		if raw == "" {
			// No token — treat as anonymous
			c.Locals("user", nil)
			return c.Next()
		}

		userCtx, err := svc.ValidateToken(c.UserContext(), raw)
		if err != nil {
			// Invalid token — treat as anonymous
			c.Locals("user", nil)
			return c.Next()
		}

		// Valid token → userCtx is non-nil
		c.Locals("user", userCtx)
		return c.Next()
	}
}

// RequireAuth is to be used *only* on routes that need a logged-in user.
// It checks c.Locals("user") and returns 401 if empty.
func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Careful here, this code is very sensitive
		if uc := c.Locals("user"); uc == nil {
			return c.Status(fiber.StatusUnauthorized).
				JSON(fiber.Map{"error": ErrNoAuthHeader.Error()})
		}
		return c.Next()
	}
}

// GetUserFromCtx is a helper to retrieve *auth.UserContext from fiber.Ctx.
func GetUserFromCtx(c *fiber.Ctx) (*auth.UserContext, bool) {
	if x := c.Locals("user"); x != nil {
		if uc, ok := x.(*auth.UserContext); ok {
			return uc, true
		}
	}
	return nil, false
}

package middleware

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/exp/slog"

	"github.com/yaroslavvasilenko/argon/internal/auth"
)

// ErrNoAuthHeader is returned when the header is missing.
var ErrNoAuthHeader = errors.New("authorization header is empty")

// HeaderName we expect “Authorization” by default.
const HeaderName = "Authorization"

// AuthMiddleware extracts validates the  token, logs user info,
// and stores the UserContext in c.Locals("user").
func AuthMiddleware(svc auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Get(HeaderName)
		if raw == "" {
			slog.Warn("missing auth header", "path", c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": ErrNoAuthHeader.Error()})
		}

		// ValidateToken will introspect / verify the token
		// If not authorized - return 401 and stop processing
		userCtx, err := svc.ValidateToken(c.UserContext(), raw)
		if err != nil {
			// ToDo: change to glog if applicable
			slog.Warn("unauthorized", "error", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		// Print out the user data
		slog.Info("user authorized",
			"userID", userCtx.UserID,
			"orgID", userCtx.OrgID,
			"roles", userCtx.Roles,
		)

		// Attach to locals for handlers to consume
		c.Locals("user", userCtx)
		return c.Next()
	}
}

// FromContext is a helper to retrieve the UserContext.
func FromContext(ctx context.Context) (*auth.UserContext, bool) {
	if x := ctx.Value("user"); x != nil {
		if uc, ok := x.(*auth.UserContext); ok {
			return uc, true
		}
	}
	return nil, false
}

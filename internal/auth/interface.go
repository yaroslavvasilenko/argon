package auth

import "context"

type Service interface {
	// ValidateToken verifies the token, returns a populated UserContext or an error.
	ValidateToken(ctx context.Context, token string) (*UserContext, error)
}

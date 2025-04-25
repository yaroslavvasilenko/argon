package auth

import "context"

//go:generate mockery init auth
type Service interface {
	// ValidateToken verifies the token, returns a populated UserContext or an error.
	ValidateToken(ctx context.Context, token string) (*UserContext, error)
}

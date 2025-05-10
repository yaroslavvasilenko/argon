package auth

import (
	"context"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization"
	"github.com/zitadel/zitadel-go/v3/pkg/authorization/oauth"
	"github.com/zitadel/zitadel-go/v3/pkg/zitadel"
)

// UserContext holds the information we need after verifying a token.
type UserContext struct {
	UserID string   // the 'sub' claim
	OrgID  string   // the organization ID (resource owner)
	Roles  []string // roles granted in the project/org
	Token  string   // the raw bearer token
}

type ZitadelAuthService struct {
	authorizer *authorization.Authorizer[*oauth.IntrospectionContext]
}

// NewZitadelAuthService wires up the ZITADEL SDK.
//   - domain: ZITADEL instance, e.g. "foo.zitadel.cloud"
//   - keyPath: path to the downloaded service account key.json ToDo: decide how to replace this method of authz
func NewZitadelAuthService(ctx context.Context, domain, keyPath string) (Service, error) {
	// create a ZITADEL client
	z := zitadel.New(domain)
	// init the Authorizer with OAuth2 introspection using JWT-profile key.json
	authZ, err := authorization.New(ctx, z, oauth.DefaultAuthorization(keyPath))
	if err != nil {
		return nil, err
	}
	return &ZitadelAuthService{authorizer: authZ}, nil
}

// ValidateToken checks the token and builds a UserContext.
func (s *ZitadelAuthService) ValidateToken(ctx context.Context, token string) (*UserContext, error) {
	// this will call the introspection endpoint, and verify active && roles, etc.
	authCtx, err := s.authorizer.CheckAuthorization(ctx, token)
	if err != nil {
		return nil, err
	}

	// build our user-facing context
	uc := &UserContext{
		UserID: authCtx.UserID(),
		OrgID:  authCtx.OrganizationID(),
		Token:  token,
	}
	return uc, nil
}

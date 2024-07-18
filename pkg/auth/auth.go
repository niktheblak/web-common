package auth

import (
	"context"
	"errors"
)

var ErrNotAuthorized = errors.New("not authorized")

type Authenticator interface {
	// Authenticate validates the given access token.
	// The implementation should return an error if the token is not valid for any reason.
	Authenticate(ctx context.Context, token string) error
}

// AlwaysAllowAuthenticator accepts any access token.
type AlwaysAllowAuthenticator struct {
}

func (a *AlwaysAllowAuthenticator) Authenticate(ctx context.Context, token string) error {
	return nil
}

func AlwaysAllow() *AlwaysAllowAuthenticator {
	return &AlwaysAllowAuthenticator{}
}

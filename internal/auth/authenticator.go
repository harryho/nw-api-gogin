package auth

import (
	"context"
	"errors"
)

type Authenticator interface {
	Authenticate(ctx context.Context, username, password string) (Principal, error)
}

type StaticAuthenticator struct {
	credentials map[string]credential
}

type credential struct {
	Password string
	Principal
}

func NewStaticAuthenticator(users map[string]struct {
	Password  string
	Principal Principal
}) (*StaticAuthenticator, error) {
	if len(users) == 0 {
		return nil, errors.New("users must not be empty")
	}
	creds := make(map[string]credential, len(users))
	for username, data := range users {
		creds[username] = credential{
			Password: data.Password,
			Principal: Principal{
				Subject: data.Principal.Subject,
				Scopes:  SanitizeScopes(data.Principal.Scopes),
			},
		}
	}
	return &StaticAuthenticator{credentials: creds}, nil
}

func (a *StaticAuthenticator) Authenticate(ctx context.Context, username, password string) (Principal, error) {
	_ = ctx
	cred, ok := a.credentials[username]
	if !ok {
		return Principal{}, NewError(ErrorInvalidCredentials, "invalid username or password", nil)
	}
	if cred.Password != password {
		return Principal{}, NewError(ErrorInvalidCredentials, "invalid username or password", nil)
	}
	return cred.Principal, nil
}

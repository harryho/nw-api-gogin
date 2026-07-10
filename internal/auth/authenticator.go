package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type Authenticator interface {
	Authenticate(ctx context.Context, username, password string) (Principal, error)
}

type StaticAuthenticator struct {
	credentials map[string]credential
}

type credential struct {
	PasswordHash []byte
	Principal
}

// MustHashPassword returns the bcrypt hash for a plaintext password using
// bcrypt's default cost. It panics on error and is intended for tests
// and startup paths only.
func MustHashPassword(plaintext string) []byte {
	h, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return h
}

func NewStaticAuthenticator(users map[string]struct {
	PasswordHash []byte
	Principal    Principal
}) (*StaticAuthenticator, error) {
	if len(users) == 0 {
		return nil, errors.New("users must not be empty")
	}
	creds := make(map[string]credential, len(users))
	for username, data := range users {
		creds[username] = credential{
			PasswordHash: data.PasswordHash,
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
	if err := bcrypt.CompareHashAndPassword(cred.PasswordHash, []byte(password)); err != nil {
		return Principal{}, NewError(ErrorInvalidCredentials, "invalid username or password", nil)
	}
	return cred.Principal, nil
}

package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Service interface {
	IssueToken(ctx context.Context, input TokenIssueRequest) (Token, error)
	ValidateToken(ctx context.Context, token string) (*Claims, error)
}

type Config struct {
	Issuer         string
	Audience       []string
	AccessTokenTTL time.Duration
}

type service struct {
	cfg           Config
	authenticator Authenticator
	keyManager    KeyManager
	now           func() time.Time
}

func NewService(cfg Config, authenticator Authenticator, keyManager KeyManager) (Service, error) {
	if authenticator == nil {
		return nil, errors.New("authenticator must not be nil")
	}
	if keyManager == nil {
		return nil, errors.New("key manager must not be nil")
	}
	if cfg.AccessTokenTTL <= 0 {
		return nil, errors.New("access token ttl must be greater than zero")
	}
	return &service{
		cfg:           cfg,
		authenticator: authenticator,
		keyManager:    keyManager,
		now:           time.Now,
	}, nil
}

func (s *service) IssueToken(ctx context.Context, input TokenIssueRequest) (Token, error) {
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	if username == "" || password == "" {
		return Token{}, NewError(ErrorInvalidCredentials, "invalid username or password", nil)
	}
	principal, err := s.authenticator.Authenticate(ctx, username, password)
	if err != nil {
		if appErr, ok := AsError(err); ok {
			return Token{}, appErr
		}
		return Token{}, NewError(ErrorInvalidCredentials, "invalid username or password", err)
	}
	allowedScopes := SanitizeScopes(principal.Scopes)
	reqScopes := SanitizeScopes(input.Scopes)
	if len(reqScopes) == 0 {
		reqScopes = allowedScopes
	}
	if !HasAllScopes(reqScopes, allowedScopes) {
		return Token{}, NewError(ErrorInvalidScope, "requested scopes are not permitted", nil)
	}
	key, err := s.keyManager.Current(ctx)
	if err != nil {
		return Token{}, NewError(ErrorInternal, "failed to load signing key", err)
	}
	now := s.now()
	expiresAt := now.Add(s.cfg.AccessTokenTTL)
	claims := Claims{
		Scopes: reqScopes,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.Issuer,
			Subject:   principal.Subject,
			Audience:  jwt.ClaimStrings(s.cfg.Audience),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			ID:        uuid.NewString(),
		},
	}
	signingMethod := jwt.GetSigningMethod(key.Algorithm)
	if signingMethod == nil {
		return Token{}, NewError(ErrorInternal, fmt.Sprintf("unsupported signing algorithm %q", key.Algorithm), nil)
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	if key.ID != "" {
		token.Header["kid"] = key.ID
	}
	signed, err := token.SignedString(key.Secret)
	if err != nil {
		return Token{}, NewError(ErrorInternal, "failed to sign token", err)
	}
	return Token{Value: signed, ExpiresAt: expiresAt, Subject: principal.Subject, Scopes: claims.Scopes}, nil
}

func (s *service) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, NewError(ErrorInvalidToken, "token is required", nil)
	}
	claims := &Claims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		kid, _ := token.Header["kid"].(string)
		key, keyErr := s.getKeyForToken(ctx, kid)
		if keyErr != nil {
			return nil, keyErr
		}
		if token.Method.Alg() != key.Algorithm {
			return nil, fmt.Errorf("unexpected signing algorithm: %s", token.Method.Alg())
		}
		return key.Secret, nil
	}, jwt.WithTimeFunc(s.now))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, NewError(ErrorInvalidToken, "token expired", err)
		}
		return nil, NewError(ErrorInvalidToken, "invalid token", err)
	}
	if !parsedToken.Valid {
		return nil, NewError(ErrorInvalidToken, "invalid token", nil)
	}
	if len(s.cfg.Issuer) > 0 && claims.Issuer != s.cfg.Issuer {
		return nil, NewError(ErrorInvalidToken, "invalid issuer", nil)
	}
	if len(s.cfg.Audience) > 0 {
		if !s.verifyAudience(claims) {
			return nil, NewError(ErrorInvalidToken, "invalid audience", nil)
		}
	}
	return claims, nil
}

func (s *service) getKeyForToken(ctx context.Context, kid string) (SigningKey, error) {
	if kid != "" {
		key, err := s.keyManager.Get(ctx, kid)
		if err == nil {
			return key, nil
		}
	}
	return s.keyManager.Current(ctx)
}

func (s *service) verifyAudience(claims *Claims) bool {
	for _, expected := range s.cfg.Audience {
		for _, actual := range []string(claims.Audience) {
			if actual == expected {
				return true
			}
		}
	}
	return false
}

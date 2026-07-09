package api

import (
	"context"

	"github.com/golang-jwt/jwt/v5"

	"github.com/harryho/nw-api-gogin/internal/auth"
)

type tokenServiceStub struct {
	issueTokenFn    func(ctx context.Context, input auth.TokenIssueRequest) (auth.Token, error)
	validateTokenFn func(ctx context.Context, token string) (*auth.Claims, error)
}

func (s *tokenServiceStub) IssueToken(ctx context.Context, input auth.TokenIssueRequest) (auth.Token, error) {
	if s != nil && s.issueTokenFn != nil {
		return s.issueTokenFn(ctx, input)
	}
	return auth.Token{}, nil
}

func (s *tokenServiceStub) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	if s != nil && s.validateTokenFn != nil {
		return s.validateTokenFn(ctx, token)
	}
	return &auth.Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: token}}, nil
}

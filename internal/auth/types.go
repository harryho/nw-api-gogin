package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenIssueRequest struct {
	Username string
	Password string
	Scopes   []string
}

type Principal struct {
	Subject string
	Scopes  []string
}

type Token struct {
	Value     string
	ExpiresAt time.Time
	Subject   string
	Scopes    []string
}

type SigningKey struct {
	ID        string
	Secret    []byte
	Algorithm string
}

type Claims struct {
	Scopes []string `json:"scp,omitempty"`
	jwt.RegisteredClaims
}

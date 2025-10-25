package auth

import (
	"context"
	"errors"
	"fmt"
)

type KeyManager interface {
	Current(ctx context.Context) (SigningKey, error)
	Get(ctx context.Context, id string) (SigningKey, error)
}

type HMACKeyManager struct {
	key SigningKey
}

func NewHMACKeyManager(secret []byte, keyID string) (*HMACKeyManager, error) {
	if len(secret) == 0 {
		return nil, errors.New("secret must not be empty")
	}
	return &HMACKeyManager{
		key: SigningKey{ID: keyID, Secret: secret, Algorithm: "HS256"},
	}, nil
}

func (m *HMACKeyManager) Current(context.Context) (SigningKey, error) {
	return m.key, nil
}

func (m *HMACKeyManager) Get(ctx context.Context, id string) (SigningKey, error) {
	if id == "" || id == m.key.ID {
		return m.key, nil
	}
	return SigningKey{}, fmt.Errorf("unknown key id: %s", id)
}

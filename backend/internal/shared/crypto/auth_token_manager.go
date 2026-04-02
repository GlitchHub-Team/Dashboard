package crypto

import (
	"errors"

	"backend/internal/shared/identity"
)

//go:generate mockgen -destination=../../../tests/shared/crypto/mocks/auth_token.go -package=mocks . AuthTokenManager

type AuthTokenManager interface {
	GenerateForRequester(requester identity.Requester) (string, error)
	GetRequesterFromToken(token string) (identity.Requester, error)
}

var ErrInvalidAuthToken = errors.New("invalid or expired authentication token")

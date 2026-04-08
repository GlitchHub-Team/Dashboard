package crypto

import (
	"time"
)

//go:generate mockgen -destination=../../../tests/shared/crypto/mocks/security_token.go -package=mocks . SecurityTokenGenerator

type SecurityTokenGenerator interface {
	/* Genera un token e la sua versione hashed */
	GenerateToken() (encodedToken string, hashedToken string, err error)

	/* Ottieni la data di scadenza a partire da adesso */
	ExpiryFromNow() time.Time
}

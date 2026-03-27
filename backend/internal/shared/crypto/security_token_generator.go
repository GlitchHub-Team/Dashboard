package crypto

import (
	"time"
)

type SecurityTokenGenerator interface {
	/* Genera un token e la sua versione hashed */
	GenerateToken() (encodedToken string, hashedToken string, err error)

	/* Ottieni la data di scadenza a partire da adesso */
	ExpiryFromNow() time.Time
}

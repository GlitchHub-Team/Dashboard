package auth

import (
	"time"

	"github.com/google/uuid"
)

/*
Token per il cambio di password dimenticata.

Possibile miglioria: dividere tra token associato a tenant member e token associato
a super admin. Al momento è tutto unito per evitare troppa duplicazione.

NOTA: ForgotPasswordToken e ConfirmAccountToken sono spezzati in caso le due struct
debbano essere modificate in modo incompatibile l'una con l'altra.
*/
type ForgotPasswordToken struct {
	// id          int
	HashedToken string
	TenantId    *uuid.UUID
	UserId      uint
	ExpiryDate  time.Time
}

func (token *ForgotPasswordToken) IsExpired() bool {
	return token.ExpiryDate.Before(time.Now())
}

/*
Token per la conferma di un account appena creato.

Possibile miglioria: dividere tra token associato a tenant member e token associato
a super admin. Al momento è tutto unito per evitare troppa duplicazione.

NOTA: ForgotPasswordToken e ConfirmAccountToken sono spezzati in caso le due struct
debbano essere modificate in modo incompatibile l'una con l'altra.
*/
type ConfirmAccountToken struct {
	// id          int
	HashedToken string
	TenantId    *uuid.UUID
	UserId      uint
	ExpiryDate  time.Time
}

func (token *ConfirmAccountToken) IsExpired() bool {
	return token.ExpiryDate.Before(time.Now())
}

package auth

import (
	"time"

	"github.com/google/uuid"
	// "backend/internal/shared/identity"
	// "github.com/google/uuid"
)

type ForgotPasswordToken struct {
	id          int
	hashedToken string
	tenantId    *uuid.UUID
	userId      uint
	expiryDate  time.Time
}

func (token *ForgotPasswordToken) IsExpired() bool {
	return token.expiryDate.Before(time.Now())
}

type ConfirmAccountToken struct {
	id          int
	hashedToken string
	tenantId    *uuid.UUID
	userId      uint
	expiryDate  time.Time
}

func (token *ConfirmAccountToken) IsExpired() bool {
	return token.expiryDate.Before(time.Now())
}

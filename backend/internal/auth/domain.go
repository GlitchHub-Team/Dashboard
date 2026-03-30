package auth

import (
	"time"

	"github.com/google/uuid"
	// "backend/internal/shared/identity"
	// "github.com/google/uuid"
)

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

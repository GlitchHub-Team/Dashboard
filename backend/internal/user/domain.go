package user

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

/*
Possibile miglioria: spezzare classe di dominio in TenantMember e SuperAdmin. Questo
richiede il refactor dell'intero package
*/
type User struct {
	Id           uint
	Name         string
	Email        string
	PasswordHash *string
	Role         identity.UserRole
	TenantId     *uuid.UUID
	Confirmed    bool
}

func (u *User) IsZero() bool {
	return *u == (User{})
}

func (u *User) SetPasswordHash(newPasswordHash string) error {
	if newPasswordHash == "" {
		return ErrEmptyPassword
	}

	if newPasswordHash == *u.PasswordHash {
		return ErrSamePassword
	}

	u.PasswordHash = &newPasswordHash
	return nil
}

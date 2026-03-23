package user

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

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

package user

import (
	"backend/internal/identity"
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

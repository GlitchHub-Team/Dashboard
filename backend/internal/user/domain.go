package user

import (
	"github.com/google/uuid"
	"backend/internal/identity"
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

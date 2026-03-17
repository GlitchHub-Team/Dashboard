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

func (u *User) SetName(name string) {
	u.Name = name
}

func (u *User) SetEmail(email string) {
	u.Email = email
}

func (u *User) SetPasswordHash(passwordHash string) {
	u.PasswordHash = &passwordHash
}

func (u *User) Confirm() {
	u.Confirmed = true
}

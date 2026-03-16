package user

import "github.com/google/uuid"

type UserRole string

const (
	ROLE_TENANT_USER  UserRole = "tenant_user"
	ROLE_TENANT_ADMIN UserRole = "tenant_admin"
	ROLE_SUPER_ADMIN  UserRole = "super_admin"
)

type User struct {
	Id           uint
	Name         string
	Email        string
	PasswordHash *string
	Role         UserRole
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

package user

import "github.com/google/uuid"

type UserRole string
const (
	ROLE_TENANT_USER  UserRole = "tenant_user"
	ROLE_TENANT_ADMIN UserRole = "tenant_admin"
	ROLE_SUPER_ADMIN  UserRole = "super_admin"
)

type User struct {
	id           int
	name         string
	email        string
	passwordHash *string
	role         UserRole
	tenantId     *uuid.UUID
	confirmed    bool
}

func (u *User) SetName(name string) {
	u.name = name
}

func (u *User) SetEmail(email string) {
	u.email = email
}

func (u *User) SetPasswordHash(passwordHash string) {
	u.passwordHash = &passwordHash
}

func (u *User) Confirm() {
	u.confirmed = true
}


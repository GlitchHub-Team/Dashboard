package user

import (
	"github.com/google/uuid"
)

type CreateTenantUserCommand struct {
	Email    string
	Username string
	TenantId uuid.UUID
}

type CreateTenantAdminCommand struct {
	Email    string
	Username string
	TenantId uuid.UUID
}

type CreateSuperAdminCommand struct {
	Email    string
	Username string
}

type DeleteUserCommand struct {
	UserId int
}

type GetUserByIdCommand struct {
	UserId int
}

type GetUsersCommand struct {
	Page  int
	Limit int
	Role  UserRole
}

type GetUsersByTenantIdCommand struct {
	Page     int
	Limit    int
	TenantId uuid.UUID
}

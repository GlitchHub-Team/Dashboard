package user

import (
	"github.com/google/uuid"
)

// Create =============================================================================================
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

// Delete =============================================================================================
type DeleteTenantUserCommand struct {
	TenantId uuid.UUID
	UserId uint
}
type DeleteTenantAdminCommand struct {
	TenantId uuid.UUID
	UserId uint
}
type DeleteSuperAdminCommand struct {
	UserId uint
}

// Get =============================================================================================
type GetTenantUserCommand struct {
	TenantId uuid.UUID
	UserId   uint
}

type GetTenantAdminCommand struct {
	TenantId uuid.UUID
	UserId   uint
}

type GetSuperAdminCommand struct {
	UserId uint
}

type GetUserByIdCommand struct {
	UserId uint
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

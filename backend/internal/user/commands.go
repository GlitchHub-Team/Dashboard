package user

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

// Create =============================================================================================
type CreateTenantUserCommand struct {
	identity.Requester
	Email    string
	Username string
	TenantId uuid.UUID
}

type CreateTenantAdminCommand struct {
	identity.Requester
	Email    string
	Username string
	TenantId uuid.UUID
}

type CreateSuperAdminCommand struct {
	identity.Requester
	Email    string
	Username string
}

// Delete =============================================================================================
type DeleteTenantUserCommand struct {
	identity.Requester
	TenantId uuid.UUID
	UserId   uint
}
type DeleteTenantAdminCommand struct {
	identity.Requester
	TenantId uuid.UUID
	UserId   uint
}
type DeleteSuperAdminCommand struct {
	identity.Requester
	UserId uint
}

// Get =============================================================================================
type GetTenantUserCommand struct {
	identity.Requester
	TenantId uuid.UUID
	UserId   uint
}

type GetTenantAdminCommand struct {
	identity.Requester
	TenantId uuid.UUID
	UserId   uint
}

type GetSuperAdminCommand struct {
	identity.Requester
	UserId uint
}

// Get multiple ---------------------------------------------------------------------------------------

type GetTenantUsersByTenantCommand struct {
	identity.Requester
	Page     int
	Limit    int
	TenantId uuid.UUID
}

type GetTenantAdminsByTenantCommand struct {
	identity.Requester
	Page     int
	Limit    int
	TenantId uuid.UUID
}

type GetSuperAdminListCommand struct {
	identity.Requester
	Page  int
	Limit int
}

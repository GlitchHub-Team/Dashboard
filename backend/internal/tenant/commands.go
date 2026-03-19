package tenant

import (
	"github.com/google/uuid"
	"backend/internal/identity"
)

type CreateTenantCommand struct {
	Name           string
	CanImpersonate bool
	identity.Requester
}

type DeleteTenantCommand struct {
	TenantId uuid.UUID
	identity.Requester
}

type GetTenantCommand struct {
	TenantId uuid.UUID
	identity.Requester
}

type GetTenantListCommand struct {
	Limit int
	Page  int
	identity.Requester
}

type GetTenantByUserCommand struct {
	UserId uuid.UUID
	identity.Requester
}

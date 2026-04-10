package tenant

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

type CreateTenantCommand struct {
	identity.Requester
	Name           string
	CanImpersonate bool
}

type DeleteTenantCommand struct {
	identity.Requester
	TenantId uuid.UUID
}

type GetTenantCommand struct {
	identity.Requester
	TenantId uuid.UUID
}

type GetTenantListCommand struct {
	Limit int
	Page  int
	// identity.Requester
}

type GetTenantByIdCommand struct {
	TenantId uuid.UUID
	identity.Requester
}

package identity

import (
	"backend/internal/tenant"

	"github.com/google/uuid"
)

type Requester struct {
	RequesterUserId   uint
	RequesterTenantId *uuid.UUID
	RequesterRole     UserRole
}

func (r *Requester) CanTenantUserAccess(accessedTenantId uuid.UUID) bool {
	return r.RequesterTenantId != nil &&
		r.RequesterRole == ROLE_TENANT_USER &&
		*r.RequesterTenantId == accessedTenantId
}

func (r *Requester) CanTenantAdminAccess(accessedTenantId uuid.UUID) bool {
	return r.RequesterTenantId != nil &&
		r.RequesterRole == ROLE_TENANT_ADMIN &&
		*r.RequesterTenantId == accessedTenantId
}

func (r *Requester) IsSuperAdmin() bool {
	return r.RequesterRole == ROLE_SUPER_ADMIN
}

func (r *Requester) CanSuperAdminAccess(tenant tenant.Tenant) bool {
	return r.IsSuperAdmin() && tenant.CanImpersonate
}

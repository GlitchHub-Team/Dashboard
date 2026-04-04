package helper

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

func NewTenantUserJWT(deps IntegrationTestDeps, tenantId uuid.UUID, userId uint, ) (string, error) {
	return deps.AuthTokenManager.GenerateForRequester(identity.Requester{
		RequesterUserId: userId,
		RequesterTenantId: &tenantId,
		RequesterRole: identity.ROLE_TENANT_USER,
	})
}

func NewTenantAdminJWT(deps IntegrationTestDeps, tenantId uuid.UUID, userId uint, ) (string, error) {
	return deps.AuthTokenManager.GenerateForRequester(identity.Requester{
		RequesterUserId: userId,
		RequesterTenantId: &tenantId,
		RequesterRole: identity.ROLE_TENANT_ADMIN,
	})
}


func NewSuperAdminJWT(deps IntegrationTestDeps, userId uint,) (string, error) {
	return deps.AuthTokenManager.GenerateForRequester(identity.Requester{
		RequesterUserId: userId,
		RequesterTenantId: nil,
		RequesterRole: identity.ROLE_SUPER_ADMIN,
	})
}

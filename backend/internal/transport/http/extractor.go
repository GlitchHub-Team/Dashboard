package http

import (
	"errors"

	"backend/internal/identity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ErrMissingIdentity = errors.New("identity missing from request context")

// TODO: inserisci questo in tutti i controller!
func ExtractRequester(ctx *gin.Context) (identity.Requester, error) {
	userId, exists := ctx.Get("requester_user_id")
	requesterUserId, ok := userId.(uint)
	if !exists || !ok {
		return identity.Requester{}, ErrMissingIdentity
	}
	

	roleString, exists := ctx.Get("requester_role")
	if !exists {
		return identity.Requester{}, ErrMissingIdentity
	}

	requesterRole := identity.UserRole(roleString.(string))

	var tenantIdPtr *uuid.UUID

	tenantIdString, exists := ctx.Get("requester_tenant_id")
	if requesterRole == identity.ROLE_SUPER_ADMIN {
		tenantIdPtr = nil
	} else if exists {
		tenantIdVal, err := uuid.Parse(tenantIdString.(string))
		tenantIdPtr = &tenantIdVal
		if err != nil {
			return identity.Requester{}, err
		}
	} else {
		return identity.Requester{}, ErrMissingIdentity
	}

	return identity.Requester{
		RequesterUserId:   requesterUserId,
		RequesterTenantId: tenantIdPtr,
		RequesterRole:     requesterRole,
	}, nil
}

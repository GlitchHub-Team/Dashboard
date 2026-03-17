package http

import (
	"backend/internal/identity"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var ErrMissingIdentity = errors.New("identity missing from request context")

// TODO: inserisci questo in tutti i controller!  
func ExtractRequester(ctx *gin.Context) (identity.Requester, error) {
	userId, exists := ctx.Get("requester_user_id")
	if !exists {
		return identity.Requester{}, ErrMissingIdentity
	}

	tenantIdString, exists := ctx.Get("requester_tenant_id")
	if !exists {
		return identity.Requester{}, ErrMissingIdentity
	}

	roleString, exists := ctx.Get("requester_role")
	if !exists {
		return identity.Requester{}, ErrMissingIdentity
	}

	tenantId, err := uuid.Parse(tenantIdString.(string))
	if err != nil {
		return identity.Requester{}, err
	}

	return identity.Requester{
		RequesterUserId:   userId.(uint),
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.UserRole(roleString.(string)),
	}, nil
}
package tenant

import (
	"errors"
	"fmt"
	"net/http"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Controller struct {
	log                  *zap.Logger
	createTenantUseCase  CreateTenantUseCase
	deleteTenantUseCase  DeleteTenantUseCase
	getTenantUseCase     GetTenantUseCase
	getTenantListUseCase GetTenantListUseCase
	getTenantByIdUseCase GetTenantByIdUseCase
}

func NewTenantController(
	log *zap.Logger,
	createTenantUseCase CreateTenantUseCase,
	deleteTenantUseCase DeleteTenantUseCase,
	getTenantUseCase GetTenantUseCase,
	getTenantListUseCase GetTenantListUseCase,
	getTenantByIdUseCase GetTenantByIdUseCase,
) *Controller {
	return &Controller{
		log:                  log,
		createTenantUseCase:  createTenantUseCase,
		deleteTenantUseCase:  deleteTenantUseCase,
		getTenantUseCase:     getTenantUseCase,
		getTenantListUseCase: getTenantListUseCase,
		getTenantByIdUseCase: getTenantByIdUseCase,
	}
}

// CREATE TENANT ======================================================================================
func (controller *Controller) CreateTenant(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto CreateTenantDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	cmd := CreateTenantCommand{
		Requester:      requester,
		Name:           bodyDto.TenantName,
		CanImpersonate: bodyDto.CanImpersonate,
	}

	createdTenant, err := controller.createTenantUseCase.CreateTenant(cmd)
	if err != nil {

		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrTenantAlreadyExists) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewTenantResponseDTO(createdTenant)
	ctx.JSON(http.StatusOK, responseDto)
}

// DELETE TENANT ======================================================================================
func (controller *Controller) DeleteTenant(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}
	var bodyDto DeleteTenantDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	tenantIdParam := ctx.Param("tenant_Id")
	tenantId, err := uuid.Parse(tenantIdParam)
	
	if err != nil {
		transportHttp.RequestError(ctx, fmt.Errorf("invalid tenant ID"))
		return
	}

	cmd := DeleteTenantCommand{
		Requester: requester,
		TenantId:  tenantId,
	}

	oldTenant, err := controller.deleteTenantUseCase.DeleteTenant(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrTenantNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewTenantResponseDTO(oldTenant)
	ctx.JSON(http.StatusOK, responseDto)
}

// GET TENANT =========================================================================================
func (controller *Controller) GetTenant(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto GetTenantDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	tenantId, _ := uuid.Parse(bodyDto.TenantId)

	cmd := GetTenantCommand{
		Requester: requester,
		TenantId:  tenantId,
	}

	tenant, err := controller.getTenantUseCase.GetTenant(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrTenantNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewTenantResponseDTO(tenant)
	ctx.JSON(http.StatusOK, responseDto)
}

// GET TENANTS ========================================================================================
func (controller *Controller) GetTenants(ctx *gin.Context) {
	// requester, err := transportHttp.ExtractRequester(ctx)
	// if err != nil {
	// 	transportHttp.RequestUnauthorized(ctx, err)
	// 	return
	// }

	var queryDto GetTenantListDTO
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	cmd := GetTenantListCommand{
		// Requester: requester,
		Page:  queryDto.Page,
		Limit: queryDto.Limit,
	}

	tenants, total, err := controller.getTenantListUseCase.GetTenantList(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDtos := NewTenantListResponseDTO(tenants, total)

	ctx.JSON(http.StatusOK, responseDtos)
}

// GET TENANT BY ID =================================================================================
// non usare :)
func (controller *Controller) GetTenantById(ctx *gin.Context) {
	var cmd GetTenantByIdCommand

	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		controller.log.Error("Error binding JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	tenant, err := controller.getTenantByIdUseCase.GetTenantById(cmd)
	if err != nil {
		controller.log.Error("Error getting tenant by user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("failed to get tenant by user"))
		return
	}

	ctx.JSON(http.StatusOK, tenant)
}

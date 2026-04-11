package tenant

import (
	"errors"
	"net/http"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/tenant/mocks/ports.go -package=mocks . CreateTenantUseCase,DeleteTenantUseCase,GetTenantUseCase,GetTenantListUseCase,GetAllTenantsUseCase

type CreateTenantUseCase interface {
	CreateTenant(cmd CreateTenantCommand) (Tenant, error)
}

type DeleteTenantUseCase interface {
	DeleteTenant(cmd DeleteTenantCommand) (Tenant, error)
}

type GetTenantUseCase interface {
	GetTenant(cmd GetTenantCommand) (Tenant, error)
}

type GetAllTenantsUseCase interface {
	GetAllTenants() ([]Tenant, error)
}

type GetTenantListUseCase interface {
	GetTenantList(cmd GetTenantListCommand) ([]Tenant, uint, error)
}

type Controller struct {
	log *zap.Logger

	createTenantUseCase  CreateTenantUseCase
	deleteTenantUseCase  DeleteTenantUseCase
	getTenantUseCase     GetTenantUseCase
	getTenantListUseCase GetTenantListUseCase
	getAllTenantsUseCase GetAllTenantsUseCase
}

func NewTenantController(
	log *zap.Logger,
	createTenantUseCase CreateTenantUseCase,
	deleteTenantUseCase DeleteTenantUseCase,
	getTenantUseCase GetTenantUseCase,
	getTenantListUseCase GetTenantListUseCase,
	getAllTenantsUseCase GetAllTenantsUseCase,
) *Controller {
	return &Controller{
		log:                  log,
		createTenantUseCase:  createTenantUseCase,
		deleteTenantUseCase:  deleteTenantUseCase,
		getTenantUseCase:     getTenantUseCase,
		getTenantListUseCase: getTenantListUseCase,
		getAllTenantsUseCase: getAllTenantsUseCase,
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
	if err := ctx.ShouldBindUri(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	tenantId, _ := uuid.Parse(bodyDto.TenantId)

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
			transportHttp.RequestNotFound(ctx, err)
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
	if err := ctx.ShouldBindUri(&bodyDto); err != nil {
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
			transportHttp.RequestNotFound(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewTenantResponseDTO(tenant)
	ctx.JSON(http.StatusOK, responseDto)
}

// GET ALL TENANTS ========================================================================================

/*
NOTA: Questo è un endpoint che ha senso solo ai fini del prototipo, in quanto è utilizzato principalmente
nella pagina di login per elencare i vari tenant a cui si può accedere. In un progetto reale, ci dev'essere
separazione completa tra i tenant, per cui ciascuno di questi avrebbe a disposizione una propria pagina
di login (su un dominio/sottodominio diverso) che permette agli utenti del tenant di accedere a quest'ultimo.
*/
func (controller *Controller) GetAllTenants(ctx *gin.Context) {
	tenants, err := controller.getAllTenantsUseCase.GetAllTenants()
	if err != nil {
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewAllTenantsResponseDTO(tenants)
	ctx.JSON(http.StatusOK, responseDto)
}

// GET TENANTS ========================================================================================

func (controller *Controller) GetTenantList(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var queryDto GetTenantListDTO
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	cmd := GetTenantListCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
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

	responseDto := NewTenantListResponseDTO(tenants, total)

	ctx.JSON(http.StatusOK, responseDto)
}

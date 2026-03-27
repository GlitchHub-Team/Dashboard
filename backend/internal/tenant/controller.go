package tenant

import (
	"errors"
	"fmt"
	"net/http"

	"backend/internal/common"
	"backend/internal/identity"
	transportHttp "backend/internal/transport/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Controller struct {
	log                    *zap.Logger
	createTenantUseCase    CreateTenantUseCase
	deleteTenantUseCase    DeleteTenantUseCase
	getTenantUseCase       GetTenantUseCase
	getTenantListUseCase   GetTenantListUseCase
	getTenantByUserUseCase GetTenantByUserUseCase
}

func NewTenantController(
	log *zap.Logger,
	createTenantUseCase CreateTenantUseCase,
	deleteTenantUseCase DeleteTenantUseCase,
	getTenantUseCase GetTenantUseCase,
	getTenantListUseCase GetTenantListUseCase,
	getTenantByUserUseCase GetTenantByUserUseCase,
) *Controller {
	return &Controller{
		log:                    log,
		createTenantUseCase:    createTenantUseCase,
		deleteTenantUseCase:    deleteTenantUseCase,
		getTenantUseCase:       getTenantUseCase,
		getTenantListUseCase:   getTenantListUseCase,
		getTenantByUserUseCase: getTenantByUserUseCase,
	}
}

// CREATE TENANT ======================================================================================
func (controller *Controller) CreateTenant(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto CreateTenantDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !common.ValidationError(ctx, err) {
			common.RequestError(ctx, err)
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
			common.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrTenantAlreadyExists) {
			common.RequestError(ctx, err)
			return
		}

		common.RequestServerError(ctx, err)
		return
	}

	responseDto := NewTenantResponseDTO(createdTenant)
	ctx.JSON(http.StatusOK, responseDto)
}

// DELETE TENANT ======================================================================================
func (controller *Controller) DeleteTenant(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}
	var bodyDto DeleteTenantDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !common.ValidationError(ctx, err) {
			common.RequestError(ctx, err)
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
			common.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrTenantNotFound) {
			common.RequestError(ctx, err)
			return
		}

		common.RequestServerError(ctx, err)
		return
	}

	responseDto := NewTenantResponseDTO(oldTenant)
	ctx.JSON(http.StatusOK, responseDto)
}

// GET TENANT =========================================================================================
func (controller *Controller) GetTenant(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto GetTenantDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !common.ValidationError(ctx, err) {
			common.RequestError(ctx, err)
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
			common.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrTenantNotFound) {
			common.RequestError(ctx, err)
			return
		}

		common.RequestServerError(ctx, err)
		return
	}

	responseDto := NewTenantResponseDTO(tenant)
	ctx.JSON(http.StatusOK, responseDto)
}

// GET TENANTS ========================================================================================
func (controller *Controller) GetTenants(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var queryDto GetTenantListDTO
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !common.ValidationError(ctx, err) {
			common.RequestError(ctx, err)
		}
		return
	}

	cmd := GetTenantListCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
	}

	tenants, err := controller.getTenantListUseCase.GetTenantList(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			common.RequestUnauthorized(ctx, err)
			return
		}

		common.RequestServerError(ctx, err)
		return
	}

	responseDtos := NewTenantListResponseDTO(tenants, 10)
	listInfo := gin.H{
		"page":  queryDto.Page,
		"limit": queryDto.Limit,
	}
	ctx.JSON(http.StatusOK, gin.H{
		"tenants":   responseDtos,
		"list_info": listInfo,
	})
}

// GET TENANT BY USER =================================================================================
// non usare :)
func (controller *Controller) GetTenantByUser(ctx *gin.Context) {
	var cmd GetTenantByUserCommand

	if err := ctx.ShouldBindJSON(&cmd); err != nil {
		controller.log.Error("Error binding JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	tenant, err := controller.getTenantByUserUseCase.GetTenantByUser(cmd)
	if err != nil {
		controller.log.Error("Error getting tenant by user", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Errorf("failed to get tenant by user"))
		return
	}

	ctx.JSON(http.StatusOK, tenant)
}

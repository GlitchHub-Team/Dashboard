package gateway

import (
	"errors"
	"net/http"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/gateway/mocks/gateway_usecases.go -package=mocks . CreateGatewayUseCase,DeleteGatewayUseCase,GetGatewayUseCase,GetAllGatewaysUseCase,GetGatewaysByTenantUseCase,CommissionGatewayUseCase,DecommissionGatewayUseCase,InterruptGatewayUseCase,ResumeGatewayUseCase,ResetGatewayUseCase,RebootGatewayUseCase,SetGatewayIntervalLimitUseCase
type CreateGatewayUseCase interface {
	CreateGateway(command CreateGatewayCommand) (Gateway, error)
}
type DeleteGatewayUseCase interface {
	DeleteGateway(cmd DeleteGatewayCommand) (Gateway, error)
}

type GetGatewayUseCase interface {
	GetGateway(cmd GetGatewayByIdCommand) (Gateway, error)
}

type GetAllGatewaysUseCase interface {
	GetAllGateways(command GetAllGatewaysCommand) ([]Gateway, uint, error)
}

type GetGatewaysByTenantUseCase interface {
	GetGatewaysByTenant(command GetGatewaysByTenantCommand) ([]Gateway, uint, error)
}

type GetGatewayByTenantIDUseCase interface {
	GetGatewayByTenantID(cmd GetGatewayByTenantIDCommand) (Gateway, error)
}

/*   =================   */

type CommissionGatewayUseCase interface {
	CommissionGateway(cmd CommissionGatewayCommand) (Gateway, error)
}

type DecommissionGatewayUseCase interface {
	DecommissionGateway(cmd DecommissionGatewayCommand) (Gateway, error)
}

type InterruptGatewayUseCase interface {
	InterruptGateway(cmd InterruptGatewayCommand) (Gateway, error)
}

type ResumeGatewayUseCase interface {
	ResumeGateway(cmd ResumeGatewayCommand) (Gateway, error)
}

type ResetGatewayUseCase interface {
	ResetGateway(cmd ResetGatewayCommand) (Gateway, error)
}

type RebootGatewayUseCase interface {
	RebootGateway(cmd RebootGatewayCommand) (Gateway, error)
}

type SetGatewayIntervalLimitUseCase interface {
	SetGatewayIntervalLimit(cmd SetGatewayIntervalLimitCommand) (Gateway, error)
}

type GatewayController struct {
	log *zap.Logger

	createGatewayUseCase           CreateGatewayUseCase
	deleteGatewayUseCase           DeleteGatewayUseCase
	getAllGatewaysUseCase          GetAllGatewaysUseCase
	getGatewaysByTenantUseCase     GetGatewaysByTenantUseCase
	commissionGatewayUseCase       CommissionGatewayUseCase
	decommissionGatewayUseCase     DecommissionGatewayUseCase
	interruptGatewayUseCase        InterruptGatewayUseCase
	resumeGatewayUseCase           ResumeGatewayUseCase
	resetGatewayUseCase            ResetGatewayUseCase
	rebootGatewayUseCase           RebootGatewayUseCase
	setGatewayIntervalLimitUseCase SetGatewayIntervalLimitUseCase
	getGatewayUseCase              GetGatewayUseCase
	getGatewayByTenantIDUseCase    GetGatewayByTenantIDUseCase
}

func NewGatewayController(
	log *zap.Logger,
	createGatewayUseCase CreateGatewayUseCase,
	deleteGatewayUseCase DeleteGatewayUseCase,
	getAllGatewaysUseCase GetAllGatewaysUseCase,
	getGatewaysByTenantUseCase GetGatewaysByTenantUseCase,
	commissionGatewayUseCase CommissionGatewayUseCase,
	decommissionGatewayUseCase DecommissionGatewayUseCase,
	interruptGatewayUseCase InterruptGatewayUseCase,
	resumeGatewayUseCase ResumeGatewayUseCase,
	resetGatewayUseCase ResetGatewayUseCase,
	rebootGatewayUseCase RebootGatewayUseCase,
	setGatewayIntervalLimitUseCase SetGatewayIntervalLimitUseCase,
	getGatewayUseCase GetGatewayUseCase,
	getGatewayByTenantIDUseCase GetGatewayByTenantIDUseCase,
) *GatewayController {
	return &GatewayController{
		log,
		createGatewayUseCase,
		deleteGatewayUseCase,
		getAllGatewaysUseCase,
		getGatewaysByTenantUseCase,
		commissionGatewayUseCase,
		decommissionGatewayUseCase,
		interruptGatewayUseCase,
		resumeGatewayUseCase,
		resetGatewayUseCase,
		rebootGatewayUseCase,
		setGatewayIntervalLimitUseCase,
		getGatewayUseCase,
		getGatewayByTenantIDUseCase,
	}
}

func (controller *GatewayController) CommissionGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto commissionGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	tenantId, err := uuid.Parse(bodyDto.TenantId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := CommissionGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
		TenantId:  tenantId,
	}

	gateway, err := controller.commissionGatewayUseCase.CommissionGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: gateway.PublicIdentifier,
	}

	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) DecommissionGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto decommissionGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := DecommissionGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.decommissionGatewayUseCase.DecommissionGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: gateway.PublicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) InterruptGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto interruptGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := InterruptGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.interruptGatewayUseCase.InterruptGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: gateway.PublicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) ResumeGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto resumeGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := ResumeGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.resumeGatewayUseCase.ResumeGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: gateway.PublicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) ResetGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto resetGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := ResetGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.resetGatewayUseCase.ResetGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: gateway.PublicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) RebootGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto rebootGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := RebootGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.rebootGatewayUseCase.RebootGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: gateway.PublicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

/*   ================================   */

func (controller *GatewayController) SetGatewayIntervalLimit(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto setGatewayIntervalLimitDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := SetGatewayIntervalLimitCommand{
		Requester:     requester,
		GatewayId:     gatewayId,
		IntervalLimit: bodyDto.IntervalLimit,
	}

	gateway, err := controller.setGatewayIntervalLimitUseCase.SetGatewayIntervalLimit(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	publicIdentifier := ""
	if gateway.PublicIdentifier != nil {
		publicIdentifier = *gateway.PublicIdentifier
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: publicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) CreateGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto createGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	cmd := CreateGatewayCommand{
		Requester: requester,
		Name:      bodyDto.GatewayName,
	}

	gateway, err := controller.createGatewayUseCase.CreateGateway(cmd)
	if err != nil {

		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayAlreadyExists) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	publicIdentifier := ""
	if gateway.PublicIdentifier != nil {
		publicIdentifier = *gateway.PublicIdentifier
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit,
		PublicIdentifier: publicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) DeleteGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}
	var bodyDto deleteGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := DeleteGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.deleteGatewayUseCase.DeleteGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit.Milliseconds(),
		PublicIdentifier: gateway.PublicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

/*   ================================   */

func (controller *GatewayController) GetAllGateways(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	queryDto := getGatewayListDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}

	if err := ctx.ShouldBindJSON(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	cmd := GetAllGatewaysCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
	}

	gateways, count, err := controller.getAllGatewaysUseCase.GetAllGateways(cmd)
	if err != nil {
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDtos := make([]gatewayAllResponseDTO, len(gateways))

	for i, gateway := range gateways {
		responseDtos[i] = gatewayAllResponseDTO{
			ListInfo: dto.ListInfo{
				Total: count,
				Count: uint(queryDto.Page),
			},
			Gateways: []AllGatewayResponseDTO{
				{
					GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
					GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
					TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
					Status:           gateway.Status,
					Interval:         gateway.IntervalLimit.Milliseconds(),
					PublicIdentifier: gateway.PublicIdentifier,
				},
			},
		}
	}

	ctx.JSON(http.StatusOK, responseDtos)
}

func (controller *GatewayController) GetGatewaysByTenant(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	tenantIdParam := ctx.Param("tenant_id")
	tenantId, err := uuid.Parse(tenantIdParam)

	queryDto := getGatewayListDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	if err != nil {
		transportHttp.RequestError(ctx, nil)
		return
	}

	cmd := GetGatewaysByTenantCommand{
		TenantId:  tenantId,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
		Requester: requester,
	}

	gateways, count, err := controller.getGatewaysByTenantUseCase.GetGatewaysByTenant(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDtos := make([]gatewayListResponseDTO, len(gateways))

	for i, gateway := range gateways {
		responseDtos[i] = gatewayListResponseDTO{
			ListInfo: dto.ListInfo{
				Total: count,
				Count: uint(queryDto.Page),
			},
			Gateways: []gatewayResponseDTO{
				{
					GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
					GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
					TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
					Status:           gateway.Status,
					Interval:         gateway.IntervalLimit.Milliseconds(),
				},
			},
		}
	}

	ctx.JSON(http.StatusOK, responseDtos)
}

func (controller *GatewayController) GetGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	gatewayPassedId := ctx.Param("gateway_id")
	gatewayId, err := uuid.Parse(gatewayPassedId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := GetGatewayByIdCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.getGatewayUseCase.GetGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit.Milliseconds(),
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) GetGatewayByTenantID(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	tenantIdParam := ctx.Param("tenant_id")
	gatewayPassedId := ctx.Param("gateway_id")

	tenantId, err := uuid.Parse(tenantIdParam)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	gatewayId, err := uuid.Parse(gatewayPassedId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := GetGatewayByTenantIDCommand{
		Requester: requester,
		TenantId:  tenantId,
		GatewayId: gatewayId,
	}

	gateway, err := controller.getGatewayByTenantIDUseCase.GetGatewayByTenantID(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			transportHttp.RequestError(ctx, err)
			return
		}

		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
		TenantIdField:    dto.TenantIdField{TenantId: gateway.TenantId.String()},
		Status:           gateway.Status,
		Interval:         gateway.IntervalLimit.Milliseconds(),
	}
	ctx.JSON(http.StatusOK, responseDto)
}

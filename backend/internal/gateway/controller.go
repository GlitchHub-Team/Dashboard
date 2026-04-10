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
	GetAllGateways() ([]Gateway, error)
}

type GetGatewaysByTenantUseCase interface {
	GetGatewaysByTenant(command GetGatewaysByTenantCommand) ([]Gateway, error)
}

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
		GatewayIdField:    dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField:  dto.GatewayNameField{GatewayName: gateway.Name},
		CommissionedToken: gateway.SigningSecret,
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
	}
	ctx.JSON(http.StatusOK, responseDto)
}

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

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
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

	responseDto := gatewayResponseDTO{
		GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
		GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
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
	}
	ctx.JSON(http.StatusOK, responseDto)
}

func (controller *GatewayController) GetAllGateways(ctx *gin.Context) {
	gateways, err := controller.getAllGatewaysUseCase.GetAllGateways()
	if err != nil {
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDtos := make([]gatewayResponseDTO, len(gateways))
	for i, gateway := range gateways {
		responseDtos[i] = gatewayResponseDTO{
			GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
			GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
			PublicIdentifier: gateway.PublicIdentifier,
		}
	}

	ctx.JSON(http.StatusOK, responseDtos)
}

func (controller *GatewayController) GetGatewaysByTenant(ctx *gin.Context) {
	var queryDto getGatewaysByTenantDTO
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	tenantId, err := uuid.Parse(queryDto.TenantId)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := GetGatewaysByTenantCommand{
		TenantId: tenantId,
	}

	gateways, err := controller.getGatewaysByTenantUseCase.GetGatewaysByTenant(cmd)
	if err != nil {
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDtos := make([]gatewayResponseDTO, len(gateways))
	for i, gateway := range gateways {
		responseDtos[i] = gatewayResponseDTO{
			GatewayIdField:   dto.GatewayIdField{GatewayId: gateway.Id.String()},
			GatewayNameField: dto.GatewayNameField{GatewayName: gateway.Name},
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

	var queryDto getGatewayByIdDTO
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(queryDto.GatewayId)
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
		PublicIdentifier: gateway.PublicIdentifier,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

package gateway

import (
	"errors"
	"net/http"

	"backend/internal/common"
	"backend/internal/common/dto"
	"backend/internal/identity"
	transportHttp "backend/internal/transport/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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

type GatewayController struct {
	log *zap.Logger

	createGatewayUseCase       CreateGatewayUseCase
	deleteGatewayUseCase       DeleteGatewayUseCase
	getAllGatewaysUseCase      GetAllGatewaysUseCase
	getGatewaysByTenantUseCase GetGatewaysByTenantUseCase
}

func NewGatewayController(
	log *zap.Logger,
	createGatewayUseCase CreateGatewayUseCase,
	deleteGatewayUseCase DeleteGatewayUseCase,
	getAllGatewaysUseCase GetAllGatewaysUseCase,
	getGatewaysByTenantUseCase GetGatewaysByTenantUseCase,
) *GatewayController {
	return &GatewayController{
		log,
		createGatewayUseCase,
		deleteGatewayUseCase,
		getAllGatewaysUseCase,
		getGatewaysByTenantUseCase,
	}
}

func (controller *GatewayController) CreateGateway(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		common.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto createGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !common.ValidationError(ctx, err) {
			common.RequestError(ctx, err)
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
			common.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayAlreadyExists) {
			common.RequestError(ctx, err)
			return
		}

		common.RequestServerError(ctx, err)
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
		common.RequestUnauthorized(ctx, err)
		return
	}
	var bodyDto deleteGatewayDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !common.ValidationError(ctx, err) {
			common.RequestError(ctx, err)
		}
		return
	}

	gatewayId, err := uuid.Parse(bodyDto.GatewayId)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	cmd := DeleteGatewayCommand{
		Requester: requester,
		GatewayId: gatewayId,
	}

	gateway, err := controller.deleteGatewayUseCase.DeleteGateway(cmd)
	if err != nil {
		if errors.Is(err, identity.ErrUnauthorizedAccess) {
			common.RequestUnauthorized(ctx, err)
			return
		} else if errors.Is(err, ErrGatewayNotFound) {
			common.RequestError(ctx, err)
			return
		}

		common.RequestServerError(ctx, err)
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
		common.RequestServerError(ctx, err)
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

func (controller *GatewayController) GetGatewaysByTenant(ctx *gin.Context) {
	var queryDto getGatewaysByTenantDTO
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !common.ValidationError(ctx, err) {
			common.RequestError(ctx, err)
		}
		return
	}

	tenantId, err := uuid.Parse(queryDto.TenantId)
	if err != nil {
		common.RequestError(ctx, err)
		return
	}

	cmd := GetGatewaysByTenantCommand{
		TenantId: tenantId,
	}

	gateways, err := controller.getGatewaysByTenantUseCase.GetGatewaysByTenant(cmd)
	if err != nil {
		common.RequestServerError(ctx, err)
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
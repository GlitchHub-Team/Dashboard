package sensor

import (
	"errors"
	"net/http"

	"backend/internal/gateway"
	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *SensorController) GetSensor(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	sensorIdParam := ctx.Param("sensor_id")
	sensorId, err := uuid.Parse(sensorIdParam)
	if err != nil {
		transportHttp.RequestError(ctx, ErrInvalidSensorID)
		return
	}

	cmd := GetSensorCommand{
		Requester: requester,
		SensorId:  sensorId,
	}

	sensor, err := c.getSensorUseCase.GetSensorById(cmd)
	if err != nil {
		if errors.Is(err, ErrSensorNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, ErrSensorNotFound)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewSensorResponseDTO(sensor)
	ctx.JSON(http.StatusOK, responseDto)
}

func (c *SensorController) GetSensorsByGateway(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	gatewayIdParam := ctx.Param("gateway_id")
	gatewayId, err := uuid.Parse(gatewayIdParam)
	if err != nil {
		transportHttp.RequestError(ctx, gateway.ErrInvalidGatewayID)
		return
	}

	queryDto := SensorQueryDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	cmd := GetSensorsByGatewayCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
		GatewayId: gatewayId,
	}

	sensors, total, err := c.getSensorsByGatewayUseCase.GetSensorsByGateway(cmd)
	if err != nil {
		if errors.Is(err, gateway.ErrGatewayNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, gateway.ErrGatewayNotFound)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewSensorsResponseDTO(sensors, total)
	ctx.JSON(http.StatusOK, responseDto)
}

func (c *SensorController) GetSensorsByTenant(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	queryDto := SensorQueryDTO{
		Pagination: dto.DEFAULT_PAGINATION,
	}
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	tenantIdParam := ctx.Param("tenant_id")
	tenantId, err := uuid.Parse(tenantIdParam)
	if err != nil {
		transportHttp.RequestError(ctx, tenant.ErrInvalidTenantID)
		return
	}

	cmd := GetSensorsByTenantCommand{
		Requester: requester,
		Page:      queryDto.Page,
		Limit:     queryDto.Limit,
		TenantId:  tenantId,
	}

	sensors, total, err := c.getSensorsByTenantUseCase.GetSensorsByTenant(cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrTenantNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestUnauthorized(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewSensorsResponseDTO(sensors, total)
	ctx.JSON(http.StatusOK, responseDto)
}

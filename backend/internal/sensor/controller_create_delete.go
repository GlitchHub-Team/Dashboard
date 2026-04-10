package sensor

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/gateway"
	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Controller) CreateSensor(ctx *gin.Context) {
	// Autorizza utente
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var bodyDto CreateSensorBodyDTO
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	cmd := CreateSensorCommand{
		Requester: requester,
		Name:      bodyDto.Name,
		Interval:  time.Duration(bodyDto.Interval) * time.Millisecond,
		Profile:   bodyDto.Profile,
		GatewayId: bodyDto.GatewayId,
	}

	sensor, err := c.createSensorUseCase.CreateSensor(cmd)
	if err != nil {
		if errors.Is(err, gateway.ErrGatewayNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, gateway.ErrGatewayNotFound)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	responseDto := NewSensorResponseDTO(sensor)
	ctx.JSON(http.StatusOK, responseDto)
}

func (c *Controller) DeleteSensor(ctx *gin.Context) {
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

	cmd := DeleteSensorCommand{
		Requester: requester,
		SensorId:  sensorId,
	}

	sensor, err := c.deleteSensorUseCase.DeleteSensor(cmd)
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

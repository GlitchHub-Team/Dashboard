package sensor

import (
	"errors"
	"net/http"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Controller) InterruptSensor(ctx *gin.Context) {
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

	cmd := InterruptSensorCommand{
		Requester: requester,
		SensorId:  sensorId,
	}

	err = c.interruptSensorUseCase.InterruptSensor(cmd)
	if err != nil {
		if errors.Is(err, ErrSensorNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, ErrSensorNotFound)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (c *Controller) ResumeSensor(ctx *gin.Context) {
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

	cmd := ResumeSensorCommand{
		Requester: requester,
		SensorId:  sensorId,
	}

	err = c.resumeSensorUseCase.ResumeSensor(cmd)
	if err != nil {
		if errors.Is(err, ErrSensorNotFound) || errors.Is(err, identity.ErrUnauthorizedAccess) {
			transportHttp.RequestNotFound(ctx, ErrSensorNotFound)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

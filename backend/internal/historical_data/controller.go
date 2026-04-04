package historical_data

import (
	"errors"
	"net/http"
	"time"

	transportHttp "backend/internal/infra/transport/http"
	transportDto "backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GetSensorHistoricalDataUseCase interface {
	GetSensorHistoricalData(cmd GetSensorHistoricalDataCommand) ([]HistoricalSample, error)
}

type Controller struct {
	log *zap.Logger

	getSensorHistoricalDataUseCase GetSensorHistoricalDataUseCase
}

func NewHistoricalDataController(
	log *zap.Logger,
	getSensorHistoricalDataUseCase GetSensorHistoricalDataUseCase,
) *Controller {
	return &Controller{
		log:                            log,
		getSensorHistoricalDataUseCase: getSensorHistoricalDataUseCase,
	}
}

func (controller *Controller) GetSensorHistoricalData(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	var uriDto GetHistoricalDataUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	tenantId, _ := uuid.Parse(uriDto.TenantId)
	sensorId, _ := uuid.Parse(uriDto.SensorId)

	var queryDto GetHistoricalDataQueryDTO
	if err := ctx.ShouldBindQuery(&queryDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	from, err := parseOptionalRFC3339(queryDto.From)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	to, err := parseOptionalRFC3339(queryDto.To)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	cmd := GetSensorHistoricalDataCommand{
		Requester: requester,
		TenantId:  tenantId,
		SensorId:  sensorId,
		From:      from,
		To:        to,
		Limit:     queryDto.Limit,
	}

	samples, err := controller.getSensorHistoricalDataUseCase.GetSensorHistoricalData(cmd)
	if err != nil {
		switch {
		case errors.Is(err, tenant.ErrTenantNotFound), errors.Is(err, identity.ErrUnauthorizedAccess):
			transportHttp.RequestNotFound(ctx, tenant.ErrTenantNotFound)
			return
		case errors.Is(err, ErrInvalidDateRange):
			transportHttp.RequestError(ctx, err)
			return
		default:
			transportHttp.RequestServerError(ctx, err)
			return
		}
	}

	ctx.JSON(http.StatusOK, NewHistoricalDataResponseDTO(samples))
}

func parseOptionalRFC3339(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, ErrInvalidTimestamp
	}

	return &parsed, nil
}

type GetHistoricalDataUriDTO struct {
	transportDto.TenantIdField
	transportDto.SensorIdField
}

package real_time_data

import (
	"errors"
	"time"

	transportHttp "backend/internal/infra/transport/http"
	transportWs "backend/internal/infra/transport/ws"
	"backend/internal/sensor"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/real_time_data/mocks/use_case.go -package=mocks . GetRealTimeDataUseCase

type GetRealTimeDataUseCase interface {
	GetRealTimeData(cmd GetRealTimeDataCommand) (
		dataChannel chan RealTimeSample, errorChannel chan RealTimeError, err error,
	)
}

type Controller struct {
	log                    *zap.Logger
	getRealTimeDataUseCase GetRealTimeDataUseCase
}

func NewController(
	log *zap.Logger,
	getRealTimeDataUseCase GetRealTimeDataUseCase,
) *Controller {
	return &Controller{
		log:                    log,
		getRealTimeDataUseCase: getRealTimeDataUseCase,
	}
}

func (c *Controller) GetRealTimeData(ctx *gin.Context) {
	// 1. Estrai requester
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, err)
		return
	}

	// 2. URI binding
	var uriDto GetRealTimeDataDTO

	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		c.log.Sugar().Errorf("err: %v", err)
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}
	sensorId, _ := uuid.Parse(uriDto.SensorId)
	tenantId, _ := uuid.Parse(uriDto.TenantId)

	// 3. Esegui comando
	cmd := GetRealTimeDataCommand{
		Requester: requester,
		SensorId:  sensorId,
		TenantId:  tenantId,
	}
	// c.log.Sugar().Errorf("err: %v", err)

	dataChannel, errChannel, err := c.getRealTimeDataUseCase.GetRealTimeData(cmd)
	if err != nil {
		if errors.Is(err, sensor.ErrSensorNotFound) || errors.Is(err, sensor.ErrSensorNotActive) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	webSocketConn, err := transportWs.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		// NOTA: gorilla aggiunge già 400 Bad Request qui
		return
	}

	defer webSocketConn.Close() //nolint:errcheck

	go c.startClientListener(webSocketConn, errChannel)

loop:
	for {
		select {
		case sample := <-dataChannel:

			c.log.Sugar().Debugf("profile: %v", sample.GetProfile())

			dataDto := MapDomainToWSDto(sample)
			if err != nil {
				c.log.Error(
					"Cannot decode sensor profile data",
					zap.Error(err),
					zap.Time("timestamp", time.Now()),
					zap.String("sensorId", sensorId.String()),
				)
				continue
			}

			err = webSocketConn.WriteJSON(dataDto)
			if err != nil {
				c.log.Error(
					"Cannot send data to client",
					zap.Error(err),
					zap.Time("timestamp", time.Now()),
					zap.String("sensorId", sensorId.String()),
					zap.Any("requester", requester),
				)
				continue
			}

			c.log.Sugar().Debugf("Sent sample @ %v", sample.GetTimestamp())

		case err := <-errChannel:
			if !errors.Is(err, ErrClientDisconnected) {
				_ = webSocketConn.WriteJSON(RealTimeErrorOutDTO{
					Error: err.Error(),
				})
			}
			break loop
		}
	}
}

func (c *Controller) startClientListener(conn *websocket.Conn, errorChannel chan RealTimeError) {
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			// TODO: togliere messaggio di DEBUG
			c.log.Sugar().Debugf("[startClientListener] Client websocket disconesso: %v", conn.NetConn().RemoteAddr())
			errorChannel <- NewErrClientDisconnected()
			break
		}
	}
}

package real_time_data

import (
	"errors"
	"time"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	transportWs "backend/internal/infra/transport/ws"
	"backend/internal/sensor"
	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type GetRealTimeDataUseCase interface {
	RetrieveRealTimeData(cmd RetrieveRealTimeDataCommand) (
		dataChannel chan RealTimeRawSample, errorChannel chan RealTimeError, err error,
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
	// requester, err := transportHttp.ExtractRequester(ctx)
	// if err != nil {
	// 	transportHttp.RequestUnauthorized(ctx, err)
	// 	c.log.Sugar().Errorf("err: %v", err)
	// 	return
	// }
	requester := identity.Requester{
		RequesterUserId:   uint(1),
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
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

	// 3. Esegui comando
	cmd := RetrieveRealTimeDataCommand{
		Requester: requester,
		SensorId:  sensorId,
	}
	// c.log.Sugar().Errorf("err: %v", err)

	dataChannel, errChannel, err := c.getRealTimeDataUseCase.RetrieveRealTimeData(cmd)
	if err != nil {
		if errors.Is(err, sensor.ErrSensorNotFound) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	webSocketConn, err := transportWs.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		transportHttp.RequestServerError(ctx, err)
		return
	}

	defer webSocketConn.Close() //nolint:errcheck

	go c.startClientListener(webSocketConn, errChannel)

loop:
	for {
		select {
		case sample := <-dataChannel:

			c.log.Sugar().Debugf("profile: %v", sample.Profile)
			dataDto, err := dto.DecodeSensorProfileData(sample.Profile, sample.Data)
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
			
			c.log.Sugar().Debugf("Sent sample @ %v", sample.Timestamp)


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
			c.log.Sugar().Debugf("[startClientListener] Client websocket disconesso: %v", conn.NetConn().RemoteAddr()) // TODO: DEBUG
			errorChannel <- NewErrClientDisconnected(time.Now())
			break
		}
	}
}

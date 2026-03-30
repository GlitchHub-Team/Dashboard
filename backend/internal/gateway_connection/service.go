package gateway_connection

import (
	"backend/internal/gateway"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GatewayHelloService interface {
	ProcessHello(msg GatewayHelloMessage) error
}

type gatewayHelloService struct {
	getGateway  gateway.GetGatewayPort
	saveGateway gateway.SaveGatewayPort
	logger      *zap.Logger
}

func NewGatewayHelloService(
	getGateway gateway.GetGatewayPort,
	saveGateway gateway.SaveGatewayPort,
	logger *zap.Logger,
) GatewayHelloService {
	return &gatewayHelloService{
		getGateway:  getGateway,
		saveGateway: saveGateway,
		logger:      logger,
	}
}

func (s *gatewayHelloService) ProcessHello(msg GatewayHelloMessage) error {
	gwID, err := uuid.Parse(msg.GatewayId)
	if err != nil {
		s.logger.Error("Invalid gateway ID format", zap.Error(err))
		return err
	}

	gw, err := s.getGateway.GetById(gwID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Error("Gateway not found in database", zap.String("gatewayId", msg.GatewayId))
			return err
		}
		s.logger.Error("Error retrieving gateway from database", zap.Error(err))
		return err
	}

	if gw.PublicIdentifier != msg.PublicIdentifier {
		gw.PublicIdentifier = msg.PublicIdentifier
		gw.Status = gateway.GATEWAY_STATUS_ACTIVE

		if err := s.saveGateway.Save(gw); err != nil {
			s.logger.Error("Error saving gateway", zap.Error(err))
			return err
		}
	}

	return nil
}

package hello

import (
	"errors"

	"backend/internal/gateway"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrPublicIdentifierRequired = errors.New("publicIdentifier is required")

type GatewayHelloUseCase interface {
	ProcessHello(msg GatewayHelloMessage) error
}

type GatewayHelloService struct {
	getGateway  gateway.GetGatewayPort
	saveGateway gateway.SaveGatewayPort
	logger      *zap.Logger
}

func NewGatewayHelloService(
	getGateway gateway.GetGatewayPort,
	saveGateway gateway.SaveGatewayPort,
	logger *zap.Logger,
) *GatewayHelloService {
	return &GatewayHelloService{
		getGateway:  getGateway,
		saveGateway: saveGateway,
		logger:      logger,
	}
}

func (s *GatewayHelloService) ProcessHello(msg GatewayHelloMessage) error {
	if msg.PublicIdentifier == "" {
		s.logger.Error("Missing public identifier in hello message")
		return ErrPublicIdentifierRequired
	}

	gwID, err := uuid.Parse(msg.GatewayId)
	if err != nil {
		s.logger.Error("Invalid gateway ID format", zap.Error(err), zap.String("gatewayId", msg.GatewayId))
		return err
	}

	gw, err := s.getGateway.GetById(gwID)
	if err != nil {
		s.logger.Error("Error retrieving gateway from database", zap.Error(err))
		return err
	}

	if gw.IsZero() {
		s.logger.Error("Gateway not found in database", zap.String("gatewayId", msg.GatewayId))
		return gateway.ErrGatewayNotFound
	}

	if gw.PublicIdentifier == nil || *gw.PublicIdentifier != msg.PublicIdentifier {
		gw.PublicIdentifier = &msg.PublicIdentifier

		if _, err := s.saveGateway.Save(gw); err != nil {
			s.logger.Error("Error saving gateway", zap.Error(err), zap.String("gatewayId", gw.Id.String()))
			return err
		}
	}

	return nil
}

package hello

import (
	"errors"

	"backend/internal/gateway"

	"go.uber.org/zap"
)

var ErrPublicIdentifierRequired = errors.New("publicIdentifier is required")

type GatewayHelloUseCase interface {
	ProcessHello(msg GatewayHelloMessageCommand) error
}

type GatewayHelloService struct {
	getGateway  gateway.GetGatewayPort
	saveGateway gateway.SaveGatewayPort
	logger      *zap.Logger
}

var _ GatewayHelloUseCase = (*GatewayHelloService)(nil)

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

func (s *GatewayHelloService) ProcessHello(cmd GatewayHelloMessageCommand) error {
	if cmd.PublicIdentifier == "" {
		s.logger.Error("Missing public identifier in hello message")
		return ErrPublicIdentifierRequired
	}

	gw, err := s.getGateway.GetById(cmd.GatewayId)
	if err != nil {
		s.logger.Error("Error retrieving gateway from database", zap.Error(err))
		return err
	}

	if gw.IsZero() {
		s.logger.Error("Gateway not found in database", zap.String("gatewayId", cmd.GatewayId.String()))
		return gateway.ErrGatewayNotFound
	}

	if gw.PublicIdentifier == nil || *gw.PublicIdentifier != cmd.PublicIdentifier {
		gw.PublicIdentifier = &cmd.PublicIdentifier

		if _, err := s.saveGateway.Save(gw); err != nil {
			s.logger.Error("Error saving gateway", zap.Error(err), zap.String("gatewayId", cmd.GatewayId.String()))
			return err
		}
	}

	return nil
}

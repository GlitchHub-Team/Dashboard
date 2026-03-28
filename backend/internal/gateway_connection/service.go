package gateway_connection

import (
	"backend/internal/gateway"
	"github.com/google/uuid"
)

type GatewayHelloService interface {
	ProcessHello(msg GatewayHelloMessage) error
}

type gatewayHelloService struct {
	saveGateway gateway.SaveGatewayPort
}

func NewGatewayHelloService(saveGateway gateway.SaveGatewayPort) GatewayHelloService {
	return &gatewayHelloService{saveGateway: saveGateway}
}

func (s *gatewayHelloService) ProcessHello(msg GatewayHelloMessage) error {
	gwID, err := uuid.Parse(msg.GatewayId)
	if err != nil {
		return err
	}
	g := gateway.Gateway{
		Id:               gwID,
		PublicIdentifier: msg.PublicIdentifier,
		Status:           gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit:    0,
	}
	return s.saveGateway.Save(g)
}

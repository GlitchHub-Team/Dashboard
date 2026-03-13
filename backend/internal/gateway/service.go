package gateway

import (
	"go.uber.org/zap"
	"github.com/google/uuid"
)

type CreateGatewayUseCase interface {
	CreateGateway(command CreateGatewayCommand) (Gateway, error) 
}

type DeleteGatewayUseCase interface {
	DeleteGateway(command DeleteGatewayCommand) error
}

// CreateGatewayService ---------------------------------------------------------------------------
type CreateGatewayService struct {
	log             *zap.Logger
	saveGatewayPort SaveGatewayPort
}

func NewCreateGatewayService(log *zap.Logger, saveGatewayPort SaveGatewayPort) *CreateGatewayService {
	return &CreateGatewayService{
		log: log,
		saveGatewayPort: saveGatewayPort,
	}
}

// CreateGatewayService ---------------------------------------------------------------------------
func (s *CreateGatewayService) CreateGateway(command CreateGatewayCommand) (Gateway, error) {
	s.log.Info("Created gateway with name" + command.Name)

	gateway := Gateway{
		Id: uuid.New(),
		Name: command.Name,
		Status: GATEWAY_STATUS_ACTIVE,
	}

	// Logica di business...

	return gateway, nil
}

func (s *CreateGatewayService) DeleteGateway(command DeleteGatewayCommand) error {
	return nil
}

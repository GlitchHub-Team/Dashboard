package gateway

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreateGatewayUseCase interface {
	CreateGateway(command CreateGatewayCommand) (Gateway, error)
}



// CreateGatewayService ---------------------------------------------------------------------------
type CreateGatewayService struct {
	log             *zap.Logger
	saveGatewayPort SaveGatewayPort
}

func NewCreateGatewayService(log *zap.Logger, saveGatewayPort SaveGatewayPort) CreateGatewayUseCase {
	return &CreateGatewayService{
		log:             log,
		saveGatewayPort: saveGatewayPort,
	}
}


func (s *CreateGatewayService) CreateGateway(command CreateGatewayCommand) (Gateway, error) {
	s.log.Info("Created gateway with name" + command.Name)

	gateway := Gateway{
		Id:     uuid.New(),
		Name:   command.Name,
		Status: GATEWAY_STATUS_ACTIVE,
	}

	// Logica di business...

	return gateway, nil
}


// DeleteGatewayService ---------------------------------------------------------------------------
type DeleteGatewayUseCase interface {
	DeleteGateway(command DeleteGatewayCommand) error
}

type DeleteGatewayService struct {
	removeGatewayPort RemoveGatewayPort
}

func NewDeleteGatewayService(removeGatewayPort RemoveGatewayPort) DeleteGatewayUseCase {
	return &DeleteGatewayService{
		removeGatewayPort: removeGatewayPort,
	}
}


func (s *DeleteGatewayService) DeleteGateway(command DeleteGatewayCommand) error {
	return nil
}


var _ CreateGatewayUseCase = (*CreateGatewayService)(nil)

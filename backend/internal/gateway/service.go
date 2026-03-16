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

// Compile-time checks
var _ CreateGatewayUseCase = (*CreateGatewayService)(nil)

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

<<<<<<< HEAD
type DeleteGatewayService struct {
	log *zap.Logger
}

func NewDeleteGatewayService(log *zap.Logger,) *DeleteGatewayService {
	return &DeleteGatewayService{
		log:             log,
	}
}
=======
// Compile-time checks
var _ DeleteGatewayUseCase = (*DeleteGatewayService)(nil)
>>>>>>> origin/issue-17

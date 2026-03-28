package gateway

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Use Cases ------------------------------------------------------------------------------------------

type CreateGatewayPort interface {
	CreateGateway(command CreateGatewayCommand) (Gateway, error)
}

type DeleteGatewayPort interface {
	DeleteGateway(command DeleteGatewayCommand) error
}

type GetGatewaysPort interface {
	GetByTenantId(tenantId string) ([]Gateway, error)
	GetAll() ([]Gateway, error)
}

type GetAllGateways interface {
	GetAllGateways() ([]Gateway, error)
}

//  Costruttore Globale -------------------------------------------------------------------------------

func NewGatewayServices(
	log *zap.Logger,
	saveGatewayPort SaveGatewayPort,
	removeGatewayPort RemoveGatewayPort,
	getGatewayPort GetGatewayPort,
	getGatewaysPort GetGatewaysPort,
) (
	CreateGatewayUseCase,
	DeleteGatewayUseCase,
	GetGatewayUseCase,
	GetAllGateways,
	GetGatewaysByTenantUseCase,
) {
	createSvc := &CreateGatewayService{log: log, saveGatewayPort: saveGatewayPort}
	deleteSvc := &DeleteGatewayService{removeGatewayPort: removeGatewayPort}
	getSvc := &GetGatewayService{getGatewayPort: getGatewayPort}
	getListSvc := &GetGatewayListService{getGatewaysPort: getGatewaysPort}
	getByTenantSvc := &GetGatewaysByTenantService{getGatewaysPort: getGatewaysPort}

	return createSvc, deleteSvc, getSvc, getListSvc, getByTenantSvc
}

// CreateGatewayService -------------------------------------------------------------------------------

type CreateGatewayService struct {
	log             *zap.Logger
	saveGatewayPort SaveGatewayPort
}

func (s *CreateGatewayService) CreateGateway(command CreateGatewayCommand) (Gateway, error) {
	if !command.Requester.IsSuperAdmin() {
		return Gateway{}, identity.ErrUnauthorizedAccess
	}

	s.log.Info("Created gateway with name " + command.Name)

	gateway := Gateway{
		Id:     uuid.New(),
		Name:   command.Name,
		Status: GATEWAY_STATUS_ACTIVE,
	}

	_, err := s.saveGatewayPort.Save(gateway)
	if err != nil {
		return Gateway{}, err
	}

	// return gateway, nil
	return Gateway{}, nil
}

// DeleteGatewayService -------------------------------------------------------------------------------

type DeleteGatewayService struct {
	removeGatewayPort RemoveGatewayPort
}

func (s *DeleteGatewayService) DeleteGateway(command DeleteGatewayCommand) (Gateway, error) {
	oldGateway, err := s.removeGatewayPort.Remove(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if oldGateway.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	return oldGateway, nil
}

type GetGatewayService struct {
	getGatewayPort GetGatewayPort
}

func (s *GetGatewayService) GetGateway(command GetGatewayByIdCommand) (Gateway, error) {
	return s.getGatewayPort.GetById(command.GatewayId.String())
}

// GetGatewayListService ------------------------------------------------------------------------------
type GetGatewayListService struct {
	getGatewaysPort GetGatewaysPort
}

func (s *GetGatewayListService) GetAllGateways() ([]Gateway, error) {
	gat, err := s.getGatewaysPort.GetAll()
	if err != nil {
		return nil, err
	}

	if gat == nil {
		return nil, ErrGatewayNotFound
	}

	return gat, nil
}

// GetGatewaysByTenantService -------------------------------------------------------------------------
type GetGatewaysByTenantService struct {
	getGatewaysPort GetGatewaysPort
}

func (s *GetGatewaysByTenantService) GetGatewaysByTenant(command GetGatewaysByTenantCommand) ([]Gateway, error) {
	if command.TenantId == uuid.Nil {
		return nil, ErrGatewayNotFound
	}

	tenantGateways, err := s.getGatewaysPort.GetByTenantId(command.TenantId.String())
	if err != nil {
		return nil, err
	}

	if tenantGateways == nil {
		return nil, ErrGatewayNotFound
	}

	return tenantGateways, nil
}

var (
	_ CreateGatewayUseCase = (*CreateGatewayService)(nil)
	_ DeleteGatewayUseCase = (*DeleteGatewayService)(nil)
	_ GetGatewayUseCase    = (*GetGatewayService)(nil)
	_ GetAllGateways       = (*GetGatewayListService)(nil)

	_ GetGatewaysByTenantUseCase = (*GetGatewaysByTenantService)(nil)
)

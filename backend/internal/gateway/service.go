package gateway

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Use Cases ------------------------------------------------------------------------------------------

type GatewayManagementService struct {
	log             *zap.Logger
	getGatewayPort  GetGatewayPort
	getGatewaysPort GetGatewaysPort
}

func NewGatewayManagementService(
	log *zap.Logger,
	getPort GetGatewayPort,
	getManyPort GetGatewaysPort,
) *GatewayManagementService {
	return &GatewayManagementService{
		log:             log,
		getGatewayPort:  getPort,
		getGatewaysPort: getManyPort,
	}
}

func (s *GatewayManagementService) GetGateway(command GetGatewayByIdCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	if !command.IsSuperAdmin() {
		if command.RequesterTenantId == nil || !gw.BelongsToTenant(*command.RequesterTenantId) {
			return Gateway{}, ErrUnauthorizedAccess
		}
	}

	return gw, nil
}

func (s *GatewayManagementService) GetAllGateways() ([]Gateway, error) {
	gat, err := s.getGatewaysPort.GetAll()
	if err != nil {
		return nil, err
	}

	if gat == nil {
		return nil, ErrGatewayNotFound
	}

	return gat, nil
}

func (s *GatewayManagementService) GetGatewaysByTenant(command GetGatewaysByTenantCommand) ([]Gateway, error) {
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

func (s *GatewayManagementService) GetGatewayById(command GetGatewayByIdCommand) (Gateway, error) {
	return s.getGatewayPort.GetById(command.GatewayId.String())
}

var (
	_ GetGatewayUseCase          = (*GatewayManagementService)(nil)
	_ GetAllGatewaysUseCase      = (*GatewayManagementService)(nil)
	_ GetGatewaysByTenantUseCase = (*GatewayManagementService)(nil)
)

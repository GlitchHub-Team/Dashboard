package gateway

import (
	"backend/internal/tenant"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Use Cases ------------------------------------------------------------------------------------------

type GatewayManagementService struct {
	log             *zap.Logger
	getGatewayPort  GetGatewayPort
	getGatewaysPort GetGatewaysPort
	getTenantPort   tenant.GetTenantPort
}

func NewGatewayManagementService(
	log *zap.Logger,
	getPort GetGatewayPort,
	getManyPort GetGatewaysPort,
	getTenantPort tenant.GetTenantPort,
) *GatewayManagementService {
	return &GatewayManagementService{
		log:             log,
		getGatewayPort:  getPort,
		getGatewaysPort: getManyPort,
		getTenantPort:   getTenantPort,
	}
}

/*  =================================   */

func (s *GatewayManagementService) GetGateway(command GetGatewayByIdCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if !command.IsSuperAdmin() {
		return Gateway{}, ErrUnauthorizedAccess
	}

	return gw, nil
}

func (s *GatewayManagementService) GetAllGateways(command GetAllGatewaysCommand) ([]Gateway, uint, error) {
	gw, count, err := s.getGatewaysPort.GetAll(command.Page, command.Limit)
	if err != nil {
		return nil, 0, err
	}

	if gw == nil {
		return nil, 0, ErrGatewayNotFound
	}

	if !command.IsSuperAdmin() {
		return nil, 0, ErrUnauthorizedAccess
	}

	return gw, count, nil
}

func (s *GatewayManagementService) GetGatewaysByTenant(command GetGatewaysByTenantCommand) ([]Gateway, uint, error) {
	if command.TenantId == uuid.Nil {
		return nil, 0, ErrGatewayNotFound
	}

	tenantFound, err := s.getTenantPort.GetTenant(command.TenantId)
	if err != nil {
		return nil, 0, err
	}

	if tenantFound.IsZero() {
		return nil, 0, tenant.ErrTenantNotFound
	}

	superAdminAccess := command.IsSuperAdmin() && tenantFound.CanImpersonate

	if !superAdminAccess && !command.CanTenantAdminAccess(command.TenantId) {
		return nil, 0, ErrUnauthorizedAccess
	}

	tenantGateways, count, err := s.getGatewaysPort.GetByTenantId(command.TenantId, command.Page, command.Limit)
	if err != nil {
		return nil, 0, err
	}

	if tenantGateways == nil {
		return nil, 0, ErrGatewayNotFound
	}

	return tenantGateways, count, nil
}

func (s *GatewayManagementService) GetGatewayByTenantID(command GetGatewayByTenantIDCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetGatewayByTenantID(command.TenantId, command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	tenantFound, err := s.getTenantPort.GetTenant(command.TenantId)
	if err != nil {
		return Gateway{}, err
	}

	if tenantFound.IsZero() {
		return Gateway{}, tenant.ErrTenantNotFound
	}

	superAdminAccess := command.IsSuperAdmin() && tenantFound.CanImpersonate

	if !superAdminAccess && !command.CanTenantAdminAccess(command.TenantId) {
		return Gateway{}, ErrUnauthorizedAccess
	}

	return gw, nil
}

var (
	_ GetGatewayUseCase           = (*GatewayManagementService)(nil)
	_ GetAllGatewaysUseCase       = (*GatewayManagementService)(nil)
	_ GetGatewaysByTenantUseCase  = (*GatewayManagementService)(nil)
	_ GetGatewayByTenantIDUseCase = (*GatewayManagementService)(nil)
)

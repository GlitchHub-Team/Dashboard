package gateway

import (
<<<<<<< issue-130
	"crypto/rand"
	"encoding/base64"

	"backend/internal/shared/identity"
	"backend/internal/tenant"

=======
>>>>>>> main
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Use Cases ------------------------------------------------------------------------------------------

type GatewayManagementService struct {
<<<<<<< issue-130
	log               *zap.Logger
	saveGatewayPort   SaveGatewayPort
	removeGatewayPort RemoveGatewayPort
	getGatewayPort    GetGatewayPort
	getGatewaysPort   GetGatewaysPort
	getTenantPort     tenant.GetTenantPort
=======
	log             *zap.Logger
	getGatewayPort  GetGatewayPort
	getGatewaysPort GetGatewaysPort
>>>>>>> main
}

func NewGatewayManagementService(
	log *zap.Logger,
	getPort GetGatewayPort,
	getManyPort GetGatewaysPort,
	getTenantPort tenant.GetTenantPort,
) *GatewayManagementService {
	return &GatewayManagementService{
<<<<<<< issue-130
		log:               log,
		saveGatewayPort:   savePort,
		removeGatewayPort: removePort,
		getGatewayPort:    getPort,
		getGatewaysPort:   getManyPort,
		getTenantPort:     getTenantPort,
	}
}

func GenerateGatewaySecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// TODO: perché non viene salvato l'interval limit???
func (s *GatewayManagementService) CreateGateway(command CreateGatewayCommand) (Gateway, error) {
	if !command.IsSuperAdmin() {
		return Gateway{}, identity.ErrUnauthorizedAccess
	}

	secret, err := GenerateGatewaySecret()
	if err != nil {
		return Gateway{}, err
	}

	s.log.Info("Created gateway with name " + command.Name)

	gateway := Gateway{
		Id:               uuid.New(),
		Name:             command.Name,
		Status:           GATEWAY_STATUS_ACTIVE,
		PublicIdentifier: command.PublicIdentifier,
		SigningSecret:    secret,
	}

	_, err = s.saveGatewayPort.Save(gateway)
	if err != nil {
		return Gateway{}, err
=======
		log:             log,
		getGatewayPort:  getPort,
		getGatewaysPort: getManyPort,
>>>>>>> main
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

<<<<<<< issue-130
	return oldGateway, nil
}

/*  =================================   */

func (s *GatewayManagementService) GetGateway(command GetGatewayByIdCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if !command.IsSuperAdmin() {
		return Gateway{}, ErrUnauthorizedAccess
=======
	if !command.IsSuperAdmin() {
		if command.RequesterTenantId == nil || !gw.BelongsToTenant(*command.RequesterTenantId) {
			return Gateway{}, ErrUnauthorizedAccess
		}
>>>>>>> main
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

	if tenantFound.IsZero() {
		return nil, 0, tenant.ErrTenantNotFound
	}

	superAdminAccess := command.Requester.IsSuperAdmin() && tenantFound.CanImpersonate

	if !superAdminAccess && !command.Requester.CanTenantAdminAccess(command.TenantId) {
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

func (s *GatewayManagementService) GetGatewayByTenantId(command GetGatewayByTenantIDCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetGatewayByTenantID(command.GatewayId, command.TenantId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	tenantFound, err := s.getTenantPort.GetTenant(command.TenantId)

	if tenantFound.IsZero() {
		return Gateway{}, tenant.ErrTenantNotFound
	}

	superAdminAccess := command.Requester.IsSuperAdmin() && tenantFound.CanImpersonate

	if !superAdminAccess && !command.Requester.CanTenantAdminAccess(command.TenantId) {
		return Gateway{}, ErrUnauthorizedAccess
	}

	return gw, nil
}

var (
	_ GetGatewayUseCase          = (*GatewayManagementService)(nil)
	_ GetAllGatewaysUseCase      = (*GatewayManagementService)(nil)
	_ GetGatewaysByTenantUseCase = (*GatewayManagementService)(nil)
)

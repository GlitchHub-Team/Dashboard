package gateway

import (
	"backend/internal/shared/identity"
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Use Cases ------------------------------------------------------------------------------------------

type GatewayManagementService struct {
	log               *zap.Logger
	saveGatewayPort   SaveGatewayPort
	removeGatewayPort RemoveGatewayPort
	getGatewayPort    GetGatewayPort
	getGatewaysPort   GetGatewaysPort
}

func NewGatewayManagementService(
	log *zap.Logger,
	savePort SaveGatewayPort,
	removePort RemoveGatewayPort,
	getPort GetGatewayPort,
	getManyPort GetGatewaysPort,

) *GatewayManagementService {
	return &GatewayManagementService{
		log:               log,
		saveGatewayPort:   savePort,
		removeGatewayPort: removePort,
		getGatewayPort:    getPort,
		getGatewaysPort:   getManyPort,
	}
}

func GenerateGatewaySecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

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
	}

	return gateway, nil
}

func (s *GatewayManagementService) DeleteGateway(command DeleteGatewayCommand) (Gateway, error) {
	oldGateway, err := s.removeGatewayPort.Remove(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if oldGateway.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	return oldGateway, nil
}

func (s *GatewayManagementService) GetGateway(command GetGatewayByIdCommand) (Gateway, error) {
	gw, err := s.getGatewayPort.GetById(command.GatewayId.String())
	if err != nil {
		return Gateway{}, err
	}

	if !command.IsSuperAdmin() && !gw.BelongsToTenant(*gw.TenantId) {
		return Gateway{}, ErrUnauthorizedAccess
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

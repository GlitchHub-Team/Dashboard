package gateway

import (
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/google/uuid"
)

type GatewayCommandPort interface {
	SendCreateGateway(gatewayId uuid.UUID, interval int64) error
	SendDeleteGateway(gatewayId uuid.UUID) error
	SendCommission(gatewayId uuid.UUID, tenantId uuid.UUID, token string) error
	SendDecommission(gatewayId uuid.UUID) error
	SendInterrupt(gatewayId uuid.UUID) error
	SendResume(gatewayId uuid.UUID) error
	SendReset(gatewayId uuid.UUID) error
	SendReboot(gatewayId uuid.UUID) error
}

type GatewayCommandService struct {
	createGatewayPort  CreateGatewayPort
	removeGatewayPort  DeleteGatewayPort
	getTenantPort      tenant.GetTenantPort
	getGatewayPort     GetGatewayPort
	saveGatewayPort    SaveGatewayPort
	gatewayCommandPort GatewayCommandPort
}

func NewGatewayCommandService(
	createGatewayPort CreateGatewayPort,
	removeGatewayPort DeleteGatewayPort,
	getGatewayPort GetGatewayPort,
	getTenantPort tenant.GetTenantPort,
	saveGatewayPort SaveGatewayPort,
	gatewayCommandPort GatewayCommandPort,
) *GatewayCommandService {
	return &GatewayCommandService{
		createGatewayPort:  createGatewayPort,
		removeGatewayPort:  removeGatewayPort,
		getGatewayPort:     getGatewayPort,
		getTenantPort:      getTenantPort,
		saveGatewayPort:    saveGatewayPort,
		gatewayCommandPort: gatewayCommandPort,
	}
}

func (s *GatewayCommandService) CommissionGateway(command CommissionGatewayCommand) (Gateway, error) {
	// Controllo che sia super admin
	if !command.IsSuperAdmin() {
		return Gateway{}, identity.ErrUnauthorizedAccess
	}

	// Controllo che il gateway esista
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	// Controllo che il gateway non sia già commissionato
	if gw.TenantId != nil {
		return Gateway{}, ErrGatewayAlreadyCommissioned
	}

	// Controllo che il tenant esista
	tenantFound, err := s.getTenantPort.GetTenant(command.TenantId)
	if err != nil {
		return Gateway{}, err
	}

	if tenantFound.IsZero() {
		return Gateway{}, tenant.ErrTenantNotFound
	}

	// Invio comando di commissionamento al gateway simulato
	err = s.gatewayCommandPort.SendCommission(gw.Id, command.TenantId, command.CommissionToken)
	if err != nil {
		return Gateway{}, err
	}

	// Aggiorno il gateway
	gw.Status = GATEWAY_STATUS_ACTIVE
	gw.TenantId = &command.TenantId

	return s.saveGatewayPort.Save(gw)
}

func (s *GatewayCommandService) DecommissionGateway(command DecommissionGatewayCommand) (Gateway, error) {
	// Controllo che sia super admin
	if !command.IsSuperAdmin() {
		return Gateway{}, identity.ErrUnauthorizedAccess
	}

	// Controllo che il gateway esista
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	// Controllo che il gateway sia commissionato
	if gw.TenantId == nil {
		return Gateway{}, ErrGatewayNotCommissioned
	}

	// Invio comando di decommissionamento al gateway simulato
	err = s.gatewayCommandPort.SendDecommission(gw.Id)
	if err != nil {
		return Gateway{}, err
	}

	// Aggiorno il gateway
	gw.Status = GATEWAY_STATUS_DECOMMISSIONED
	gw.TenantId = nil

	return s.saveGatewayPort.Save(gw)
}

func (s *GatewayCommandService) InterruptGateway(command InterruptGatewayCommand) (Gateway, error) {
	// Controllo che il gateway esista
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	// Controllo che l'utente abbia i permessi per interrompere il gateway
	if !command.IsSuperAdmin() {
		if command.RequesterTenantId == nil || !gw.BelongsToTenant(*command.RequesterTenantId) || command.RequesterRole != identity.ROLE_TENANT_ADMIN {
			return Gateway{}, identity.ErrUnauthorizedAccess
		}
	}

	// Controllo che il gateway non sia già interrotto o decommissionato
	if gw.Status != GATEWAY_STATUS_ACTIVE {
		return Gateway{}, ErrGatewayNotActive
	}

	err = s.gatewayCommandPort.SendInterrupt(gw.Id)
	if err != nil {
		return Gateway{}, err
	}

	gw.Status = GATEWAY_STATUS_INACTIVE
	return s.saveGatewayPort.Save(gw)
}

func (s *GatewayCommandService) ResumeGateway(command ResumeGatewayCommand) (Gateway, error) {
	// Controllo che il gateway esista
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	// Controllo che l'utente abbia i permessi per riattivare il gateway
	if !command.IsSuperAdmin() {
		if command.RequesterTenantId == nil || !gw.BelongsToTenant(*command.RequesterTenantId) || command.RequesterRole != identity.ROLE_TENANT_ADMIN {
			return Gateway{}, identity.ErrUnauthorizedAccess
		}
	}

	// Controllo che il gateway non sia attivo o decommissionato
	if gw.Status != GATEWAY_STATUS_INACTIVE {
		return Gateway{}, ErrGatewayNotInactive
	}

	// Invio comando di riattivazione al gateway simulato
	if err := s.gatewayCommandPort.SendResume(gw.Id); err != nil {
		return Gateway{}, err
	}

	gw.Status = GATEWAY_STATUS_ACTIVE
	return s.saveGatewayPort.Save(gw)
}

func (s *GatewayCommandService) ResetGateway(command ResetGatewayCommand) (Gateway, error) {
	// Controllo che il gateway esista
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	// Controllo che l'utente abbia i permessi per resettare il gateway
	if !command.IsSuperAdmin() {
		if command.RequesterTenantId == nil || !gw.BelongsToTenant(*command.RequesterTenantId) || command.RequesterRole != identity.ROLE_TENANT_ADMIN {
			return Gateway{}, identity.ErrUnauthorizedAccess
		}
	}

	// Invio comando di reset al gateway simulato
	if err := s.gatewayCommandPort.SendReset(gw.Id); err != nil {
		return Gateway{}, err
	}

	gw.IntervalLimit = DEFAULT_INTERVAL_LIMIT
	return s.saveGatewayPort.Save(gw)
}

func (s *GatewayCommandService) RebootGateway(command RebootGatewayCommand) (Gateway, error) {
	// Controllo che il gateway esista
	gw, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if gw.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	// Controllo che l'utente abbia i permessi per riavviare il gateway
	if !command.IsSuperAdmin() {
		if command.RequesterTenantId == nil || !gw.BelongsToTenant(*command.RequesterTenantId) || command.RequesterRole != identity.ROLE_TENANT_ADMIN {
			return Gateway{}, identity.ErrUnauthorizedAccess
		}
	}

	// Invio comando di riavvio al gateway simulato
	if err := s.gatewayCommandPort.SendReboot(gw.Id); err != nil {
		return Gateway{}, err
	}

	return gw, nil
}

func (s *GatewayCommandService) CreateGateway(command CreateGatewayCommand) (Gateway, error) {
	// Controllo che sia super admin
	if !command.IsSuperAdmin() {
		return Gateway{}, identity.ErrUnauthorizedAccess
	}

	gateway := Gateway{
		Id:               uuid.New(),
		Name:             command.Name,
		IntervalLimit:    command.Interval,
		Status:           GATEWAY_STATUS_DECOMMISSIONED,
		PublicIdentifier: nil,
		TenantId:         nil,
	}

	// Invio comando di creazione al gateway simulato
	if err := s.gatewayCommandPort.SendCreateGateway(gateway.Id, gateway.IntervalLimit.Milliseconds()); err != nil {
		return Gateway{}, err
	}

	// Salvo il gateway nel database
	return s.createGatewayPort.Create(gateway)
}

func (s *GatewayCommandService) DeleteGateway(command DeleteGatewayCommand) (Gateway, error) {
	// Controllo che l'utente sia super admin
	if !command.IsSuperAdmin() {
		return Gateway{}, identity.ErrUnauthorizedAccess
	}

	// Controllo che il gateway esista
	oldGateway, err := s.getGatewayPort.GetById(command.GatewayId)
	if err != nil {
		return Gateway{}, err
	}

	if oldGateway.IsZero() {
		return Gateway{}, ErrGatewayNotFound
	}

	// Invio comando di cancellazione al gateway simulato
	if err := s.gatewayCommandPort.SendDeleteGateway(oldGateway.Id); err != nil {
		return Gateway{}, err
	}

	// Rimuovo il gateway dal database
	return s.removeGatewayPort.Delete(oldGateway.Id)
}

var (
	_ CreateGatewayUseCase       = (*GatewayCommandService)(nil)
	_ DeleteGatewayUseCase       = (*GatewayCommandService)(nil)
	_ CommissionGatewayUseCase   = (*GatewayCommandService)(nil)
	_ DecommissionGatewayUseCase = (*GatewayCommandService)(nil)
	_ InterruptGatewayUseCase    = (*GatewayCommandService)(nil)
	_ ResumeGatewayUseCase       = (*GatewayCommandService)(nil)
	_ ResetGatewayUseCase        = (*GatewayCommandService)(nil)
	_ RebootGatewayUseCase       = (*GatewayCommandService)(nil)
)

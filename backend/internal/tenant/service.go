package tenant

import (
	"backend/internal/shared/identity"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/tenant/mocks/ports.go -package=mocks . CreateTenantPort,DeleteTenantPort,GetTenantPort,GetTenantsPort,GetTenantByUserPort

type CreateTenantPort interface {
	CreateTenant(tenant Tenant) (Tenant, error)
}

type DeleteTenantPort interface {
	DeleteTenant(tenantId uuid.UUID) (Tenant, error)
}

type GetTenantPort interface {
	GetTenant(tenantId uuid.UUID) (Tenant, error)
}

type GetTenantsPort interface {
	GetTenants(page, limit int) ([]Tenant, uint, error)
	GetAllTenants() ([]Tenant, error)
}


type TenantService struct {
	createTenantPort    CreateTenantPort
	deleteTenantPort    DeleteTenantPort
	getTenantPort       GetTenantPort
	getTenantsPort      GetTenantsPort
}

func NewCreateTenantService(
	createTenantPort CreateTenantPort,
	deleteTenantPort DeleteTenantPort,
	getTenantPort GetTenantPort,
	getTenantsPort GetTenantsPort,
) *TenantService {
	return &TenantService{
		createTenantPort:    createTenantPort,
		deleteTenantPort:    deleteTenantPort,
		getTenantPort:       getTenantPort,
		getTenantsPort:      getTenantsPort,
	}
}

/*
	Crea un tenant come specificato in cmd
*/
func (service *TenantService) CreateTenant(cmd CreateTenantCommand) (Tenant, error) {
	if !cmd.IsSuperAdmin() {
		return Tenant{}, identity.ErrUnauthorizedAccess
	}

	newTenantId := uuid.New()

	newTenant := Tenant{
		Id:             newTenantId,
		Name:           cmd.Name,
		CanImpersonate: cmd.CanImpersonate,
	}

	tenant, err := service.createTenantPort.CreateTenant(newTenant)
	if err != nil {
		return Tenant{}, err
	}

	return tenant, nil
}

/*
	Elimina un tenant come specificato in cmd
*/
func (service *TenantService) DeleteTenant(cmd DeleteTenantCommand) (Tenant, error) {
	if !cmd.IsSuperAdmin() {
		return Tenant{}, identity.ErrUnauthorizedAccess
	}

	_, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return Tenant{}, err
	}

	oldTenant, err := service.deleteTenantPort.DeleteTenant(cmd.TenantId)
	if err != nil {
		return Tenant{}, err
	}
	return oldTenant, err
}

/*
	Ottiene un tenant per TenantId come specificato in cmd
*/
func (service *TenantService) GetTenant(cmd GetTenantCommand) (Tenant, error) {
	tenant, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return Tenant{}, err
	}

	if tenant.IsZero() {
		return Tenant{}, ErrTenantNotFound
	}

	if !cmd.IsSuperAdmin() && !cmd.CanTenantAdminAccess(tenant.Id) {
		return Tenant{}, identity.ErrUnauthorizedAccess
	}

	return tenant, nil
}

/*
	Ottiene la lista NON PAGINATA di tutti i tenant presenti
*/
func (service *TenantService) GetAllTenants() ([]Tenant, error) {
	return service.getTenantsPort.GetAllTenants()
}

/*
	Ottiene la lista paginata dei tenant come specificato in cmd
*/
func (service *TenantService) GetTenantList(cmd GetTenantListCommand) ([]Tenant, uint, error) {
	if !cmd.IsSuperAdmin() {
		return nil, 0, identity.ErrUnauthorizedAccess
	}

	tenants, total, err := service.getTenantsPort.GetTenants(cmd.Page, cmd.Limit)
	if err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}


// Compile-time checks
var (
	_ CreateTenantUseCase  = (*TenantService)(nil)
	_ DeleteTenantUseCase  = (*TenantService)(nil)
	_ GetTenantUseCase     = (*TenantService)(nil)
	_ GetTenantListUseCase = (*TenantService)(nil)
	_ GetAllTenantsUseCase = (*TenantService)(nil)
)

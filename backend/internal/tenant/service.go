package tenant

import "github.com/google/uuid"

//go:generate mockgen -destination=../../tests/tenant/mocks/use_cases_crud.go -package=mocks . CreateTenantUseCase,DeleteTenantUseCase,GetTenantUseCase,GetTenantListUseCase,GetTenantByUserUseCase

type CreateTenantPort interface {
	CreateTenant(tenant Tenant) (Tenant, error)
}
type CreateTenantUseCase interface {
	CreateTenant(cmd CreateTenantCommand) (Tenant, error)
}

type DeleteTenantUseCase interface {
	DeleteTenant(cmd DeleteTenantCommand) (Tenant, error)
}

type GetTenantUseCase interface {
	GetTenant(cmd GetTenantCommand) (Tenant, error)
}

type GetTenantListUseCase interface {
	GetTenantList(cmd GetTenantListCommand) ([]Tenant, uint, error)
}

type GetTenantByUserUseCase interface {
	GetTenantByUser(cmd GetTenantByUserCommand) (Tenant, error)
}

type TenantService struct {
	createTenantPort    CreateTenantPort
	deleteTenantPort    DeleteTenantPort
	getTenantPort       GetTenantPort
	getTenantsPort      GetTenantsPort
	getTenantByUserPort GetTenantByUserPort
}

func NewCreateTenantService(
	createTenantPort CreateTenantPort,
	deleteTenantPort DeleteTenantPort,
	getTenantPort GetTenantPort,
	getTenantsPort GetTenantsPort,
	getTenantByUserPort GetTenantByUserPort,
) (CreateTenantUseCase, DeleteTenantUseCase, GetTenantUseCase, GetTenantListUseCase, GetTenantByUserUseCase) {
	service := &TenantService{
		createTenantPort:    createTenantPort,
		deleteTenantPort:    deleteTenantPort,
		getTenantPort:       getTenantPort,
		getTenantsPort:      getTenantsPort,
		getTenantByUserPort: getTenantByUserPort,
	}
	return service, service, service, service, service
}

func (service *TenantService) CreateTenant(cmd CreateTenantCommand) (Tenant, error) {
	if !cmd.IsSuperAdmin() {
		return Tenant{}, ErrUnauthorized
	}

	newTenantId := uuid.New()

	newTenant := Tenant{
		Id: newTenantId,
		Name:           cmd.Name,
		CanImpersonate: cmd.CanImpersonate,
	}

	tenant, err := service.createTenantPort.CreateTenant(newTenant)
	if err != nil {
		return Tenant{}, err
	}

	return tenant, nil
}

func (service *TenantService) DeleteTenant(cmd DeleteTenantCommand) (Tenant, error) {
	tenant, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return Tenant{}, err
	}

	if tenant.IsZero() {
		return Tenant{}, ErrTenantNotFound
	}

	if !cmd.IsSuperAdmin() {
		return Tenant{}, ErrUnauthorized
	}

	if !tenant.CanImpersonate {
		return Tenant{}, ErrImpersonationFailed
	}

	oldTenant, err := service.deleteTenantPort.DeleteTenant(cmd.TenantId)

	return oldTenant, err
}

func (service *TenantService) GetTenant(cmd GetTenantCommand) (Tenant, error) {
	tenant, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return Tenant{}, err
	}

	if tenant.IsZero() {
		return Tenant{}, ErrTenantNotFound
	}

	if !cmd.IsSuperAdmin() && !cmd.CanTenantAdminAccess(tenant.Id) {
		return Tenant{}, ErrUnauthorized
	}

	return tenant, nil
}

func (service *TenantService) GetTenantList(cmd GetTenantListCommand) ([]Tenant, uint, error) {
	tenants, total, err := service.getTenantsPort.GetTenants()
	if err != nil {
		return nil, 0, err
	}

	if tenants == nil {
		tenants = make([]Tenant, 0)
	}

	return tenants, total, nil
}

func (service *TenantService) GetTenantByUser(cmd GetTenantByUserCommand) (Tenant, error) {
	tenant, err := service.getTenantByUserPort.GetTenantByUser(cmd.UserId)
	if err != nil {
		return Tenant{}, err
	}

	if tenant.IsZero() {
		return Tenant{}, ErrTenantNotFound
	}

	if !cmd.IsSuperAdmin() && !cmd.CanTenantAdminAccess(tenant.Id) {
		return Tenant{}, ErrUnauthorized
	}

	return tenant, nil
}

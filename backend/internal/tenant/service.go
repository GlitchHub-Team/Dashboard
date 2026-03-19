package tenant

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
	GetTenantList(cmd GetTenantListCommand) ([]Tenant, error)
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
	
	if !cmd.Requester.IsSuperAdmin() {
		return Tenant{}, ErrUnauthorized
	}

	if !cmd.CanImpersonate {
		return Tenant{}, ErrImpersonationFailded
	}

	newTenant := Tenant{
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

	if !cmd.Requester.IsSuperAdmin() {
		return Tenant{}, ErrUnauthorized
	}

	if !tenant.CanImpersonate {
		return Tenant{}, ErrImpersonationFailded
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

	if !cmd.Requester.IsSuperAdmin() && !cmd.Requester.CanTenantAdminAccess(tenant.Id) {
		return Tenant{}, ErrUnauthorized
	}

	return tenant, nil
}

func (service *TenantService) GetTenantList(cmd GetTenantListCommand) ([]Tenant, error) {

	if !cmd.Requester.IsSuperAdmin() {
		return nil, ErrUnauthorized
	}

	tenants, err := service.getTenantsPort.GetTenants()

	if err != nil {
		return nil, err
	}

	for _, tenant := range tenants {
		if !tenant.CanImpersonate {
			return nil, ErrUnauthorized
		}
	}

	return tenants, nil
}

func (service *TenantService) GetTenantByUser(cmd GetTenantByUserCommand) (Tenant, error) {
	tenant, err := service.getTenantByUserPort.GetTenantByUser(cmd.UserId)
	if err != nil {
		return Tenant{}, err
	}

	if tenant.IsZero() {
		return Tenant{}, ErrTenantNotFound
	}

	if !cmd.Requester.IsSuperAdmin() && !cmd.Requester.CanTenantAdminAccess(tenant.Id) {
		return Tenant{}, ErrUnauthorized
	}


	return tenant, nil
}

package user

import (
	"backend/internal/shared/identity"
	"backend/internal/tenant"
)

/*
Servizio usato per ottenere informazioni su uno o più utenti.

Possibile miglioria: Validare l'input, non affidandosi a validazione in controller
*/
type GetUserService struct {
	getUserPort   GetUserPort
	getTenantPort tenant.GetTenantPort
}

func NewGetUserService(getUserPort GetUserPort, getTenantPort tenant.GetTenantPort) (
	GetTenantUserUseCase,
	GetTenantAdminUseCase,
	GetSuperAdminUseCase,

	GetTenantUsersByTenantUseCase,
	GetTenantAdminsByTenantUseCase,
	GetSuperAdminListUseCase,
) {
	service := &GetUserService{
		getUserPort:   getUserPort,
		getTenantPort: getTenantPort,
	}
	return service, service, service, service, service, service
}

// Get single -----------------------------------------------------------------------------------------

func (service *GetUserService) GetTenantUser(cmd GetTenantUserCommand) (User, error) {
	// TODO: Ottimizzare controllo autorizz. (metti qua controllo per tenant user/admin)

	// 1) Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// Controlla autorizzazione
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate                           //nolint:staticcheck
	tenantAdminAccess := cmd.Requester.CanTenantAdminAccess(cmd.TenantId)                                    //nolint:staticcheck
	tenantUserAccess := cmd.Requester.CanTenantUserAccess(cmd.TenantId) && cmd.RequesterUserId == cmd.UserId //nolint:staticcheck
	if !superAdminAccess && !tenantAdminAccess && !tenantUserAccess {                                        //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 2) Get tenant user
	user, err := service.getUserPort.GetTenantUser(cmd.TenantId, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	return user, nil
}

func (service *GetUserService) GetTenantAdmin(cmd GetTenantAdminCommand) (User, error) {
	// TODO: Ottimizzare controllo autorizz. (metti qua controllo per tenant user/admin)

	// 1) Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// Controlla autorizzazione
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate //nolint:staticcheck
	tenantAdminAccess := cmd.Requester.CanTenantAdminAccess(cmd.TenantId)          //nolint:staticcheck
	if !superAdminAccess && !tenantAdminAccess {
		return User{}, identity.ErrUnauthorizedAccess
	}
	// 2) Get tenant admin
	user, err := service.getUserPort.GetTenantAdmin(cmd.TenantId, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

func (service *GetUserService) GetSuperAdmin(cmd GetSuperAdminCommand) (User, error) {
	// Controlla autorizzazione
	// NOTA: rimosso static check per chiarezza
	if !cmd.Requester.IsSuperAdmin() { //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 1) Get super admin
	user, err := service.getUserPort.GetSuperAdmin(cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

// Get multiple ---------------------------------------------------------------------------------------

func (service *GetUserService) GetTenantUsersByTenant(cmd GetTenantUsersByTenantCommand) (
	tenantUsers []User, total uint, err error,
) {
	// TODO: Ottimizzare controllo autorizz. (metti qua controllo per tenant user/admin)

	// 1) Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return nil, 0, err
	}
	if tenantFound.IsZero() {
		return nil, 0, tenant.ErrTenantNotFound
	}

	// Controlla autorizzazione
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate //nolint:staticcheck
	tenantAdminAccess := cmd.Requester.CanTenantAdminAccess(cmd.TenantId)          //nolint:staticcheck
	if !superAdminAccess && !tenantAdminAccess {                                   //nolint:staticcheck
		return nil, 0, identity.ErrUnauthorizedAccess
	}

	// 2) Get tenant users
	tenantUsers, total, err = service.getUserPort.GetTenantUsersByTenant(cmd.TenantId, cmd.Page, cmd.Limit)
	if err != nil {
		return nil, 0, err
	}

	return tenantUsers, total, nil
}

func (service *GetUserService) GetTenantAdminsByTenant(cmd GetTenantAdminsByTenantCommand) (
	tenantUsers []User, total uint, err error,
) {
	// TODO: Ottimizzare controllo autorizz. (metti qua controllo per tenant user/admin)
	// 1) Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return nil, 0, err
	}
	if tenantFound.IsZero() {
		return nil, 0, tenant.ErrTenantNotFound
	}

	// 2) Controlla autorizzazione
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate //nolint:staticcheck
	tenantAdminAccess := cmd.Requester.CanTenantAdminAccess(cmd.TenantId)          //nolint:staticcheck
	if !superAdminAccess && !tenantAdminAccess {                                   //nolint:staticcheck
		return nil, 0, identity.ErrUnauthorizedAccess
	}

	// 3) Get tenant users
	tenantUsers, total, err = service.getUserPort.GetTenantAdminsByTenant(cmd.TenantId, cmd.Page, cmd.Limit)
	if err != nil {
		return nil, 0, err
	}

	return tenantUsers, total, nil
}

func (service *GetUserService) GetSuperAdminList(cmd GetSuperAdminListCommand) (
	superAdmins []User, total uint, err error,
) {
	// 1) Controlla autorizzazione
	// NOTA: rimosso static check per chiarezza
	if !cmd.Requester.IsSuperAdmin() { //nolint:staticcheck
		return nil, 0, identity.ErrUnauthorizedAccess
	}

	// 2) Get super admin
	superAdmins, total, err = service.getUserPort.GetSuperAdminList(cmd.Page, cmd.Limit)
	if err != nil {
		return nil, 0, err
	}
	return superAdmins, total, nil
}

// Compile-time checks
var (
	_ GetTenantUserUseCase           = (*GetUserService)(nil)
	_ GetTenantAdminUseCase          = (*GetUserService)(nil)
	_ GetSuperAdminUseCase           = (*GetUserService)(nil)
	_ GetTenantUsersByTenantUseCase  = (*GetUserService)(nil)
	_ GetTenantAdminsByTenantUseCase = (*GetUserService)(nil)
	_ GetSuperAdminListUseCase       = (*GetUserService)(nil)
)

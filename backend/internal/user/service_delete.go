package user

import (
	"backend/internal/shared/identity"
	"backend/internal/tenant"
)

// Delete User ====================================================================================
/*
	Servizio di eliminazione utente.
*/
type DeleteUserService struct {
	deleteUserPort DeleteUserPort
	getUserPort    GetUserPort
	getTenantPort  tenant.GetTenantPort
}

func NewDeleteUserService(
	deleteUserPort DeleteUserPort,
	getUserPort GetUserPort,
	getTenantPort tenant.GetTenantPort,
) *DeleteUserService {
	service := &DeleteUserService{
		deleteUserPort: deleteUserPort,
		getUserPort:    getUserPort,
		getTenantPort:  getTenantPort,
	}
	return service
}

func (service *DeleteUserService) DeleteTenantUser(cmd DeleteTenantUserCommand) (User, error) {
	// TODO: Ottimizzare controllo autorizz. (metti qua controllo per tenant user/admin)

	// 1. Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// Controlla autorizzazione tenant
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate //nolint:staticcheck
	if !superAdminAccess && !cmd.Requester.CanTenantAdminAccess(cmd.TenantId) {    //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 2. Controlla user
	user, err := service.getUserPort.GetUser(&cmd.TenantId, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	// 3. Elimina user
	oldUser, err := service.deleteUserPort.DeleteTenantUser(cmd.TenantId, cmd.UserId)
	return oldUser, err
}

func (service *DeleteUserService) DeleteTenantAdmin(cmd DeleteTenantAdminCommand) (User, error) {
	// TODO: Ottimizzare controllo autorizz. (metti qua controllo per tenant user/admin)

	// 1. Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// Controlla autorizzazione tenant
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin()                            //nolint:staticcheck
	if !superAdminAccess && !cmd.Requester.CanTenantAdminAccess(cmd.TenantId) { //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 2. Controlla user
	user, err := service.getUserPort.GetUser(&cmd.TenantId, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	// 3. Controlla che non sia l'ultimo tenant admin
	total, err := service.getUserPort.CountTenantAdminsByTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if total <= 1 {
		return User{}, ErrCannotDeleteLastAdmin
	}

	// 4. Elimina user
	oldUser, err := service.deleteUserPort.DeleteTenantAdmin(cmd.TenantId, cmd.UserId)
	return oldUser, err
}

func (service *DeleteUserService) DeleteSuperAdmin(cmd DeleteSuperAdminCommand) (User, error) {
	// NOTA: rimosso static check per chiarezza
	if !cmd.Requester.IsSuperAdmin() { //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 1. Controlla user
	user, err := service.getUserPort.GetUser(nil, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	// 2. Controlla che non sia l'ultimo tenant admin
	total, err := service.getUserPort.CountSuperAdmins()
	if err != nil {
		return User{}, err
	}
	if total == 1 {
		return User{}, ErrCannotDeleteLastAdmin
	}

	// 3. Elimina user
	oldUser, err := service.deleteUserPort.DeleteSuperAdmin(cmd.UserId)
	return oldUser, err
}

// Compile-time checks
var (
	_ DeleteTenantUserUseCase  = (*DeleteUserService)(nil)
	_ DeleteTenantAdminUseCase = (*DeleteUserService)(nil)
	_ DeleteSuperAdminUseCase  = (*DeleteUserService)(nil)
)

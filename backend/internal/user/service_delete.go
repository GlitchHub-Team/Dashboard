package user

import (
	"backend/internal/shared/identity"
	"backend/internal/tenant"
)

// Delete User ====================================================================================
type DeleteUserService struct {
	deleteUserPort DeleteUserPort
	getUserPort    GetUserPort
	getTenantPort  tenant.GetTenantPort
}

func NewDeleteUserService(
	deleteUserPort DeleteUserPort,
	getUserPort GetUserPort,
	getTenantPort tenant.GetTenantPort,
) (DeleteTenantUserUseCase, DeleteTenantAdminUseCase, DeleteSuperAdminUseCase) {
	service := &DeleteUserService{
		deleteUserPort: deleteUserPort,
		getUserPort:    getUserPort,
		getTenantPort:  getTenantPort,
	}
	return service, service, service
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
	user, err := service.getUserPort.GetTenantUser(cmd.TenantId, cmd.UserId)
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

	// 2. Controlla autorizzazione tenant
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate //nolint:staticcheck
	if !superAdminAccess && !cmd.Requester.CanTenantAdminAccess(cmd.TenantId) {    //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 3. Controlla user
	user, err := service.getUserPort.GetTenantAdmin(cmd.TenantId, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
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
	user, err := service.getUserPort.GetSuperAdmin(cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	// 2. Elimina user
	oldUser, err := service.deleteUserPort.DeleteSuperAdmin(cmd.UserId)
	return oldUser, err
}

// Compile-time checks
var (
	_ DeleteTenantUserUseCase  = (*DeleteUserService)(nil)
	_ DeleteTenantAdminUseCase = (*DeleteUserService)(nil)
	_ DeleteSuperAdminUseCase  = (*DeleteUserService)(nil)
)

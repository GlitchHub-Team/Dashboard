package user

import (
	"backend/internal/tenant"
)

// Delete User ====================================================================================
type DeleteTenantUserUseCase interface {
	DeleteTenantUser(cmd DeleteTenantUserCommand) (User, error)
}

type DeleteTenantAdminUseCase interface {
	DeleteTenantAdmin(cmd DeleteTenantAdminCommand) (User, error)
}

type DeleteSuperAdminUseCase interface {
	DeleteSuperAdmin(cmd DeleteSuperAdminCommand) (User, error)
}

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
	// Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// Controlla user
	user, err := service.getUserPort.GetTenantUser(cmd.TenantId, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	// Elimina user
	oldUser, err := service.deleteUserPort.DeleteTenantUser(cmd.TenantId, cmd.UserId)
	return oldUser, err
}

func (service *DeleteUserService) DeleteTenantAdmin(cmd DeleteTenantAdminCommand) (User, error) {
	// Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// Controlla user
	user, err := service.getUserPort.GetTenantAdmin(cmd.TenantId, cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	// Elimina user
	oldUser, err := service.deleteUserPort.DeleteTenantAdmin(cmd.TenantId, cmd.UserId)
	return oldUser, err
}

func (service *DeleteUserService) DeleteSuperAdmin(cmd DeleteSuperAdminCommand) (User, error) {
	// Controlla user
	user, err := service.getUserPort.GetSuperAdmin(cmd.UserId)
	if err != nil {
		return User{}, err
	}
	if user.IsZero() {
		return User{}, ErrUserNotFound
	}

	// Elimina user
	oldUser, err := service.deleteUserPort.DeleteSuperAdmin(cmd.UserId)
	return oldUser, err
}

// Compile-time checks
var (
	_ DeleteTenantUserUseCase  = (*DeleteUserService)(nil)
	_ DeleteTenantAdminUseCase = (*DeleteUserService)(nil)
	_ DeleteSuperAdminUseCase  = (*DeleteUserService)(nil)
)

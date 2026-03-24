package user

import (
	"backend/internal/auth"
	"backend/internal/email"
	"backend/internal/identity"
	"backend/internal/tenant"
)

//go:generate mockgen -destination=../../tests/user/mocks/use_cases_create.go -package=mocks . CreateTenantUserUseCase,CreateTenantAdminUseCase,CreateSuperAdminUseCase

type CreateTenantUserUseCase interface {
	CreateTenantUser(cmd CreateTenantUserCommand) (User, error)
}

type CreateTenantAdminUseCase interface {
	CreateTenantAdmin(cmd CreateTenantAdminCommand) (User, error)
}

type CreateSuperAdminUseCase interface {
	CreateSuperAdmin(cmd CreateSuperAdminCommand) (User, error)
}

type CreateUserService struct {
	createUserPort          CreateUserPort
	deleteUserPort          DeleteUserPort
	getUserPort             GetUserPort
	getTenantPort           tenant.GetTenantPort
	confirmAccountTokenPort auth.ConfirmTokenPort
	sendEmailPort           email.SendEmailPort
}

func NewCreateUserService(
	createUserPort CreateUserPort,
	deleteUserPort DeleteUserPort,
	getUserPort GetUserPort,
	getTenantPort tenant.GetTenantPort,
	confirmAccountTokenPort auth.ConfirmTokenPort,
	sendEmailPort email.SendEmailPort,
) (CreateTenantUserUseCase, CreateTenantAdminUseCase, CreateSuperAdminUseCase) {
	service := &CreateUserService{
		createUserPort:          createUserPort,
		deleteUserPort:          deleteUserPort,
		getUserPort:             getUserPort,
		getTenantPort:           getTenantPort,
		confirmAccountTokenPort: confirmAccountTokenPort,
		sendEmailPort:           sendEmailPort,
	}
	return service, service, service
}

func (service *CreateUserService) CreateTenantUser(cmd CreateTenantUserCommand) (User, error) {
	// TODO: Ottimizzare controllo autorizz. (metti qua controllo per tenant user/admin)

	// 1. Controlla esistenza tenant
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
	checkedUser, err := service.getUserPort.GetTenantUserByEmail(cmd.TenantId, cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !checkedUser.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// 4. Crea user
	user, err := service.createUserPort.CreateUser(User{
		Name:      cmd.Username,
		Email:     cmd.Email,
		Role:      identity.ROLE_TENANT_USER,
		TenantId:  &cmd.TenantId,
		Confirmed: false,
	})
	if err != nil {
		return User{}, err
	}

	// 5. Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(user.Id)
	if err != nil {
		return User{}, err
	}

	// 6. Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(user.Email, confirmAccountToken)
	if err != nil {
		// 7. Elimina account se invio mail fallisce
		_, deletionErr := service.deleteUserPort.DeleteTenantUser(*user.TenantId, user.Id)
		if deletionErr != nil {
			return User{}, deletionErr
		}
		return User{}, ErrCannotSendEmail
	}

	// Ritorna user
	return user, nil
}

func (service *CreateUserService) CreateTenantAdmin(cmd CreateTenantAdminCommand) (User, error) {
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
	if !superAdminAccess && !cmd.CanTenantAdminAccess(cmd.TenantId) {              //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 3. Controlla user
	user, err := service.getUserPort.GetTenantAdminByEmail(cmd.TenantId, cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !user.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// 4. Crea user
	user, err = service.createUserPort.CreateUser(User{
		Name:      cmd.Username,
		Email:     cmd.Email,
		Role:      identity.ROLE_TENANT_ADMIN,
		TenantId:  &cmd.TenantId,
		Confirmed: false,
	})
	if err != nil {
		return User{}, err
	}

	// 5. Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(user.Id)
	if err != nil {
		return User{}, err
	}

	// 6. Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(user.Email, confirmAccountToken)
	if err != nil {
		// 7. Elimina account se invio mail fallisce
		_, deletionErr := service.deleteUserPort.DeleteTenantAdmin(*user.TenantId, user.Id)
		if deletionErr != nil {
			return User{}, deletionErr
		}
		return User{}, ErrCannotSendEmail
	}

	// Ritorna user
	return user, nil
}

func (service *CreateUserService) CreateSuperAdmin(cmd CreateSuperAdminCommand) (User, error) {
	// Controlla autorizzazione tenant
	// NOTA: rimosso static check per chiarezza
	if !cmd.Requester.IsSuperAdmin() { //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 1. Controlla user
	user, err := service.getUserPort.GetSuperAdminByEmail(cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !user.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// 2. Crea user
	user, err = service.createUserPort.CreateUser(User{
		Name:      cmd.Username,
		Email:     cmd.Email,
		Role:      identity.ROLE_SUPER_ADMIN,
		TenantId:  nil,
		Confirmed: false,
	})
	if err != nil {
		return User{}, err
	}

	// 3. Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(user.Id)
	if err != nil {
		return User{}, err
	}

	// 4. Invia token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(user.Email, confirmAccountToken)
	if err != nil {
		// 5. Elimina account se invio mail fallisce
		_, deletionErr := service.deleteUserPort.DeleteSuperAdmin(user.Id)
		if deletionErr != nil {
			return User{}, deletionErr
		}
		return User{}, ErrCannotSendEmail
	}

	// Ritorna user
	return user, nil
}

// Compile-time checks
var (
	_ CreateTenantUserUseCase  = (*CreateUserService)(nil)
	_ CreateTenantAdminUseCase = (*CreateUserService)(nil)
	_ CreateSuperAdminUseCase  = (*CreateUserService)(nil)
)

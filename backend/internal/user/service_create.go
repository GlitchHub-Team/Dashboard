package user

import (
	"fmt"

	"backend/internal/auth"
	"backend/internal/email"
	"backend/internal/tenant"
)

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
	// 1) Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// 2) Controlla user
	user, err := service.getUserPort.GetTenantUserByEmail(cmd.TenantId, cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !user.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// 3) Crea user
	user, err = service.createUserPort.CreateUser(User{
		Name:      cmd.Username,
		Email:     cmd.Email,
		Role:      ROLE_TENANT_USER,
		TenantId:  &cmd.TenantId,
		Confirmed: false,
	})
	if err != nil {
		return User{}, err
	}

	// 4) Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(user.Id)
	if err != nil {
		return User{}, err
	}

	// 5) Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(user.Email, confirmAccountToken)
	if err != nil {
		// 6) Elimina account se invio mail fallisce
		// TODO: Gestire eliminazione dell'account se invio di email fallisce
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
	// Controlla tenant
	tenantFound, err := service.getTenantPort.GetTenant(cmd.TenantId)
	if err != nil {
		return User{}, err
	}
	if tenantFound.IsZero() {
		return User{}, tenant.ErrTenantNotFound
	}

	// Controlla user
	user, err := service.getUserPort.GetTenantAdminByEmail(cmd.TenantId, cmd.Email)
	if err != nil {
		return User{}, fmt.Errorf("error obtaining tenant user %v @%v: %v", cmd.Email, cmd.TenantId, err)
	}
	if !user.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// Crea user
	newUser, err := service.createUserPort.CreateUser(User{
		Name:      cmd.Username,
		Email:     cmd.Email,
		Role:      ROLE_TENANT_ADMIN,
		TenantId:  &cmd.TenantId,
		Confirmed: false,
	})
	if err != nil {
		return User{}, err
	}

	// Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(newUser.Id)
	if err != nil {
		return User{}, err
	}

	// Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(newUser.Email, confirmAccountToken)
	if err != nil {
		// TODO: Come gestisco errore di invio email? Cancello account?
	}

	// Ritorna user
	return newUser, nil
}

func (service *CreateUserService) CreateSuperAdmin(cmd CreateSuperAdminCommand) (User, error) {
	// Controlla user
	user, err := service.getUserPort.GetSuperAdminByEmail(cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !user.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// Crea user
	newUser, err := service.createUserPort.CreateUser(User{
		Name:      cmd.Username,
		Email:     cmd.Email,
		Role:      ROLE_SUPER_ADMIN,
		TenantId:  nil,
		Confirmed: false,
	})
	if err != nil {
		return User{}, err
	}

	// Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(newUser.Id)
	if err != nil {
		return User{}, err
	}

	// Invia token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(newUser.Email, confirmAccountToken)
	if err != nil {
		// TODO: Come gestisco errore di invio email? Cancello account?
		return User{}, err
	}

	// Ritorna user
	return newUser, err
}

// Compile-time checks
var (
	_ CreateTenantUserUseCase  = (*CreateUserService)(nil)
	_ CreateTenantAdminUseCase = (*CreateUserService)(nil)
	_ CreateSuperAdminUseCase  = (*CreateUserService)(nil)
)

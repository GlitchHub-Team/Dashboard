package user

import (
	"backend/internal/auth"
	"backend/internal/email"
)

// Create User ====================================================================================
type CreateTenantUserUseCase interface {
	CreateTenantUser(cmd CreateTenantUserCommand) (*User, error)
}

type CreateTenantAdminUseCase interface {
	CreateTenantAdmin(cmd CreateTenantAdminCommand) (*User, error)
}

type CreateSuperAdminUseCase interface {
	CreateSuperAdmin(cmd CreateSuperAdminCommand) (*User, error)
}

type CreateUserService struct {
	createUserPort          CreateUserPort
	getUserPort             GetUserPort
	confirmAccountTokenPort auth.ConfirmTokenPort
	sendEmailPort           email.SendEmailPort
}

func NewCreateUserService(
	createUserPort CreateUserPort,
	getUserPort GetUserPort,
	confirmAccountTokenPort auth.ConfirmTokenPort,
	sendEmailPort email.SendEmailPort,
) (CreateTenantUserUseCase, CreateTenantAdminUseCase, CreateSuperAdminUseCase) {
	service := &CreateUserService{
		createUserPort:          createUserPort,
		getUserPort:             getUserPort,
		confirmAccountTokenPort: confirmAccountTokenPort,
		sendEmailPort:           sendEmailPort,
	}
	return service, service, service
}

func (service *CreateUserService) CreateTenantUser(cmd CreateTenantUserCommand) (*User, error) {
	// Controlla user
	user, err := service.getUserPort.GetUserByEmail(cmd.Email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, errUserAlreadyExists
	}

	// Crea user
	newUser, err := service.createUserPort.CreateUser(User{
		name:      cmd.Username,
		email:     cmd.Email,
		role:      ROLE_TENANT_USER,
		tenantId:  &cmd.TenantId,
		confirmed: false,
	})
	if err != nil {
		return nil, err
	}

	// Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(newUser.id)
	if err != nil {
		return nil, err
	}

	// Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(newUser.email, confirmAccountToken)
	if err != nil {
		// TODO: Come gestisco errore di invio email? Cancello account?
	}

	// Ritorna user
	return newUser, nil
}

func (service *CreateUserService) CreateTenantAdmin(cmd CreateTenantAdminCommand) (*User, error) {
	// Controlla user
	user, err := service.getUserPort.GetUserByEmail(cmd.Email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, errUserAlreadyExists
	}

	// Crea user
	newUser, err := service.createUserPort.CreateUser(User{
		name:      cmd.Username,
		email:     cmd.Email,
		role:      ROLE_TENANT_ADMIN,
		tenantId:  &cmd.TenantId,
		confirmed: false,
	})
	if err != nil {
		return nil, err
	}

	// Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(newUser.id)
	if err != nil {
		return nil, err
	}

	// Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(newUser.email, confirmAccountToken)
	if err != nil {
		// TODO: Come gestisco errore di invio email? Cancello account?
	}

	// Ritorna user
	return newUser, err
}

func (service *CreateUserService) CreateSuperAdmin(cmd CreateSuperAdminCommand) (*User, error) {
	// Controlla user
	user, err := service.getUserPort.GetUserByEmail(cmd.Email)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return nil, errUserAlreadyExists
	}

	// Crea user
	newUser, err := service.createUserPort.CreateUser(User{
		name:      cmd.Username,
		email:     cmd.Email,
		role:      ROLE_SUPER_ADMIN,
		tenantId:  nil,
		confirmed: false,
	})
	if err != nil {
		return nil, err
	}

	// Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(newUser.id)
	if err != nil {
		return nil, err
	}

	// Invia token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(newUser.email, confirmAccountToken)
	if err != nil {
		// TODO: Come gestisco errore di invio email? Cancello account?
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

// Delete User ====================================================================================
type DeleteUserUseCase interface {
	DeleteUser(cmd DeleteUserCommand) (*User, error)
}

type DeleteUserService struct {
	deleteUserPort DeleteUserPort
	getUserPort    GetUserPort
}

func NewDeleteUserService(
	deleteUserPort DeleteUserPort,
	getUserPort GetUserPort,
) DeleteUserUseCase {
	return &DeleteUserService{
		deleteUserPort: deleteUserPort,
		getUserPort:    getUserPort,
	}
}

func (service *DeleteUserService) DeleteUser(cmd DeleteUserCommand) (*User, error) {
	// Controlla user
	user, err := service.getUserPort.GetUserById(cmd.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errInexistentUser
	}

	// Elimina user
	oldUser, err := service.deleteUserPort.DeleteUser(cmd.UserId)
	return oldUser, err
}

// Compile-time checks
var _ DeleteUserUseCase = (*DeleteUserService)(nil)

// Get User ====================================================================================
type GetUserByIdUseCase interface {
	GetUserById(cmd GetUserByIdCommand) (*User, error)
}

type GetUsersUseCase interface {
	GetUsers(cmd GetUsersCommand) (users []User, total int, err error)
}

type GetUsersByTenantIdUseCase interface {
	GetUsersByTenantId(cmd GetUsersByTenantIdCommand) (users []User, total int, err error)
}

type GetUserService struct {
	getUserPort GetUserPort
}

func NewGetUserService(getUserPort GetUserPort) (GetUserByIdUseCase, GetUsersUseCase, GetUsersByTenantIdUseCase) {
	service := &GetUserService{
		getUserPort: getUserPort,
	}
	return service, service, service
}

func (service *GetUserService) GetUserById(cmd GetUserByIdCommand) (*User, error) {
	user, err := service.getUserPort.GetUserById(cmd.UserId)
	return user, err
}

func (service *GetUserService) GetUsers(cmd GetUsersCommand) ([]User, int, error) {
	user, total, err := service.getUserPort.GetUsers(cmd.Page, cmd.Limit)
	return user, total, err
}

func (service *GetUserService) GetUsersByTenantId(cmd GetUsersByTenantIdCommand) ([]User, int, error) {
	user, total, err := service.getUserPort.GetUsersByTenantId(cmd.TenantId)
	return user, total, err
}

// Compile-time checks
var (
	_ GetUserByIdUseCase        = (*GetUserService)(nil)
	_ GetUsersUseCase           = (*GetUserService)(nil)
	_ GetUsersByTenantIdUseCase = (*GetUserService)(nil)
)

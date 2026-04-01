package user

import (
	// "backend/internal/auth"

	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/user/mocks/ports_create.go -package=mocks . GenerateTokenPort,SendConfirmAccountEmailPort

type GenerateTokenPort interface {
	NewConfirmAccountToken(user User) (string, error)
}

type SendConfirmAccountEmailPort interface {
	SendConfirmAccountEmail(toAddr string, tenantId *uuid.UUID, tokenString string) error
}

/*
Servizio di creazione utente.

Possibile miglioria: Validare l'input, non affidandosi a validazione in controller
*/
type CreateUserService struct {
	createUserPort          SaveUserPort
	deleteUserPort          DeleteUserPort
	getUserPort             GetUserPort
	getTenantPort           tenant.GetTenantPort // TODO: definire interfaccia localmente
	confirmAccountTokenPort GenerateTokenPort
	sendEmailPort           SendConfirmAccountEmailPort
}

func NewCreateUserService(
	createUserPort SaveUserPort,
	deleteUserPort DeleteUserPort,
	getUserPort GetUserPort,
	getTenantPort tenant.GetTenantPort,
	confirmAccountTokenPort GenerateTokenPort,
	sendEmailPort SendConfirmAccountEmailPort,
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

	// Controlla autorizzazione tenant
	// NOTA: rimosso static check per chiarezza
	superAdminAccess := cmd.Requester.IsSuperAdmin() && tenantFound.CanImpersonate //nolint:staticcheck
	if !superAdminAccess && !cmd.Requester.CanTenantAdminAccess(cmd.TenantId) {    //nolint:staticcheck
		return User{}, identity.ErrUnauthorizedAccess
	}

	// 2. Controlla user
	checkedUser, err := service.getUserPort.GetUserByEmail(&cmd.TenantId, cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !checkedUser.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// 3. Crea user
	user, err := service.createUserPort.SaveUser(User{
		Name:      cmd.Username,
		Email:     cmd.Email,
		Role:      identity.ROLE_TENANT_USER,
		TenantId:  &cmd.TenantId,
		Confirmed: false,
	})
	if err != nil {
		return User{}, err
	}

	// 4. Crea token di conferma
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(user)
	if err != nil {
		return User{}, err
	}

	// 5. Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(user.Email, &cmd.TenantId, confirmAccountToken)
	if err != nil {
		// 6. Elimina account se invio mail fallisce
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
	user, err := service.getUserPort.GetUserByEmail(&cmd.TenantId, cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !user.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// 4. Crea user
	user, err = service.createUserPort.SaveUser(User{
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
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(user)
	if err != nil {
		return User{}, err
	}

	// 6. Invia email per token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(user.Email, &cmd.TenantId, confirmAccountToken)
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
	user, err := service.getUserPort.GetUserByEmail(nil, cmd.Email)
	if err != nil {
		return User{}, err
	}
	if !user.IsZero() {
		return User{}, ErrUserAlreadyExists
	}

	// 2. Crea user
	user, err = service.createUserPort.SaveUser(User{
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
	confirmAccountToken, err := service.confirmAccountTokenPort.NewConfirmAccountToken(user)
	if err != nil {
		return User{}, err
	}

	// 4. Invia token di conferma
	err = service.sendEmailPort.SendConfirmAccountEmail(user.Email, nil, confirmAccountToken)
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

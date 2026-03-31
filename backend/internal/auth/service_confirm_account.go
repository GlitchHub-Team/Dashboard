package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/user"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/auth/mocks/ports_confirm_account.go -package=mocks . ConfirmAccountTokenPort

type ConfirmAccountTokenPort interface {
	NewConfirmAccountToken(user user.User) (string, error)
	DeleteConfirmAccountToken(token ConfirmAccountToken) error

	GetTenantMemberByConfirmAccountToken(tenantId uuid.UUID, tokenString string) (
		userFound user.User, err error,
	)
	GetSuperAdminByConfirmAccountToken(tokenString string) (
		userFound user.User, err error,
	)

	GetTenantConfirmAccountToken(tenantId uuid.UUID, tokenString string) (
		token ConfirmAccountToken, err error,
	)
	GetSuperAdminConfirmAccountToken(tokenString string) (
		token ConfirmAccountToken, err error,
	)
}

/* Service che gestisce la conferma degli account */
type ConfirmUserAccountService struct {
	log                     *zap.Logger
	hasher                  crypto.SecretHasher
	confirmAccountTokenPort ConfirmAccountTokenPort
	saveUserPort            user.SaveUserPort
}

// Compile-time checks
var (
	_ ConfirmAccountUseCase            = (*ConfirmUserAccountService)(nil)
	_ VerifyConfirmAccountTokenUseCase = (*ConfirmUserAccountService)(nil)
)

func NewConfirmUserAccountService(
	log *zap.Logger,
	hasher crypto.SecretHasher,
	confirmAccountTokenPort ConfirmAccountTokenPort,
	saveUserPort user.SaveUserPort,
) *ConfirmUserAccountService {
	return &ConfirmUserAccountService{
		log:                     log,
		hasher:                  hasher,
		confirmAccountTokenPort: confirmAccountTokenPort,
		saveUserPort:            saveUserPort,
	}
}

func (service *ConfirmUserAccountService) getValidTenantToken(tenantId uuid.UUID, token string) (ConfirmAccountToken, error) {
	tokenObj, err := service.confirmAccountTokenPort.GetTenantConfirmAccountToken(tenantId, token)
	if err != nil {
		return ConfirmAccountToken{}, err
	}
	if tokenObj == (ConfirmAccountToken{}) {
		return ConfirmAccountToken{}, ErrTokenNotFound
	}
	if tokenObj.IsExpired() {
		return ConfirmAccountToken{}, ErrTokenExpired
	}

	return tokenObj, err
}

func (service *ConfirmUserAccountService) getValidSuperAdminToken(token string) (ConfirmAccountToken, error) {
	tokenObj, err := service.confirmAccountTokenPort.GetSuperAdminConfirmAccountToken(token)
	if err != nil {
		return ConfirmAccountToken{}, err
	}
	if tokenObj == (ConfirmAccountToken{}) {
		return ConfirmAccountToken{}, ErrTokenNotFound
	}
	if tokenObj.IsExpired() {
		return ConfirmAccountToken{}, ErrTokenExpired
	}
	
	return tokenObj, err
}

func (service *ConfirmUserAccountService) ConfirmAccount(cmd ConfirmAccountCommand) (
	confirmedUser user.User, err error,
) {
	// 1. Get token
	var tokenObj ConfirmAccountToken

	// - Super Admin
	if cmd.TenantId == nil {
		tokenObj, err = service.getValidSuperAdminToken(cmd.Token)
	} else
	// - Tenant Member
	{
		tokenObj, err = service.getValidTenantToken(*cmd.TenantId, cmd.Token)
	}

	if err != nil {
		return
	}

	// 2. Get user
	// TODO: Non so se fare questa query o chiamare getUserPort.Get -> problema di questo approccio =
	// dover rifare sempre la logica relativa ai ruoli

	// - Super Admin
	if cmd.TenantId == nil {
		confirmedUser, err = service.confirmAccountTokenPort.GetSuperAdminByConfirmAccountToken(cmd.Token)
	} else
	// - Tenant Member
	{
		confirmedUser, err = service.confirmAccountTokenPort.GetTenantMemberByConfirmAccountToken(*cmd.TenantId, cmd.Token)
	}

	if confirmedUser.Confirmed {
		return user.User{}, ErrAccountAlreadyConfirmed
	}

	if err != nil {
		return user.User{}, ErrTokenNotFound
	}


	// 3. Create hash
	hash, err := service.hasher.HashSecret(cmd.NewPassword)
	if err != nil {
		return user.User{}, err
	}

	// Imposta password e confirmed status

	// - Password
	err = confirmedUser.SetPasswordHash(hash)
	if err != nil {
		return user.User{}, err
	}

	// - Confirmed status
	confirmedUser.Confirmed = true

	// 4. Save user
	confirmedUser, err = service.saveUserPort.SaveUser(confirmedUser)
	if err != nil {
		return user.User{}, err
	}

	// 5. Delete token
	// NOTA: questo errore non è bloccante
	deleteErr := service.confirmAccountTokenPort.DeleteConfirmAccountToken(tokenObj)
	if deleteErr != nil {
		service.log.Error(
			"Cannot delete token", 
			zap.String("token", cmd.Token), 
			zap.String("type", "confirm_account"), 
			zap.Error(err),
		)
	}

	return
}

/* Verifica esistenza del token di conferma account. Se nil, allora il token esiste, altrimenti non esiste o è scaduto.*/
func (service *ConfirmUserAccountService) VerifyConfirmAccountToken(cmd VerifyConfirmAccountTokenCommand) (err error) {
	if cmd.TenantId == nil {
		_, err = service.getValidSuperAdminToken(cmd.Token)
	} else {
		_, err = service.getValidTenantToken(*cmd.TenantId, cmd.Token)
	}
	return err
}

package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/user"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/auth/mocks/ports.go -package=mocks . ConfirmAccountTokenPort

// TODO: Non sono sicuro che i metodi NewConfirmAccountToken e GetConfirmAccountTokenByUserId
// debbano prendere tenantId in input
type ConfirmAccountTokenPort interface {
	NewConfirmAccountToken(user user.User) (string, error)
	DeleteConfirmAccountToken(token ConfirmAccountToken) error
	GetConfirmAccountToken(tokenString string) (ConfirmAccountToken, error)
	GetUserByConfirmAccountToken(tokenString string) (user.User, error)
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

func (service *ConfirmUserAccountService) getToken(token string) (ConfirmAccountToken, error) {
	tokenObj, err := service.confirmAccountTokenPort.GetConfirmAccountToken(token)
	if err != nil {
		return ConfirmAccountToken{}, nil
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
	tokenObj, err := service.getToken(cmd.Token)
	if err != nil {
		return user.User{}, err
	}

	// 2. Get user
	// TODO: Non so se fare questa query o chiamare getUserPort.Get -> problema di questo approccio =
	// dover rifare sempre la logica relativa ai ruoli
	confirmedUser, err = service.confirmAccountTokenPort.GetUserByConfirmAccountToken(cmd.Token)
	if err != nil {
		return confirmedUser, ErrTokenNotFound
	}
	if confirmedUser.Confirmed {
		return user.User{}, ErrAccountAlreadyConfirmed
	}

	// 3. Set password and confirmed status
	hash, err := service.hasher.HashSecret(cmd.NewPassword)
	if err != nil {
		return user.User{}, err
	}

	err = confirmedUser.SetPasswordHash(hash)
	if err != nil {
		return user.User{}, err
	}
	confirmedUser.Confirmed = true

	// 4. Save user
	confirmedUser, err = service.saveUserPort.SaveUser(confirmedUser)
	if err != nil {
		return user.User{}, err
	}

	// 5. Delete token
	err = service.confirmAccountTokenPort.DeleteConfirmAccountToken(tokenObj)
	if err != nil {
		service.log.Error("Cannot delete token", zap.String("token", cmd.Token), zap.Error(err))
	}

	return
}

/* Verifica esistenza del token di conferma account. Se nil, allora il token esiste, altrimenti non esiste o è scaduto.*/
func (service *ConfirmUserAccountService) VerifyConfirmAccountToken(cmd VerifyConfirmAccountTokenCommand) error {
	_, err := service.getToken(cmd.Token)
	return err
}

package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/user"
)

/*
Service che gestisce le sessioni utente (login/logout)
*/
type SessionService struct {
	// log         *zap.Logger
	hasher      crypto.SecretHasher
	getUserPort user.GetUserPort
}

// Compile-time checks
var (
	_ LoginUserUseCase  = (*SessionService)(nil)
	_ LogoutUserUseCase = (*SessionService)(nil)
)

func NewAuthSessionService(
	// log *zap.Logger,
	hasher crypto.SecretHasher,
	getUserPort user.GetUserPort,
) *SessionService {
	return &SessionService{
		// log:         log,
		hasher:      hasher,
		getUserPort: getUserPort,
	}
}

/*
Date le credenziali specificati in cmd, ritorna il relativo user.

Se il ruolo specificato non è stato trovato, ritorna come errore ErrAccountNotConfirmed.

Se l'account non è stato confermato, ritorna come errore ErrAccountNotConfirmed

Se le credenziali specificate non sono corrette, ritorna come errore ErrWrongCredentials.
*/
func (service *SessionService) LoginUser(cmd LoginUserCommand) (
	foundUser user.User, err error,
) {
	// Get user
	foundUser, err = service.getUserPort.GetUserByEmail(cmd.TenantId, cmd.Email)
	if err != nil {
		return user.User{}, err
	}

	// Check confirmed
	if !foundUser.Confirmed {
		return user.User{}, ErrAccountNotConfirmed
	}

	// Check password
	hash := ""
	if foundUser.PasswordHash == nil {
		return user.User{}, ErrWrongCredentials
	}
	hash = *foundUser.PasswordHash

	err = service.hasher.CompareHashAndSecret(hash, cmd.Password)
	if err != nil {
		return user.User{}, ErrWrongCredentials
	}

	return foundUser, nil
}

func (service *SessionService) LogoutUser(cmd LogoutUserCommand) error {
	// NOTA: corpo tenuto vuoto in caso di implementazione di audit log
	return nil
}

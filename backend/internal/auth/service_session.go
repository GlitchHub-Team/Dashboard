package auth

import (
	"backend/internal/shared/identity"
	"backend/internal/shared/crypto"
	"backend/internal/user"
)

/*
Service che gestisce le sessioni utente (login/logout)
*/
type SessionService struct {
	hasher      crypto.SecretHasher
	getUserPort user.GetUserPort
}

// Compile-time checks
var (
	_ LoginUserUseCase  = (*SessionService)(nil)
	_ LogoutUserUseCase = (*SessionService)(nil)
)

func NewAuthSessionService(
	hasher crypto.SecretHasher,
	getUserPort user.GetUserPort,
) *SessionService {
	return &SessionService{
		hasher:      hasher,
		getUserPort: getUserPort,
	}
}

func (service *SessionService) LoginUser(cmd LoginUserCommand) (
	foundUser user.User, err error,
) {
	// Get user
	switch cmd.Role {
	case identity.ROLE_TENANT_USER:
		foundUser, err = service.getUserPort.GetTenantUserByEmail(*cmd.TenantId, cmd.Email)
	case identity.ROLE_TENANT_ADMIN:
		foundUser, err = service.getUserPort.GetTenantAdminByEmail(*cmd.TenantId, cmd.Email)
	case identity.ROLE_SUPER_ADMIN:
		foundUser, err = service.getUserPort.GetSuperAdminByEmail(cmd.Email)
	default:
		err = identity.ErrUnknownRole
	}
	if err != nil {
		return user.User{}, err
	}

	// Check confirmed
	if !foundUser.Confirmed {
		return user.User{}, ErrAccountNotConfirmed
	}

	// Check password
	err = service.hasher.CompareHashAndSecret(*foundUser.PasswordHash, cmd.Password)
	if err != nil {
		return user.User{}, ErrWrongCredentials
	}

	return foundUser, nil
}

func (service *SessionService) LogoutUser(cmd LogoutUserCommand) error {
	return nil
}

package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/user"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/auth/mocks/ports.go -package=mocks . ForgotPasswordTokenPort,SendChangePasswordEmailPort

type ForgotPasswordTokenPort interface {
	NewForgotPasswordToken(user user.User) (string, error)
	DeleteForgotPasswordToken(token ForgotPasswordToken) error

	GetTenantMemberByForgotPasswordToken(tenantId uuid.UUID, tokenString string) (
		userFound user.User, err error,
	)
	GetSuperAdminByForgotPasswordToken(tokenString string) (
		userFound user.User, err error,
	)

	GetTenantForgotPasswordToken(tenantId string, tokenString string) (
		token ForgotPasswordToken, err error,
	)
	GetSuperAdminForgotPasswordToken(tokenString string) (
		token ForgotPasswordToken, err error,
	)
}

type SendChangePasswordEmailPort interface {
	SendChangePasswordEmail(toAddr string, token string) error
}

/*
Service che gestisce i cambi password (che siano forgot password o richiesti da utenti loggati)
*/
type ChangePasswordService struct {
	tokenGenerator crypto.SecurityTokenGenerator
	hasher         crypto.SecretHasher // NOTA: dev'essere lo stesso hasher con cui si è creato il token

	forgotPasswordTokenPort     ForgotPasswordTokenPort
	sendChangePasswordEmailPort SendChangePasswordEmailPort
	getUserPort                 user.GetUserPort
	saveUserPort                user.SaveUserPort
}

var (
	_ VerifyForgotPasswordTokenUseCase = (*ChangePasswordService)(nil)
	_ RequestForgotPasswordUseCase     = (*ChangePasswordService)(nil)
	_ ConfirmForgotPasswordUseCase     = (*ChangePasswordService)(nil)
	_ ChangePasswordUseCase            = (*ChangePasswordService)(nil)
)

func NewChangePasswordService(
	tokenGenerator crypto.SecurityTokenGenerator,
	hasher crypto.SecretHasher,
	changePasswordTokenPort ForgotPasswordTokenPort,
	sendChangePasswordEmailPort SendChangePasswordEmailPort,
	getUserPort user.GetUserPort,
	saveUserPort user.SaveUserPort,
) *ChangePasswordService {
	return &ChangePasswordService{
		tokenGenerator:              tokenGenerator,
		hasher:                      hasher,
		forgotPasswordTokenPort:     changePasswordTokenPort,
		sendChangePasswordEmailPort: sendChangePasswordEmailPort,
		getUserPort:                 getUserPort,
		saveUserPort:                saveUserPort,
	}
}

/*
Ritorna l'oggetto ForgotPasswordToken relativo al token plain tokenString nel tenant con id tenantId.
Se il token è scaduto, elimina il token e ritorna tokenObj vuoto ed ErrTokenExpired
*/
func (service *ChangePasswordService) getValidTenantToken(tenantId uuid.UUID, tokenString string) (tokenObj ForgotPasswordToken, err error) {
	tokenObj, err = service.forgotPasswordTokenPort.GetTenantForgotPasswordToken(tenantId.String(), tokenString)
	if err != nil {
		return ForgotPasswordToken{}, err
	}
	if tokenObj.IsExpired() {
		return ForgotPasswordToken{}, ErrTokenExpired
	}
	return tokenObj, err
}

/*
Ritorna l'oggetto ForgotPasswordToken relativo al token plain tokenString per super admin.
Se il token è scaduto, elimina il token e ritorna tokenObj vuoto ed ErrTokenExpired
*/
func (service *ChangePasswordService) getValidSuperAdminToken(tokenString string) (tokenObj ForgotPasswordToken, err error) {
	tokenObj, err = service.forgotPasswordTokenPort.GetSuperAdminForgotPasswordToken(tokenString)
	if err != nil {
		return ForgotPasswordToken{}, err
	}
	if tokenObj.IsExpired() {
		return ForgotPasswordToken{}, ErrTokenExpired
	}
	return tokenObj, err
}

/*
Verifica esistenza del token forgot password.
Se il token esiste ritorna nil, altrimenti ritorna errore non-nil.
*/
func (service *ChangePasswordService) VerifyForgotPasswordToken(cmd VerifyForgotPasswordTokenCommand) (err error) {
	// Super Admin
	if cmd.TenantId == nil {
		_, err = service.getValidSuperAdminToken(cmd.Token)
	} else
	// Tenant Member
	{
		_, err = service.getValidTenantToken(*cmd.TenantId, cmd.Token)
	}
	return
}

/*
Richiede il cambio di password dimenticata, come descritto in cmd.
Se la procedura va a buon fine ritorna nil, altrimenti ritorna errore non-nil.
*/
func (service *ChangePasswordService) RequestForgotPassword(cmd RequestForgotPasswordCommand) error {
	// 1. Controlla utente
	userFound, err := service.getUserPort.GetUserByEmail(cmd.TenantId, cmd.Email)
	if err != nil {
		return err
	}
	if userFound.IsZero() {
		return user.ErrUserNotFound
	}
	if !userFound.Confirmed {
		return ErrAccountNotConfirmed
	}

	// 2. Crea token
	tokenString, err := service.forgotPasswordTokenPort.NewForgotPasswordToken(userFound)
	if err != nil {
		return err
	}

	// 3. Invia mail cambio password
	err = service.sendChangePasswordEmailPort.SendChangePasswordEmail(cmd.Email, tokenString)
	if err != nil {
		return err
	}

	return nil
}

/*
Conferma la richiesta di cambio password dimenticata.
*/
func (service *ChangePasswordService) ConfirmForgotPassword(cmd ConfirmForgotPasswordCommand) (err error) {
	// 1. Get token:
	var tokenObj ForgotPasswordToken

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

	// 2. Controlla token con plaintext ricevuto
	err = service.hasher.CompareHashAndSecret(tokenObj.hashedToken, cmd.Token)
	if err != nil {
		return err
	}

	// 3. Get user
	userFound, err := service.getUserPort.GetUser(tokenObj.tenantId, tokenObj.userId)
	if err != nil {
		return err
	}
	if !userFound.Confirmed {
		return ErrAccountNotConfirmed
	}

	// 4. Crea hash password
	newPasswordHash, err := service.hasher.HashSecret(cmd.NewPassword)
	if err != nil {
		return err
	}

	// 5. Cambia password (controllo dominio)
	err = userFound.SetPasswordHash(newPasswordHash)
	if err != nil {
		return err
	}

	// 6. Elimina token
	err = service.forgotPasswordTokenPort.DeleteForgotPasswordToken(tokenObj)
	if err != nil {
		return err
	}

	// 7. Salva user con password modificata
	_, err = service.saveUserPort.SaveUser(userFound)
	if err != nil {
		return err
	}

	return nil
}

func (service *ChangePasswordService) ChangePassword(cmd ChangePasswordCommand) error {
	// 1. Get user
	userFound, err := service.getUserPort.GetUser(cmd.RequesterTenantId, cmd.RequesterUserId)
	if err != nil {
		return err
	}

	// 2. Check conferma account
	if !userFound.Confirmed {
		return ErrAccountNotConfirmed
	}

	// 3. Controlla password vecchia
	passwordHash := ""
	if userFound.PasswordHash != nil {
		passwordHash = *userFound.PasswordHash
	}
	err = service.hasher.CompareHashAndSecret(passwordHash, cmd.OldPassword)
	if err != nil {
		return ErrWrongCredentials
	}

	// 4. Genera nuovo hash
	newHash, err := service.hasher.HashSecret(cmd.NewPassword)
	if err != nil {
		return err
	}

	// 5. Cambia password (controllo di dominio)
	err = userFound.SetPasswordHash(newHash)
	if err != nil {
		return err
	}

	// 6. Salva user
	_, err = service.saveUserPort.SaveUser(userFound)
	if err != nil {
		return err
	}

	return nil
}

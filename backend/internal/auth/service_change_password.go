package auth

import (
	"backend/internal/shared/crypto"
	"backend/internal/user"

	"github.com/google/uuid"
)

type ChangePasswordTokenPort interface {
	SaveChangePasswordToken(token ForgotPasswordToken) (ForgotPasswordToken, error)
	DeleteChangePasswordToken(token ForgotPasswordToken) error
	GetChangePasswordTokenByUser(tenantId *uuid.UUID, userId uint) (ForgotPasswordToken, error)
	GetChangePasswordToken(hashedTokenString string) (ForgotPasswordToken, error)
}

type SendChangePasswordEmailPort interface {
	SendChangePasswordEmail(toAddr string, token string) error
}

/*
Service che gestisce i cambi password (che siano forgot password o richiesti da utenti loggati)
*/
type ChangePasswordService struct {
	tokenGenerator crypto.TokenGenerator
	hasher         crypto.SecretHasher // NOTA: dev'essere lo stesso hasher con cui si è creato il token

	changePasswordTokenPort     ChangePasswordTokenPort
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
	tokenGenerator crypto.TokenGenerator,
	hasher crypto.SecretHasher,
	changePasswordTokenPort ChangePasswordTokenPort,
	sendChangePasswordEmailPort SendChangePasswordEmailPort,
	getUserPort user.GetUserPort,
	saveUserPort user.SaveUserPort,
) *ChangePasswordService {
	return &ChangePasswordService{
		tokenGenerator:              tokenGenerator,
		hasher:                      hasher,
		changePasswordTokenPort:     changePasswordTokenPort,
		sendChangePasswordEmailPort: sendChangePasswordEmailPort,
		getUserPort:                 getUserPort,
		saveUserPort:                saveUserPort,
	}
}

/*
Ritorna l'oggetto ForgotPasswordToken relativo al token plain tokenString.
Se il token è scaduto, elimina il token e ritorna tokenObj vuoto ed ErrTokenExpired
*/
func (service *ChangePasswordService) getValidToken(tokenString string) (tokenObj ForgotPasswordToken, err error) {
	tokenObj, err = service.changePasswordTokenPort.GetChangePasswordToken(tokenString)
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
func (service *ChangePasswordService) VerifyForgotPasswordToken(cmd VerifyForgotPasswordTokenCommand) error {
	_, err := service.getValidToken(cmd.Token)
	return err
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
	tokenString, hashedTokenString, err := service.tokenGenerator.GenerateToken()
	expiryDate := service.tokenGenerator.ExpiryFromNow()
	if err != nil {
		return err
	}
	_, err = service.changePasswordTokenPort.SaveChangePasswordToken(ForgotPasswordToken{
		hashedToken: hashedTokenString,
		tenantId:    cmd.TenantId,
		userId:      userFound.Id,
		expiryDate:  expiryDate,
	})
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

func (service *ChangePasswordService) ConfirmForgotPassword(cmd ConfirmForgotPasswordCommand) error {
	// 1. Get token
	tokenObj, err := service.getValidToken(cmd.Token)
	if err != nil {
		return err
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
	err = service.changePasswordTokenPort.DeleteChangePasswordToken(tokenObj)
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

	// 2. Controlla password vecchia
	err = service.hasher.CompareHashAndSecret(*userFound.PasswordHash, cmd.OldPassword)
	if err != nil {
		return ErrWrongCredentials
	}

	// 3. Genera nuovo hash
	newHash, err := service.hasher.HashSecret(cmd.NewPassword)
	if err != nil {
		return err
	}

	// 4. Cambia password (controllo di dominio)
	err = userFound.SetPasswordHash(newHash)
	if err != nil {
		return err
	}

	// 5. Salva user
	_, err = service.saveUserPort.SaveUser(userFound)
	if err != nil {
		return err
	}

	return nil
}

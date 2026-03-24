package auth

import (
	"errors"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/crypto"
	"backend/internal/shared/identity"
	"backend/internal/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../tests/auth/mocks/use_cases.go -package=mocks . LoginUserUseCase,LogoutUserUseCase,ConfirmAccountUseCase,VerifyConfirmAccountTokenUseCase,VerifyForgotPasswordTokenUseCase,RequestForgotPasswordUseCase,ConfirmForgotPasswordUseCase,ChangePasswordUseCase

// Session

type LoginUserUseCase interface {
	LoginUser(LoginUserCommand) (user.User, error)
}
type LogoutUserUseCase interface {
	LogoutUser(LogoutUserCommand) error
}

// Confirm account

type ConfirmAccountUseCase interface {
	ConfirmAccount(ConfirmAccountCommand) (user.User, error)
}
type VerifyConfirmAccountTokenUseCase interface {
	VerifyConfirmAccountToken(VerifyConfirmAccountTokenCommand) error
}

// Change password

type VerifyForgotPasswordTokenUseCase interface {
	VerifyForgotPasswordToken(VerifyForgotPasswordTokenCommand) error
}
type RequestForgotPasswordUseCase interface {
	RequestForgotPassword(RequestForgotPasswordCommand) error
}
type ConfirmForgotPasswordUseCase interface {
	ConfirmForgotPassword(ConfirmForgotPasswordCommand) error
}
type ChangePasswordUseCase interface {
	ChangePassword(ChangePasswordCommand) error
}

type Controller struct {
	log *zap.Logger

	authTokenManager crypto.AuthTokenManager

	loginUserUseCase  LoginUserUseCase
	logoutUserUseCase LogoutUserUseCase

	confirmAccountUseCase            ConfirmAccountUseCase
	verifyConfirmAccountTokenUseCase VerifyConfirmAccountTokenUseCase

	verifyForgotPasswordTokenUseCase VerifyForgotPasswordTokenUseCase
	requestForgotPasswordUseCase     RequestForgotPasswordUseCase
	confirmForgotPasswordUseCase     ConfirmForgotPasswordUseCase
	changePasswordUseCase            ChangePasswordUseCase
}

func NewController(
	log *zap.Logger,

	authTokenManager crypto.AuthTokenManager,

	loginUserUseCase LoginUserUseCase,
	logoutUserUseCase LogoutUserUseCase,

	confirmAccountUseCase ConfirmAccountUseCase,
	verifyConfirmAccountTokenUseCase VerifyConfirmAccountTokenUseCase,

	verifyForgotPasswordTokenUseCase VerifyForgotPasswordTokenUseCase,
	requestForgotPasswordUseCase RequestForgotPasswordUseCase,
	confirmForgotPasswordUseCase ConfirmForgotPasswordUseCase,
	changePasswordUseCase ChangePasswordUseCase,
) *Controller {
	return &Controller{
		log:              log,
		authTokenManager: authTokenManager,

		loginUserUseCase:  loginUserUseCase,
		logoutUserUseCase: logoutUserUseCase,

		confirmAccountUseCase:            confirmAccountUseCase,
		verifyConfirmAccountTokenUseCase: verifyConfirmAccountTokenUseCase,

		verifyForgotPasswordTokenUseCase: verifyForgotPasswordTokenUseCase,
		requestForgotPasswordUseCase:     requestForgotPasswordUseCase,
		confirmForgotPasswordUseCase:     confirmForgotPasswordUseCase,
		changePasswordUseCase:            changePasswordUseCase,
	}
}

// Sessione ===========================================================================================

func (controller *Controller) LoginUser(ctx *gin.Context) {
	// 1. Binding JSON body
	var bodyDto LoginUserDto
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// 2. Esegui login
	var tenantId *uuid.UUID
	if bodyDto.TenantId != nil {
		parsed, _ := uuid.Parse(*bodyDto.TenantId)
		tenantId = &parsed
	}
	cmd := LoginUserCommand{
		TenantId: tenantId,
		Email:    bodyDto.Email,
		Password: bodyDto.Password,
		Role:     identity.UserRole(bodyDto.UserRole),
	}
	userLogged, err := controller.loginUserUseCase.LoginUser(cmd)
	if err != nil {
		if errors.Is(err, user.ErrUnknownRole) {
			transportHttp.RequestError(ctx, err)
			return
		} else if errors.Is(err, ErrAccountNotConfirmed) || errors.Is(err, ErrWrongCredentials) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// 3. Crea JWT da restituire all'utente
	jwtToken, err := controller.authTokenManager.GenerateForRequester(identity.Requester{
		RequesterUserId:   userLogged.Id,
		RequesterTenantId: userLogged.TenantId,
		RequesterRole:     userLogged.Role,
	})
	if err != nil {
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// 4. Rispondi all'utente
	responseDto := LoginResponseDto{
		JWT: jwtToken,
	}
	transportHttp.RequestOk(ctx, responseDto)
}

func (controller *Controller) LogoutUser(ctx *gin.Context) {
	// 1. Extract requester
	_, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestError(ctx, err)
		return
	}

	// NOTA: corpo vuoto in caso ci vadano audit log

	transportHttp.RequestOk(ctx, gin.H{
		"result": "ok",
	})
}

// Conferma account ===================================================================================

func (controller *Controller) VerifyConfirmAccountToken(ctx *gin.Context) {
	// 1. Binding URI
	var uriDto dto.TokenUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		transportHttp.RequestNotFound(ctx, ErrTokenNotFound)
		return
	}

	// 2. Check token
	err := controller.verifyConfirmAccountTokenUseCase.VerifyConfirmAccountToken(VerifyConfirmAccountTokenCommand{
		Token: uriDto.Token,
	})
	if err != nil {
		ctx.JSON(404, gin.H{
			"error":  ErrTokenNotFound.Error(),
			"result": false,
		})
		return
	}

	transportHttp.RequestOk(ctx, gin.H{
		"result": true,
	})
}

func (controller *Controller) ConfirmAccount(ctx *gin.Context) {
	// 1. Binding JSON
	var bodyDto ConfirmUserAccountDto
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// 2. Verifica token di confirma
	err := controller.verifyConfirmAccountTokenUseCase.VerifyConfirmAccountToken(VerifyConfirmAccountTokenCommand{
		Token: bodyDto.Token,
	})
	if err != nil {
		transportHttp.RequestNotFound(ctx, ErrTokenNotFound)
		return
	}

	// 3. Esegui comando
	confirmedUser, err := controller.confirmAccountUseCase.ConfirmAccount(ConfirmAccountCommand{
		Token:       bodyDto.Token,
		NewPassword: bodyDto.NewPassword,
	})
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) || errors.Is(err, ErrAccountAlreadyConfirmed) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// 4. Crea token di autenticazione da restituire all'utente
	authToken, err := controller.authTokenManager.GenerateForRequester(identity.Requester{
		RequesterUserId:   confirmedUser.Id,
		RequesterTenantId: confirmedUser.TenantId,
		RequesterRole:     confirmedUser.Role,
	})
	if err != nil {
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// 5. Invia risposta
	responseDto := LoginResponseDto{
		JWT: authToken,
	}
	transportHttp.RequestOk(ctx, responseDto)
}

// FORGOT PASSWORD ====================================================================================

func (controller *Controller) VerifyForgotPasswordToken(ctx *gin.Context) {
	// 1. Binding JSON
	var uriDto dto.TokenUriDTO
	if err := ctx.ShouldBindUri(&uriDto); err != nil {
		transportHttp.RequestNotFound(ctx, ErrTokenNotFound)
		return
	}

	// 2. Esegui comando
	err := controller.verifyForgotPasswordTokenUseCase.VerifyForgotPasswordToken(VerifyForgotPasswordTokenCommand{
		Token: uriDto.Token,
	})
	if err != nil {
		transportHttp.RequestNotFound(ctx, ErrTokenNotFound)
		return
	}

	// 3. Rispondi
	transportHttp.RequestOk(ctx, ResultDto{Result: "ok"})
}

func (controller *Controller) RequestForgotPasswordToken(ctx *gin.Context) {
	// 1. Binding JSON
	var bodyDto RequestForgotPasswordDto
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// 2. Esegui comando
	var tenantId *uuid.UUID
	if bodyDto.TenantId != nil {
		parsed, _ := uuid.Parse(*bodyDto.TenantId)
		tenantId = &parsed
	}
	err := controller.requestForgotPasswordUseCase.RequestForgotPassword(RequestForgotPasswordCommand{
		TenantId: tenantId,
		Email:    bodyDto.Email,
	})
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) || errors.Is(err, ErrAccountNotConfirmed) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// 3. Rispondi
	transportHttp.RequestOk(ctx, ResultDto{Result: "ok"})
}

func (controller *Controller) ConfirmForgotPasswordToken(ctx *gin.Context) {
	// 1. Binding JSON
	var bodyDto ConfirmForgotPasswordDto
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// 2. Esegui comando
	err := controller.confirmForgotPasswordUseCase.ConfirmForgotPassword(ConfirmForgotPasswordCommand{
		Token:       bodyDto.Token,
		NewPassword: bodyDto.NewPassword,
	})
	if err != nil {
		if errors.Is(err, ErrAccountNotConfirmed) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// 3. Rispondi
	transportHttp.RequestOk(ctx, ResultDto{Result: "ok"})
}

func (controller *Controller) ChangePassword(ctx *gin.Context) {
	requester, err := transportHttp.ExtractRequester(ctx)
	if err != nil {
		transportHttp.RequestUnauthorized(ctx, transportHttp.ErrMissingIdentity)
		return
	}

	// 1. Binding JSON
	var bodyDto ChangePasswordDto
	if err := ctx.ShouldBindJSON(&bodyDto); err != nil {
		if !transportHttp.ValidationError(ctx, err) {
			transportHttp.RequestError(ctx, err)
		}
		return
	}

	// 2. Esegui comando
	err = controller.changePasswordUseCase.ChangePassword(ChangePasswordCommand{
		Requester:   requester,
		OldPassword: bodyDto.OldPassword,
		NewPassword: bodyDto.NewPassword,
	})
	if err != nil {
		if errors.Is(err, ErrWrongCredentials) || errors.Is(err, ErrAccountNotConfirmed) {
			transportHttp.RequestNotFound(ctx, err)
			return
		}
		transportHttp.RequestServerError(ctx, err)
		return
	}

	// 3. Rispondi
	transportHttp.RequestOk(ctx, ResultDto{Result: "ok"})
}

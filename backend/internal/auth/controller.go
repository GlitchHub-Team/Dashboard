package auth

import "backend/internal/user"

//go:generate mockgen -destination=../../tests/auth/mocks/use_cases.go -package=mocks . LoginUserUseCase,LogoutUserUseCase,ConfirmAccountUseCase,VerifyConfirmAccountTokenUseCase

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

type Controller struct{}

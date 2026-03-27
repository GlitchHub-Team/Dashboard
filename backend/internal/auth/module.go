package auth

import (
	"backend/internal/user"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"auth",

	// Metodi pubblici
	fx.Provide(
		// Inbound ports
		NewController,

		// Servizi
		fx.Annotate(
			NewAuthSessionService,
			fx.As(new(LoginUserUseCase)),
			fx.As(new(LogoutUserUseCase)),
		),
		fx.Annotate(
			NewConfirmUserAccountService,
			fx.As(new(ConfirmAccountUseCase)),
			fx.As(new(VerifyConfirmAccountTokenUseCase)),
		),
		fx.Annotate(
			NewChangePasswordService,
			fx.As(new(VerifyForgotPasswordTokenUseCase)),
			fx.As(new(RequestForgotPasswordUseCase)),
			fx.As(new(ConfirmForgotPasswordUseCase)),
			fx.As(new(ChangePasswordUseCase)),
		),

		// Outbound ports
		fx.Annotate(
			NewConfirmAccountTokenPostgreAdapter,
			fx.As(new(ConfirmAccountTokenPort)),
			fx.As(new(user.GenerateTokenPort)),
		),
		fx.Annotate(
			NewChangePasswordTokenPostgreAdapter,
			fx.As(new(ForgotPasswordTokenPort)),
		),
	),

	// Metodi privati
	fx.Provide(
		fx.Private,
		
		newTenantPasswordTokenPgRepository,
		newSuperAdminPasswordTokenPgRepository,

		newTenantConfirmTokenPgRepository,
		newSuperAdminConfirmTokenPgRepository,
	),
)

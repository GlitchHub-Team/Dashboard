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
			NewConfirmAccountTokenPgAdapter,
			fx.As(new(ConfirmAccountTokenPort)),
			fx.As(new(user.GenerateTokenPort)),
		),
		fx.Annotate(
			NewChangePasswordTokenPgAdapter,
			fx.As(new(ForgotPasswordTokenPort)),
		),

		// Repository
		fx.Annotate(
			newTenantConfirmTokenPgRepository,
			fx.As(new(TenantConfirmTokenRepository)),
		),

		fx.Annotate(
			newSuperAdminConfirmTokenPgRepository,
			fx.As(new(SuperAdminConfirmTokenRepository)),
		),

		fx.Annotate(
			newTenantPasswordTokenPgRepository,
			fx.As(new(TenantPasswordTokenRepository)),
		),

		fx.Annotate(
			newSuperAdminPasswordTokenPgRepository,
			fx.As(new(SuperAdminPasswordTokenRepository)),
		),
	),

	// Metodi privati
	fx.Provide(
		fx.Private,

		newTenantConfirmTokenPgRepository,
		newSuperAdminConfirmTokenPgRepository,

		newTenantPasswordTokenPgRepository,
		newSuperAdminPasswordTokenPgRepository,
	),
)

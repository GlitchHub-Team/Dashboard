package user

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",

	// Metodi pubblici
	fx.Provide(
		// Controller
		NewUserController,

		// Use Cases
		fx.Annotate(
			NewCreateUserService,
			fx.As(new(CreateTenantUserUseCase)),
			fx.As(new(CreateTenantAdminUseCase)),
			fx.As(new(CreateSuperAdminUseCase)),
		),
		fx.Annotate(
			NewDeleteUserService,
			fx.As(new(DeleteTenantUserUseCase)),
			fx.As(new(DeleteTenantAdminUseCase)),
			fx.As(new(DeleteSuperAdminUseCase)),
		),
		fx.Annotate(
			NewGetUserService,
			fx.As(new(GetTenantUserUseCase)),
			fx.As(new(GetTenantAdminUseCase)),
			fx.As(new(GetSuperAdminUseCase)),
			fx.As(new(GetTenantUsersByTenantUseCase)),
			fx.As(new(GetTenantAdminsByTenantUseCase)),
			fx.As(new(GetSuperAdminListUseCase)),
		),

		// Outbound ports
		fx.Annotate(
			NewUserPostgreAdapter,
			fx.As(new(SaveUserPort)),
			fx.As(new(DeleteUserPort)),
			fx.As(new(GetUserPort)),
		),

		// Repositories (faccio provide delle interfacce per TU)
		fx.Annotate(
			newTenantMemberPgRepository,
			fx.As(new(TenantMemberRepository)),
		),
		fx.Annotate(
			newSuperAdminPgRepository,
			fx.As(new(SuperAdminRepository)),
		),
	),
)

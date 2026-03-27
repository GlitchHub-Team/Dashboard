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
		NewCreateUserService,
		NewDeleteUserService,
		NewGetUserService,

		// Outbound ports
		NewUserPostgreAdapter,

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

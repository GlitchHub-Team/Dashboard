package user

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",

	// Metodi pubblici
	fx.Provide(
		NewUserController,
		// Use Cases
		NewCreateUserService,
		NewDeleteUserService,
		NewGetUserService,

		// Outbound ports
		NewUserPostgreAdapter,
	),

	// Metodi privati
	fx.Provide(
		fx.Private,
		newUserPostgreRepository,
	),
)

package gateway

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway",

	// Metodi pubblici
	fx.Provide(
		NewGatewayController,

		fx.Annotate(
			NewCreateGatewayService,
			fx.As(new(CreateGatewayUseCase)),
		),

		fx.Annotate(
			NewDeleteGatewayService,
			fx.As(new(DeleteGatewayUseCase)),
		),

		fx.Annotate(
			NewGetGatewayService,
			fx.As(new(GetGatewayUseCase)),
		),

		fx.Annotate(
			NewGetAllGatewaysService,
			fx.As(new(GetAllGatewaysUseCase)),
		),

		fx.Annotate(
			NewGetGatewaysByTenantService,
			fx.As(new(GetGatewaysByTenantUseCase)),
		),

		fx.Annotate(
			NewGatewayPostgreAdapter,
			fx.As(new(SaveGatewayPort)),
			fx.As(new(RemoveGatewayPort)),
			fx.As(new(GetGatewayPort)),
			fx.As(new(GetGatewaysPort)),
		),

		NewGatewayPostgreRepository,
	),
)

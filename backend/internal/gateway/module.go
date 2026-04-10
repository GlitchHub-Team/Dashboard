package gateway

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway",

	fx.Provide(
		NewGatewayController,

		fx.Annotate(
			NewGatewayPostgreRepository,
			fx.As(new(GatewayRepository)),
		),

		fx.Annotate(
			NewGatewayPostgreAdapter,
			fx.As(new(SaveGatewayPort)),
			fx.As(new(RemoveGatewayPort)),
			fx.As(new(GetGatewayPort)),
			fx.As(new(GetGatewaysPort)),
		),

		fx.Annotate(
			NewGatewayCommandNATSAdapter,
			fx.As(new(GatewayCommandPort)),
		),

		fx.Annotate(
			NewGatewayManagementService,
			fx.As(new(CreateGatewayUseCase)),
			fx.As(new(DeleteGatewayUseCase)),
			fx.As(new(GetGatewayUseCase)),
			fx.As(new(GetAllGatewaysUseCase)),
			fx.As(new(GetGatewaysByTenantUseCase)),
		),

		fx.Annotate(
			NewGatewayCommandService,
			fx.As(new(CommissionGatewayUseCase)),
			fx.As(new(DecommissionGatewayUseCase)),
			fx.As(new(InterruptGatewayUseCase)),
			fx.As(new(ResumeGatewayUseCase)),
			fx.As(new(ResetGatewayUseCase)),
			fx.As(new(RebootGatewayUseCase)),
			fx.As(new(SetGatewayIntervalLimitUseCase)),
		),
	),
)

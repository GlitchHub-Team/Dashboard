package gateway

import (
	"time"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway",

	fx.Supply(TimeoutNATSClient(time.Second*10)),
	fx.Provide(
		NewGatewayController,

		fx.Annotate(
			NewGatewayPostgreRepository,
			fx.As(new(GatewayRepository)),
		),

		fx.Annotate(
			NewGatewayPostgreAdapter,
			fx.As(new(CreateGatewayPort)),
			fx.As(new(SaveGatewayPort)),
			fx.As(new(DeleteGatewayPort)),
			fx.As(new(GetGatewayPort)),
			fx.As(new(GetGatewaysPort)),
		),

		fx.Annotate(
			NewGatewayCommandNATSRepository,
			fx.As(new(GatewayCommandPort)),
		),

		fx.Annotate(
			NewGatewayManagementService,
			fx.As(new(GetGatewayUseCase)),
			fx.As(new(GetAllGatewaysUseCase)),
			fx.As(new(GetGatewaysByTenantUseCase)),
		),

		fx.Annotate(
			NewGatewayCommandService,
			fx.As(new(CreateGatewayUseCase)),
			fx.As(new(DeleteGatewayUseCase)),
			fx.As(new(CommissionGatewayUseCase)),
			fx.As(new(DecommissionGatewayUseCase)),
			fx.As(new(InterruptGatewayUseCase)),
			fx.As(new(ResumeGatewayUseCase)),
			fx.As(new(ResetGatewayUseCase)),
			fx.As(new(RebootGatewayUseCase)),
		),
	),
)

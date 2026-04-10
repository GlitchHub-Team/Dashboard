package sensor

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"sensor",

	fx.Provide(
		NewSensorController,

		fx.Annotate(
			NewCreateSensorService,
			fx.As(new(CreateSensorUseCase)),
		),
		fx.Annotate(
			NewDeleteSensorService,
			fx.As(new(DeleteSensorUseCase)),
		),
		fx.Annotate(
			NewGetSensorsByGatewayIdService,
			fx.As(new(GetSensorsByGatewayUseCase)),
		),
		fx.Annotate(
			NewGetSensorByIdService,
			fx.As(new(GetSensorUseCase)),
		),
		fx.Annotate(
			NewGetSensorByTenantIdService,
			fx.As(new(GetSensorsByTenantUseCase)),
		),
		fx.Annotate(
			NewInterruptSensorService,
			fx.As(new(InterruptSensorUseCase)),
		),
		fx.Annotate(
			NewResumeSensorService,
			fx.As(new(ResumeSensorUseCase)),
		),

		fx.Annotate(
			NewDbSensorAdapter,
			fx.As(new(CreateSensorPort)),
			fx.As(new(DeleteSensorPort)),
			fx.As(new(GetSensorByIdPort)),
			fx.As(new(GetSensorByTenantPort)),
			fx.As(new(GetSensorsByTenantIdPort)),
			fx.As(new(GetSensorsByGatewayIdPort)),
			fx.As(new(UpdateSensorStatusPort)),
		),
		fx.Annotate(
			NewSendCmdAdapter,
			fx.As(new(CreateSensorCmdPort)),
			fx.As(new(DeleteSensorCmdPort)),
			fx.As(new(SendResumeCmdPort)),
			fx.As(new(SendInterruptCmdPort)),
		),

		fx.Annotate(
			NewSensorPostgreRepository,
			fx.As(new(DatabaseRepository)),
		),
		fx.Annotate(
			NewSensorNatsRepository,
			fx.As(new(MessageBrokerRepository)),
		),
	),
)

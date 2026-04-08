package real_time_data

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"real_time_data",

	// Metodi pubblici
	fx.Provide(
		// Inbound port
		NewController,

		// Business logic
		fx.Annotate(
			NewRealTimeDataService,
			fx.As(new(GetRealTimeDataUseCase)),
		),
		
		// Outbound port
		fx.Annotate(
			NewRealTimeDataNATSAdapter,
			fx.As(new(RealTimeDataPort)),
		),

		// Reader
		fx.Annotate(
			newConcreteRealTimeDataNATSReader,
			fx.As(new(RealTimeDataNATSReader)),
		),
	),

	// Metodi privati
	// fx.Provide(
	// 	fx.Private,
	// 	// ...
	// ),
)

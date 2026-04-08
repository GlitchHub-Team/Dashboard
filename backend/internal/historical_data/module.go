package historical_data

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"historical_data",

	// Metodi pubblici
	fx.Provide(
		NewHistoricalDataController,
		fx.Annotate(
			NewGetHistoricalDataService,
			fx.As(new(GetSensorHistoricalDataUseCase)),
		),
		fx.Annotate(
			NewHistoricalDataTimescaleAdapter,
			fx.As(new(GetHistoricalDataPort)),
		),
	),

	// Metodi privati
	fx.Provide(
		fx.Private,
		newHistoricalDataTimescaleRepository,
	),
)

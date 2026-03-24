package historical_data

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"historical_data",

	// Metodi pubblici
	fx.Provide(
	//..
	),

	// Metodi privati
	fx.Provide(
		fx.Private,
		// ...
	),
)

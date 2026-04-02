package real_time_data

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"real_time_data",

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

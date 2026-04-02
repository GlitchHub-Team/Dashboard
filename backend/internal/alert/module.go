package alert

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"alert",

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

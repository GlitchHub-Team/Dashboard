package tenant

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"tenant",

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
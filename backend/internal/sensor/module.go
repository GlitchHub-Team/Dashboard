package sensor

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"sensor",

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
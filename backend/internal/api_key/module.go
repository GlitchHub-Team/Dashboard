package api_key

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"api_key",

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
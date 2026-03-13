package auth

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"auth",

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
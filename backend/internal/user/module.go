package user

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"user",

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
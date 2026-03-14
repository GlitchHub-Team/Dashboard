package auth

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"auth",

    // Metodi pubblici
    fx.Provide(
        NewConfirmAccountTokenPostgreAdapter,
    ),

    // Metodi privati
    fx.Provide(
        fx.Private,
        newPasswordTokenPostgreRepository,
        newConfirmTokenPostgreRepository,
   ),
)
package tenant

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"tenant",

    // Metodi pubblici
    fx.Provide(
        NewTenantPostgreRepository,
        NewGetTenantsPostgreAdapter,
    ),

    // Metodi privati
    fx.Provide(
        fx.Private,
        // ...
   ),
)
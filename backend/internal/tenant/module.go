package tenant

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"tenant",

    // Metodi pubblici
    fx.Provide(
        // controller
        NewTenantController,

        //UseCase
        NewCreateTenantService,
      
        //Outbound Port
        NewTenantPostgreAdapter,
    ),

    // Metodi privati
    fx.Provide(
        fx.Private,
        
   ),
)
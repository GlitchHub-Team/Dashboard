package gateway

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway",

	// Metodi pubblici
	fx.Provide(

		// Use Cases (inbound ports)
<<<<<<< HEAD
		common.FxAs[CreateGatewayUseCase](NewCreateGatewayService),
		common.FxAs[DeleteGatewayUseCase](NewDeleteGatewayService),
=======
		NewGatewayController,

		NewCreateGatewayService,
		NewDeleteGatewayService,
>>>>>>> origin/issue-17

		// Outbound ports
		NewGatewayPostgreAdapter,
	),

	// Metodi privati
	fx.Provide(
		fx.Private,
		newGatewayPostgreRepository,
	),
)

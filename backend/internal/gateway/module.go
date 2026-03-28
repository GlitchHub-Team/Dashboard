package gateway

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway",

	// Metodi pubblici
	fx.Provide(

		// Use Cases (inbound ports)
		NewGatewayController,

		NewCreateGatewayService,
		NewDeleteGatewayService,

		// Outbound ports
		NewGatewayPostgreAdapter,
	),

	// Metodi privati
	fx.Provide(
		fx.Private,
		NewGatewayPostgreRepository,
	),
)

package gateway

import (
	"backend/internal/common"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway",

	// Metodi pubblici
	fx.Provide(
		NewGatewayController,
		
		// Use Cases (inbound ports)
		common.FxAs[CreateGatewayUseCase](NewCreateGatewayService),
		common.FxAs[DeleteGatewayUseCase](NewCreateGatewayService),

		// Outbound ports
		common.FxAs[SaveGatewayPort](NewGatewayPostgreAdapter),

	),

	// Metodi privati
	fx.Provide(
		fx.Private,
		newGatewayPostgreRepository,
	),
)

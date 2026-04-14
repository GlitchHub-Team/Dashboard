package hello

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway_connection",
	fx.Provide(
		NewGatewayHelloService,
		NewConsumer,
		NewNATSWorker,
	),
	fx.Invoke(
		func(worker *NATSWorker, lc fx.Lifecycle) {
			worker.Run(lc)
		},
	),
)

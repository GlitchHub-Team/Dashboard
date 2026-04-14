package hello

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"gateway_hello",
	fx.Provide(
		fx.Annotate(
			NewGatewayHelloService,
			fx.As(new(GatewayHelloUseCase)),
		),
		NewConsumer,
		NewNATSWorker,
	),
	fx.Invoke(
		func(worker *NATSWorker, lc fx.Lifecycle) {
			worker.Run(lc)
		},
	),
)

package middlewares

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"http_middlewares",
	fx.Provide(
		NewAuthzMiddleware,
	),
)

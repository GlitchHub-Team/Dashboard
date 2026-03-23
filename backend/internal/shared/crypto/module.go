package crypto

import (
	"go.uber.org/fx"
)


var Module = fx.Module(
	"crypto",
	fx.Provide(
		fx.Annotate(
			NewBcryptHasher,
			fx.As(new(SecretHasher)),
		),
		fx.Annotate(
			NewMainTokenGenerator,
			fx.As(new(TokenGenerator)),
		),
	),
)

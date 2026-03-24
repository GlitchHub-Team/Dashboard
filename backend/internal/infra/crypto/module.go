package crypto

import (
	"backend/internal/shared/crypto"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"crypto",
	fx.Provide(
		fx.Annotate(
			NewBcryptHasher,
			fx.As(new(crypto.SecretHasher)),
		),
		fx.Annotate(
			NewMainTokenGenerator,
			fx.As(new(crypto.SecurityTokenGenerator)),
		),
		fx.Annotate(
			NewJWTManager,
			fx.As(new(crypto.AuthTokenManager)),
		),
	),
)

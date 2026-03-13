package common

import (
	"go.uber.org/fx"
)


func FxAs[T any](f any) any {
	return fx.Annotate(
		f,
		fx.As(new(T)),
	)
}
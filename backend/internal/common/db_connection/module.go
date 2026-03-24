package db_connection

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"db_connection",
	fx.Provide(
		NewDatabaseConnection,
	),
)

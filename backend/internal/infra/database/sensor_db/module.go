package sensordb

import (
	"backend/internal/infra/database/sensor_db/connection"
	"backend/internal/infra/database/sensor_db/migrate"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"sensordb",

	fx.Provide(
		connection.NewTimescaleDBConnection,

		fx.Annotate(
			migrate.NewSensorDBMigrator,
			fx.As(new(migrate.Migrator)),
		),
	),

	fx.Invoke(connection.SetSensorDbLifecycle),
)

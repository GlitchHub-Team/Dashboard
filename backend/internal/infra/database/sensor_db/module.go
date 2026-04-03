package sensordb

import (
	"os"

	"backend/internal/infra/utils"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"sensordb",

	fx.Supply(SensorDBAddress(os.Getenv("POSTGRES_HOST"))),
	fx.Supply(SensorDBPort(utils.EnvInt("POSTGRES_PORT", 5432))),
	fx.Supply(SensorDBUsername(os.Getenv("POSTGRES_USER"))),
	fx.Supply(SensorDBPassword(os.Getenv("POSTGRES_PASSWORD"))),
	fx.Supply(SensorDBName(os.Getenv("POSTGRES_DB"))),

	fx.Provide(NewTimescaleDBConnection),
)

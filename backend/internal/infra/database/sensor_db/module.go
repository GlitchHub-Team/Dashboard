package sensordb

import (
	"go.uber.org/fx"
)

var Module = fx.Module(
	"sensordb",

	fx.Provide(NewTimescaleDBConnection),
	fx.Invoke(SetSensorDbLifecycle),
)

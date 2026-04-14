package modules

import (
	"backend/internal/alert"
	"backend/internal/api_key"
	"backend/internal/auth"
	"backend/internal/email"
	"backend/internal/gateway"
	gateway_connection "backend/internal/gateway/hello"
	"backend/internal/historical_data"
	"backend/internal/infra/crypto"
	"backend/internal/infra/database/cloud_db"
	sensordb "backend/internal/infra/database/sensor_db"
	"backend/internal/infra/transport/http/router"
	"backend/internal/real_time_data"
	"backend/internal/sensor"
	"backend/internal/shared/config"
	"backend/internal/tenant"
	"backend/internal/user"

	httpMiddlewares "backend/internal/infra/transport/http/middlewares"
	transportNats "backend/internal/infra/transport/nats"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func AppModules() fx.Option {
	return fx.Options(
		// Moduli infrastrutturali
		config.Module,
		crypto.Module,
		cloud_db.Module, // NOTA: Questo esegue la migrazione PRIMA di eseguire NewGinEngine()
		email.Module,
		httpMiddlewares.Module,
		sensordb.Module,
		transportNats.Module,

		// Moduli funzionalità
		alert.Module,   // NOTA: Desiderabile
		api_key.Module, // NOTA: Desiderabile
		auth.Module,
		gateway.Module,
		gateway_connection.Module,
		historical_data.Module,
		real_time_data.Module,
		sensor.Module,
		tenant.Module,
		user.Module,

		fx.Provide(
			router.NewGinEngine,
			zap.NewExample,
		),
	)
}

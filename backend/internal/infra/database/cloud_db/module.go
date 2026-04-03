package cloud_db

import (
	"fmt"
	"os"

	"backend/internal/infra/database/cloud_db/connection"
	"backend/internal/infra/utils"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"cloud_db",
	fx.Supply(connection.CloudDBAddress(os.Getenv("CLOUD_POSTGRES_HOST"))),
	fx.Supply(connection.CloudDBPort(utils.EnvInt("CLOUD_POSTGRES_PORT", 5432))),
	fx.Supply(connection.CloudDBUsername(os.Getenv("CLOUD_POSTGRES_USER"))),
	fx.Supply(connection.CloudDBPassword(os.Getenv("CLOUD_POSTGRES_PASSWORD"))),
	fx.Supply(connection.CloudDBName(os.Getenv("CLOUD_POSTGRES_DB"))),

	fx.Provide(
		connection.NewDatabaseConnection,
		NewPostgreMigrator,
	),
	fx.Invoke(
		func(migrator Migrator) {
			err := migrator.Migrate()
			if err != nil {
				panic(fmt.Errorf("migrator error: %v", err))
			}
		},
	),
)

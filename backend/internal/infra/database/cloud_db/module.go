package cloud_db

import (
	"fmt"

	"backend/internal/infra/database/cloud_db/connection"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"cloud_db",

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

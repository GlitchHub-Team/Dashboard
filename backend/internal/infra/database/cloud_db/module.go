package cloud_db

import (
	"fmt"

	"backend/internal/infra/database/cloud_db/connection"
	"backend/internal/infra/database/cloud_db/migrate"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"cloud_db",
	fx.Provide(
		connection.NewDatabaseConnection,
		migrate.NewPostgreMigrator,
	),
	fx.Invoke(
		func(migrator migrate.Migrator) {
			err := migrator.Migrate()
			if err != nil {
				panic(fmt.Errorf("migrator error: %v", err))
			}
		},
	),
)

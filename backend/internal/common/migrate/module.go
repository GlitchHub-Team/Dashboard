package migrate

import (
	"fmt"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"migrate",
	fx.Provide(
		NewPostgreMigrator,
	),
	fx.Invoke(
		func(migrator DbMigrator) {
			err := migrator.Migrate()
			if err != nil {
				panic(fmt.Errorf("Migrator error: %v", err))
			}
		},
	),
)
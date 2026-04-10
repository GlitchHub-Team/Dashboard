package cloud_db

import (
	"fmt"

	"backend/internal/infra/database/cloud_db/connection"
	"backend/internal/infra/database/cloud_db/migrate"
	"backend/internal/shared/config"
	"backend/internal/tenant"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"cloud_db",

	fx.Provide(
		connection.NewCloudDbConnection,
		fx.Annotate(
			migrate.NewCloudDBMigrator,
			fx.As(new(migrate.Migrator)),
			fx.As(new(localCloudMigrator)),
			fx.As(new(tenant.LocalCloudMigrator)),
		),
	),

	fx.Invoke(connection.SetCloudDbLifecycle),
	fx.Invoke(
		func(tenantRepo *tenant.TenantPostgreRepository, cfg *config.Config, migrator localCloudMigrator) {
			err := migrateAll(tenantRepo, migrator, !cfg.CloudDBTest) // Imposta dati di default solo se non sto testando
			if err != nil {
				panic(fmt.Errorf("migrator error: %v", err))
			}
		},
	),
)

package migrate

import (
	"fmt"

	"backend/internal/config"
	"backend/internal/tenant"
	"backend/internal/user"

	// "backend/internal/user"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	// "fmt"
)

type DbMigrator interface {
	Migrate() error
}

type PostgreMigrator struct {
	log *zap.Logger
	cfg *config.Config
	// getTenantsPort tenant.GetTenantsPort
	getTenantsRepo *tenant.TenantPostgreRepository
}

func NewPostgreMigrator(
	log *zap.Logger,
	cfg *config.Config,
	getTenantsRepo *tenant.TenantPostgreRepository,
) DbMigrator {
	return &PostgreMigrator{
		log:            log,
		cfg:            cfg,
		getTenantsRepo: getTenantsRepo,
	}
}

func (migrator *PostgreMigrator) Migrate() error {
	/* Entity che sono associate allo schema public */
	publicEntities := []any{
		&user.SuperAdminEntity{},
		&tenant.TenantEntity{},
	}

	/* Entity da associare a uno schema tenant specifico */
	tenantSchemaEntities := [](interface{ TableName() string }){
		&user.TenantMemberEntity{},
	}

	migrator.log.Info("[Migrator] started")

	db, err := gorm.Open(
		postgres.Open(migrator.cfg.CloudDBUrl),
		&gorm.Config{},
	)
	if err != nil {
		return fmt.Errorf("error open migrator connection: %v", err)
	}
	migrator.log.Info("[Migrator] Open connection")

	// Migrate entities for public schema
	err = db.AutoMigrate(publicEntities...)
	if err != nil {
		return err
	}

	// Get tenants
	tenants, err := migrator.getTenantsRepo.GetAllTenants()
	if err != nil {
		return err
	}

	// Create schemas
	for _, tenant := range tenants {
		schemaName := fmt.Sprintf("tenant_%v", tenant.ID)
		if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%v\"", schemaName)).Error; err != nil {
			return fmt.Errorf("error creating schema %v: %v", schemaName, err)
		}
		migrator.log.Sugar().Infof("[Migrator] Migrated schema %v", schemaName)
	}

	// Migrate entities for each schema
	for _, tenant := range tenants {
		tenantId := tenant.ID
		if tenantId == "" {
			continue
		}

		schemaName := fmt.Sprintf("tenant_%v", tenantId)

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(fmt.Sprintf("set local search_path to \"%s\"", schemaName)).Error; err != nil {
				return fmt.Errorf("failed to set search_path to %s: %v", schemaName, err)
			}

			for _, entity := range tenantSchemaEntities {
				migrator.log.Sugar().Infof("Migrating %v", entity.TableName())

				if err = tx.AutoMigrate(entity); err != nil {
					return fmt.Errorf("error migrating table %v: %v", entity.TableName(), err)
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error migrating tenant %v: %v", tenantId, err)
		}
	}

	return nil
}

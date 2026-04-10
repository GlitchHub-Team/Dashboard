package migrate

import (
	"fmt"

	"backend/internal/auth"
	"backend/internal/gateway"
	"backend/internal/infra/database/cloud_db/connection"
	"backend/internal/sensor"
	"backend/internal/shared/config"
	"backend/internal/tenant"
	"backend/internal/user"

	"backend/internal/shared/crypto"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

/* Entity che sono associate allo schema public */
var publicEntities = []any{
	&tenant.TenantEntity{},
	&gateway.GatewayEntity{},
	&sensor.SensorEntity{},
	&user.SuperAdminEntity{},
	&auth.SuperAdminConfirmTokenEntity{},
	&auth.SuperAdminPasswordTokenEntity{},
}

/* Entity da associare a uno schema tenant specifico */
var tenantSchemaEntities = [](interface{ TableName() string }){
	&auth.TenantConfirmTokenEntity{},
	&auth.TenantPasswordTokenEntity{},
	&user.TenantMemberEntity{},
}

func GetPublicEntities() []any {
	return publicEntities
}

func GetTenantSchemaEntities() [](interface{ TableName() string }) {
	return tenantSchemaEntities
}

type Migrator interface {
	// MigrateAll(setDefaultData bool) error

	Logger() *zap.Logger

	MigratePublic() error
	MigrateTenantSchema(tenantId string, shouldLog bool) error
}

type CloudDBMigrator struct {
	log    *zap.Logger
	cfg    *config.Config
	db     *gorm.DB
	hasher crypto.SecretHasher
}

var _ Migrator = (*CloudDBMigrator)(nil)

func NewCloudDBMigrator(
	log *zap.Logger,
	cfg *config.Config,
	db connection.CloudDBConnection,
	hasher crypto.SecretHasher,
) *CloudDBMigrator {
	return &CloudDBMigrator{
		log:    log,
		cfg:    cfg,
		db:     (*gorm.DB)(db),
		hasher: hasher,
	}
}

func (migrator *CloudDBMigrator) DB() *gorm.DB                { return migrator.db }
func (migrator *CloudDBMigrator) Logger() *zap.Logger         { return migrator.log }
func (migrator *CloudDBMigrator) Hasher() crypto.SecretHasher { return migrator.hasher }

func (migrator *CloudDBMigrator) MigratePublic() error {
	// Migra entity per schema public
	err := migrator.db.AutoMigrate(GetPublicEntities()...)
	return err
}

func (migrator *CloudDBMigrator) MigrateTenantSchema(tenantId string, shouldLog bool) error {
	schemaName := fmt.Sprintf("tenant_%v", tenantId)

	// Crea schema
	if err := migrator.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%v\"", schemaName)).Error; err != nil {
		return fmt.Errorf("error creating schema %v: %v", schemaName, err)
	}

	if shouldLog {
		migrator.log.Sugar().Infof("[Migrator] Migrated schema %v", schemaName)
	}

	// Migra tutte le tabelle
	err := migrator.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(fmt.Sprintf("set local search_path to \"%s\"", schemaName)).Error; err != nil {
			return fmt.Errorf("failed to set search_path to %s: %v", schemaName, err)
		}

		for _, entity := range GetTenantSchemaEntities() {
			if shouldLog {
				migrator.log.Sugar().Infof("Migrating %v", entity.TableName())
			}

			if err := tx.AutoMigrate(entity); err != nil {
				return fmt.Errorf("error migrating table %v: %v", entity.TableName(), err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error migrating tenant %v: %v", tenantId, err)
	}

	return nil
}

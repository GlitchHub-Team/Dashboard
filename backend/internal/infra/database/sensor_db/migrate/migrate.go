package migrate

import (
	"backend/internal/infra/database/schema"
	"backend/internal/infra/database/sensor_db/connection"
	"backend/internal/shared/config"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Migrator interface {
	/*
		Crea uno schema per il sensor DB
	*/
	MigrateTenantSchema(tenantId string) error
	DeleteTenantSchema(tenantId string) error
}

type SensorDBMigrator struct {
	log *zap.Logger
	cfg *config.Config
	db  *gorm.DB
}

var _ Migrator = (*SensorDBMigrator)(nil)

func NewSensorDBMigrator(
	log *zap.Logger,
	cfg *config.Config,
	db connection.SensorDBConnection,
) *SensorDBMigrator {
	return &SensorDBMigrator{
		log: log,
		cfg: cfg,
		db:  (*gorm.DB)(db),
	}
}

/*
Migra lo schema del sensor DB per il tenant con id tenantId.

Possibile miglioria: utilizzare GORM (anche in backend/internal/historical_data/repository.go) e usare AutoMigrate()
*/
func (migrator *SensorDBMigrator) MigrateTenantSchema(tenantId string) error {
	schemaName := schema.GetSchemaName(tenantId)

	if err := schema.CreateSchema(migrator.db, schemaName); err != nil {
		return err
	}

	// NOTA: Qui si potrebbe utilizzare AutoMigrate di GORM
	migrator.db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%v".sensor_data (
			sensor_id UUID NOT NULL,
			gateway_id UUID NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL,
			tenant_id UUID NOT NULL,
			profile VARCHAR(255) NOT NULL,
			data JSONB NOT NULL,
			PRIMARY KEY (sensor_id, gateway_id, timestamp)
		);
	`, schemaName))

	return nil
}

func (migrator *SensorDBMigrator) DeleteTenantSchema(tenantId string) error {
	schemaName := schema.GetSchemaName(tenantId)

	if err := schema.DropSchema(migrator.db, schemaName); err != nil {
		return err
	}

	return nil
}

package migrate

import (
	"backend/internal/infra/database/schema"
	"backend/internal/infra/database/sensor_db/connection"
	"backend/internal/shared/config"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Migrator interface {
	/*
		Crea uno schema per il sensor DB
	*/
	CreateTenantSchema(tenantId string) error
	DeleteTenantSchema(tenantId string) error
}

type SensorDBMigrator struct {
	log *zap.Logger
	cfg *config.Config
	db  *gorm.DB
}

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

func (migrator *SensorDBMigrator) CreateTenantSchema(tenantId string) error {
	schemaName := schema.GetSchemaName(tenantId)

	if err := schema.CreateSchema(migrator.db, schemaName); err != nil {
		return err
	}

	return nil
}

func (migrator *SensorDBMigrator) DeleteTenantSchema(tenantId string) error {
	schemaName := schema.GetSchemaName(tenantId)

	if err := schema.DropSchema(migrator.db, schemaName); err != nil {
		return err
	}

	return nil
}

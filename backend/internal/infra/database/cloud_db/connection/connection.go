package connection

import (
	"fmt"

	"backend/internal/shared/config"
	dbPackage "backend/internal/infra/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDatabaseConnection(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.CloudDBUrl
	db, err := gorm.Open(
		postgres.Open(dsn), &gorm.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire DB: %v", err)
	}

	return db, nil
}



func WithTenantSchema(tenantId string, table dbPackage.Tabler) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Table(
			fmt.Sprintf("\"tenant_%s\".\"%s\"", tenantId, table.TableName()),
		)
	}
}

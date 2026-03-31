package connection

import (
	"context"
	"fmt"
	"time"

	dbPackage "backend/internal/infra/database"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CloudDBConnection *gorm.DB

type (
	CloudDBAddress  string
	CloudDBPort     int
	CloudDBUsername string
	CloudDBPassword string
	CloudDBName     string
)

func NewDatabaseConnection(addr CloudDBAddress, port CloudDBPort, user CloudDBUsername, pass CloudDBPassword, dbname CloudDBName) (CloudDBConnection, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		addr, port, user, pass, dbname,
	)
	db, err := gorm.Open(
		postgres.Open(dsn), &gorm.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire connessione Postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("impossibile ottenere connessione SQL da GORM: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("impossibile raggiungere Postgres: %w", err)
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

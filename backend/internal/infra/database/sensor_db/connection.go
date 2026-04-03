package sensordb

import (
	"backend/internal/shared/config"
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SensorDBConnection *gorm.DB

func NewTimescaleDBConnection(
	// addr SensorDBAddress, port SensorDBPort, user SensorDBUsername, pass SensorDBPassword, dbname SensorDBName,
	cfg *config.Config,
) (SensorDBConnection, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.SensorDBHost, int(cfg.SensorDBPort), cfg.SensorDBUser, cfg.SensorDBPassword, cfg.SensorDBName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire connessione TimescaleDB: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("impossibile ottenere connessione SQL da GORM: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("impossibile raggiungere TimescaleDB: %w", err)
	}
	return db, nil
}

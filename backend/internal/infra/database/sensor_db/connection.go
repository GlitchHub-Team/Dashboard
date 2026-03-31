package sensordb

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SensorDBConnection *gorm.DB

type (
	SensorDBAddress  string
	SensorDBPort     int
	SensorDBUsername string
	SensorDBPassword string
	SensorDBName     string
)

func NewTimescaleDBConnection(addr SensorDBAddress, port SensorDBPort, user SensorDBUsername, pass SensorDBPassword, dbname SensorDBName) (SensorDBConnection, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		addr, port, user, pass, dbname,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

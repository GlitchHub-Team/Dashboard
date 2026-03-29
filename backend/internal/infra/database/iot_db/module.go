package iot_db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
)

func NewDatabaseConnection(cfg *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresUser,
		cfg.PostgresPass,
		cfg.PostgresDB,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire connessione IoT DB: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("impossibile raggiungere IoT DB: %w", err)
	}

	return db, nil
}

var Module = fx.Module(
	"iot_db",
	fx.Provide(
		NewDatabaseConnection,
	),
)

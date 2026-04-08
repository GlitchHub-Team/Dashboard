package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
Crea una nuova connessione all'engine Postgres del database con host, port, user e password specificati.
*/
func NewPostgresEngineConnection(host string, port int, user, password string) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire connessione Postgres: %w", err)
	}
	return db, err
}

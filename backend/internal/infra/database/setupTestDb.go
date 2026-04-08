package database

import (
	"fmt"
	"strings"
	"time"

	"backend/internal/shared/config"

	"go.uber.org/zap"
)

/*
Quale database impostare con SetupTestDatabase
*/
type SetupTestDbEnum string

const (
	SETUP_TEST_CLOUD_DB  SetupTestDbEnum = "cloud"  // Imposta Cloud DB di test
	SETUP_TEST_SENSOR_DB SetupTestDbEnum = "sensor" // Imposta Sensor DB di test
)

func getTestDbName(oldName string) string {
	dateTime := strings.ReplaceAll(time.Now().Format("20060102_150405_.999999999"), ".", "")
	return fmt.Sprintf("test_%v__%v", dateTime, oldName)
}

/*
Crea un database di test e cambia la configurazione cfg in modo tale che cfg.CloudDBName rispecchi
il nome del database di test appena creato.

In caso venga aggiunto un nuovo database, bisogna aggiungere un valore a SetupTestDbEnum e aggiungere il
case nello switch dentro la funzione

NOTA: L'ultimo passaggio è necessario perché il nome del database di test comprende la data di esecuzione al
nanosecondo.
*/
func SetupTestDatabase(
	log *zap.Logger,
	cfg *config.Config,
	whichDb SetupTestDbEnum,
) error {
	var host, user, password string
	var port int
	switch whichDb {
	case SETUP_TEST_CLOUD_DB:
		host = cfg.CloudDBHost
		port = int(cfg.CloudDBPort)
		user = cfg.CloudDBUser
		password = cfg.CloudDBPassword

	case SETUP_TEST_SENSOR_DB:
		host = cfg.SensorDBHost
		port = int(cfg.SensorDBPort)
		user = cfg.SensorDBUser
		password = cfg.SensorDBPassword

	default:
		return fmt.Errorf("valore sconosciuto per SetupTestDbEnum: %v", whichDb)
	}

	db, err := NewPostgresEngineConnection(host, port, user, password)
	if err != nil {
		return err
	}

	var inputName string
	switch whichDb {
	case SETUP_TEST_CLOUD_DB:
		inputName = cfg.CloudDBName

	case SETUP_TEST_SENSOR_DB:
		inputName = cfg.SensorDBName
	}

	newDatabaseName := getTestDbName(inputName)
	log.Info("", zap.String("newDatabaseName", newDatabaseName))

	switch whichDb {
	case SETUP_TEST_CLOUD_DB:
		cfg.CloudDBName = newDatabaseName
	case SETUP_TEST_SENSOR_DB:
		cfg.SensorDBName = newDatabaseName
	}

	if err := db.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", newDatabaseName)).Error; err != nil {
		return fmt.Errorf("impossibile creare database di test %v (%v): %v", whichDb, newDatabaseName, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("impossibile ottenere SQL DB (DB generico): %w", err)
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("impossibile chiudere SQL DB (DB generico): %w", err)
	}

	log.Info(
		"Creato database di test con successo",
		zap.String("type", string(whichDb)),
		zap.String("dbName", newDatabaseName),
	)

	return nil
}

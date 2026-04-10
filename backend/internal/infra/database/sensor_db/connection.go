package sensordb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/shared/config"

	dbPackage "backend/internal/infra/database"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SensorDBConnection *gorm.DB

func NewTimescaleDBConnection(
	// addr SensorDBAddress, port SensorDBPort, user SensorDBUsername, pass SensorDBPassword, dbname SensorDBName,
	log *zap.Logger,
	cfg *config.Config,
) (SensorDBConnection, error) {
	if cfg.CloudDBTest {
		err := dbPackage.SetupTestDatabase(log, cfg, dbPackage.SETUP_TEST_SENSOR_DB)
		if err != nil {
			return nil, err
		}
	}

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

func SetSensorDbLifecycle(
	lc fx.Lifecycle,
	log *zap.Logger,
	cfg *config.Config,
	sensorDB SensorDBConnection,
) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Start Sensor DB")
			return nil
		},
		OnStop: func(context.Context) error {
			log.Info("Stop Sensor DB", zap.Bool("isTest", cfg.SensorDBTest))

			if sensorDB != nil {
				sqlDB, err := (*gorm.DB)(sensorDB).DB()
				if err != nil {
					log.Warn("impossibile ottenere connessione SQL da GORM durante stop Sensor DB", zap.Error(err))
				} else if err := sqlDB.Close(); err != nil {
					log.Warn("impossibile chiudere connessione SQL Sensor DB", zap.Error(err))
				}
			}

			// Se modalità di test, allora elimina database perché non serve più.
			if cfg.SensorDBTest {
				db, err := dbPackage.NewPostgresEngineConnection(
					cfg.SensorDBHost,
					int(cfg.SensorDBPort),
					cfg.SensorDBUser,
					cfg.SensorDBPassword,
				)
				if err != nil {
					return fmt.Errorf("impossibile eliminare CSensoroud DB di test: %v", err)
				}

				sqlDB, err := db.DB()
				if err != nil {
					log.Warn("impossibile ottenere connessione SQL engine per Sensor DB cleanup", zap.Error(err))
				} else {
					defer func() {
						if closeErr := sqlDB.Close(); closeErr != nil {
							log.Warn("impossibile chiudere connessione SQL engine Sensor DB", zap.Error(closeErr))
						}
					}()
				}

				if !strings.HasPrefix(cfg.SensorDBName, "test_") {
					return fmt.Errorf(
						"/!\\ ATTENZIONE: è stata attivata la modalità di test su Cloud DB (cfg.CloudDbTest == true),"+
							" ma cfg.CloudDbName == \"%v\" (non inizia con 'test_')."+
							" Se questo errore viene mostrato, probabilmente cfg.CloudDbTest è stato impostato a true per errore."+
							" Per evitare eliminazioni sgradevoli, il database %v non verrà eliminato",
						cfg.SensorDBName,
						cfg.SensorDBName,
					)
				}

				err = dbPackage.SeverDropDatabase(log, db, cfg.CloudDBName, "Cloud DB Test")
				if err != nil {
					return err
				}
			}

			return nil
		},
	})
}

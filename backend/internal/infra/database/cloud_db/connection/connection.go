package connection

import (
	"context"
	"fmt"
	"time"

	"backend/internal/infra/database"
	dbPackage "backend/internal/infra/database"
	"backend/internal/shared/config"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type CloudDBConnection *gorm.DB

func NewCloudDbConnection(
	log *zap.Logger,
	cfg *config.Config,
) (CloudDBConnection, error) {
	dbConfig := &gorm.Config{TranslateError: true}

	// 1. Se uso modalità test, modifica la configurazione e crea il DB temporaneo ====================
	if cfg.CloudDBTest {
		err := dbPackage.SetupTestDatabase(log, cfg, dbPackage.SETUP_TEST_CLOUD_DB)
		if err != nil {
			return nil, err
		}
	}

	// 2. Apri connessione database =========================================================
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.CloudDBHost, int(cfg.CloudDBPort), cfg.CloudDBUser, cfg.CloudDBPassword, cfg.CloudDBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), dbConfig)
	if err != nil {
		return nil, fmt.Errorf("impossibile aprire connessione Postgres: %w", err)
	}

	// 3. Verifica connessione con DB =================================================================
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("impossibile ottenere connessione SQL da GORM: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("impossibile raggiungere Postgres: %w", err)
	}

	// 4. Ritorna =====================================================================================
	return db, nil
}

func SetCloudDbLifecycle(
    lc fx.Lifecycle,
    log *zap.Logger,
    cfg *config.Config,
) {
    targetDBName := cfg.CloudDBName

    lc.Append(fx.Hook{
        OnStart: func(context.Context) error {
            log.Info("Start Cloud DB")
            return nil
        },
        OnStop: func(context.Context) error {
            log.Info("Stop Cloud DB", zap.Bool("isTest", cfg.CloudDBTest))

            if cfg.CloudDBTest {
                db, err := dbPackage.NewPostgresEngineConnection(
                    cfg.CloudDBHost,
                    int(cfg.CloudDBPort),
                    cfg.CloudDBUser,
                    cfg.CloudDBPassword,
                )
                if err != nil {
                    return fmt.Errorf("impossibile eliminare Cloud DB di test: %v", err)
                }

                defer func() {
                    if engineDb, err := db.DB(); err == nil {
                        _ = engineDb.Close()
                    }
                }()

                if targetDBName[:5] != "test_" {
                    return fmt.Errorf( //nolint:staticcheck
                        "/!\\ ATTENZIONE: è stata attivata la modalità di test su Cloud DB (cfg.CloudDbTest == true),"+
                            " ma cfg.CloudDbName == \"%v\" (non inizia con 'test_')."+
                            " Se questo errore viene mostrato, probabilmente cfg.CloudDbTest è stato impostato a true per errore."+
                            " Per evitare eliminazioni sgradevoli, il database %v non verrà eliminato.",
                        targetDBName,
                        targetDBName,
                    )
                }

				err = database.SeverDropDatabase(log, db, targetDBName, "Cloud DB test")
				if err != nil {
					return err
				}
            }

            return nil
        },
    })
}

func WithTenantSchema(tenantId string, table dbPackage.Tabler) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Table(
			fmt.Sprintf("\"tenant_%s\".\"%s\"", tenantId, table.TableName()),
		)
	}
}

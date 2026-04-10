package connection

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	cloudDB CloudDBConnection,
) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Info("Start Cloud DB")
			return nil
		},
		OnStop: func(context.Context) error {
			log.Info("Stop Cloud DB", zap.Bool("isTest", cfg.CloudDBTest))

			if cloudDB != nil {
				sqlDB, err := (*gorm.DB)(cloudDB).DB()
				if err != nil {
					log.Warn("impossibile ottenere connessione SQL da GORM durante stop Cloud DB", zap.Error(err))
				} else if err := sqlDB.Close(); err != nil {
					log.Warn("impossibile chiudere connessione SQL Cloud DB", zap.Error(err))
				}
			}

			// Se modalità di test, allora elimina database perché non serve più.
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

				sqlDB, err := db.DB()
				if err != nil {
					log.Warn("impossibile ottenere connessione SQL engine per Cloud DB cleanup", zap.Error(err))
				} else {
					defer func() {
						if closeErr := sqlDB.Close(); closeErr != nil {
							log.Warn("impossibile chiudere connessione SQL engine Cloud DB", zap.Error(closeErr))
						}
					}()
				}

				if !strings.HasPrefix(cfg.CloudDBName, "test_") {
					return fmt.Errorf("ATTENZIONE: è stata attivata la modalità di test su Cloud DB (cfg.CloudDbTest == true),"+
						" ma cfg.CloudDbName == \"%v\" (non inizia con 'test_')."+
						" Se questo errore viene mostrato, probabilmente cfg.CloudDbTest è stato impostato a true per errore."+
						" Per evitare eliminazioni sgradevoli, il database %v non verrà eliminato",
						cfg.CloudDBName,
						cfg.CloudDBName,
					)
				}
            }

				if err := db.Exec(
					"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = ? AND pid <> pg_backend_pid()",
					cfg.CloudDBName,
				).Error; err != nil {
					log.Warn("impossibile terminare connessioni attive Cloud DB", zap.String("name", cfg.CloudDBName), zap.Error(err))
				}

				err = db.Exec(fmt.Sprintf("DROP DATABASE \"%s\"", cfg.CloudDBName)).Error
				if err != nil {
					return fmt.Errorf("impossibile eliminare Cloud DB di test %v: %w", cfg.CloudDBName, err)
				}

				log.Info("Eliminato cloud db di test", zap.String("name", cfg.CloudDBName))
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

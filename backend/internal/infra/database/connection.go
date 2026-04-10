package database

import (
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Elimina forzatamente tutte le connessioni attive al database db di nome targetDBName ("sever") prima di eliminarlo ("drop").

NOTA: Questa funzione dev'essere chiamata SINCRONAMENTE e non dovrebbe causare problemi con OnStop nel lifecycle Fx.
*/
func SeverDropDatabase(log *zap.Logger, db *gorm.DB, targetDBName, dbType string) error {
	terminateQuery := fmt.Sprintf(`
		SELECT pg_terminate_backend(pg_stat_activity.pid)
		FROM pg_stat_activity
		WHERE pg_stat_activity.datname = '%s'
			AND pid <> pg_backend_pid();
	`, targetDBName)

	err := db.Exec(terminateQuery).Error
	if err != nil {
		log.Sugar().Warnf("impossibile forzare chiusura di connessioni a %v \"%v\": %v", err)
	}

	err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\" ", targetDBName)).Error
	if err != nil {
		return fmt.Errorf("impossibile eliminare %v \"%v\": %w", dbType, targetDBName, err)
	}

	log.Info("Eliminato con successo %v \"%v\"")

	return nil
}

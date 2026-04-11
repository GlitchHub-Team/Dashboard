package schema

/*
Funzioni comuni alle funzioni di migration
*/

import (
	"fmt"

	"gorm.io/gorm"
)

func GetSchemaName(tenantId string) string {
	return fmt.Sprintf("tenant_%s", tenantId)
}

func CreateSchema(db *gorm.DB, schemaName string) error {
	if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%v\"", schemaName)).Error; err != nil {
		return fmt.Errorf("error creating schema %v: %w", schemaName, err)
	}
	return nil
}

func DropSchema(db *gorm.DB, schemaName string) error {
	if err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS \"%v\" CASCADE", schemaName)).Error; err != nil {
		return fmt.Errorf("error dropping schema %v: %w", schemaName, err)
	}
	return nil
}

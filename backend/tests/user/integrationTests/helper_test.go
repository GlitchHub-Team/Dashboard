package user_integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/auth"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"

	clouddb "backend/internal/infra/database/cloud_db/connection"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func mustJSONBody(t *testing.T, payload any) *bytes.Reader {
	t.Helper()

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal json payload: %v", err)
	}
	return bytes.NewReader(jsonBody)
}

func authHeader(jwt string) http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+jwt)
	return header
}

func checkNoTenant(tenantId string) helper.IntegrationTestCheck {
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		var row *struct{ SchemaName string }

		schemaName := "tenant_" + tenantId
		// https://stackoverflow.com/questions/7016419/postgresql-check-if-schema-exists
		_ = db.Exec(
			fmt.Sprintf("SELECT schema_name FROM information_schema.schemata WHERE schema_name = '%v'", schemaName),
		).Find(row)

		return row == nil
	}
}

func checkNoTenantMember(email string, tenantId string) helper.IntegrationTestCheck {
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		if checkNoTenant(tenantId)(respRecorder, deps) {
			return true
		}

		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		db.Scopes(clouddb.WithTenantSchema(tenantId, &user.TenantMemberEntity{})).
			Model(&user.TenantMemberEntity{}).
			Where("email = ?", email).
			Count(&count)
		return count == 0
	}
}

func checkNoSuperAdmin(email string) helper.IntegrationTestCheck {
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		db.
			Model(&user.SuperAdminEntity{}).
			Where("email = ?", email).
			Count(&count)
		return count == 0
	}
}

func checkTenantMemberInserted(email string, tenantId string) helper.IntegrationTestCheck {
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		db.
			Scopes(clouddb.WithTenantSchema(tenantId, &user.TenantMemberEntity{})).
			Model(&user.TenantMemberEntity{}).Where("email = ?", email).
			Count(&count)
		return count == 1
	}
}

func checkSuperAdminInserted(email string) helper.IntegrationTestCheck {
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		db.
			Model(&user.SuperAdminEntity{}).Where("email = ?", email).
			Count(&count)
		return count == 1
	}
}


func preSetupCreateTenant(tenantId uuid.UUID) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		tenantEntity := tenant.TenantEntity{ID: tenantId.String(), Name: "test tenant", CanImpersonate: false}
		if err := db.Clauses().Create(&tenantEntity).Error; err != nil {
			return false
		}
		// create schema
		schemaName := "tenant_" + tenantId.String()
		if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", schemaName)).Error; err != nil {
			return false
		}
		// migrate tenant_members in this schema
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(fmt.Sprintf("set local search_path to \"%s\"", schemaName)).Error; err != nil {
				return err
			}
			return tx.AutoMigrate(&user.TenantMemberEntity{}, &auth.TenantConfirmTokenEntity{}, &auth.TenantPasswordTokenEntity{})
		}); err != nil {
			return false
		}
		return true
	}
}

func preSetupAddTenantMember(tenantId uuid.UUID, entity *user.TenantMemberEntity) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		err := db.Scopes(clouddb.WithTenantSchema(tenantId.String(), &user.TenantMemberEntity{})).
			Create(entity).
			Error
		return err == nil
	}
}

func preSetupAddSuperAdmin(entity *user.SuperAdminEntity) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		err := db.Create(entity).Error
		return err == nil
	}
}

func postSetupDeleteTenant(t *testing.T, tenantId uuid.UUID) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)
		schemaName := "tenant_" + tenantId.String()
		err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" CASCADE", schemaName)).Error
		if err != nil {
			t.Logf("Errore eliminando schema %v: %v", schemaName, err)
		}

		err = db.Where("id = ?", tenantId.String()).Delete(&tenant.TenantEntity{}).Error
		if err != nil {
			t.Logf("Errore eliminando tenant %v: %v", schemaName, err)
		}
	}
}

func postSetupDeleteSuperAdmin(email string) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := user.SuperAdminEntity{}
		_ = db.Where("email = ?", email).Delete(&entity).Error
	}
}
package user_integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/user"
	"backend/tests/helper"

	clouddb "backend/internal/infra/database/cloud_db/connection"

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

package integration

import (
	"fmt"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"backend/internal/user"
	"backend/tests/helper"

	clouddb "backend/internal/infra/database/cloud_db/connection"

	"gorm.io/gorm"
)

func CheckNoTenant(tenantId string) helper.IntegrationTestCheck {
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

func CheckNoTenantMember(email string, tenantId string) helper.IntegrationTestCheck {
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		if CheckNoTenant(tenantId)(respRecorder, deps) {
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

func CheckNoSuperAdmin(email string) helper.IntegrationTestCheck {
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

func CheckTenantMemberInserted(email string, tenantId string) helper.IntegrationTestCheck {
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

func CheckSuperAdminInserted(email string) helper.IntegrationTestCheck {
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		db.
			Model(&user.SuperAdminEntity{}).Where("email = ?", email).
			Count(&count)
		return count == 1
	}
}

func CheckSMTPMessageForToken(t *testing.T, tokenType string, exists bool) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		messages, err := deps.MockSMTPServer.WaitForMessagesAndPurge(1, 50*time.Millisecond)
		if err != nil {
			if exists {
				t.Errorf("error waiting for SMTP messages: %v", err)
				return false
			} else {
				return true
			}
		}

		lineEndingRe := regexp.MustCompile(`=\s*\n`)
		contentEncodingRe := regexp.MustCompile(`Content-Transfer-Encoding: quoted-printable`)
		tokenRe := regexp.MustCompile(
			fmt.Sprintf(
				`/%v/[\d\w-_]+(\?tid=[0-9a-fA-F-]+)?`, tokenType,
			),
		)

		message := messages[0]
		body := message.MsgRequest()

		// Handle quoted-printable
		if contentEncodingRe.MatchString(body) {
			body = lineEndingRe.ReplaceAllString(body, "")
		}

		// Check token
		return tokenRe.MatchString(body)
	}
}

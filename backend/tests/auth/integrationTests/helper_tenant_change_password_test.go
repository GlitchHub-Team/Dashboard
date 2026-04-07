package auth_integration_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/auth"
	cloud_db "backend/internal/infra/database/cloud_db/connection"
	"backend/tests/helper"

	"gorm.io/gorm"
)

func CheckTenantForgotPasswordTokenExists(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantPasswordTokenEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &auth.TenantPasswordTokenEntity{})).
			Where("token = ?", hashedToken).
			Find(entity).
			Error

		return err == nil
	}
}

func CheckTenantForgotPasswordTokenExistsForUser(t *testing.T, tenantId string, userId uint) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantPasswordTokenEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &auth.TenantPasswordTokenEntity{})).
			Where("user_id = ?", userId).
			Find(entity).
			Error

		return err == nil
	}
}

func CheckNoTenantForgotPasswordTokenForUser(t *testing.T, tenantId string, userId uint) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantPasswordTokenEntity{}
		var count int64
		db.
			Scopes(cloud_db.WithTenantSchema(tenantId, entity)).
			Where("user_id = ?", userId).
			Count(&count)

		return count == 0
	}
}


func CheckTenantForgotPasswordTokenExpired(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantPasswordTokenEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &auth.TenantPasswordTokenEntity{})).
			Where("token = ?", hashedToken).
			Find(entity).
			Error
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
			return false
		}

		return entity.ExpiresAt.Before(time.Now())
	}
}

func CheckNoTenantForgotPasswordToken(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantPasswordTokenEntity{}
		var count int64
		db.
			Scopes(cloud_db.WithTenantSchema(tenantId, entity)).
			Where("token = ?", hashedToken).
			Count(&count)

		return count == 0
	}
}

func PreSetupAddTenantForgotPasswordToken(t *testing.T, entity auth.TenantPasswordTokenEntity) helper.IntegrationTestPreSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) bool {
		if entity.TenantId == "" {
			t.Logf("Impossibile creare token conferma per tenant nullo")
			return false
		}
		db := (*gorm.DB)(deps.CloudDB)

		err := db.
			Scopes(cloud_db.WithTenantSchema(entity.TenantId, &auth.TenantPasswordTokenEntity{})).
			Create(&entity).
			Error

		return err == nil
	}
}

func PostSetupDeleteTenantForgotPasswordToken(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestPostSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantPasswordTokenEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, entity)).
			Where("token = ?", hashedToken).
			Delete(entity).
			Error
		if err != nil {
			t.Logf("error deleting tenant confirm token: %v", err)
		}
	}
}

func PostSetupDeleteTenantForgotPasswordTokensForUser(t *testing.T, tenantId string, userId uint) helper.IntegrationTestPostSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := []auth.TenantPasswordTokenEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &auth.TenantPasswordTokenEntity{})).
			Where("user_id = ?", userId).
			Delete(entity).
			Error
		if err != nil {
			t.Logf("error deleting tenant confirm token: %v", err)
		}
	}
}

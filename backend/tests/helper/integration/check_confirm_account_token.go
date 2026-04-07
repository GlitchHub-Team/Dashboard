package integration

import (
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/auth"
	cloud_db "backend/internal/infra/database/cloud_db/connection"
	"backend/tests/helper"

	"gorm.io/gorm"
)

// TENANT MEMBER ================================================================================================================

func CheckCountTenantConfirmAccountTokens(t *testing.T, tenantId string, expectedCount int64) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		var count int64
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &auth.TenantConfirmTokenEntity{})).
			Count(&count).
			Error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		return count == expectedCount

	}
}

func CheckTenantConfirmAccountTokenExists(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantConfirmTokenEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &auth.TenantConfirmTokenEntity{})).
			Where("token = ?", hashedToken).
			Find(entity).
			Error

		return err == nil
	}
}

func CheckTenantConfirmAccountTokenExpired(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantConfirmTokenEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &auth.TenantConfirmTokenEntity{})).
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

func CheckNoTenantConfirmAccountToken(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantConfirmTokenEntity{}
		var count int64
		db.
			Scopes(cloud_db.WithTenantSchema(tenantId, entity)).
			Where("token = ?", hashedToken).
			Count(&count)

		return count == 0
	}
}

// SUPER ADMIN ================================================================================================================

func CheckCountSuperAdminConfirmAccountTokens(t *testing.T, expectedCount int64) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		var count int64
		err := db.
			Model(&auth.SuperAdminConfirmTokenEntity{}).
			Count(&count).
			Error
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		return count == expectedCount

	}
}

func CheckSuperAdminConfirmAccountTokenExists(t *testing.T, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.SuperAdminConfirmTokenEntity{}
		err := db.
			Model(&entity).
			Where("token = ?", hashedToken).
			Find(&entity).
			Error

		return err == nil
	}
}

func CheckNoSuperAdminConfirmAccountToken(t *testing.T, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		var count int64
		db.Model(&auth.SuperAdminConfirmTokenEntity{}).Where("token = ?", hashedToken).Count(&count)

		return count == 0
	}
}

func CheckSuperAdminConfirmAccountTokenExpired(t *testing.T, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.SuperAdminConfirmTokenEntity{}
		err := db.
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
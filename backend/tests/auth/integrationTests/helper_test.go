package auth_integration_test

import (
	"net/http/httptest"
	"testing"

	cloud_db "backend/internal/infra/database/cloud_db/connection"
	"backend/internal/user"
	"backend/tests/helper"

	"gorm.io/gorm"
)

func CheckTenantMemberConfirmed(t *testing.T, tenantId string, userId uint, confirmed bool) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &user.TenantMemberEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &user.TenantMemberEntity{})).
			Where("id = ?", userId).
			Where("confirmed = ?", confirmed).
			Find(entity).
			Error
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
			return false
		}

		return true
	}
}

func CheckSuperAdminConfirmed(t *testing.T, userId uint, confirmed bool) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &user.SuperAdminEntity{}
		err := db.
			Where("id = ?", userId).
			Where("confirmed = ?", confirmed).
			Find(entity).
			Error
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
			return false
		}

		return true
	}
}

func CheckTenantMemberPassword(t *testing.T, tenantId string, userId uint, expectedPassword string, expectedResult bool) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &user.TenantMemberEntity{}
		err := db.
			Scopes(cloud_db.WithTenantSchema(tenantId, &user.TenantMemberEntity{})).
			Where("id = ?", userId).
			Find(entity).
			Error
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
			return false
		}

		// treat nil password as empty string
		var storedPassword string
		if entity.Password != nil {
			storedPassword = *entity.Password
		}

		if expectedPassword == "" || storedPassword == "" {
			if expectedResult && (expectedPassword != storedPassword) {
				t.Errorf("Expected empty password, got %v", storedPassword)
				return false
			}

			if !expectedResult && (expectedPassword == storedPassword) {
				t.Errorf("Expected non-empty password, got empty password",)
				return false
			}

			return true
		}

		t.Logf("stored: %#v, exp: %#v", storedPassword, expectedPassword)

		err = deps.SecretHasher.CompareHashAndSecret(storedPassword, expectedPassword)

		if expectedResult && err != nil {
			t.Errorf("Password hash (%v) does not match expected password (%v): %v", storedPassword, expectedPassword, err)
			return false
		}

		if !expectedResult && err == nil {
			t.Errorf("Password hash (%v) DOES match expected password (%v)", storedPassword, expectedPassword,)
			return false
		}

		return true
	}
}

func CheckSuperAdminPassword(t *testing.T, userId uint, expectedPassword string, expectedResult bool) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &user.SuperAdminEntity{}
		err := db.
			Where("id = ?", userId).
			Find(entity).
			Error
		if err != nil {
			t.Errorf("Expected nil error, got %v", err)
			return false
		}

		var storedPassword string
		if entity.Password != nil {
			storedPassword = *entity.Password
		}

		if expectedPassword == "" || storedPassword == "" {
			if expectedResult && (expectedPassword != storedPassword) {
				t.Errorf("Expected empty password, got %v", storedPassword)
				return false
			}

			if !expectedResult && (expectedPassword == storedPassword) {
				t.Errorf("Expected non-empty password, got empty password",)
				return false
			}

			return true
		}

		t.Logf("stored: %#v, exp: %#v", storedPassword, expectedPassword)

		err = deps.SecretHasher.CompareHashAndSecret(storedPassword, expectedPassword)

		if expectedResult && err != nil {
			t.Errorf("Password hash (%v) does not match expected password (%v): %v", storedPassword, expectedPassword, err)
			return false
		}

		if !expectedResult && err == nil {
			t.Errorf("Password hash (%v) DOES match expected password (%v)", storedPassword, expectedPassword,)
			return false
		}

		return true
	}
}

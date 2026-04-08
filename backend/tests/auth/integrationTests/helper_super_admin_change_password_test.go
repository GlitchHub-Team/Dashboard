package auth_integration_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/auth"
	"backend/tests/helper"

	"gorm.io/gorm"
)

// helper: hash password using same pre-hash + bcrypt approach used in production
func CheckSuperAdminForgotPasswordTokenExists(t *testing.T, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.SuperAdminPasswordTokenEntity{}
		err := db.
			Model(&entity).
			Where("token = ?", hashedToken).
			Find(&entity).
			Error

		return err == nil
	}
}

func CheckSuperAdminForgotPasswordTokenExistsForUser(t *testing.T, tenantId string, userId uint) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.SuperAdminPasswordTokenEntity{}
		err := db.
			Where("user_id = ?", userId).
			Find(entity).
			Error

		return err == nil
	}
}

func CheckNoSuperAdminForgotPasswordToken(t *testing.T, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		var count int64
		db.
			Model(&auth.SuperAdminPasswordTokenEntity{}).
			Where("token = ?", hashedToken).
			Count(&count)

		return count == 0
	}
}

func CheckNoSuperAdminForgotPasswordTokenForUser(t *testing.T, userId uint) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		var count int64
		db.
			Model(&auth.SuperAdminPasswordTokenEntity{}).
			Where("user_id = ?", userId).
			Count(&count)

		return count == 0
	}
}

func CheckSuperAdminForgotPasswordTokenExpired(t *testing.T, hashedToken string) helper.IntegrationTestCheck {
	t.Helper()
	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.SuperAdminPasswordTokenEntity{}
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

func PreSetupAddSuperAdminForgotPasswordToken(t *testing.T, entity auth.SuperAdminPasswordTokenEntity) helper.IntegrationTestPreSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		err := db.
			Create(&entity).
			Error

		return err == nil
	}
}

func PostSetupDeleteSuperAdminForgotPasswordToken(t *testing.T, hashedToken string) helper.IntegrationTestPostSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.SuperAdminPasswordTokenEntity{}
		err := db.
			Where("token = ?", hashedToken).
			Delete(entity).
			Error
		if err != nil {
			t.Logf("error deleting super admin confirm token: %v", err)
		}
	}
}

func PostSetupDeleteSuperAdminForgotPasswordTokensForUser(t *testing.T, tenantId string, userId uint) helper.IntegrationTestPostSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := []auth.SuperAdminPasswordTokenEntity{}
		err := db.
			Where("user_id = ?", userId).
			Delete(entity).
			Error
		if err != nil {
			t.Logf("error deleting tenant confirm token: %v", err)
		}
	}
}

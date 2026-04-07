package auth_integration_test

import (
	"crypto/sha512"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"backend/internal/auth"
	"backend/internal/shared/identity"
	"backend/tests/helper"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// helper: hash password using same pre-hash + bcrypt approach used in production
func hashPasswordForTest(plaintext string) (string, error) {
	pre := sha512.Sum512([]byte(plaintext))
	h, err := bcrypt.GenerateFromPassword(pre[:], bcrypt.DefaultCost)
	return string(h), err
}

func CheckValidJWTInResponse(t *testing.T, expectedRequester identity.Requester) helper.IntegrationTestCheck {
	t.Helper()

	return func(r *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		// Unmarshal response body
		var resp map[string]any
		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			t.Logf("invalid json: %v", err)
			return false
		}

		// Ottieni JWT dal corpo
		jwtStr, ok := resp["jwt"].(string)
		if !ok || jwtStr == "" {
			t.Logf("missing jwt in response")
			return false
		}

		// Ottieni requester dal JWT
		requester, err := deps.AuthTokenManager.GetRequesterFromToken(jwtStr)
		if err != nil {
			t.Logf("invalid token: %v", err)
			return false
		}

		// Paragona con requester atteso
		if !reflect.DeepEqual(requester, expectedRequester) {
			t.Logf("Expected requester %#v, got %#v", expectedRequester, requester)
			return false
		}

		return true
	}
}

func PreSetupAddSuperAdminConfirmAccountToken(t *testing.T, entity auth.SuperAdminConfirmTokenEntity) helper.IntegrationTestPreSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		err := db.
			Create(&entity).
			Error

		return err == nil
	}
}

func PostSetupDeleteSuperAdminConfirmAccountToken(t *testing.T, hashedToken string) helper.IntegrationTestPostSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.SuperAdminConfirmTokenEntity{}
		err := db.
			Where("token = ?", hashedToken).
			Delete(entity).
			Error
		if err != nil {
			t.Logf("error deleting super admin confirm token: %v", err)
		}
	}
}

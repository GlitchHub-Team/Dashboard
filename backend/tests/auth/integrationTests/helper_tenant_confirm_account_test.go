package auth_integration_test

import (
	"testing"

	"backend/internal/auth"
	cloud_db "backend/internal/infra/database/cloud_db/connection"
	"backend/tests/helper"

	"gorm.io/gorm"
)

func PreSetupAddTenantConfirmAccountToken(t *testing.T, entity auth.TenantConfirmTokenEntity) helper.IntegrationTestPreSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) bool {
		if entity.TenantId == "" {
			t.Logf("Impossibile creare token conferma per tenant nullo")
			return false
		}
		db := (*gorm.DB)(deps.CloudDB)

		err := db.
			Scopes(cloud_db.WithTenantSchema(entity.TenantId, &auth.TenantConfirmTokenEntity{})).
			Create(&entity).
			Error

		return err == nil
	}
}

func PostSetupDeleteTenantConfirmAccountToken(t *testing.T, tenantId, hashedToken string) helper.IntegrationTestPostSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := &auth.TenantConfirmTokenEntity{}
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

package tenant_integration_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/auth"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func PreSetupCreateTenantWithName(tenantID uuid.UUID, tenantName string, canImpersonate bool) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		tenantEntity := tenant.TenantEntity{
			ID:             tenantID.String(),
			Name:           tenantName,
			CanImpersonate: canImpersonate,
		}
		if err := db.Create(&tenantEntity).Error; err != nil {
			return false
		}

		schemaName := "tenant_" + tenantID.String()
		if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", schemaName)).Error; err != nil {
			return false
		}

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

func PostSetupDeleteTenantByName(t *testing.T, tenantName string) helper.IntegrationTestPostSetup {
	t.Helper()

	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := tenant.TenantEntity{}
		err := db.Where("name = ?", tenantName).First(&entity).Error
		if err != nil {
			return
		}

		schemaName := "tenant_" + entity.ID
		if err := db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" CASCADE", schemaName)).Error; err != nil {
			t.Logf("error dropping schema %s: %v", schemaName, err)
		}

		if err := db.Where("id = ?", entity.ID).Delete(&tenant.TenantEntity{}).Error; err != nil {
			t.Logf("error deleting tenant %s: %v", entity.ID, err)
		}
	}
}

func CheckTenantInsertedByName(t *testing.T, tenantName string, canImpersonate bool) helper.IntegrationTestCheck {
	t.Helper()

	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		entity := tenant.TenantEntity{}
		err := db.Where("name = ?", tenantName).First(&entity).Error
		if err != nil {
			t.Errorf("cannot find tenant by name %s: %v", tenantName, err)
			return false
		}

		if entity.CanImpersonate != canImpersonate {
			t.Errorf("expected can_impersonate=%v for tenant %s, got %v", canImpersonate, tenantName, entity.CanImpersonate)
			return false
		}

		return true
	}
}

func CheckNoTenantByName(t *testing.T, tenantName string) helper.IntegrationTestCheck {
	t.Helper()

	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&tenant.TenantEntity{}).Where("name = ?", tenantName).Count(&count).Error
		if err != nil {
			t.Errorf("cannot count tenants by name %s: %v", tenantName, err)
			return false
		}
		return count == 0
	}
}

func CheckTenantExistsByID(t *testing.T, tenantID uuid.UUID) helper.IntegrationTestCheck {
	t.Helper()

	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&tenant.TenantEntity{}).Where("id = ?", tenantID.String()).Count(&count).Error
		if err != nil {
			t.Errorf("cannot count tenant by id %s: %v", tenantID.String(), err)
			return false
		}
		return count == 1
	}
}

func CheckNoTenantByID(t *testing.T, tenantID uuid.UUID) helper.IntegrationTestCheck {
	t.Helper()

	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&tenant.TenantEntity{}).Where("id = ?", tenantID.String()).Count(&count).Error
		if err != nil {
			t.Errorf("cannot count tenant by id %s: %v", tenantID.String(), err)
			return false
		}
		return count == 0
	}
}

func CheckResponseBodyContains(t *testing.T, values ...string) helper.IntegrationTestCheck {
	t.Helper()

	return func(respRecorder *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		body := respRecorder.Body.String()
		for _, value := range values {
			if !strings.Contains(body, value) {
				t.Errorf("expected response to contain %q, got %s", value, body)
				return false
			}
		}
		return true
	}
}

func tenantPath(tenantID uuid.UUID) string {
	return "/api/v1/tenant/" + tenantID.String()
}

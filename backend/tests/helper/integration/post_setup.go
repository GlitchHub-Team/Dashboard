package integration

import (
	"fmt"
	"testing"

	clouddb "backend/internal/infra/database/cloud_db/connection"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func PostSetupDeleteTenant(t *testing.T, tenantId uuid.UUID) helper.IntegrationTestPostSetup {
	t.Helper()
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

func PostSetupDeleteTenantMember(tenantId uuid.UUID, email string) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := user.TenantMemberEntity{}
		_ = db.
			Scopes(clouddb.WithTenantSchema(tenantId.String(), &user.TenantMemberEntity{})).
			Where("email = ?", email).
			Delete(&entity).
			Error
	}
}

func PostSetupDeleteSuperAdmin(email string) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)

		entity := user.SuperAdminEntity{}
		_ = db.Where("email = ?", email).Delete(&entity).Error
	}
}

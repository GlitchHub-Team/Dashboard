package cloud_db

import (
	"fmt"

	conn "backend/internal/infra/database/cloud_db/connection"
	"backend/internal/shared/crypto"
	"backend/internal/user"

	"backend/internal/tenant"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NOTA: SERVIREBBERO 4 e 8 servono per avere UUID valido (es. 11111111-1111-4111-8111-111111111111)
const (
	tenant1Id = "11111111-1111-1111-1111-111111111111"
	tenant2Id = "22222222-2222-2222-2222-222222222222"
)

/*
NOTA: Quest'interfaccia dev'essere uguale a [migrate.CloudDBMigrator]
*/
type localCloudMigrator interface {
	DB() *gorm.DB
	Logger() *zap.Logger
	Hasher() crypto.SecretHasher

	MigratePublic() error
	MigrateTenantSchema(tenantId string, shouldLog bool) error
}

func populatePublicDefaultData(migrator localCloudMigrator) error {
	// Password per utenti test
	encrPass, err := migrator.Hasher().HashSecret("12345678")
	if err != nil {
		return fmt.Errorf("failed to hash secret: %v", err)
	}

	// Super admin
	superAdmin := user.SuperAdminEntity{
		ID:        1,
		Email:     "super@admin.com",
		Name:      "Super Admin",
		Password:  &encrPass,
		Confirmed: true,
	}

	if err := migrator.DB().Clauses(clause.OnConflict{DoNothing: true}).Create(&superAdmin).Error; err != nil {
		return fmt.Errorf("failed to create super admin: %v", err)
	}

	tenantMembers := []user.TenantMemberEntity{
		{
			ID:        1,
			Email:     "tenant1@admin.com",
			Name:      "Tenant 1 Admin",
			Password:  &encrPass,
			Confirmed: true,
			Role:      "tenant_admin",
			TenantId:  tenant1Id,
		},
		{
			ID:        2,
			Email:     "tenant1@user.com",
			Name:      "Tenant 1 User",
			Password:  &encrPass,
			Confirmed: true,
			Role:      "tenant_user",
			TenantId:  tenant1Id,
		},
		{
			ID:        1,
			Email:     "tenant2@admin.com",
			Name:      "Tenant 2 Admin",
			Password:  &encrPass,
			Confirmed: true,
			Role:      "tenant_admin",
			TenantId:  tenant2Id,
		},
		{
			ID:        2,
			Email:     "tenant2@user.com",
			Name:      "Tenant 2 User",
			Password:  &encrPass,
			Confirmed: true,
			Role:      "tenant_user",
			TenantId:  tenant2Id,
		},
	}

	for _, tenantMember := range tenantMembers {
		if err := migrator.DB().
			Scopes(conn.WithTenantSchema(tenantMember.TenantId, &user.TenantMemberEntity{})).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&tenantMember).
			Error; err != nil {
			return fmt.Errorf("failed to create tenant member: %v", err)
		}
	}

	return nil
}

func populateTenantDefaultData(migrator localCloudMigrator) error {
	tenants := []tenant.TenantEntity{
		{ID: tenant1Id, Name: "Tenant 1", CanImpersonate: true},
		{ID: tenant2Id, Name: "Tenant 2", CanImpersonate: false},
	}

	// Tenant 1 e 2
	for _, tenant := range tenants {
		if err := migrator.DB().Clauses(clause.OnConflict{DoNothing: true}).Create(&tenant).Error; err != nil {
			return fmt.Errorf("failed to create tenant %v: %v", tenant.ID, err)
		}
	}
	return nil
}

func migrateAll(
	tenantRepo tenant.TenantRepository,
	migrator localCloudMigrator,
	setDefaultData bool,
) (err error) {
	// 1. Migrazione pubblica
	err = migrator.MigratePublic()
	if err != nil {
		return
	}

	// 2. Impostazione dati pubblici (opzionale)
	if setDefaultData {
		migrator.Logger().Sugar().Infof("[Migrator] Add default data")
		err = populatePublicDefaultData(migrator)
		if err != nil {
			return
		}
	}

	// 3. Migrazione per tutti i tenant
	tenants, err := tenantRepo.GetAllTenants()
	if err != nil {
		return fmt.Errorf("cannot get tenant list: %w", err)
	}

	for _, tenant := range tenants {
		err = migrator.MigrateTenantSchema(tenant.ID, true)
		if err != nil {
			return err
		}
	}

	// 4. Impostazione dati tenant (opzionale)
	if setDefaultData {
		migrator.Logger().Sugar().Infof("[Migrator] add tenant data")
		err = populateTenantDefaultData(migrator)
		if err != nil {
			return err
		}
	}

	return
}

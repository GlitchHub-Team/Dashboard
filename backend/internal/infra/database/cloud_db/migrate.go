package cloud_db

import (
	"fmt"

	"backend/internal/auth"
	"backend/internal/gateway"
	"backend/internal/sensor"
	"backend/internal/tenant"
	"backend/internal/user"

	clouddb "backend/internal/infra/database/cloud_db/connection"
	hasher "backend/internal/shared/crypto"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NOTA: SERVIREBBERO 4 e 8 servono per avere UUID valido (es. 11111111-1111-4111-8111-111111111111)
const (
	tenant1Id = "11111111-1111-1111-1111-111111111111"
	tenant2Id = "22222222-2222-2222-2222-222222222222"
)

type Migrator interface {
	Migrate() error
}

type PostgreMigrator struct {
	log *zap.Logger
	db  clouddb.CloudDBConnection
	// getTenantsPort tenant.GetTenantsPort
	getTenantsRepo *tenant.TenantPostgreRepository
	hasher         hasher.SecretHasher
}

func NewPostgreMigrator(
	log *zap.Logger,
	db clouddb.CloudDBConnection,
	getTenantsRepo *tenant.TenantPostgreRepository,
	hasher hasher.SecretHasher,
) Migrator {
	return &PostgreMigrator{
		log:            log,
		db:             db,
		getTenantsRepo: getTenantsRepo,
		hasher:         hasher,
	}
}

func (migrator *PostgreMigrator) Migrate() error {
	/* Entity che sono associate allo schema public */
	publicEntities := []any{
		&tenant.TenantEntity{},
		&gateway.GatewayEntity{},
		&sensor.SensorEntity{},
		&user.SuperAdminEntity{},
		&auth.SuperAdminConfirmTokenEntity{},
		&auth.SuperAdminPasswordTokenEntity{},
	}

	/* Entity da associare a uno schema tenant specifico */
	tenantSchemaEntities := [](interface{ TableName() string }){
		&user.TenantMemberEntity{},
	}

	migrator.log.Info("[Migrator] started on cloud DB")

	// Migrate entities for public schema
	db := (*gorm.DB)(migrator.db)
	err := db.AutoMigrate(publicEntities...)
	if err != nil {
		return err
	}

	err = populateTenantTestData(db)
	if err != nil {
		return fmt.Errorf("failed to populate tenant test data: %v", err)
	}

	// Get tenants
	tenants, err := migrator.getTenantsRepo.GetAllTenants()
	if err != nil {
		return err
	}

	// Create schemas
	for _, tenant := range tenants {
		schemaName := fmt.Sprintf("tenant_%v", tenant.ID)
		if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%v\"", schemaName)).Error; err != nil {
			return fmt.Errorf("error creating schema %v: %v", schemaName, err)
		}
		migrator.log.Sugar().Infof("[Migrator] Migrated schema %v", schemaName)
	}

	// Migrate entities for each schema
	for _, tenant := range tenants {
		tenantId := tenant.ID
		if tenantId == "" {
			continue
		}

		schemaName := fmt.Sprintf("tenant_%v", tenantId)

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(fmt.Sprintf("set local search_path to \"%s\"", schemaName)).Error; err != nil {
				return fmt.Errorf("failed to set search_path to %s: %v", schemaName, err)
			}

			for _, entity := range tenantSchemaEntities {
				migrator.log.Sugar().Infof("Migrating %v", entity.TableName())

				if err = tx.AutoMigrate(entity); err != nil {
					return fmt.Errorf("error migrating table %v: %v", entity.TableName(), err)
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("error migrating tenant %v: %v", tenantId, err)
		}
	}

	err = populateWithTestData(db, migrator.hasher)
	if err != nil {
		return fmt.Errorf("failed to populate test data: %v", err)
	}

	return nil
}

func populateTenantTestData(db *gorm.DB) error {
	tenants := []tenant.TenantEntity{
		{ID: tenant1Id, Name: "Tenant 1", CanImpersonate: true},
		{ID: tenant2Id, Name: "Tenant 2", CanImpersonate: false},
	}

	// Tenant 1 e 2
	for _, tenant := range tenants {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tenant).Error; err != nil {
			return fmt.Errorf("failed to create tenant %v: %v", tenant.ID, err)
		}
	}
	return nil
}

func populateWithTestData(db *gorm.DB, hasher hasher.SecretHasher) error {
	// Password per utenti test
	encrPass, err := hasher.HashSecret("12345678")
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

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&superAdmin).Error; err != nil {
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
		if err := db.Scopes(clouddb.WithTenantSchema(tenantMember.TenantId, &user.TenantMemberEntity{})).Clauses(clause.OnConflict{DoNothing: true}).Create(&tenantMember).Error; err != nil {
			return fmt.Errorf("failed to create tenant member: %v", err)
		}
	}

	return nil
}

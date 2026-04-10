package tenant

import (
	clouddb "backend/internal/infra/database/cloud_db/connection"
	// "backend/internal/infra/database/cloud_db/migrate"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TenantEntity struct {
	ID             string `gorm:"size:36"`
	Name           string `gorm:"size:256"`
	CanImpersonate bool
}

/*
NOTA: Interfaccia locale che deve rispecchiare "backend/internal/infra/database/cloud_db/migrate".Migrator

Da non confondere con [TenantMigrator], che invece è il tipo utilizzato internamente al tenant package
*/
type LocalCloudMigrator interface {
	MigrateTenantSchema(tenantId string, shouldLog bool) error
}

type TenantPostgreRepository struct {
	log      *zap.Logger
	db       clouddb.CloudDBConnection
	migrator LocalCloudMigrator
}

func NewTenantPostgreRepository(
	log *zap.Logger,
	db clouddb.CloudDBConnection,
	migrator LocalCloudMigrator,
) *TenantPostgreRepository {
	return &TenantPostgreRepository{
		log:      log,
		db:       db,
		migrator: migrator,
	}
}

func (repo *TenantPostgreRepository) GetTenant(tenantId string) (*TenantEntity, error) {
	var entity *TenantEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", tenantId).
		Find(&entity).
		Error
	return entity, err
}

func (repo *TenantPostgreRepository) GetTenants(offset, limit int) ([]TenantEntity, int64, error) {
	db := (*gorm.DB)(repo.db)

	var entities []TenantEntity
	err := db.
		Offset(offset).
		Limit(limit).
		Find(&entities).
		Error
	if err != nil {
		return nil, 0, err
	}

	var total int64
	if err := db.Model(&TenantEntity{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, err
}

func (repo *TenantPostgreRepository) GetAllTenants() ([]TenantEntity, error) {
	db := (*gorm.DB)(repo.db)

	var tenants []TenantEntity
	if err := db.Find(&tenants).Error; err != nil {
		return nil, err
	}

	return tenants, nil
}

func (repo *TenantPostgreRepository) SaveTenant(tenant *TenantEntity) error {
	db := (*gorm.DB)(repo.db)

	err := repo.migrator.MigrateTenantSchema(tenant.ID, false)
	if err != nil {
		return err
	}

	return db.Save(tenant).Error
}

func (repo *TenantPostgreRepository) DeleteTenant(tenant *TenantEntity) error {
	db := (*gorm.DB)(repo.db)
	err := db.
		Clauses(clause.Returning{}).
		Delete(tenant).
		Error

	return err
}

func (repo *TenantPostgreRepository) GetTenantById(tenantId string) (*TenantEntity, error) {
	var tenant TenantEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", tenantId).
		Find(&tenant).
		Error
	return &tenant, err
}

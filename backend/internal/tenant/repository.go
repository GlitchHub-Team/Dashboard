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
	Name           string `gorm:"index:,unique;size:256"`
	CanImpersonate bool
}

func (TenantEntity) TableName() string { return "tenants" }

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
		First(&entity, "id = ?", tenantId).
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

func (repo *TenantPostgreRepository) SaveTenant(entity *TenantEntity) error {
	db := (*gorm.DB)(repo.db)

	err := repo.migrator.MigrateTenantSchema(entity.ID, false)
	if err != nil {
		return err
	}

	return db.Save(entity).Error
}

/*
Elimina tenant, rappresentato da entity. E' importante che entity.ID != ""
*/
func (repo *TenantPostgreRepository) DeleteTenant(entity *TenantEntity) (err error) {
	db := (*gorm.DB)(repo.db)
	result := db.
		Clauses(clause.Returning{}).
		Delete(entity, "id = ?", entity.ID)
	err = result.Error
	if err != nil {
		return
	}

	if result.RowsAffected == 0 {
		err = ErrTenantNotFound
		return
	}

	return
}

func (repo *TenantPostgreRepository) GetTenantByUser(userId string) (*TenantEntity, error) {
	var tenant TenantEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", userId).
		Find(&tenant).
		Error
	return &tenant, err
}

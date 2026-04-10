package tenant

import (
	clouddb "backend/internal/infra/database/cloud_db/connection"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TenantEntity struct {
	ID             string `gorm:"size:36"`
	Name           string `gorm:"size:256"`
	CanImpersonate bool
}

type TenantPostgreRepository struct {
	log *zap.Logger
	db  clouddb.CloudDBConnection
}

func NewTenantPostgreRepository(
	log *zap.Logger,
	db clouddb.CloudDBConnection,
) *TenantPostgreRepository {
	return &TenantPostgreRepository{
		log: log,
		db:  db,
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

func (repo *TenantPostgreRepository) GetAllTenants() ([]TenantEntity, error) {
	var users []TenantEntity
	db := (*gorm.DB)(repo.db)
	if err := db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *TenantPostgreRepository) SaveTenant(tenant *TenantEntity) error {
	db := (*gorm.DB)(repo.db)
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

func (repo *TenantPostgreRepository) GetTenantByUser(userId string) (*TenantEntity, error) {
	var tenant TenantEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", userId).
		Find(&tenant).
		Error
	return &tenant, err
}

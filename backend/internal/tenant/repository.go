package tenant

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TenantEntity struct {
	ID             string `gorm:"size:256"`
	Name           string `gorm:"size:256"`
	CanImpersonate bool
}

type TenantPostgreRepository struct {
	log *zap.Logger
	db  *gorm.DB
}

func NewTenantPostgreRepository(
	log *zap.Logger,
	db *gorm.DB,
) *TenantPostgreRepository {
	return &TenantPostgreRepository{
		log: log,
		db:  db,
	}
}

func (repo *TenantPostgreRepository) GetTenant(tenantId string) (*TenantEntity, error) {
	var entity *TenantEntity
	err := repo.db.
		Where("id = ?", tenantId).
		Find(&entity).
		Error
	return entity, err
}

func (repo *TenantPostgreRepository) GetAllTenants() ([]TenantEntity, error) {
	var users []TenantEntity
	if err := repo.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *TenantPostgreRepository) SaveTenant(tenant *TenantEntity) error {
	return repo.db.Save(tenant).Error
}

func (repo *TenantPostgreRepository) DeleteTenant(tenant *TenantEntity) (Tenant, error) {
	oldTenant, err := tenant.toTenant()
	if err != nil {
		return Tenant{}, err
	}

	err = repo.db.Delete(tenant).Error
	if err != nil {
		return Tenant{}, err
	}

	return oldTenant, nil
}

func (repo *TenantPostgreRepository) GetTenantByUser(userId string) (*TenantEntity, error) {
	var tenant TenantEntity
	err := repo.db.
		Where("id = ?", userId).
		Find(&tenant).
		Error
	return &tenant, err
}

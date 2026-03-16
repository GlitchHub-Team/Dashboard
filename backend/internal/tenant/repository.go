package tenant

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TenantEntity struct {
	ID             string `gorm:"size:256"`
	Name           string `gorm:"size:256"`
	CanImpersonate bool
}

func (TenantEntity) TableName() string { return "public.tenants" }

func (entity *TenantEntity) toTenant() (Tenant, error) {
	tenantId, err := uuid.Parse(entity.ID)

	return Tenant{
		Id: tenantId,
		Name: entity.Name,
		CanImpersonate: entity.CanImpersonate,
	}, err

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

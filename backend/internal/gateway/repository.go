package gateway

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GatewayRepository interface {
	GetById(id uuid.UUID) (Gateway, error)
	Save(g Gateway) error
}

type gatewayEntity struct {
	ID               string `gorm:"primaryKey;size:36"`
	Name             string `gorm:"size:128;not null"`
	TenantId         string `gorm:"size:36;index"`
	Status           string `gorm:"not null;size:32"`
	IntervalLimit    int64  `gorm:"not null"`
	PublicIdentifier string `gorm:"size:128;not null"`
}

func (gatewayEntity) TableName() string { return "gateways" }

func (e *gatewayEntity) toDomain() Gateway {
	id, _ := uuid.Parse(e.ID)
	var tenantId *uuid.UUID
	if e.TenantId != "" {
		tid, _ := uuid.Parse(e.TenantId)
		tenantId = &tid
	}
	return Gateway{
		Id:               id,
		Name:             e.Name,
		PublicIdentifier: e.PublicIdentifier,
		TenantId:         tenantId,
		Status:           GatewayStatus(e.Status),
		IntervalLimit:    e.IntervalLimit,
	}
}

type gatewayPostgreRepository struct {
	db *gorm.DB
}

func NewGatewayPostgreRepository(db *gorm.DB) GatewayRepository {
	return &gatewayPostgreRepository{db: db}
}

func MigrateGateway(db *gorm.DB) error {
	return db.AutoMigrate(&gatewayEntity{})
}

func (repo *gatewayPostgreRepository) GetById(id uuid.UUID) (Gateway, error) {
	var entity gatewayEntity
	err := repo.db.Where("id = ?", id.String()).First(&entity).Error
	if err != nil {
		return Gateway{}, err
	}
	return entity.toDomain(), nil
}

func (repo *gatewayPostgreRepository) Save(g Gateway) error {
	var existing gatewayEntity

	err := repo.db.Where("id = ?", g.Id.String()).First(&existing).Error

	if err != nil {
		return err
	}

	tenantVal := ""
	if g.TenantId != nil {
		tenantVal = g.TenantId.String()
	}

	return repo.db.Model(&gatewayEntity{}).
		Where("id = ?", g.Id.String()).
		Updates(map[string]interface{}{
			"name":              g.Name,
			"public_identifier": g.PublicIdentifier,
			"status":            string(g.Status),
			"interval_limit":    g.IntervalLimit,
			"tenant_id":         tenantVal,
		}).Error
}

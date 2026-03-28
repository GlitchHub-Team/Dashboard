package gateway

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gatewayEntity struct {
	ID               uint   `gorm:"primaryKey;autoIncrement"`
	Name             string `gorm:"size:128;not null"`
	TenantId         string `gorm:"size:36"`
	Status           string `gorm:"not null;size:32"`
	IntervalLimit    int64  `gorm:"not null"`
	PublicIdentifier string `gorm:"unique;size:36;not null"`
}

func (gatewayEntity) TableName() string { return "gateways" }

func (e *gatewayEntity) fromDomain(g Gateway) {
	e.Name = g.Name
	if g.TenantId != nil {
		e.TenantId = g.TenantId.String()
	}
	e.Status = string(g.Status)
	e.IntervalLimit = g.IntervalLimit
	e.PublicIdentifier = g.PublicIdentifier
}

func (e *gatewayEntity) toDomain() Gateway {
	var tenantId *uuid.UUID
	if e.TenantId != "" {
		parsed, _ := uuid.Parse(e.TenantId)
		tenantId = &parsed
	}
	return Gateway{
		Id:               uuid.Nil,
		Name:             e.Name,
		TenantId:         tenantId,
		Status:           GatewayStatus(e.Status),
		IntervalLimit:    e.IntervalLimit,
		PublicIdentifier: e.PublicIdentifier,
	}
}

type gatewayPostgreRepository struct {
	db *gorm.DB
}

func NewGatewayPostgreRepository(db *gorm.DB) gatewayPostgreRepository {
	return gatewayPostgreRepository{db: db}
}

func MigrateGateway(db *gorm.DB) error {
	return db.AutoMigrate(&gatewayEntity{})
}

func (repo *gatewayPostgreRepository) Save(g Gateway) error {
	var existing gatewayEntity
	err := repo.db.Where("public_identifier = ?", g.PublicIdentifier).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}
	var entity gatewayEntity
	entity.fromDomain(g)
	return repo.db.Create(&entity).Error
}

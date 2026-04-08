package gateway

import (
	"errors"
	"time"

	"backend/internal/tenant"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	clouddb "backend/internal/infra/database/cloud_db/connection"
)

type DB any // TODO: solo per test

// per il commissionig // risoista  requst replay,

// type gatewayEntity struct{}

// entity =============================================================================================

type GatewayEntity struct {
	ID       string  `gorm:"type:uuid;primaryKey"`
	Name     string  `gorm:"type:varchar(255);not null"`
	TenantId *string `gorm:"type:uuid;index"`
	// il modo giusto per fare il fk per assurdo
	Tenant           *tenant.Tenant `gorm:"foreignKey:TenantId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status           string         `gorm:"type:varchar(50);not null"`
	PublicIdentifier string         `gorm:"type:varchar(255)"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (GatewayEntity) TableName() string { return "gateways" }

type gatewayPostgreRepository struct {
	log *zap.Logger
	db  clouddb.CloudDBConnection
}

func NewGatewayPostgreRepository(log *zap.Logger, db clouddb.CloudDBConnection) *gatewayPostgreRepository {
	return &gatewayPostgreRepository{
		log: log,
		db:  db,
	}
}

// methods ============================================================================================

func GatewayEntityToDomain(entity *GatewayEntity) (Gateway, error) {
	if entity == nil {
		return Gateway{}, nil
	}

	gatewayId, err := uuid.Parse(entity.ID)
	if err != nil {
		return Gateway{}, err
	}

	var tenantId *uuid.UUID

	if entity.TenantId != nil {
		parsed, err := uuid.Parse(*entity.TenantId)
		tenantId = &parsed
		if err != nil {
			return Gateway{}, err
		}
	}

	return Gateway{
		Id: gatewayId,
		Name: entity.Name,
		TenantId: tenantId,
		Status: GatewayStatus(entity.Status),
		// IntervalLimit: entity.,
	}, nil
}

func (entity *GatewayEntity) FromGateway(g Gateway) {
	entity.ID = g.Id.String()
	entity.Name = g.Name
	entity.Status = string(g.Status)
	entity.PublicIdentifier = g.PublicIdentifier

	if g.TenantId != nil {
		tenantIdStr := g.TenantId.String()
		entity.TenantId = &tenantIdStr
	} else {
		entity.TenantId = nil
	}
}

func (entity *GatewayEntity) ToGateway() Gateway {
	id, _ := uuid.Parse(entity.ID)
	var tenantId *uuid.UUID
	if entity.TenantId != nil {
		parsed, _ := uuid.Parse(*entity.TenantId)
		tenantId = &parsed
	}
	return Gateway{
		Id:               id,
		Name:             entity.Name,
		Status:           (GatewayStatus)(entity.Status),
		TenantId:         tenantId,
		PublicIdentifier: entity.PublicIdentifier,
	}
}

func (repo *gatewayPostgreRepository) SaveGateway(gateway Gateway) error {
	entity := &GatewayEntity{}
	entity.FromGateway(gateway)

	existing := &GatewayEntity{}
	db := (*gorm.DB)(repo.db)
	err := db.Where("id = ? AND tenant_id IS NOT NULL", entity.ID).First(existing).Error

	if err == nil {
		return ErrGatewayAlreadyAssigned
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return db.Save(entity).Error
}

func (repo *gatewayPostgreRepository) DeleteGateway(gateway Gateway) error {
	entity := &GatewayEntity{}

	db := (*gorm.DB)(repo.db)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", gateway.Id).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(entity).Error; err != nil {
			return err
		}

		return tx.Delete(entity).Error
	})
}

func (repo *gatewayPostgreRepository) GetGatewayById(gatewayId string) (GatewayEntity, error) {
	var entity GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", gatewayId).
		First(&entity).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return GatewayEntity{}, nil
	}
	return entity, err
}

func (repo *gatewayPostgreRepository) GetGatewaysByTenantId(tenantId string) ([]GatewayEntity, error) {
	var entities []GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.Where("tenant_id = ?", tenantId).Find(&entities).Error
	return entities, err
}

func (repo *gatewayPostgreRepository) GetAllGateways() ([]GatewayEntity, error) {
	var entities []GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.Find(&entities).Error
	return entities, err
}

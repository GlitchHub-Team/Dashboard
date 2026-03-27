package gateway

import (
	"errors"
	"time"

	"backend/internal/tenant"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB any // TODO: solo per test

// per il commissionig // risoista  requst replay,

type gatewayEntity struct{}

// entity =============================================================================================

type GatewayEntity struct {
	GatewayId string  `gorm:"type:uuid;primaryKey"`
	Name      string  `gorm:"type:varchar(255);not null"`
	TenantId  *string `gorm:"type:uuid;index"`
	// il modo giusto per fare il fk per assurdo
	Tenant    *tenant.Tenant `gorm:"foreignKey:TenantId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status    string         `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (GatewayEntity) TableName() string { return "gateways" }

type gatewayPostgreRepository struct {
	log *zap.Logger
	db  *gorm.DB
}

func NewGatewayPostgreRepository(log *zap.Logger, db *gorm.DB) *gatewayPostgreRepository {
	return &gatewayPostgreRepository{
		log: log,
		db:  db,
	}
}

// methods ============================================================================================

func (entity *GatewayEntity) fromGateway(g Gateway) {
	entity.GatewayId = g.Id.String()
	entity.Name = g.Name
	entity.Status = string(g.Status)

	if g.TenantId != nil {
		tenantIdStr := g.TenantId.String()
		entity.TenantId = &tenantIdStr
	} else {
		entity.TenantId = nil
	}
}

func (entity *GatewayEntity) toGateway() Gateway {
	id, _ := uuid.Parse(entity.GatewayId)
	var tenantId *uuid.UUID
	if entity.TenantId != nil {
		parsed, _ := uuid.Parse(*entity.TenantId)
		tenantId = &parsed
	}
	return Gateway{
		Id:       id,
		Name:     entity.Name,
		Status:   (GatewayStatus)(entity.Status),
		TenantId: tenantId,
	}
}

func (repo *gatewayPostgreRepository) SaveGateway(gateway Gateway) error {
	entity := &GatewayEntity{}
	entity.fromGateway(gateway)

	existing := &GatewayEntity{}
	err := repo.db.Where("gateway_id = ? AND tenant_id IS NOT NULL", entity.GatewayId).First(existing).Error

	if err == nil {
		return ErrGatewayAlreadyAssigned
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return repo.db.Save(entity).Error
}

func (repo *gatewayPostgreRepository) DeleteGateway(gateway Gateway) error {
	entity := &GatewayEntity{}

	return repo.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("gateway_id = ?", gateway.Id).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(entity).Error; err != nil {
			return err
		}

		return tx.Delete(entity).Error
	})
}

func (repo *gatewayPostgreRepository) GetGatewayById(gatewayId string) (GatewayEntity, error) {
    var entity GatewayEntity
    err := repo.db.
        Where("gateway_id = ?", gatewayId).
        First(&entity). 
        Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return GatewayEntity{}, nil 
    }
    return entity, err
}

func (repo *gatewayPostgreRepository) GetGatewaysByTenantId(tenantId string) ([]GatewayEntity, error) {
    var entities []GatewayEntity
    err := repo.db.Where("tenant_id = ?", tenantId).Find(&entities).Error
    return entities, err
}

func (repo *gatewayPostgreRepository) GetAllGateways() ([]GatewayEntity, error) {
    var entities []GatewayEntity
    err := repo.db.Find(&entities).Error
    return entities, err
}

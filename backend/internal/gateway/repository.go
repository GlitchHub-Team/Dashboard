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

// per il commissionig // risoista  requst replay,

// type gatewayEntity struct{}

// entity =============================================================================================

type GatewayEntity struct {
	ID               string         `gorm:"type:uuid;primaryKey"`
	Name             string         `gorm:"type:varchar(255);not null"`
	TenantId         *string        `gorm:"type:uuid;index"`
	Interval         int64          `gorm:"not null"`
	Tenant           *tenant.Tenant `gorm:"foreignKey:TenantId;references:Id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status           string         `gorm:"type:varchar(50);not null"`
	PublicIdentifier *string        `gorm:"type:varchar(255)"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (GatewayEntity) TableName() string { return "gateways" }

type gatewayPostgreRepository struct {
	log *zap.Logger
	db  clouddb.CloudDBConnection
}

func NewGatewayPostgreRepository(log *zap.Logger, db clouddb.CloudDBConnection) GatewayRepository {
	return &gatewayPostgreRepository{
		log: log,
		db:  db,
	}
}

type GatewayRepository interface {
	SaveGateway(entity *GatewayEntity) error
	DeleteGateway(entity *GatewayEntity) error
	GetGatewayById(gatewayId string) (Gateway, error)
	GetGatewaysByTenantId(tenantId string) ([]Gateway, error)
	GetAllGateways() ([]Gateway, error)
	CreateGateway(entity *GatewayEntity) (Gateway, error)
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
		Id:               gatewayId,
		Name:             entity.Name,
		TenantId:         tenantId,
		IntervalLimit:    time.Duration(entity.Interval) * time.Millisecond,
		Status:           GatewayStatus(entity.Status),
		PublicIdentifier: entity.PublicIdentifier,
	}, nil
}

func (entity *GatewayEntity) FromGateway(g Gateway) {
	entity.ID = g.Id.String()
	entity.Name = g.Name
	entity.Status = string(g.Status)
	entity.Interval = g.IntervalLimit.Milliseconds()

	if g.TenantId != nil {
		tenantIdStr := g.TenantId.String()
		entity.TenantId = &tenantIdStr
	} else {
		entity.TenantId = nil
	}
	entity.PublicIdentifier = g.PublicIdentifier
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
		IntervalLimit:    time.Duration(entity.Interval) * time.Millisecond,
		PublicIdentifier: entity.PublicIdentifier,
	}
}

func (repo *gatewayPostgreRepository) CreateGateway(entity *GatewayEntity) (Gateway, error) {
	if err := (*gorm.DB)(repo.db).Clauses(clause.Returning{}).Create(entity).Error; err != nil {
		return Gateway{}, err
	}
	return entity.ToGateway(), nil
}

func (repo *gatewayPostgreRepository) SaveGateway(entity *GatewayEntity) error {
	db := (*gorm.DB)(repo.db)
	return db.Clauses(clause.Returning{}).Save(entity).Error
}

func (repo *gatewayPostgreRepository) DeleteGateway(entity *GatewayEntity) error {
	db := (*gorm.DB)(repo.db)
	err := db.
		Clauses(clause.Returning{}).
		Delete(entity).
		Error
	if err != nil {
		repo.log.Error("Failed to delete gateway", zap.Error(err))
		return err
	}
	return nil
}

// TODO: hexagonal sbagliato, repo non può ritornare classi di dominio
func (repo *gatewayPostgreRepository) GetGatewayById(gatewayId string) (Gateway, error) {
	var entity GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", gatewayId).
		First(&entity).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Gateway{}, ErrGatewayNotFound
	}
	if err != nil {
		return Gateway{}, err
	}

	gateway, err := GatewayEntityToDomain(&entity)
	return gateway, err
}

// TODO: hexagonal sbagliato, repo non può ritornare classi di dominio
func (repo *gatewayPostgreRepository) GetGatewaysByTenantId(tenantId string) ([]Gateway, error) {
	var entities []GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.Where("tenant_id = ?", tenantId).Find(&entities).Error
	if err != nil {
		return nil, err
	}

	gateways := make([]Gateway, len(entities))
	for i, entity := range entities {
		gateways[i], err = GatewayEntityToDomain(&entity)
		if err != nil {
			return nil, err
		}
	}

	return gateways, nil
}

// TODO: hexagonal sbagliato, repo non può ritornare classi di dominio
func (repo *gatewayPostgreRepository) GetAllGateways() ([]Gateway, error) {
	var entities []GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.Find(&entities).Error
	if err != nil {
		return nil, err
	}
	gateways := make([]Gateway, len(entities))
	for i, entity := range entities {
		gateways[i], err = GatewayEntityToDomain(&entity)
		if err != nil {
			return nil, err
		}
	}
	return gateways, nil
}

var _ GatewayRepository = (*gatewayPostgreRepository)(nil)

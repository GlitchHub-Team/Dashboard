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

// methods ============================================================================================

func GatewayEntityToDomain(entity *GatewayEntity) (Gateway, error) {
	if entity == nil {
		return Gateway{}, errors.New("entity is nil")
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

func (repo *gatewayPostgreRepository) GetGatewayByTenantID(tenantId string, gatewayId string) (GatewayEntity, error) {
	var gatewayEntity GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("tenant_id = ? AND id = ?", tenantId, gatewayId).
		First(&gatewayEntity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return GatewayEntity{}, ErrGatewayNotFound
		}
		return GatewayEntity{}, err
	}
	return gatewayEntity, nil
}

// TODO: hexagonal sbagliato, repo non può ritornare classi di dominio
func (repo *gatewayPostgreRepository) GetGatewayById(gatewayId string) (GatewayEntity, error) {
	var entity GatewayEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", gatewayId).
		First(&entity).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return GatewayEntity{}, ErrGatewayNotFound
	}
	if err != nil {
		return GatewayEntity{}, err
	}
	return entity, err
}

// TODO: hexagonal sbagliato, repo non può ritornare classi di dominio
func (repo *gatewayPostgreRepository) GetGatewaysByTenantId(tenantId string, offset int, limit int) ([]GatewayEntity, uint, error) {
	var entities []GatewayEntity
	var count int64

	db := (*gorm.DB)(repo.db)

	baseQuery := db.
		Where("tenant_id = ?", tenantId)

	err := baseQuery.Offset(offset).Limit(limit).Find(&entities).Error
	if err != nil {
		return nil, 0, err
	}

	err = baseQuery.
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	return entities, uint(count), err
}

// TODO: hexagonal sbagliato, repo non può ritornare classi di dominio
func (repo *gatewayPostgreRepository) GetAllGateways(offset int, limit int) ([]GatewayEntity, uint, error) {
	var entities []GatewayEntity
	var totalCount int64
	db := (*gorm.DB)(repo.db)

	if err := db.Model(&GatewayEntity{}).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, uint(totalCount), nil
}

type GatewayRepository interface {
	SaveGateway(entity *GatewayEntity) error
	CreateGateway(entity *GatewayEntity) (Gateway, error)
	DeleteGateway(entity *GatewayEntity) error
	GetGatewayById(gatewayId string) (GatewayEntity, error)
	GetGatewaysByTenantId(tenantId string, offset int, limit int) ([]GatewayEntity, uint, error)
	GetAllGateways(offset int, limit int) ([]GatewayEntity, uint, error)
	GetGatewayByTenantID(tenantId string, gatewayId string) (GatewayEntity, error)
}

var _ GatewayRepository = (*gatewayPostgreRepository)(nil)

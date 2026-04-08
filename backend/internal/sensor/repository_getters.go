package sensor

import (
	"errors"

	"gorm.io/gorm"
)

func (repo *sensorPostgreRepository) GetSensorsByGatewayId(gatewayId string, offset int, limit int) ([]SensorEntity, uint, error) {
	var sensorEntities []SensorEntity
	var count int64
	var err error

	db := (*gorm.DB)(repo.db)
	baseQuery := db.
		Where("gateway_id = ?", gatewayId)

	err = baseQuery.
		Order("name ASC").
		Offset(offset).
		Limit(limit).
		Find(&sensorEntities).Error
	if err != nil {
		return nil, 0, err
	}

	err = baseQuery.
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	return sensorEntities, uint(count), err
}

func (repo *sensorPostgreRepository) GetSensorById(sensorId string) (SensorEntity, error) {
	var sensorEntity SensorEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Where("id = ?", sensorId).
		First(&sensorEntity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SensorEntity{}, ErrSensorNotFound
		}
		return SensorEntity{}, err
	}
	return sensorEntity, nil
}

func (repo *sensorPostgreRepository) GetSensorByTenant(tenantId, sensorId string) (SensorEntity, error) {
	var sensorEntity SensorEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Joins("Gateway").
		Joins("INNER JOIN tenants on tenants.id = ?", tenantId).
		Where("id = ?", sensorId).
		First(&sensorEntity).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SensorEntity{}, ErrSensorNotFound
		}
		return SensorEntity{}, err
	}
	return sensorEntity, nil
}

func (repo *sensorPostgreRepository) GetSensorWithGateway(sensorId string) (SensorEntity, error) {
	var sensorEntity SensorEntity
	db := (*gorm.DB)(repo.db)
	err := db.
		Joins("Gateway").
		Find(&sensorEntity, "sensors.id = ?", sensorId).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return SensorEntity{}, ErrSensorNotFound
		}
		return SensorEntity{}, err
	}

	return sensorEntity, nil
}

func (repo *sensorPostgreRepository) GetSensorsByTenantId(tenantId string, offset int, limit int) ([]SensorEntity, uint, error) {
	var sensorEntities []SensorEntity
	var count int64

	db := (*gorm.DB)(repo.db)
	query := db.Model(&SensorEntity{}).
		Joins("JOIN gateways ON gateways.id = sensors.gateway_id").
		Where("gateways.tenant_id = ?", tenantId)

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("sensors.name ASC").
		Offset(offset).
		Limit(limit).
		Find(&sensorEntities).Error
	if err != nil {
		return nil, 0, err
	}

	return sensorEntities, uint(count), nil
}

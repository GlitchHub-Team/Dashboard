package sensor

func (repo *sensorPostgreRepository) GetSensorsByGatewayId(gatewayId string, offset int, limit int) ([]SensorEntity, uint, error) {
	var sensorEntities []SensorEntity
	var count int64
	var err error

	baseQuery := repo.db.
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
	err := repo.db.
		Where("id = ?", sensorId).
		First(&sensorEntity).Error
	if err != nil {
		return SensorEntity{}, err
	}
	return sensorEntity, nil
}

func (repo *sensorPostgreRepository) GetSensorsByTenantId(tenantId string, offset int, limit int) ([]SensorEntity, uint, error) {
	var sensorEntities []SensorEntity
	var count int64

	query := repo.db.Model(&SensorEntity{}).
		Joins("JOIN gateways ON gateways.gateway_id = sensors.gateway_id").
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

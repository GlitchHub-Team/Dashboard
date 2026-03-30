package sensor

import (
	"backend/internal/infra/database/pagination"

	"github.com/google/uuid"
)

// Compile-time checks
var (
	_ GetSensorByIdPort         = (*DbSensorAdapter)(nil)
	_ GetSensorsByTenantIdPort  = (*DbSensorAdapter)(nil)
	_ GetSensorsByGatewayIdPort = (*DbSensorAdapter)(nil)
)

func (adapater *DbSensorAdapter) GetSensorsByGatewayId(gatewayId uuid.UUID, page int, limit int) ([]Sensor, uint, error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	entities, count, err := adapater.repo.GetSensorsByGatewayId(gatewayId.String(), offset, limit)
	if err != nil {
		return nil, 0, err
	}

	sensors := make([]Sensor, len(entities))
	for i, entity := range entities {
		sensors[i] = entity.ToSensor()
	}

	return sensors, count, nil
}

func (adapter *DbSensorAdapter) GetSensorById(sensorId uuid.UUID) (Sensor, error) {
	entity, err := adapter.repo.GetSensorById(sensorId.String())
	if err != nil {
		return Sensor{}, err
	}

	return entity.ToSensor(), nil
}

func (adapter *DbSensorAdapter) GetSensorsByTenant(tenantId uuid.UUID, page int, limit int) ([]Sensor, uint, error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	entities, count, err := adapter.repo.GetSensorsByTenantId(tenantId.String(), offset, limit)
	if err != nil {
		return nil, 0, err
	}

	sensors := make([]Sensor, len(entities))
	for i, entity := range entities {
		sensors[i] = entity.ToSensor()
	}

	return sensors, count, nil
}

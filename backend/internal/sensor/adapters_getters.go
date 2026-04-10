package sensor

import (
	"backend/internal/infra/database"
	"backend/internal/infra/database/pagination"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/sensor/mocks/ports_getters.go -package=mocks . GetSensorByTenantPort

type GetSensorByTenantPort interface {
	/*
		Ritorna al sensore con ID sensorId associato al tenant con ID tenantId. Questa funzione applica nativamente il controllo
		sul Repository di associazione del sensore al tenant, ritornando ErrSensorNotFound in caso il sensore non sia trovato
		oppure non sia associato al tenant specificato.
	*/
	GetSensorByTenant(tenantId, sensorId uuid.UUID) (
		sensor Sensor, sensorTenantId *uuid.UUID, err error,
	)
}

// Compile-time checks
var (
	_ GetSensorByIdPort         = (*DbSensorAdapter)(nil)
	_ GetSensorByTenantPort     = (*DbSensorAdapter)(nil)
	_ GetSensorsByTenantIdPort  = (*DbSensorAdapter)(nil)
	_ GetSensorsByGatewayIdPort = (*DbSensorAdapter)(nil)
)

func (adapter *DbSensorAdapter) GetSensorsByGatewayId(gatewayId uuid.UUID, page int, limit int) ([]Sensor, uint, error) {
	offset, err := pagination.PageLimitToOffset(page, limit)
	if err != nil {
		return nil, 0, err
	}

	entities, count, err := adapter.repo.GetSensorsByGatewayId(gatewayId.String(), offset, limit)
	if err != nil {
		return nil, 0, err
	}

	sensors, err := database.MapEntityListToDomain(entities, SensorEntityToDomain)
	if err != nil {
		return nil, 0, err
	}
	return sensors, count, nil
}

func (adapter *DbSensorAdapter) GetSensorById(sensorId uuid.UUID) (Sensor, error) {
	entity, err := adapter.repo.GetSensorById(sensorId.String())
	if err != nil {
		return Sensor{}, err
	}

	sensor, err := SensorEntityToDomain(&entity)
	return sensor, err
}

func (adapter *DbSensorAdapter) extractTenantFromSensorEntity(entity *SensorEntity) (
	sensorTenantId *uuid.UUID, err error,
) {
	if entity == nil {
		return nil, ErrSensorNotFound
	}

	if entity.Gateway.TenantId == nil {
		return nil, nil
	}

	tenantId, err := uuid.Parse(*entity.Gateway.TenantId)
	sensorTenantId = &tenantId

	return
}

/*
Implementazione di GetSensorByTenantPort.GetSensorByTenant().
Vedere commento nell'interfaccia per ulteriori dettagli.
*/
func (adapter *DbSensorAdapter) GetSensorByTenant(tenantId, sensorId uuid.UUID) (
	sensor Sensor, sensorTenantId *uuid.UUID, err error,
) {
	entity, err := adapter.repo.GetSensorByTenant(tenantId.String(), sensorId.String())
	if err != nil {
		return
	}

	sensor, err = SensorEntityToDomain(&entity)
	if err != nil {
		return
	}
	sensorTenantId, err = adapter.extractTenantFromSensorEntity(&entity)
	return
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

	sensors, err := database.MapEntityListToDomain(entities, SensorEntityToDomain)
	if err != nil {
		return nil, 0, err
	}
	return sensors, count, nil
}

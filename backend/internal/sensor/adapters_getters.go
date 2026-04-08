package sensor

import (
	"backend/internal/gateway"
	"backend/internal/infra/database/pagination"
	"backend/internal/tenant"

	"github.com/google/uuid"
)

type GetSensorByTenantPort interface {
	/*
		Ritorna al sensore con ID sensorId associato al tenant con ID tenantId. Questa funzione applica nativamente il controllo
		sul Repository di associazione del sensore al tenant, ritornando ErrSensorNotFound in caso il sensore non sia trovato
		oppure non sia associato al tenant specificato.
	*/
	GetSensorByTenant(tenantId, sensorId uuid.UUID) (Sensor, tenant.Tenant, error)
}

type GetSensorWithGatewayPort interface {
	/*
		Ritorna il sensore con ID sensorId e, se esiste, il tenant associato. Se il sensore non è associato ad alcun
		tenant allora tenant è == tenant.Tenant{} (zero-value dello struct)
	*/
	GetSensorWithGateway(sensorId uuid.UUID) (sensor Sensor, gateway gateway.Gateway, err error)
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

func (adapter *DbSensorAdapter) extractTenantFromSensorEntity(entity *SensorEntity) (tenant.Tenant, error) {
	if entity == nil {
		return tenant.Tenant{}, ErrSensorNotFound
	}

	tenantEntity := entity.Gateway.Tenant
	if tenantEntity == nil {
		return tenant.Tenant{}, nil
	}

	tenantObj, err := tenant.TenantEntityToDomain(tenantEntity)
	if err != nil {
		return tenant.Tenant{}, err
	}

	return tenantObj, nil
}

/*
Implementazione di GetSensorByTenantPort.GetSensorByTenant().
Vedere commento nell'interfaccia per ulteriori dettagli.
*/
func (adapter *DbSensorAdapter) GetSensorByTenant(tenantId, sensorId uuid.UUID) (Sensor, tenant.Tenant, error) {
	entity, err := adapter.repo.GetSensorByTenant(tenantId.String(), sensorId.String())
	if err != nil {
		return Sensor{}, tenant.Tenant{}, err
	}

	sensor := entity.ToSensor()
	tenantObj, err := adapter.extractTenantFromSensorEntity(&entity)
	return sensor, tenantObj, err
}

func (adapter *DbSensorAdapter) GetSensorWithGateway(sensorId uuid.UUID) (Sensor, gateway.Gateway, error) {
	entity, err := adapter.repo.GetSensorWithGateway(sensorId.String())
	if err != nil {
		return Sensor{}, gateway.Gateway{}, err
	}

	sensor := entity.ToSensor()
	gatewayObj, err := gateway.GatewayEntityToDomain(&entity.Gateway)
	return sensor, gatewayObj, err
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

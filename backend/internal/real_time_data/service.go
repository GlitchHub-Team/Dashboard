package real_time_data

import (
	"backend/internal/gateway"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/google/uuid"
)

type RealTimeDataPort interface {
	StartDataRetriever(
		tenantId uuid.UUID, sensor sensor.Sensor, dataChan chan RealTimeRawSample, errorChan chan RealTimeError,
	) error
}

// Service ------------------------------------------------------------------------------------------------------

type RealTimeDataService struct {
	sensorWithGatewayPort sensor.GetSensorWithGatewayPort
	sensorByTenantPort   sensor.GetSensorByTenantPort
	realTimeDataPort     RealTimeDataPort
}

func NewRealTimeDataService(
	sensorWithTenantPort sensor.GetSensorWithGatewayPort,
	sensorByTenantPort sensor.GetSensorByTenantPort,
	realTimeDataPort RealTimeDataPort,
) *RealTimeDataService {
	return &RealTimeDataService{
		sensorWithGatewayPort: sensorWithTenantPort,
		sensorByTenantPort:   sensorByTenantPort,
		realTimeDataPort:     realTimeDataPort,
	}
}

func (s *RealTimeDataService) RetrieveRealTimeData(cmd RetrieveRealTimeDataCommand) (
	dataChannel chan RealTimeRawSample, errChannel chan RealTimeError, err error,
) {
	// 1. Controllo business logic
	var sensorObj sensor.Sensor
	var tenantId *uuid.UUID

	// - Super Admin
	if cmd.Requester.RequesterRole == identity.ROLE_SUPER_ADMIN {
		var gateway gateway.Gateway
		sensorObj, gateway, err = s.sensorWithGatewayPort.GetSensorWithGateway(cmd.SensorId)
		tenantId = gateway.TenantId
	} else
	// - Tenant member (controllo integrato di accesso)
	{
		var tenant tenant.Tenant
		sensorObj, tenant, err = s.sensorByTenantPort.GetSensorByTenant(*cmd.Requester.RequesterTenantId, cmd.SensorId)
		tenantId = &tenant.Id
	}

	if err != nil {
		return nil, nil, err
	}

	if tenantId == nil {
		return nil, nil, sensor.ErrSensorNotActive
	}

	// 2. Creazione canali
	dataChannel = make(chan RealTimeRawSample, 1024)
	errChannel = make(chan RealTimeError, 1024)

	dataPort := s.realTimeDataPort
	err = dataPort.StartDataRetriever(*tenantId, sensorObj, dataChannel, errChannel)
	if err != nil {
		return nil, nil, err
	}

	return dataChannel, errChannel, nil
}

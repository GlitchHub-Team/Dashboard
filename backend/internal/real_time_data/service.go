package real_time_data

import (
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/google/uuid"
)

//go:generate mockgen -destination=../../tests/real_time_data/mocks/ports.go -package=mocks . RealTimeDataPort

type RealTimeDataPort interface {
	StartDataRetriever(
		tenantId uuid.UUID, sensor sensor.Sensor, dataChan chan RealTimeSample, errorChan chan RealTimeError,
	) error
}

// Service ------------------------------------------------------------------------------------------------------

type RealTimeDataService struct {
	tenantPort         tenant.GetTenantPort
	sensorByTenantPort sensor.GetSensorByTenantPort
	realTimeDataPort   RealTimeDataPort
}

var _ GetRealTimeDataUseCase = (*RealTimeDataService)(nil)

func NewRealTimeDataService(
	tenantPort tenant.GetTenantPort,
	sensorByTenantPort sensor.GetSensorByTenantPort,
	realTimeDataPort RealTimeDataPort,
) *RealTimeDataService {
	return &RealTimeDataService{
		tenantPort:         tenantPort,
		sensorByTenantPort: sensorByTenantPort,
		realTimeDataPort:   realTimeDataPort,
	}
}

func (s *RealTimeDataService) GetRealTimeData(cmd GetRealTimeDataCommand) (
	dataChannel chan RealTimeSample, errChannel chan RealTimeError, err error,
) {
	// 1. Ottieni sensore
	sensorObj, sensorTenantId, err := s.sensorByTenantPort.GetSensorByTenant(cmd.TenantId, cmd.SensorId)
	if err != nil {
		return nil, nil, err
	}

	// 2. Ottieni tenant
	sensorTenant, err := s.tenantPort.GetTenant(*sensorTenantId)
	if err != nil {
		return nil, nil, err
	}

	// Check accesso
	// - Super Admin
	if cmd.Requester.RequesterRole == identity.ROLE_SUPER_ADMIN && !sensorTenant.CanImpersonate {
		return nil, nil, tenant.ErrImpersonationFailed
	} else
	// - Tenant Member
	{
		if (sensorTenantId == nil) || (cmd.Requester.RequesterTenantId != nil && *cmd.Requester.RequesterTenantId != *sensorTenantId) {
			return nil, nil, sensor.ErrSensorNotFound
		}
	}

	// 3. Creazione canali
	dataChannel = make(chan RealTimeSample, 16)
	errChannel = make(chan RealTimeError, 1)

	dataPort := s.realTimeDataPort
	err = dataPort.StartDataRetriever(cmd.TenantId, sensorObj, dataChannel, errChannel)
	if err != nil {
		return nil, nil, err
	}

	return dataChannel, errChannel, nil
}

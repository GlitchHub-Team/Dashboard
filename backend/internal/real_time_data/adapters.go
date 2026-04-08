package real_time_data

import (
	"fmt"

	"backend/internal/sensor"

	"github.com/google/uuid"
)

type RealTimeDataNATSAdapter struct {
	reader RealTimeDataNATSReader
}

var _ RealTimeDataPort = (*RealTimeDataNATSAdapter)(nil)

func NewRealTimeDataNATSAdapter(
	reader RealTimeDataNATSReader,
) *RealTimeDataNATSAdapter {
	return &RealTimeDataNATSAdapter{
		reader: reader,
	}
}


func (adapter *RealTimeDataNATSAdapter) getSubject(tenantId, gatewayId, sensorId uuid.UUID) (string, error) {
	if tenantId == uuid.Nil {
		return "", sensor.ErrSensorNotFound
	}

	return fmt.Sprintf("sensor.%s.%s.%s", tenantId.String(), gatewayId.String(), sensorId.String()), nil
}

func (adapter *RealTimeDataNATSAdapter) StartDataRetriever(
	tenantId uuid.UUID, sensor sensor.Sensor,
	dataChan chan RealTimeRawSample, errorChan chan RealTimeError,
) (err error) {
	subject, err := adapter.getSubject(tenantId, sensor.GatewayId, sensor.Id)
	if err != nil {
		return
	}

	go adapter.reader.StartSubscriber(subject, sensor.Profile, dataChan, errorChan)

	return
}

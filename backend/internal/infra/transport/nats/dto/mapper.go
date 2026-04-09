package dto

import (
	"encoding/json"
	"fmt"
	"reflect"

	httpDto "backend/internal/infra/transport/http/dto"
	sensorProfile "backend/internal/sensor/profile"
)

/*
Mappa raw a un oggetto di tipo ConcreteDataSampleNATSDto[T] (rappresentato dall'interfaccia [DataSampleNATSDto]),
dove T è determinato in modo forzato a seconda del parametro profile.
*/
func MapRawToDataSampleNATSDto(profile sensorProfile.SensorProfile, raw json.RawMessage) (DataSampleNATSDto, error) {
	switch profile {
	case sensorProfile.ECG_CUSTOM:
		return mapRawToDto[httpDto.ECGData](raw)
	case sensorProfile.ENVIRONMENTAL_SENSING:
		return mapRawToDto[httpDto.EnvironmentalSensingData](raw)
	case sensorProfile.HEART_RATE:
		return mapRawToDto[httpDto.HeartRateData](raw)
	case sensorProfile.HEALTH_THERMOMETER:
		return mapRawToDto[httpDto.HealthThermometerData](raw)
	case sensorProfile.PULSE_OXIMETER:
		return mapRawToDto[httpDto.PulseOximeterData](raw)
	default:
		return nil, sensorProfile.ErrUnknownProfile
	}
}

func mapRawToDto[T any](raw json.RawMessage) (*ConcreteDataSampleNATSDto[T], error) {
	var decoded ConcreteDataSampleNATSDto[T]
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return &decoded, fmt.Errorf("cannot decode sensor profile data for %v: %w", reflect.TypeFor[T](), err)
	}
	return &decoded, nil
}

package dto

import (
	"encoding/json"
	"fmt"
	"reflect"

	sensorProfile "backend/internal/sensor/profile"
)

type ECGData struct {
	Waveform []int `json:"Waveform" binding:"required,min=1"`
}

type EnvironmentalSensingData struct {
	TemperatureValue float64 `json:"TemperatureValue" binding:"required"`
	HumidityValue    float64 `json:"HumidityValue" binding:"required"`
	PressureValue    float64 `json:"PressureValue" binding:"required"`
}

type HealthThermometerData struct {
	TemperatureValue float64 `json:"TemperatureValue" binding:"required"`
}

type HeartRateData struct {
	BpmValue int `json:"BpmValue" binding:"required"`
}

type PulseOximeterData struct {
	Spo2Value      float64 `json:"Spo2Value" binding:"required"`
	PulseRateValue int     `json:"PulseRateValue" binding:"required"`
}

func DecodeSensorProfileData(profile sensorProfile.SensorProfile, raw json.RawMessage) (any, error) {
	switch profile {
	case sensorProfile.ECG_CUSTOM:
		return decodeSensorProfileData[ECGData](raw)
	case sensorProfile.ENVIRONMENTAL_SENSING:
		return decodeSensorProfileData[EnvironmentalSensingData](raw)
	case sensorProfile.HEALTH_THERMOMETER:
		return decodeSensorProfileData[HealthThermometerData](raw)
	case sensorProfile.HEART_RATE:
		return decodeSensorProfileData[HeartRateData](raw)
	case sensorProfile.PULSE_OXIMETER:
		return decodeSensorProfileData[PulseOximeterData](raw)
	default:
		return nil, fmt.Errorf("unsupported sensor profile %q", profile)
	}
}

func decodeSensorProfileData[T any](raw json.RawMessage) (T, error) {
	var decoded T
	fmt.Printf(">>>>>>>>>>>> %s <<<<<<<<<<<<<<< \n", string(raw))
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return decoded, fmt.Errorf("cannot decode sensor profile data for %v: %w", reflect.TypeFor[T](), err)
	}
	fmt.Printf(">>>>>>>>>>>> %#v <<<<<<<<<<<<<<< \n", (decoded))
	return decoded, nil
}

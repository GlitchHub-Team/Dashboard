package dto

import (
	"encoding/json"
	"fmt"
)

const (
	heartRateProfile            = "heart_rate"
	pulseOximeterProfile        = "pulse_oximeter"
	ecgCustomProfile            = "ecg_custom"
	healthThermometerProfile    = "health_thermometer"
	environmentalSensingProfile = "environmental_sensing"
)

type ECGData struct {
	Waveform []int `json:"Waveform" binding:"required,min=250,max=250"`
}

type EnvironmentalSensingData struct {
	TemperatureValue float64 `json:"TemperatureValue" binding:"required"`
	HumidityValue    float64 `json:"HumidityValue" binding:"required"`
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

func DecodeSensorProfileData(profile string, raw json.RawMessage) (any, error) {
	switch profile {
	case ecgCustomProfile:
		return decodeSensorProfileData[ECGData](profile, raw)
	case environmentalSensingProfile:
		return decodeSensorProfileData[EnvironmentalSensingData](profile, raw)
	case healthThermometerProfile:
		return decodeSensorProfileData[HealthThermometerData](profile, raw)
	case heartRateProfile:
		return decodeSensorProfileData[HeartRateData](profile, raw)
	case pulseOximeterProfile:
		return decodeSensorProfileData[PulseOximeterData](profile, raw)
	default:
		return nil, fmt.Errorf("unsupported sensor profile %q", profile)
	}
}

func decodeSensorProfileData[T any](profile string, raw json.RawMessage) (T, error) {
	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return decoded, fmt.Errorf("decode sensor profile data for %q: %w", profile, err)
	}
	return decoded, nil
}

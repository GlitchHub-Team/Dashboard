package profile

type SensorProfile string

const (
	HEART_RATE            SensorProfile = "heart_rate"
	PULSE_OXIMETER        SensorProfile = "pulse_oximeter"
	ECG_CUSTOM            SensorProfile = "ecg_custom"
	HEALTH_THERMOMETER    SensorProfile = "health_thermometer"
	ENVIRONMENTAL_SENSING SensorProfile = "environmental_sensing"
)

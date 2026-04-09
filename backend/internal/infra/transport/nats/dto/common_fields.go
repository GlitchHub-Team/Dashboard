package dto

type GatewayIdField struct {
	GatewayId string `json:"gatewayId" binding:"required,uuid"`
}

type SensorIdField struct {
	SensorId string `json:"sensorId" binding:"required,uuid"`
}

type TenantIdField struct {
	TenantId string `json:"tenantId" binding:"required,uuid"`
}

type TimestampField struct {
	Timestamp string `json:"timestamp" binding:"required,datetime=2006-01-02T15:04:05.999999999Z07:00"` // NOTA: formato definito da time.RFC3339Nano
}

type ProfileField struct {
	Profile string `json:"profile" binding:"required,one_of=ECG EnvironmentalSensing HealthThermometer HeartRate PulseOximeter"`
}
package dto

type DataSampleNATSDto interface {
	GetTimestamp() string
}

type ConcreteDataSampleNATSDto[T any] struct {
	SensorIdField
	GatewayIdField
	TenantIdField
	TimestampField
	ProfileField
	Data T `json:"data" binding:"required"`
}

var _ DataSampleNATSDto = (*ConcreteDataSampleNATSDto[any])(nil)

func (dto *ConcreteDataSampleNATSDto[T]) GetTimestamp() string {
	return dto.Timestamp
}

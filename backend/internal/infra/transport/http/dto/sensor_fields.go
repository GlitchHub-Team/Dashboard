package dto

type SensorIdField struct {
	SensorId string `uri:"sensor_id" form:"sensor_id" json:"sensor_id" binding:"required,uuid4"`
}

package sensor

import (
	"time"

	"backend/internal/gateway"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	CREATE_SENSOR_CMD_SUBJECT    = "commands.addsensor"
	DELETE_SENSOR_CMD_SUBJECT    = "commands.deletesensor"
	INTERRUPT_SENSOR_CMD_SUBJECT = "commands.interruptsensor"
	RESUME_SENSOR_CMD_SUBJECT    = "commands.resumesensor"
	TIMEOUT_DURATION             = 10 * time.Second
)

type sensorPostgreRepository struct {
	log *zap.Logger
	db  *gorm.DB
}

type CommandResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type sensorNatsRepository struct {
	log *zap.Logger
	nc  *nats.Conn
}

func NewSensorPostgreRepository(log *zap.Logger, db *gorm.DB) *sensorPostgreRepository {
	return &sensorPostgreRepository{
		log: log,
		db:  db,
	}
}

func NewSensorNatsRepository(log *zap.Logger, nc *nats.Conn) *sensorNatsRepository {
	return &sensorNatsRepository{
		log: log,
		nc:  nc,
	}
}

type SensorEntity struct {
	Id        string                `gorm:"primaryKey"`
	GatewayId string                `gorm:"column:gateway_id;index"`
	Gateway   gateway.GatewayEntity `gorm:"foreignKey:GatewayId;not null;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name      string                `gorm:"size:255;not null"`
	Interval  int64                 `gorm:"not null"`
	Profile   string                `gorm:"size:50;not null"`
	Status    string                `gorm:"size:50;not null,check:status = 'active' or status = 'inactive',default:'active'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateSensorCmdEntity struct {
	SensorId  string `json:"sensorId"`
	GatewayId string `json:"gatewayId"`
	Interval  int64  `json:"interval"`
	Profile   string `json:"profile"`
}

type DeleteSensorCmdEntity struct {
	SensorId  string `json:"sensorId"`
	GatewayId string `json:"gatewayId"`
}

type InterruptSensorCmdEntity struct {
	SensorId  string `json:"sensorId"`
	GatewayId string `json:"gatewayId"`
}

type ResumeSensorCmdEntity struct {
	SensorId  string `json:"sensorId"`
	GatewayId string `json:"gatewayId"`
}

func (SensorEntity) TableName() string { return "sensors" }

func FromSensor(s Sensor) *SensorEntity {
	entity := &SensorEntity{}
	entity.Id = s.Id.String()
	entity.GatewayId = s.GatewayId.String()
	entity.Name = s.Name
	entity.Interval = s.Interval.Milliseconds()
	entity.Profile = string(s.Profile)
	entity.Status = string(s.Status)
	return entity
}

func (entity *SensorEntity) ToSensor() Sensor {
	return Sensor{
		Id:        uuid.MustParse(entity.Id),
		GatewayId: uuid.MustParse(entity.GatewayId),
		Name:      entity.Name,
		Interval:  time.Duration(entity.Interval) * time.Millisecond,
		Profile:   SensorProfile(entity.Profile),
		Status:    SensorStatus(entity.Status),
	}
}

package sensor

import "go.uber.org/zap"

//go:generate mockgen -destination=../../tests/sensor/mocks/use_cases_create_delete.go -package=mocks . CreateSensorUseCase,DeleteSensorUseCase
//go:generate mockgen -destination=../../tests/sensor/mocks/use_cases_getters.go -package=mocks . GetSensorUseCase,GetSensorsByGatewayUseCase,GetSensorsByTenantUseCase
//go:generate mockgen -destination=../../tests/sensor/mocks/use_cases_interrupt_resume.go -package=mocks . InterruptSensorUseCase,ResumeSensorUseCase

type CreateSensorUseCase interface {
	CreateSensor(cmd CreateSensorCommand) (Sensor, error)
}

type DeleteSensorUseCase interface {
	DeleteSensor(cmd DeleteSensorCommand) (Sensor, error)
}

/* Getters */

type GetSensorUseCase interface {
	GetSensorById(cmd GetSensorCommand) (Sensor, error)
}

type GetSensorsByGatewayUseCase interface {
	GetSensorsByGateway(cmd GetSensorsByGatewayCommand) ([]Sensor, uint, error)
}

type GetSensorsByTenantUseCase interface {
	GetSensorsByTenant(cmd GetSensorsByTenantCommand) ([]Sensor, uint, error)
}

/* Comandi */

type InterruptSensorUseCase interface {
	InterruptSensor(cmd InterruptSensorCommand) error
}

type ResumeSensorUseCase interface {
	ResumeSensor(cmd ResumeSensorCommand) error
}

type SensorController struct {
	log *zap.Logger

	createSensorUseCase CreateSensorUseCase
	deleteSensorUseCase DeleteSensorUseCase

	getSensorUseCase           GetSensorUseCase
	getSensorsByGatewayUseCase GetSensorsByGatewayUseCase
	getSensorsByTenantUseCase  GetSensorsByTenantUseCase

	interruptSensorUseCase InterruptSensorUseCase
	resumeSensorUseCase    ResumeSensorUseCase
}

func NewSensorController(
	log *zap.Logger,

	createSensorUseCase CreateSensorUseCase,
	deleteSensorUseCase DeleteSensorUseCase,

	getSensorUseCase GetSensorUseCase,
	getSensorsByGatewayUseCase GetSensorsByGatewayUseCase,
	getSensorsByTenantUseCase GetSensorsByTenantUseCase,

	interruptSensorUseCase InterruptSensorUseCase,
	resumeSensorUseCase ResumeSensorUseCase,
) *SensorController {
	return &SensorController{
		log: log,

		createSensorUseCase: createSensorUseCase,
		deleteSensorUseCase: deleteSensorUseCase,

		getSensorUseCase:           getSensorUseCase,
		getSensorsByGatewayUseCase: getSensorsByGatewayUseCase,
		getSensorsByTenantUseCase:  getSensorsByTenantUseCase,

		interruptSensorUseCase: interruptSensorUseCase,
		resumeSensorUseCase:    resumeSensorUseCase,
	}
}

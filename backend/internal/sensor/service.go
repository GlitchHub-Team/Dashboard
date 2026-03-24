package sensor

import (
	"backend/internal/gateway"
)

type SensorService struct {
	getGatewayPort gateway.GetGatewayPort
}

type GetSensorByGatewayCommand struct{}

func (service *SensorService) GetSensorByGateway(command GetSensorByGatewayCommand) {
	service.getGatewayPort.GetAll()
}

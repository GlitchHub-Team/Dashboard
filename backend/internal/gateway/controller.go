package gateway




type GatewayController struct {
	createGatewayUseCase CreateGatewayUseCase
	deleteGatewayUseCase DeleteGatewayUseCase
	// ...
}

func (this *GatewayController) CreateGateway(command createGatewayDTO) error {

	return nil
}

func (this *GatewayController) DeleteGateway(command deleteGatewayDTO) error{
	
	return nil
}
package gateway

type CreateGatewayUseCase interface  {
	CreateGateway(command CreateGatewayCommand) error
}

type DeleteGatewayUseCase interface  {
	DeleteGateway(command DeleteGatewayCommand) error

}


type GatewayService struct {
	
}
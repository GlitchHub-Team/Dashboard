package gateway

// import gorm

type DB any // TODO: solo per test

type gatewayPostgreRepository struct {
	// db DB
}

func newGatewayPostgreRepository() gatewayPostgreRepository { return gatewayPostgreRepository{} }

type gatewayEntity struct{}

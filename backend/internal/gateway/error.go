package gateway

import (
	"errors"
)

var (
	ErrGatewayAlreadyAssigned = errors.New("gateway ha già un tenant assegnato")
	ErrSaveGateway            = errors.New("errore salvataggio gateway")

)
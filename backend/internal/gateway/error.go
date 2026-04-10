package gateway

import (
	"errors"
)

var (
	ErrGatewayAlreadyAssigned       = errors.New("gateway ha già un tenant assegnato")
	ErrSaveGateway                  = errors.New("errore salvataggio gateway")
	ErrGatewayNotFound              = errors.New("gateway non trovato")
	ErrUnauthorizedAccess           = errors.New("accesso non autorizzato")
	ErrGatewayAlreadyExists         = errors.New("gateway con lo stesso nome già esistente")
	ErrInvalidGatewayID             = errors.New("ID gateway non valido")
	ErrGatewayAlreadyCommissioned   = errors.New("gateway già commissionato")
	ErrGatewayAlreadyDecommissioned = errors.New("gateway già decommissionato")
	ErrGatewayNotCommissioned       = errors.New("gateway non commissionato")
	ErrComunicationNats             = errors.New("errore comunicazione con NATS")
	ErrGatewayAlreadyInterrupted    = errors.New("gateway già interrotto")
	ErrMissingGatewaySecret         = errors.New("gateway non ha un segreto di firma")
)

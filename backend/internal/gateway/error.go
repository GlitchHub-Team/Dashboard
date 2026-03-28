package gateway

import (
	"errors"
)

var (
	ErrGatewayAlreadyAssigned = errors.New("gateway ha già un tenant assegnato")
	ErrSaveGateway            = errors.New("errore salvataggio gateway")
	ErrGatewayNotFound        = errors.New("gateway non trovato")
	ErrUnauthorizedAccess     = errors.New("accesso non autorizzato")
	ErrGatewayAlreadyExists   = errors.New("gateway con lo stesso nome già esistente")
	ErrInvalidGatewayID       = errors.New("ID gateway non valido")
)

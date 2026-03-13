package gateway

type GatewayStatus int

const (
	GATEWAY_STATUS_ACTIVE 	GatewayStatus = iota
	GATEWAY_STATUS_INACTIVE
)


type Gateway struct {
	Id string
	Name string
	// Tenant *uuid.UUID
	Status GatewayStatus
}


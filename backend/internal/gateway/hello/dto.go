package hello

type GatewayHelloMessageDTO struct {
	GatewayId        string `json:"gatewayId"`
	PublicIdentifier string `json:"publicIdentifier"`
}

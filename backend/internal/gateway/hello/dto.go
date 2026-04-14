package hello

type GatewayHelloMessage struct {
	GatewayId        string `json:"gatewayId"`
	PublicIdentifier string `json:"publicIdentifier"`
}

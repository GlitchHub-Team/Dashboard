package gateway_connection

type GatewayHelloMessage struct {
	GatewayId        string `json:"gatewayid"`
	PublicIdentifier string `json:"publicIdentifier"`
}

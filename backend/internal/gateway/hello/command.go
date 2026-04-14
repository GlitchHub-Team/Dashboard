package hello

import "github.com/google/uuid"

type GatewayHelloMessageCommand struct {
	GatewayId        uuid.UUID
	PublicIdentifier string
}

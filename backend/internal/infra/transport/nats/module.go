package nats

import (
	"os"

	"backend/internal/infra/utils"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"nats",
	fx.Supply(NatsAddress(os.Getenv("NATS_HOST"))),
	fx.Supply(NatsPort(utils.EnvInt("NATS_PORT", 4222))),
	fx.Supply(NatsCredsPath(os.Getenv("DASHBOARD_CREDS_PATH"))),
	fx.Supply(NatsTestCredsPath(os.Getenv("TEST_CREDS_PATH"))),
	fx.Supply(NatsCAPemPath(os.Getenv("CA_PEM_PATH"))),
	fx.Provide(NewNATSConnection),
	fx.Provide(NewNATSTestConnection),
	fx.Provide(NewJetStreamContext),
)

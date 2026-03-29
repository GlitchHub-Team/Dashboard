package gateway_connection

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NATSWorker struct {
	nc      *nats.Conn
	service GatewayHelloService
	logger  *zap.Logger
}

func NewNATSWorker(nc *nats.Conn, service GatewayHelloService, logger *zap.Logger) *NATSWorker {
	return &NATSWorker{nc: nc, service: service, logger: logger}
}

func (w *NATSWorker) Run(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go w.ListenHelloMessages(context.Background())
			return nil
		},
	})
}

func (w *NATSWorker) ProcessMsg(msg jetstream.Msg) {
	var helloMsg GatewayHelloMessage
	if err := json.Unmarshal(msg.Data(), &helloMsg); err != nil {
		msg.Term()
		return
	}

	if err := w.service.ProcessHello(helloMsg); err != nil {
		msg.Nak()
		return
	}

	msg.Ack()
}

func (w *NATSWorker) ListenHelloMessages(ctx context.Context) {
	js, _ := jetstream.New(w.nc)

	cons, _ := js.Consumer(ctx, "HELLO_STREAM", "gateway_hello_consumer")

	cons.Consume(w.ProcessMsg, jetstream.PullMaxMessages(1))

	<-ctx.Done()
}

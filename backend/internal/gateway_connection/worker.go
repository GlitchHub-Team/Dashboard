package gateway_connection

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NATSWorker struct {
	js      jetstream.JetStream
	service GatewayHelloService
	logger  *zap.Logger
}

func NewNATSWorker(js jetstream.JetStream, service GatewayHelloService, logger *zap.Logger) *NATSWorker {
	return &NATSWorker{js: js, service: service, logger: logger}
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
		if err := msg.Term(); err != nil {
			w.logger.Error("failed to Term message", zap.Error(err))
		}
		return
	}

	if err := w.service.ProcessHello(helloMsg); err != nil {
		if err := msg.Nak(); err != nil {
			w.logger.Error("failed to Nak message", zap.Error(err))
		}
		return
	}

	if err := msg.Ack(); err != nil {
		w.logger.Error("failed to Ack message", zap.Error(err))
	}
}

func (w *NATSWorker) ListenHelloMessages(ctx context.Context) {
	cons, err := w.js.Consumer(ctx, "HELLO_STREAM", "gateway_hello_consumer")
	if err != nil {
		w.logger.Error("js.Consumer failed", zap.Error(err))
		return
	}

	_, err = cons.Consume(w.ProcessMsg, jetstream.PullMaxMessages(1))
	if err != nil {
		w.logger.Error("failed to start consume", zap.Error(err))
		return
	}

	<-ctx.Done()
}

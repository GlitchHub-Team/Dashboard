package gateway_connection

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NATSWorker struct {
	js      nats.JetStreamContext
	service GatewayHelloService
	logger  *zap.Logger
}

func NewNATSWorker(js nats.JetStreamContext, service GatewayHelloService, logger *zap.Logger) *NATSWorker {
	return &NATSWorker{js: js, service: service, logger: logger}
}

func (w *NATSWorker) Run(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go w.ListenHelloMessages()
			return nil
		},
	})
}

func (w *NATSWorker) ProcessMsg(msg *nats.Msg) error {
	var helloMsg GatewayHelloMessage
	if err := json.Unmarshal(msg.Data, &helloMsg); err != nil {
		w.logger.Error("Failed to unmarshal hello message", zap.Error(err))
		return err
	}
	if err := w.service.ProcessHello(helloMsg); err != nil {
		w.logger.Error("Failed to process hello message", zap.Error(err))
		return err
	}

	if msg.Reply != "" { // Se c'è un destinatario (NATS reale), allora fai l'Ack
		if err := msg.Ack(); err != nil {
			w.logger.Error("Failed to ack message", zap.Error(err))
			return err
		}
	}
	return nil
}

func (w *NATSWorker) ListenHelloMessages() {
	subject := "gateway.hello.*"
	stream := "HELLO_STREAM"
	consumer := "gateway_hello_consumer"

	for {
		msgs, err := w.js.PullSubscribe(subject, consumer, nats.BindStream(stream))
		if err != nil {
			w.logger.Error("NATS PullSubscribe error", zap.Error(err))
			time.Sleep(2 * time.Second)
			continue
		}
		for {
			messages, err := msgs.Fetch(10, nats.MaxWait(2*time.Second))
			if err != nil && err != nats.ErrTimeout {
				w.logger.Error("NATS Fetch error", zap.Error(err))
				break
			}
			for _, msg := range messages {
				w.ProcessMsg(msg) // errors already logged
			}
		}
	}
}

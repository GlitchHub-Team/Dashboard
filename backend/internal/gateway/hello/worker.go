package hello

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NATSWorker struct {
	consumer            jetstream.Consumer
	gatewayHelloUseCase GatewayHelloUseCase
	logger              *zap.Logger
}

func NewConsumer(js jetstream.JetStream) (jetstream.Consumer, error) {
	return js.CreateOrUpdateConsumer(context.Background(), "HELLO_STREAM", jetstream.ConsumerConfig{
		Durable:       "gateway_hello_consumer",
		FilterSubject: "gateway.hello.*",
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
}

func NewNATSWorker(consumer jetstream.Consumer, service GatewayHelloUseCase, logger *zap.Logger) *NATSWorker {
	return &NATSWorker{consumer: consumer, gatewayHelloUseCase: service, logger: logger}
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
	var helloDto GatewayHelloMessageDTO
	if err := json.Unmarshal(msg.Data(), &helloDto); err != nil {
		if err := msg.Term(); err != nil {
			w.logger.Error("failed to Term message", zap.Error(err))
		}
		return
	}

	gatewayId, err := uuid.Parse(helloDto.GatewayId)
	if err != nil {
		w.logger.Error("failed to parse UUID", zap.Error(err))
		return
	}

	cmd := GatewayHelloMessageCommand{
		GatewayId:        gatewayId,
		PublicIdentifier: helloDto.PublicIdentifier,
	}
	if err := w.gatewayHelloUseCase.ProcessHello(cmd); err != nil {
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
	_, err := w.consumer.Consume(w.ProcessMsg, jetstream.PullMaxMessages(1))
	if err != nil {
		w.logger.Error("failed to start consume", zap.Error(err))
		return
	}

	<-ctx.Done()
}

package nats_utils

import (
	"log"
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
)

type (
	NatsAddress   string
	NatsPort      int
	NatsToken     string
	NatsSeed      string
	NatsCredsPath string
	NatsCAPemPath string
)

func NewNATSConnection(address NatsAddress, port NatsPort, credsPath NatsCredsPath, caPemPath NatsCAPemPath) *nats.Conn {
	options := make([]nats.Option, 0, 2)
	if string(credsPath) != "" {
		options = append(options, CredsFileAuth(string(credsPath)))
	}
	options = append(options, CAPemAuth(string(caPemPath)))

	nc, err := nats.Connect("nats://"+string(address)+":"+strconv.Itoa(int(port)), options...)
	if err != nil {
		log.Fatalf("Error while connecting to NATS server: %v", err)
	}

	return nc
}

func NewJetStreamContext(nc *nats.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func JWTAuth(token, seed string) nats.Option {
	return nats.UserJWT(
		func() (string, error) { return token, nil },
		func(nonce []byte) ([]byte, error) {
			kp, err := nkeys.FromSeed([]byte(seed))
			if err != nil {
				return nil, err
			}
			return kp.Sign(nonce)
		},
	)
}

func CredsFileAuth(credsPath string) nats.Option {
	return nats.UserCredentials(credsPath)
}

func CAPemAuth(caPemPath string) nats.Option {
	if caPemPath == "" {
		return func(_ *nats.Options) error { return nil }
	}

	return nats.RootCAs(caPemPath)
}

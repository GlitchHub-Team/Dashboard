package nats

import (
	"log"
	"strconv"

	natsPkg "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nkeys"
)

type (
	NatsAddress        string
	NatsPort           int
	NatsToken          string
	NatsSeed           string
	NatsCredsPath      string
	NatsTestCredsPath  string
	NatsCAPemPath      string
	NatsTestConnection *natsPkg.Conn
)

func NewNATSConnection(address NatsAddress, port NatsPort, credsPath NatsCredsPath, caPemPath NatsCAPemPath) *natsPkg.Conn {
	options := make([]natsPkg.Option, 0, 2)
	if string(credsPath) != "" {
		options = append(options, CredsFileAuth(string(credsPath)))
	}
	options = append(options, CAPemAuth(string(caPemPath)))

	nc, err := natsPkg.Connect("nats://"+string(address)+":"+strconv.Itoa(int(port)), options...)
	if err != nil {
		log.Fatalf("Error while connecting to NATS server: %v", err)
	}

	return nc
}

func NewNATSTestConnection(address NatsAddress, port NatsPort, credsPath NatsTestCredsPath, caPemPath NatsCAPemPath) NatsTestConnection {
	options := make([]natsPkg.Option, 0, 2)
	if string(credsPath) != "" {
		options = append(options, CredsFileAuth(string(credsPath)))
	}
	options = append(options, CAPemAuth(string(caPemPath)))

	nc, err := natsPkg.Connect("nats://"+string(address)+":"+strconv.Itoa(int(port)), options...)
	if err != nil {
		log.Fatalf("Error while connecting to NATS server: %v", err)
	}

	return nc
}

func NewJetStreamContext(nc *natsPkg.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func JWTAuth(token, seed string) natsPkg.Option {
	return natsPkg.UserJWT(
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

func CredsFileAuth(credsPath string) natsPkg.Option {
	return natsPkg.UserCredentials(credsPath)
}

func CAPemAuth(caPemPath string) natsPkg.Option {
	if caPemPath == "" {
		return func(_ *natsPkg.Options) error { return nil }
	}

	return natsPkg.RootCAs(caPemPath)
}

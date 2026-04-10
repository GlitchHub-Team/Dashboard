package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type (
	NatsAddress   string
	NatsPort      int
	NatsToken     string
	NatsSeed      string
	NatsCredsPath string
	NatsCAPemPath string
)

type SensorMessage struct {
	SensorID  string          `json:"sensorId"`
	GatewayID string          `json:"gatewayId"`
	TenantID  string          `json:"tenantId"`
	Timestamp time.Time       `json:"timestamp"`
	Profile   string          `json:"profile"`
	Data      json.RawMessage `json:"data"`
}

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

func CredsFileAuth(credsPath string) nats.Option {
	return nats.UserCredentials(credsPath)
}

func CAPemAuth(caPemPath string) nats.Option {
	if caPemPath == "" {
		return func(_ *nats.Options) error { return nil }
	}

	return nats.RootCAs(caPemPath)
}

func mustBeUUID(value, fieldName string) {
	if _, err := uuid.Parse(value); err != nil {
		log.Fatalf("%s must be a valid UUID: %v", fieldName, err)
	}
}

func main() {
	tenantID := flag.String("tenantId", "", "Tenant UUID (required)")
	gatewayID := flag.String("gatewayId", "", "Gateway UUID (required)")
	sensorID := flag.String("sensorId", "", "Sensor UUID (required)")
	profile := flag.String("profile", "HeartRate", "Sensor profile")
	publishEvery := flag.Duration("publishEvery", time.Second, "Publish interval (e.g. 500ms, 1s)")

	natsAddress := flag.String("natsAddress", "localhost", "NATS server address")
	natsPort := flag.Int("natsPort", 4222, "NATS server port")
	natsCreds := flag.String("natsCreds", "admin_test.creds", "NATS creds file path")
	natsCA := flag.String("natsCA", "ca.pem", "NATS CA pem file path")

	flag.Parse()

	if *tenantID == "" || *gatewayID == "" || *sensorID == "" {
		log.Fatal("tenantId, gatewayId, and sensorId are required")
	}
	mustBeUUID(*tenantID, "tenantId")
	mustBeUUID(*gatewayID, "gatewayId")
	mustBeUUID(*sensorID, "sensorId")
	if *publishEvery <= 0 {
		log.Fatal("publishEvery must be greater than 0")
	}

	nc := NewNATSConnection(
		NatsAddress(*natsAddress),
		NatsPort(*natsPort),
		NatsCredsPath(*natsCreds),
		NatsCAPemPath(*natsCA),
	)
	defer nc.Close()

	subject := fmt.Sprintf("sensor.%s.%s.%s", *tenantID, *gatewayID, *sensorID)
	ticker := time.NewTicker(*publishEvery)
	defer ticker.Stop()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	log.Printf("Publishing on subject %q every %s. Press Ctrl-C to stop.", subject, publishEvery.String())

	profileStrategy := ParseSensorProfile(*profile, NewRand())

	for {
		select {
		case <-sigCh:
			log.Println("Shutdown signal received, stopping publisher")
			return
		case tickTime := <-ticker.C:
			data, _ := profileStrategy.Generate().Data.Serialize()

			msg := SensorMessage{
				SensorID:  *sensorID,
				GatewayID: *gatewayID,
				TenantID:  *tenantID,
				Timestamp: tickTime,
				Profile:   *profile,
				Data:      data,
			}

			payload, err := json.Marshal(msg)
			if err != nil {
				log.Printf("failed to marshal payload: %v", err)
				continue
			}

			if err := nc.Publish(subject, payload); err != nil {
				log.Printf("failed to publish message: %v", err)
				continue
			}

			log.Printf("published: %s", payload)
		}
	}
}

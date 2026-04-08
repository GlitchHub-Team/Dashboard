package real_time_data

import (
	"encoding/json"
	"fmt"
	"time"

	sensorProfile "backend/internal/sensor/profile"

	"github.com/nats-io/nats.go"
)

type RealTimeDataNATSReader interface {
	StartSubscriber(
		subject string, profile sensorProfile.SensorProfile,
		receivingChannel chan RealTimeRawSample, errorChannel chan RealTimeError,
	) error
}

type concreteRealTimeDataNATSReader struct {
	nc *nats.Conn
}

var _ RealTimeDataNATSReader = (*concreteRealTimeDataNATSReader)(nil) // Compile-time check

func newConcreteRealTimeDataNATSReader(nc *nats.Conn) *concreteRealTimeDataNATSReader {
	return &concreteRealTimeDataNATSReader{
		nc: nc,
	}
}

func (reader *concreteRealTimeDataNATSReader) StartSubscriber(
	subject string,
	profile sensorProfile.SensorProfile,
	receivingChannel chan RealTimeRawSample,
	errorChannel chan RealTimeError,
) error {
	sub, err := reader.nc.Subscribe(subject, func(msg *nats.Msg) {
		sample := RealTimeRawSample{
			Profile:   profile,
			Data:      msg.Data,
			Timestamp: time.Now(),
		}

		receivingChannel <- sample
	})
	defer sub.Unsubscribe() //nolint:errcheck

	if err != nil {
		return err
	}

loop:
	for err := range errorChannel {
		fmt.Printf("[concreteRealTimeDataNATSReader] Interrupting (%v)", err)
		break loop
	}

	return nil
}

// TODO: funzione di test da eliminare prima o poi!!
func mockDataGenerator(
	receivingChannel chan RealTimeRawSample,
	errChannel chan RealTimeError,
) {
	t := time.NewTicker(500 * time.Millisecond)

loop:
	for {
		select {
		case <-t.C:
			b := ([]byte)("\"ok\"")
			receivingChannel <- RealTimeRawSample{
				Data:      json.RawMessage(b),
				Timestamp: time.Now(),
			}
			fmt.Printf("[mockDataGenerator] Generated @ %v\n", time.Now())

		case <-errChannel:
			fmt.Println("[mockDataGenerator] Interrupting")
			break loop
		}
	}
}

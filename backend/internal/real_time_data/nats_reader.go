package real_time_data

import (
	"sync"
	"time"

	sensorProfile "backend/internal/sensor/profile"

	"github.com/nats-io/nats.go"
)

//go:generate mockgen -destination=../../tests/real_time_data/mocks/nats_reader.go -package=mocks . RealTimeDataNATSReader

type RealTimeDataNATSReader interface {
	StartSubscriber(
		subject string, profile sensorProfile.SensorProfile,
		receivingChannel chan RealTimeSample, errorChannel chan RealTimeError,
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

type lastTimestampContainer struct {
	mu    sync.Mutex
	value time.Time
}

/*
Fa controllo thread-safe (usando mutex) su newTime rispetto a t.value: se newTime è più recente di t.value, allora
imposta newTime a t.value e ritorna true, altrimenti ritorna false
*/
func (t *lastTimestampContainer) CompareAndSet(newTime time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// t.value < newTime => non sono fuori ordine
	if t.value.Before(newTime) {
		t.value = newTime
		return true
	}

	return false
}

func (reader *concreteRealTimeDataNATSReader) StartSubscriber(
	subject string,
	profile sensorProfile.SensorProfile,
	receivingChannel chan RealTimeSample,
	errorChannel chan RealTimeError,
) error {
	lastTimestamp := lastTimestampContainer{
		value: time.Now(),
	}

	sub, err := reader.nc.Subscribe(subject, func(msg *nats.Msg) {
		sample, err := MapNATSRawToDomain(profile, msg.Data)
		if err != nil {
			errorChannel <- NewErrMappingError(err)
			return
		}

		if lastTimestamp.CompareAndSet(sample.GetTimestamp()) {
			receivingChannel <- sample
		}
	})
	defer sub.Unsubscribe() //nolint:errcheck

	if err != nil {
		return err
	}

loop:
	for range errorChannel {
		break loop
	}

	return nil
}

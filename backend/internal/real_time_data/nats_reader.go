package real_time_data

import (
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

func (reader *concreteRealTimeDataNATSReader) StartSubscriber(
	subject string,
	profile sensorProfile.SensorProfile,
	receivingChannel chan RealTimeSample,
	errorChannel chan RealTimeError,
) error {
	sub, err := reader.nc.Subscribe(subject, func(msg *nats.Msg) {
		sample, err := MapNATSRawToDomain(profile, msg.Data)

		if err != nil {
			errorChannel <- NewErrMappingError(err)
			return
		}

		receivingChannel <- sample
	})
	defer sub.Unsubscribe() //nolint:errcheck

	if err != nil {
		return err
	}

loop:
	for _ = range errorChannel { 
		break loop
	}

	return nil
}

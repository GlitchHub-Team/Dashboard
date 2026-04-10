package main

import (
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type SensorProfile interface {
	Generate() *GeneratedSensorData
	String() string
}

type SerializableData interface {
	Serialize() ([]byte, error)
}

type GeneratedSensorData struct {
	SensorId  uuid.UUID
	Timestamp time.Time
	Profile   string
	Data      SerializableData
}

type Rand interface {
	Intn(n int) int
	Float64() float64
}

func NewRand() Rand {
	source := rand.NewSource(time.Now().UnixNano())
	return rand.New(source)
}

func ParseSensorProfile(profileType string, rand Rand) SensorProfile {
	switch profileType {
	case "ECG":
		return NewEcgProfile(uuid.New(), rand)
	case "EnvironmentalSensing":
		return NewEnvironmentalSensingProfile(uuid.New(), rand)
	case "HealthThermometer":
		return NewHealthThermometerProfile(uuid.New(), rand)
	case "HeartRate":
		return NewHeartRateProfile(uuid.New(), rand)
	case "PulseOximeter":
		return NewPulseOximeterProfile(uuid.New(), rand)
	default:
		return nil
	}
}

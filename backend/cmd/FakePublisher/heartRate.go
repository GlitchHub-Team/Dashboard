package main

import (
	"time"

	"github.com/google/uuid"
)

type HeartRateProfile struct {
	sensorId uuid.UUID
	rand     Rand
}

func NewHeartRateProfile(sensorId uuid.UUID, rand Rand) *HeartRateProfile {
	return &HeartRateProfile{
		sensorId: sensorId,
		rand:     rand,
	}
}

type HeartRateData struct {
	BpmValue int
}

func generateHeartRate(rand Rand) *HeartRateData {
	bpm := 60 + rand.Intn(41) // da 60 a 100 bpm, range tipico per un adulto a riposo
	return &HeartRateData{
		BpmValue: bpm,
	}
}

func (g *HeartRateProfile) Generate() *GeneratedSensorData {
	data := generateHeartRate(g.rand)

	return &GeneratedSensorData{
		SensorId:  g.sensorId,
		Timestamp: time.Now(),
		Profile:   g.String(),
		Data:      data,
	}
}

func (g *HeartRateProfile) String() string {
	return "HeartRate"
}

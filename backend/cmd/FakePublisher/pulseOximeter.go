package main

import (
	"time"

	"github.com/google/uuid"
)

type PulseOximeterProfile struct {
	sensorId uuid.UUID
	rand     Rand
}

func NewPulseOximeterProfile(sensorId uuid.UUID, rand Rand) *PulseOximeterProfile {
	return &PulseOximeterProfile{
		sensorId: sensorId,
		rand:     rand,
	}
}

type PulseOximeterData struct {
	SpO2Value      float64
	PulseRateValue int
}

func generatePulseOximeter(rand Rand) *PulseOximeterData {
	spO2 := 95.0 + rand.Float64()*5.0 // SpO2 tra 95% e 100%
	pulseRate := 60 + rand.Intn(41)   // da 60 a 100 bpm, range tipico per un adulto a riposo

	return &PulseOximeterData{
		SpO2Value:      spO2,
		PulseRateValue: pulseRate,
	}
}

func (g *PulseOximeterProfile) Generate() *GeneratedSensorData {
	data := generatePulseOximeter(g.rand)

	return &GeneratedSensorData{
		SensorId:  g.sensorId,
		Timestamp: time.Now(),
		Profile:   g.String(),
		Data:      data,
	}
}

func (g *PulseOximeterProfile) String() string {
	return "PulseOximeter"
}

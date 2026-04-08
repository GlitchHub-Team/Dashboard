package main

import (
	"time"

	"github.com/google/uuid"
)

type HealthThermometerProfile struct {
	sensorId uuid.UUID
	rand     Rand
}

func NewHealthThermometerProfile(sensorId uuid.UUID, rand Rand) *HealthThermometerProfile {
	return &HealthThermometerProfile{
		sensorId: sensorId,
		rand:     rand,
	}
}

type HealthThermometerData struct {
	TemperatureValue float64
}

func generateHealthThermometer(rand Rand) *HealthThermometerData {
	temperature := 36.0 + rand.Float64()*2 // Celsius, temperatura corporea

	return &HealthThermometerData{
		TemperatureValue: temperature,
	}
}

func (g *HealthThermometerProfile) Generate() *GeneratedSensorData {
	data := generateHealthThermometer(g.rand)

	return &GeneratedSensorData{
		SensorId:  g.sensorId,
		Timestamp: time.Now(),
		Profile:   g.String(),
		Data:      data,
	}
}

func (g *HealthThermometerProfile) String() string {
	return "HealthThermometer"
}

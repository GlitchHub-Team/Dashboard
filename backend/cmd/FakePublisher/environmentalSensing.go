package main

import (
	"time"

	"github.com/google/uuid"
)

type EnvironmentalSensingProfile struct {
	sensorId uuid.UUID
	rand     Rand
}

func NewEnvironmentalSensingProfile(sensorId uuid.UUID, rand Rand) *EnvironmentalSensingProfile {
	return &EnvironmentalSensingProfile{
		sensorId: sensorId,
		rand:     rand,
	}
}

type EnvironmentalSensingData struct {
	TemperatureValue float64
	HumidityValue    float64
	PressureValue    float64
}

func generateEnvironmentalSensing(rand Rand) *EnvironmentalSensingData {
	temperature := 15.0 + rand.Float64()*15.0 // Float64 è tra 0.0 e 1.0, 1.0 escluso
	humidity := 30.0 + rand.Float64()*40.0
	pressure := 980.0 + rand.Float64()*50.0

	return &EnvironmentalSensingData{
		TemperatureValue: temperature,
		HumidityValue:    humidity,
		PressureValue:    pressure,
	}
}

func (g *EnvironmentalSensingProfile) Generate() *GeneratedSensorData {
	data := generateEnvironmentalSensing(g.rand)

	return &GeneratedSensorData{
		SensorId:  g.sensorId,
		Timestamp: time.Now(),
		Profile:   g.String(),
		Data:      data,
	}
}

func (g *EnvironmentalSensingProfile) String() string {
	return "EnvironmentalSensing"
}

package main

import (
	"time"

	"github.com/google/uuid"
)

const (
	N_SAMPLES = 250
	MIN_UV    = -500
	MAX_UV    = 1500
)

type EcgProfile struct {
	sensorId uuid.UUID
	rand     Rand
}

func NewEcgProfile(sensorId uuid.UUID, rand Rand) *EcgProfile {
	return &EcgProfile{
		sensorId: sensorId,
		rand:     rand,
	}
}

type EcgData struct {
	Waveform []int
}

func generateEcg(rand Rand) []int {
	waveform := make([]int, N_SAMPLES)

	rangeSize := MAX_UV - MIN_UV + 1

	for i := 0; i < N_SAMPLES; i++ {
		waveform[i] = rand.Intn(rangeSize) + MIN_UV
	}

	return waveform
}

func (g *EcgProfile) Generate() *GeneratedSensorData {
	data := &EcgData{Waveform: generateEcg(g.rand)}

	return &GeneratedSensorData{
		SensorId:  g.sensorId,
		Timestamp: time.Now(),
		Profile:   g.String(),
		Data:      data,
	}
}

func (g *EcgProfile) String() string {
	return "ECG"
}

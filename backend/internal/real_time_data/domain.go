package real_time_data

import (
	// "encoding/json"
	"fmt"
	"time"

	sensorProfile "backend/internal/sensor/profile"
)

type RealTimeSample interface {
	GetData() any
	GetProfile() sensorProfile.SensorProfile
	GetTimestamp() time.Time
}

type BaseSample struct {
	Profile   sensorProfile.SensorProfile
	Timestamp time.Time
}

// ECG -------------------------------------------------------------
type EcgSampleData struct {
	Waveform []int
}
type ECGSample struct {
	BaseSample
	Data EcgSampleData
}

var _ RealTimeSample = (*ECGSample)(nil)

func (s *ECGSample) GetData() any                            { return s.Data }
func (s *ECGSample) GetProfile() sensorProfile.SensorProfile { return s.Profile }
func (s *ECGSample) GetTimestamp() time.Time                 { return s.Timestamp }

// ESS -------------------------------------------------------------
type EnvironmentalSensingSampleData struct {
	Temperature float64
	Humidity    float64
	Pressure    float64
}
type EnvironmentalSensingSample struct {
	BaseSample
	Data EnvironmentalSensingSampleData
}

var _ RealTimeSample = (*EnvironmentalSensingSample)(nil)

func (s *EnvironmentalSensingSample) GetData() any                            { return s.Data }
func (s *EnvironmentalSensingSample) GetProfile() sensorProfile.SensorProfile { return s.Profile }
func (s *EnvironmentalSensingSample) GetTimestamp() time.Time                 { return s.Timestamp }

// Health Thermometer ----------------------------------------------
type HealthThermometerSampleData struct {
	Temperature float64
}
type HealthThermometerSample struct {
	BaseSample
	Data HealthThermometerSampleData
}

var _ RealTimeSample = (*HealthThermometerSample)(nil)

func (s *HealthThermometerSample) GetData() any                            { return s.Data }
func (s *HealthThermometerSample) GetProfile() sensorProfile.SensorProfile { return s.Profile }
func (s *HealthThermometerSample) GetTimestamp() time.Time                 { return s.Timestamp }

// Heart Rate ------------------------------------------------------
type HeartRateSampleData struct {
	BpmValue int
}
type HeartRateSample struct {
	BaseSample
	Data HeartRateSampleData
}

var _ RealTimeSample = (*HeartRateSample)(nil)

func (s *HeartRateSample) GetData() any                            { return s.Data }
func (s *HeartRateSample) GetProfile() sensorProfile.SensorProfile { return s.Profile }
func (s *HeartRateSample) GetTimestamp() time.Time                 { return s.Timestamp }

// Pulse Oximeter --------------------------------------------------
type PulseOximeterSampleData struct {
	Spo2      float64
	PulseRate int
}
type PulseOximeterSample struct {
	BaseSample
	Data PulseOximeterSampleData
}

var _ RealTimeSample = (*PulseOximeterSample)(nil)

func (s *PulseOximeterSample) GetData() any                            { return s.Data }
func (s *PulseOximeterSample) GetProfile() sensorProfile.SensorProfile { return s.Profile }
func (s *PulseOximeterSample) GetTimestamp() time.Time                 { return s.Timestamp }

// Error ---------------------------------------------------------

type RealTimeError struct {
	Err       error
	Timestamp time.Time
}

func (e RealTimeError) Error() string {
	return e.Err.Error()
}

func (e RealTimeError) Unwrap() error {
	return e.Err
}

func NewErrClientDisconnected() RealTimeError {
	return RealTimeError{
		Err:       ErrClientDisconnected,
		Timestamp: time.Now(),
	}
}

func NewErrMappingError(err error) RealTimeError {
	return RealTimeError{
		Err:       fmt.Errorf("%w: %w", ErrMappingError, err),
		Timestamp: time.Now(),
	}
}

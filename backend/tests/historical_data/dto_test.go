package historical_data_test

import (
	"encoding/json"
	"testing"
	"time"

	"backend/internal/historical_data"
	transportDto "backend/internal/infra/transport/http/dto"
	sensorProfile "backend/internal/sensor/profile"

	"github.com/google/uuid"
)

func TestNewHistoricalDataResponseDTO_DecodesDataByProfile(t *testing.T) {
	timestamp := time.Now().UTC()

	waveform := make([]int, 250)
	for i := range waveform {
		waveform[i] = i
	}

	tests := []struct {
		name         string
		profile      string
		rawData      string
		assertSample func(t *testing.T, sample historical_data.HistoricalSampleResponseDTO)
	}{
		{
			name:    "heart rate",
			profile: string(sensorProfile.HEART_RATE),
			rawData: `{"BpmValue":72}`,
			assertSample: func(t *testing.T, sample historical_data.HistoricalSampleResponseDTO) {
				t.Helper()
				data, ok := sample.Data.(transportDto.HeartRateData)
				if !ok {
					t.Fatalf("unexpected data type %T", sample.Data)
				}
				if data.BpmValue != 72 {
					t.Fatalf("unexpected bpm value: got %d want 72", data.BpmValue)
				}
			},
		},
		{
			name:    "pulse oximeter",
			profile: string(sensorProfile.PULSE_OXIMETER),
			rawData: `{"Spo2Value":98,"PulseRateValue":71}`,
			assertSample: func(t *testing.T, sample historical_data.HistoricalSampleResponseDTO) {
				t.Helper()
				data, ok := sample.Data.(transportDto.PulseOximeterData)
				if !ok {
					t.Fatalf("unexpected data type %T", sample.Data)
				}
				if data.Spo2Value != 98 || data.PulseRateValue != 71 {
					t.Fatalf("unexpected pulse oximeter data: got %+v", data)
				}
			},
		},
		{
			name:    "environmental sensing",
			profile: string(sensorProfile.ENVIRONMENTAL_SENSING),
			rawData: `{"TemperatureValue":21.4,"HumidityValue":56.2}`,
			assertSample: func(t *testing.T, sample historical_data.HistoricalSampleResponseDTO) {
				t.Helper()
				data, ok := sample.Data.(transportDto.EnvironmentalSensingData)
				if !ok {
					t.Fatalf("unexpected data type %T", sample.Data)
				}
				if data.TemperatureValue != 21.4 || data.HumidityValue != 56.2 {
					t.Fatalf("unexpected environmental data: got %+v", data)
				}
			},
		},
		{
			name:    "health thermometer",
			profile: string(sensorProfile.HEALTH_THERMOMETER),
			rawData: `{"TemperatureValue":36.7}`,
			assertSample: func(t *testing.T, sample historical_data.HistoricalSampleResponseDTO) {
				t.Helper()
				data, ok := sample.Data.(transportDto.HealthThermometerData)
				if !ok {
					t.Fatalf("unexpected data type %T", sample.Data)
				}
				if data.TemperatureValue != 36.7 {
					t.Fatalf("unexpected temperature value: got %v want 36.7", data.TemperatureValue)
				}
			},
		},
		{
			name:    "ecg custom",
			profile: string(sensorProfile.ECG_CUSTOM),
			rawData: mustMarshalJSON(t, transportDto.ECGData{Waveform: waveform}),
			assertSample: func(t *testing.T, sample historical_data.HistoricalSampleResponseDTO) {
				t.Helper()
				data, ok := sample.Data.(transportDto.ECGData)
				if !ok {
					t.Fatalf("unexpected data type %T", sample.Data)
				}
				if len(data.Waveform) != 250 {
					t.Fatalf("unexpected waveform length: got %d want 250", len(data.Waveform))
				}
				if data.Waveform[0] != 0 || data.Waveform[249] != 249 {
					t.Fatalf("unexpected waveform bounds: got first=%d last=%d", data.Waveform[0], data.Waveform[249])
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := historical_data.NewHistoricalDataResponseDTO([]historical_data.HistoricalSample{
				{
					SensorId:  uuid.New(),
					GatewayId: uuid.New(),
					TenantId:  uuid.New(),
					Profile:   sensorProfile.SensorProfile(test.profile),
					Timestamp: timestamp,
					Data:      json.RawMessage(test.rawData),
				},
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if response.Count != 1 {
				t.Fatalf("unexpected count: got %d want 1", response.Count)
			}
			if len(response.Samples) != 1 {
				t.Fatalf("unexpected sample length: got %d want 1", len(response.Samples))
			}

			test.assertSample(t, response.Samples[0])
		})
	}
}

func TestNewHistoricalDataResponseDTO_ReturnsErrorForUnsupportedProfile(t *testing.T) {
	_, err := historical_data.NewHistoricalDataResponseDTO([]historical_data.HistoricalSample{
		{
			SensorId:  uuid.New(),
			GatewayId: uuid.New(),
			TenantId:  uuid.New(),
			Profile:   "unsupported_profile",
			Timestamp: time.Now().UTC(),
			Data:      json.RawMessage(`{"foo":"bar"}`),
		},
	})
	if err == nil {
		t.Fatal("expected error for unsupported profile")
	}
}

func TestNewHistoricalDataResponseDTO_ReturnsErrorForMalformedJSON(t *testing.T) {
	_, err := historical_data.NewHistoricalDataResponseDTO([]historical_data.HistoricalSample{
		{
			SensorId:  uuid.New(),
			GatewayId: uuid.New(),
			TenantId:  uuid.New(),
			Profile:   sensorProfile.HEART_RATE,
			Timestamp: time.Now().UTC(),
			Data:      json.RawMessage(`{"BpmValue":`),
		},
	})
	if err == nil {
		t.Fatal("expected error for malformed json")
	}
}

func mustMarshalJSON(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(raw)
}

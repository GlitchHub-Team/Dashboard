package historical_data_test

import (
	"encoding/json"
	"testing"
	"time"

	"backend/internal/historical_data"

	"github.com/google/uuid"
)

func TestNewHistoricalDataResponseDTO(t *testing.T) {
	t.Run("maps samples into response dto", func(t *testing.T) {
		sensorId := uuid.New()
		gatewayId := uuid.New()
		tenantId := uuid.New()
		ts := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
		payload := json.RawMessage(`{"value":72}`)

		response := historical_data.NewHistoricalDataResponseDTO([]historical_data.HistoricalSample{
			{
				SensorId:  sensorId,
				GatewayId: gatewayId,
				TenantId:  tenantId,
				Profile:   "HeartRate",
				Timestamp: ts,
				Data:      payload,
			},
		})

		if response.Count != 1 {
			t.Fatalf("expected count 1, got %d", response.Count)
		}
		if len(response.Samples) != 1 {
			t.Fatalf("expected 1 sample, got %d", len(response.Samples))
		}
		if response.Samples[0].SensorId != sensorId.String() {
			t.Fatalf("unexpected sensor id: %s", response.Samples[0].SensorId)
		}
		if response.Samples[0].GatewayId != gatewayId.String() {
			t.Fatalf("unexpected gateway id: %s", response.Samples[0].GatewayId)
		}
		if response.Samples[0].TenantId != tenantId.String() {
			t.Fatalf("unexpected tenant id: %s", response.Samples[0].TenantId)
		}
		if response.Samples[0].Profile != "HeartRate" {
			t.Fatalf("unexpected profile: %s", response.Samples[0].Profile)
		}
	})

	t.Run("returns empty slice info for empty input", func(t *testing.T) {
		response := historical_data.NewHistoricalDataResponseDTO(nil)
		if response.Count != 0 {
			t.Fatalf("expected count 0, got %d", response.Count)
		}
		if len(response.Samples) != 0 {
			t.Fatalf("expected zero samples, got %d", len(response.Samples))
		}
	})
}

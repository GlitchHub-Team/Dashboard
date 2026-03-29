package integrationtests

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"backend/internal/historical_data"
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/google/uuid"
)

func TestGetSensorHistoricalData_ValidData_ReturnsStoredSamples(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	db := setupTimescaleDB(t)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sensorID := uuid.New()
	gatewayID := uuid.New()
	ts := time.Now().UTC().Truncate(time.Microsecond)
	payload := json.RawMessage(`{"value":72}`)

	insertSensorDataRow(t, db, tenantID, sensorID, gatewayID, ts, "HeartRate", payload)
	t.Cleanup(func() {
		cleanupSensorDataRow(t, db, tenantID, sensorID, gatewayID, ts)
	})

	useCase := setupHistoricalDataUseCase(t, db, &staticTenantPort{
		tenant: tenant.Tenant{
			Id:             tenantID,
			CanImpersonate: false,
		},
	})

	samples, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &tenantID,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: tenantID,
		SensorId: sensorID,
	})
	if err != nil {
		t.Fatalf("GetSensorHistoricalData() unexpected error: %v", err)
	}

	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	if samples[0].SensorId != sensorID {
		t.Fatalf("unexpected sensor id: got %v want %v", samples[0].SensorId, sensorID)
	}
	if samples[0].GatewayId != gatewayID {
		t.Fatalf("unexpected gateway id: got %v want %v", samples[0].GatewayId, gatewayID)
	}
	if samples[0].TenantId != tenantID {
		t.Fatalf("unexpected tenant id: got %v want %v", samples[0].TenantId, tenantID)
	}
}

func TestGetSensorHistoricalData_FilterByTimeRange_ReturnsSubset(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	db := setupTimescaleDB(t)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sensorID := uuid.New()
	gatewayID := uuid.New()

	ts1 := time.Now().UTC().Truncate(time.Microsecond)
	ts2 := ts1.Add(2 * time.Minute)

	insertSensorDataRow(t, db, tenantID, sensorID, gatewayID, ts1, "HeartRate", []byte(`{"value":70}`))
	insertSensorDataRow(t, db, tenantID, sensorID, gatewayID, ts2, "HeartRate", []byte(`{"value":75}`))
	t.Cleanup(func() {
		cleanupSensorDataRow(t, db, tenantID, sensorID, gatewayID, ts1)
		cleanupSensorDataRow(t, db, tenantID, sensorID, gatewayID, ts2)
	})

	useCase := setupHistoricalDataUseCase(t, db, &staticTenantPort{
		tenant: tenant.Tenant{
			Id:             tenantID,
			CanImpersonate: false,
		},
	})

	from := ts2.Add(-time.Second)
	to := ts2.Add(time.Second)

	samples, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &tenantID,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: tenantID,
		SensorId: sensorID,
		From:     &from,
		To:       &to,
	})
	if err != nil {
		t.Fatalf("GetSensorHistoricalData() unexpected error: %v", err)
	}

	if len(samples) != 1 {
		t.Fatalf("expected 1 filtered sample, got %d", len(samples))
	}
	if !samples[0].Timestamp.Equal(ts2) {
		t.Fatalf("unexpected timestamp: got %v want %v", samples[0].Timestamp, ts2)
	}
}

func TestGetSensorHistoricalData_NoData_ReturnsEmptySlice(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	db := setupTimescaleDB(t)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	sensorID := uuid.New()

	useCase := setupHistoricalDataUseCase(t, db, &staticTenantPort{
		tenant: tenant.Tenant{
			Id:             tenantID,
			CanImpersonate: false,
		},
	})

	samples, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &tenantID,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: tenantID,
		SensorId: sensorID,
	})
	if err != nil {
		t.Fatalf("GetSensorHistoricalData() unexpected error: %v", err)
	}
	if len(samples) != 0 {
		t.Fatalf("expected empty result, got %d samples", len(samples))
	}
}

func TestGetSensorHistoricalData_UnauthorizedTenant_ReturnsError(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	db := setupTimescaleDB(t)
	tenantID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	otherTenantID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	useCase := setupHistoricalDataUseCase(t, db, &staticTenantPort{
		tenant: tenant.Tenant{
			Id:             tenantID,
			CanImpersonate: false,
		},
	})

	_, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &otherTenantID,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: tenantID,
		SensorId: uuid.New(),
	})
	if !errors.Is(err, identity.ErrUnauthorizedAccess) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

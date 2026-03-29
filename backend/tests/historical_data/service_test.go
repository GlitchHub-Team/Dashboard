package historical_data_test

import (
	"errors"
	"testing"
	"time"

	"backend/internal/historical_data"
	"backend/internal/shared/identity"
	"backend/internal/tenant"

	"github.com/google/uuid"
)

type fakeGetHistoricalDataPort struct {
	samples        []historical_data.HistoricalSample
	err            error
	called         bool
	gotTenantId    uuid.UUID
	gotSensorId    uuid.UUID
	gotFilterLimit int
}

func (f *fakeGetHistoricalDataPort) GetSensorHistoricalData(
	tenantId uuid.UUID,
	sensorId uuid.UUID,
	filter historical_data.HistoricalDataFilter,
) ([]historical_data.HistoricalSample, error) {
	f.called = true
	f.gotTenantId = tenantId
	f.gotSensorId = sensorId
	f.gotFilterLimit = filter.Limit
	return f.samples, f.err
}

type fakeGetTenantPort struct {
	tenant tenant.Tenant
	err    error
}

func (f *fakeGetTenantPort) GetTenant(uuid.UUID) (tenant.Tenant, error) {
	return f.tenant, f.err
}

func TestGetHistoricalDataService_GetSensorHistoricalData_Success(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()

	expectedSamples := []historical_data.HistoricalSample{
		{
			SensorId:  targetSensorId,
			TenantId:  targetTenantId,
			Profile:   "HeartRate",
			Timestamp: time.Now().UTC(),
		},
	}

	getHistoricalDataPort := &fakeGetHistoricalDataPort{samples: expectedSamples}
	getTenantPort := &fakeGetTenantPort{
		tenant: tenant.Tenant{
			Id:             targetTenantId,
			CanImpersonate: false,
		},
	}

	service := historical_data.NewGetHistoricalDataService(getHistoricalDataPort, getTenantPort)

	samples, err := service.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &targetTenantId,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: targetTenantId,
		SensorId: targetSensorId,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !getHistoricalDataPort.called {
		t.Fatalf("expected port to be called")
	}

	if getHistoricalDataPort.gotTenantId != targetTenantId {
		t.Fatalf("unexpected tenant id: got %v want %v", getHistoricalDataPort.gotTenantId, targetTenantId)
	}

	if getHistoricalDataPort.gotSensorId != targetSensorId {
		t.Fatalf("unexpected sensor id: got %v want %v", getHistoricalDataPort.gotSensorId, targetSensorId)
	}

	if getHistoricalDataPort.gotFilterLimit != historical_data.DefaultHistoricalDataLimit {
		t.Fatalf(
			"unexpected filter limit: got %d want %d",
			getHistoricalDataPort.gotFilterLimit,
			historical_data.DefaultHistoricalDataLimit,
		)
	}

	if len(samples) != 1 {
		t.Fatalf("unexpected sample count: got %d want 1", len(samples))
	}
}

func TestGetHistoricalDataService_GetSensorHistoricalData_Unauthorized(t *testing.T) {
	targetTenantId := uuid.New()
	otherTenantId := uuid.New()

	getHistoricalDataPort := &fakeGetHistoricalDataPort{}
	getTenantPort := &fakeGetTenantPort{
		tenant: tenant.Tenant{
			Id:             targetTenantId,
			CanImpersonate: false,
		},
	}

	service := historical_data.NewGetHistoricalDataService(getHistoricalDataPort, getTenantPort)

	_, err := service.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &otherTenantId,
			RequesterRole:     identity.ROLE_TENANT_USER,
		},
		TenantId: targetTenantId,
		SensorId: uuid.New(),
	})
	if !errors.Is(err, identity.ErrUnauthorizedAccess) {
		t.Fatalf("unexpected error: got %v want %v", err, identity.ErrUnauthorizedAccess)
	}

	if getHistoricalDataPort.called {
		t.Fatalf("expected port not to be called")
	}
}

func TestGetHistoricalDataService_GetSensorHistoricalData_InvalidDateRange(t *testing.T) {
	targetTenantId := uuid.New()

	getHistoricalDataPort := &fakeGetHistoricalDataPort{}
	getTenantPort := &fakeGetTenantPort{}

	service := historical_data.NewGetHistoricalDataService(getHistoricalDataPort, getTenantPort)

	from := time.Now().UTC()
	to := from.Add(-time.Hour)

	_, err := service.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &targetTenantId,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: targetTenantId,
		SensorId: uuid.New(),
		From:     &from,
		To:       &to,
	})
	if !errors.Is(err, historical_data.ErrInvalidDateRange) {
		t.Fatalf("unexpected error: got %v want %v", err, historical_data.ErrInvalidDateRange)
	}

	if getHistoricalDataPort.called {
		t.Fatalf("expected port not to be called")
	}
}

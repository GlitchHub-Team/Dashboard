package historical_data_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"backend/internal/historical_data"
	sensordb "backend/internal/infra/database/sensor_db/connection"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	tenantmocks "backend/tests/tenant/mocks"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	stubDriverOnce sync.Once
	activeRecorder *queryRecorder
)

type queryRecorder struct {
	query string
	args  []driver.NamedValue
	rows  driver.Rows
	err   error
}

type stubDriver struct{}

type stubConn struct{}

type stubRows struct {
	columns []string
	values  [][]driver.Value
	index   int
	nextErr error
}

func registerStubDriver() {
	stubDriverOnce.Do(func() {
		sql.Register("historical_data_stub", stubDriver{})
	})
}

func (stubDriver) Open(string) (driver.Conn, error) {
	return stubConn{}, nil
}

func (stubConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (stubConn) Close() error                        { return nil }
func (stubConn) Begin() (driver.Tx, error)           { return nil, errors.New("not implemented") }

func (stubConn) QueryContext(
	_ context.Context,
	query string,
	args []driver.NamedValue,
) (driver.Rows, error) {
	activeRecorder.query = query
	activeRecorder.args = args
	if activeRecorder.err != nil {
		return nil, activeRecorder.err
	}
	return activeRecorder.rows, nil
}

func (r *stubRows) Columns() []string { return r.columns }

func (r *stubRows) Close() error { return nil }

func (r *stubRows) Next(dest []driver.Value) error {
	if r.index < len(r.values) {
		copy(dest, r.values[r.index])
		r.index++
		return nil
	}
	if r.nextErr != nil {
		return r.nextErr
	}
	return io.EOF
}

func newTestDB(t *testing.T, recorder *queryRecorder) *sql.DB {
	t.Helper()
	registerStubDriver()
	activeRecorder = recorder
	db, err := sql.Open("historical_data_stub", "")
	if err != nil {
		t.Fatalf("failed to open stub db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func newTestSensorDBConnection(t *testing.T, recorder *queryRecorder) sensordb.SensorDBConnection {
	t.Helper()

	sqlDB := newTestDB(t, recorder)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return sensordb.SensorDBConnection(gormDB)
}

func buildUseCaseWithModule(
	t *testing.T,
	db sensordb.SensorDBConnection,
	getTenantPort tenant.GetTenantPort,
) historical_data.GetSensorHistoricalDataUseCase {
	t.Helper()

	var useCase historical_data.GetSensorHistoricalDataUseCase
	app := fx.New(
		historical_data.Module,
		fx.Provide(
			func() sensordb.SensorDBConnection { return db },
			func() tenant.GetTenantPort { return getTenantPort },
			func() *zap.Logger { return zap.NewNop() },
		),
		fx.Populate(&useCase),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("failed to build fx app: %v", err)
	}
	return useCase
}

func TestModule_GetSensorHistoricalData_Success(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()
	targetGatewayId := uuid.New()
	from := time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)

	db := newTestSensorDBConnection(t, &queryRecorder{
		rows: &stubRows{
			columns: []string{"sensor_id", "gateway_id", "tenant_id", "profile", "timestamp", "data"},
			values: [][]driver.Value{
				{
					targetSensorId.String(),
					targetGatewayId.String(),
					targetTenantId.String(),
					"HeartRate",
					from,
					[]byte(`{"value":72}`),
				},
			},
		},
	})

	ctrl := gomock.NewController(t)
	getTenantPort := tenantmocks.NewMockGetTenantPort(ctrl)
	getTenantPort.EXPECT().
		GetTenant(targetTenantId).
		Return(tenant.Tenant{
			Id:             targetTenantId,
			CanImpersonate: false,
		}, nil).
		Times(1)

	useCase := buildUseCaseWithModule(t, db, getTenantPort)

	samples, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &targetTenantId,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: targetTenantId,
		SensorId: targetSensorId,
		From:     &from,
		To:       &to,
		Limit:    10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(samples) != 1 {
		t.Fatalf("expected 1 sample, got %d", len(samples))
	}
	if samples[0].GatewayId != targetGatewayId {
		t.Fatalf("unexpected gateway id: %v", samples[0].GatewayId)
	}
	if !strings.Contains(activeRecorder.query, targetTenantId.String()) {
		t.Fatalf("expected tenant schema in query, got %q", activeRecorder.query)
	}
	if !strings.Contains(activeRecorder.query, "AND timestamp >= $3") {
		t.Fatalf("expected from filter in query, got %q", activeRecorder.query)
	}
	if !strings.Contains(activeRecorder.query, "AND timestamp <= $4") {
		t.Fatalf("expected to filter in query, got %q", activeRecorder.query)
	}
	if !strings.Contains(activeRecorder.query, "LIMIT $5") {
		t.Fatalf("expected limit in query, got %q", activeRecorder.query)
	}
	if len(activeRecorder.args) != 5 {
		t.Fatalf("expected 5 query args, got %d", len(activeRecorder.args))
	}
}

func TestModule_GetSensorHistoricalData_QueryError(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()

	db := newTestSensorDBConnection(t, &queryRecorder{err: errors.New("query failed")})

	ctrl := gomock.NewController(t)
	getTenantPort := tenantmocks.NewMockGetTenantPort(ctrl)
	getTenantPort.EXPECT().
		GetTenant(targetTenantId).
		Return(tenant.Tenant{
			Id:             targetTenantId,
			CanImpersonate: false,
		}, nil).
		Times(1)

	useCase := buildUseCaseWithModule(t, db, getTenantPort)

	_, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &targetTenantId,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: targetTenantId,
		SensorId: targetSensorId,
	})
	if err == nil || err.Error() != "query failed" {
		t.Fatalf("expected query error, got %v", err)
	}
}

func TestModule_GetSensorHistoricalData_ScanError(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()

	db := newTestSensorDBConnection(t, &queryRecorder{
		rows: &stubRows{
			columns: []string{"sensor_id", "gateway_id", "tenant_id", "profile", "timestamp", "data"},
			values: [][]driver.Value{
				{"bad-uuid", uuid.New().String(), targetTenantId.String(), "HeartRate", time.Now(), []byte(`{"value":1}`)},
			},
		},
	})

	ctrl := gomock.NewController(t)
	getTenantPort := tenantmocks.NewMockGetTenantPort(ctrl)
	getTenantPort.EXPECT().
		GetTenant(targetTenantId).
		Return(tenant.Tenant{
			Id:             targetTenantId,
			CanImpersonate: false,
		}, nil).
		Times(1)

	useCase := buildUseCaseWithModule(t, db, getTenantPort)

	_, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &targetTenantId,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: targetTenantId,
		SensorId: targetSensorId,
	})
	if err == nil {
		t.Fatalf("expected scan error")
	}
}

func TestModule_GetSensorHistoricalData_RowsError(t *testing.T) {
	targetTenantId := uuid.New()
	targetSensorId := uuid.New()

	db := newTestSensorDBConnection(t, &queryRecorder{
		rows: &stubRows{
			columns: []string{"sensor_id", "gateway_id", "tenant_id", "profile", "timestamp", "data"},
			nextErr: errors.New("rows error"),
		},
	})

	ctrl := gomock.NewController(t)
	getTenantPort := tenantmocks.NewMockGetTenantPort(ctrl)
	getTenantPort.EXPECT().
		GetTenant(targetTenantId).
		Return(tenant.Tenant{
			Id:             targetTenantId,
			CanImpersonate: false,
		}, nil).
		Times(1)

	useCase := buildUseCaseWithModule(t, db, getTenantPort)

	_, err := useCase.GetSensorHistoricalData(historical_data.GetSensorHistoricalDataCommand{
		Requester: identity.Requester{
			RequesterTenantId: &targetTenantId,
			RequesterRole:     identity.ROLE_TENANT_ADMIN,
		},
		TenantId: targetTenantId,
		SensorId: targetSensorId,
	})
	if err == nil || err.Error() != "rows error" {
		t.Fatalf("expected rows error, got %v", err)
	}
}

package integrationtests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"backend/internal/historical_data"
	"backend/internal/tenant"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type staticTenantPort struct {
	tenant tenant.Tenant
	err    error
}

func (s *staticTenantPort) GetTenant(uuid.UUID) (tenant.Tenant, error) {
	return s.tenant, s.err
}

func envInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return fallback
}

func getEnvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func setupTimescaleDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		getEnvOrDefault("POSTGRES_HOST", "timescale"),
		envInt("POSTGRES_PORT", 5432),
		getEnvOrDefault("POSTGRES_USER", "admin"),
		getEnvOrDefault("POSTGRES_PASSWORD", "admin"),
		getEnvOrDefault("POSTGRES_DB", "sensor_db"),
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Skipf("Timescale non raggiungibile (%v), integrazione saltata", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func setupHistoricalDataUseCase(
	t *testing.T,
	db *sql.DB,
	tenantPort tenant.GetTenantPort,
) historical_data.GetSensorHistoricalDataUseCase {
	t.Helper()

	var useCase historical_data.GetSensorHistoricalDataUseCase

	app := fx.New(
		historical_data.Module,
		fx.Provide(
			func() *sql.DB { return db },
			func() tenant.GetTenantPort { return tenantPort },
			func() *zap.Logger { return zap.NewNop() },
		),
		fx.Populate(&useCase),
	)
	if err := app.Err(); err != nil {
		t.Fatalf("fx.New() unexpected error: %v", err)
	}

	return useCase
}

func insertSensorDataRow(
	t *testing.T,
	db *sql.DB,
	tenantID, sensorID, gatewayID uuid.UUID,
	ts time.Time,
	profile string,
	payload []byte,
) {
	t.Helper()

	query := fmt.Sprintf(
		`INSERT INTO "%s".sensor_data (sensor_id, gateway_id, timestamp, tenant_id, profile, data) VALUES ($1,$2,$3,$4,$5,$6)`,
		tenantID.String(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, query, sensorID, gatewayID, ts, tenantID, profile, payload); err != nil {
		t.Fatalf("insertSensorDataRow() unexpected error: %v", err)
	}
}

func cleanupSensorDataRow(
	t *testing.T,
	db *sql.DB,
	tenantID, sensorID, gatewayID uuid.UUID,
	ts time.Time,
) {
	t.Helper()

	query := fmt.Sprintf(
		`DELETE FROM "%s".sensor_data WHERE sensor_id=$1 AND gateway_id=$2 AND timestamp=$3`,
		tenantID.String(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, query, sensorID, gatewayID, ts); err != nil {
		t.Fatalf("cleanupSensorDataRow() unexpected error: %v", err)
	}
}

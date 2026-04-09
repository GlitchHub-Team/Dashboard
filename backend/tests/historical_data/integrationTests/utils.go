package integrationtests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	transportDto "backend/internal/infra/transport/http/dto"
	sensorProfile "backend/internal/sensor/profile"
	"backend/internal/tenant"
	"backend/tests/helper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type historicalDataResponse struct {
	Count   uint                           `json:"count"`
	Samples []historicalDataSampleResponse `json:"samples"`
}

type historicalDataSampleResponse struct {
	SensorID  string          `json:"sensor_id"`
	GatewayID string          `json:"gateway_id"`
	TenantID  string          `json:"tenant_id"`
	Profile   string          `json:"profile"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type historicalDataExpectedSample struct {
	SensorID   uuid.UUID
	GatewayID  uuid.UUID
	TenantID   uuid.UUID
	Profile    string
	Timestamp  time.Time
	HeartRate  int
	ExpectData bool
}

func authHeader(jwt string) http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+jwt)
	return header
}

func historicalDataPath(tenantID uuid.UUID, sensorID uuid.UUID) string {
	return "/api/v1/tenant/" + tenantID.String() + "/sensor/" + sensorID.String() + "/historical_data"
}

func preSetupInsertSensorDataRow(
	tenantID, sensorID, gatewayID uuid.UUID,
	ts time.Time,
	profile string,
	payload []byte,
) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		sqlDB, err := sensorSQLDB(deps)
		if err != nil {
			return false
		}

		query := fmt.Sprintf(
			`INSERT INTO "%s".sensor_data (sensor_id, gateway_id, timestamp, tenant_id, profile, data) VALUES ($1,$2,$3,$4,$5,$6)`,
			tenantID.String(),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, err = sqlDB.ExecContext(ctx, query, sensorID, gatewayID, ts, tenantID, profile, payload)
		return err == nil
	}
}

func postSetupDeleteSensorDataRow(
	tenantID, sensorID, gatewayID uuid.UUID,
	ts time.Time,
) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		sqlDB, err := sensorSQLDB(deps)
		if err != nil {
			return
		}

		query := fmt.Sprintf(
			`DELETE FROM "%s".sensor_data WHERE sensor_id=$1 AND gateway_id=$2 AND timestamp=$3`,
			tenantID.String(),
		)

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		_, _ = sqlDB.ExecContext(ctx, query, sensorID, gatewayID, ts)
	}
}

func checkHistoricalDataEmptyResponse() helper.IntegrationTestCheck {
	return func(r *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var response historicalDataResponse
		if err := json.Unmarshal(r.Body.Bytes(), &response); err != nil {
			return false
		}

		return response.Count == 0 && len(response.Samples) == 0
	}
}

func checkHistoricalDataResponse(
	expectedCount uint,
	expected historicalDataExpectedSample,
) helper.IntegrationTestCheck {
	return func(r *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var response historicalDataResponse
		if err := json.Unmarshal(r.Body.Bytes(), &response); err != nil {
			return false
		}

		if response.Count != expectedCount || len(response.Samples) != int(expectedCount) {
			return false
		}

		sample := response.Samples[0]
		if sample.SensorID != expected.SensorID.String() {
			return false
		}
		if sample.GatewayID != expected.GatewayID.String() {
			return false
		}
		if sample.TenantID != expected.TenantID.String() {
			return false
		}
		if sample.Profile != expected.Profile {
			return false
		}
		if !sample.Timestamp.Equal(expected.Timestamp) {
			return false
		}

		if !expected.ExpectData {
			return len(sample.Data) == 0
		}

		switch expected.Profile {
		case string(sensorProfile.HEART_RATE):
			var decoded transportDto.HeartRateData
			if err := json.Unmarshal(sample.Data, &decoded); err != nil {
				return false
			}
			return decoded.BpmValue == expected.HeartRate
		default:
			return false
		}
	}
}

func sensorSQLDB(deps helper.IntegrationTestDeps) (*sql.DB, error) {
	db := (*gorm.DB)(deps.SensorDB)
	return db.DB()
}

func setupHistoricalDataTenantTestContext(
	cloudDB *gorm.DB,
	sensorDB *gorm.DB,
	tenantID uuid.UUID,
	canImpersonate bool,
) error {
	tenantEntity := tenant.TenantEntity{
		ID:             tenantID.String(),
		Name:           "Historical Data Test Tenant",
		CanImpersonate: canImpersonate,
	}

	if err := cloudDB.Create(&tenantEntity).Error; err != nil {
		return fmt.Errorf("failed to create tenant %v: %v", tenantID, err)
	}

	if err := sensorDB.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`, tenantID.String())).Error; err != nil {
		return fmt.Errorf("failed to create tenant schema %v: %v", tenantID, err)
	}

	if err := sensorDB.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS "%s".sensor_data (
			sensor_id UUID NOT NULL,
			gateway_id UUID NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL,
			tenant_id UUID NOT NULL,
			profile VARCHAR(255) NOT NULL,
			data JSONB NOT NULL,
			PRIMARY KEY (sensor_id, gateway_id, timestamp)
		)
	`, tenantID.String())).Error; err != nil {
		return fmt.Errorf("failed to create sensor_data for tenant %v: %v", tenantID, err)
	}

	return nil
}

func cleanupHistoricalDataTenantTestContext(
	cloudDB *gorm.DB,
	sensorDB *gorm.DB,
	tenantID uuid.UUID,
) {
	_ = sensorDB.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS "%s" CASCADE`, tenantID.String())).Error
	_ = cloudDB.Where("id = ?", tenantID.String()).Delete(&tenant.TenantEntity{}).Error
}

package integrationtests

import (
	"net/http"
	"testing"
	"time"

	"backend/internal/historical_data"
	sensorProfile "backend/internal/sensor/profile"
	"backend/internal/tenant"
	"backend/tests/helper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestGetSensorHistoricalDataIntegration(t *testing.T) {
	// Questo package ha una profondita diversa dagli altri integration test,
	// quindi i path delle credenziali condivise vanno riallineati esplicitamente.

	t.Setenv("DASHBOARD_CREDS_PATH", "dashboard.creds")
	t.Setenv("TEST_CREDS_PATH", "admin_test.creds")
	t.Setenv("CA_PEM_PATH", "ca.pem")

	deps := helper.SetupIntegrationTest(t)
	tenantIDOne := uuid.New()
	tenantIDTwo := uuid.New()

	if err := setupHistoricalDataTenantTestContext((*gorm.DB)(deps.CloudDB), (*gorm.DB)(deps.SensorDB), tenantIDOne, true); err != nil {
		t.Fatalf("failed to setup tenant one context: %v", err)
	}
	t.Cleanup(func() {
		cleanupHistoricalDataTenantTestContext((*gorm.DB)(deps.CloudDB), (*gorm.DB)(deps.SensorDB), tenantIDOne)
	})

	if err := setupHistoricalDataTenantTestContext((*gorm.DB)(deps.CloudDB), (*gorm.DB)(deps.SensorDB), tenantIDTwo, false); err != nil {
		t.Fatalf("failed to setup tenant two context: %v", err)
	}
	t.Cleanup(func() {
		cleanupHistoricalDataTenantTestContext((*gorm.DB)(deps.CloudDB), (*gorm.DB)(deps.SensorDB), tenantIDTwo)
	})

	tenantAdminTenantOneJWT, err := helper.NewTenantAdminJWT(deps, tenantIDOne, 999)
	if err != nil {
		t.Fatalf("failed to generate tenant admin jwt: %v", err)
	}

	tenantAdminTenantTwoJWT, err := helper.NewTenantAdminJWT(deps, tenantIDTwo, 1000)
	if err != nil {
		t.Fatalf("failed to generate tenant admin jwt: %v", err)
	}

	sensorIDValid := uuid.New()
	gatewayIDValid := uuid.New()
	tsValid := time.Now().UTC().Truncate(time.Microsecond)

	sensorIDRange := uuid.New()
	gatewayIDRange := uuid.New()
	tsRangeOne := tsValid.Add(1 * time.Minute)
	tsRangeTwo := tsValid.Add(3 * time.Minute)

	sensorIDNoData := uuid.New()
	sensorIDUnauthorized := uuid.New()

	from := tsRangeTwo.Add(-time.Second).Format(time.RFC3339)
	to := tsRangeTwo.Add(time.Second).Format(time.RFC3339)

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupInsertSensorDataRow(
					tenantIDOne,
					sensorIDValid,
					gatewayIDValid,
					tsValid,
					string(sensorProfile.HEART_RATE),
					[]byte(`{"BpmValue":72}`),
				),
			},
			Name:   "Richiesta dati storici valida",
			Method: http.MethodGet,
			Path:   historicalDataPath(tenantIDOne, sensorIDValid),
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: `"samples"`,
			ResponseChecks: []helper.IntegrationTestCheck{
				checkHistoricalDataResponse(
					1,
					historicalDataExpectedSample{
						SensorID:   sensorIDValid,
						GatewayID:  gatewayIDValid,
						TenantID:   tenantIDOne,
						Profile:    string(sensorProfile.HEART_RATE),
						Timestamp:  tsValid,
						HeartRate:  72,
						ExpectData: true,
					},
				),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensorDataRow(tenantIDOne, sensorIDValid, gatewayIDValid, tsValid),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupInsertSensorDataRow(
					tenantIDOne,
					sensorIDRange,
					gatewayIDRange,
					tsRangeOne,
					string(sensorProfile.HEART_RATE),
					[]byte(`{"BpmValue":70}`),
				),
				preSetupInsertSensorDataRow(
					tenantIDOne,
					sensorIDRange,
					gatewayIDRange,
					tsRangeTwo,
					string(sensorProfile.HEART_RATE),
					[]byte(`{"BpmValue":75}`),
				),
			},
			Name:   "Filtro per intervallo temporale",
			Method: http.MethodGet,
			Path:   historicalDataPath(tenantIDOne, sensorIDRange) + "?from=" + from + "&to=" + to,
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: `"count":1`,
			ResponseChecks: []helper.IntegrationTestCheck{
				checkHistoricalDataResponse(
					1,
					historicalDataExpectedSample{
						SensorID:   sensorIDRange,
						GatewayID:  gatewayIDRange,
						TenantID:   tenantIDOne,
						Profile:    string(sensorProfile.HEART_RATE),
						Timestamp:  tsRangeTwo,
						HeartRate:  75,
						ExpectData: true,
					},
				),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensorDataRow(tenantIDOne, sensorIDRange, gatewayIDRange, tsRangeTwo),
				postSetupDeleteSensorDataRow(tenantIDOne, sensorIDRange, gatewayIDRange, tsRangeOne),
			},
		},
		{
			PreSetups: nil,
			Name:      "Nessun dato disponibile",
			Method:    http.MethodGet,
			Path:      historicalDataPath(tenantIDOne, sensorIDNoData),
			Header:    authHeader(tenantAdminTenantOneJWT),
			Body:      nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: `"count":0`,
			ResponseChecks: []helper.IntegrationTestCheck{
				checkHistoricalDataEmptyResponse(),
			},

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Tenant non autorizzato",
			Method:    http.MethodGet,
			Path:      historicalDataPath(tenantIDOne, sensorIDUnauthorized),
			Header:    authHeader(tenantAdminTenantTwoJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Timestamp query invalido",
			Method:    http.MethodGet,
			Path:      historicalDataPath(tenantIDOne, sensorIDNoData) + "?from=not-a-timestamp",
			Header:    authHeader(tenantAdminTenantOneJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: helper.ErrJsonString(historical_data.ErrInvalidTimestamp),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

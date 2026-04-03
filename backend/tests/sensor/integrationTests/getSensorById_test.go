package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	clouddb "backend/internal/infra/database/cloud_db/connection"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/helper"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type getSensorResponse struct {
	GatewayID      string `json:"gateway_id"`
	Profile        string `json:"profile"`
	SensorID       string `json:"sensor_id"`
	DataInterval   int64  `json:"data_interval"`
	SensorInterval int64  `json:"sensor_interval"`
	SensorName     string `json:"sensor_name"`
	Status         string `json:"status"`
}

func TestGetSensorByIdIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT := mustGenerateJWTForRequester(t, deps.AuthTokenManager, identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	})

	tenantIDs := mustLoadAtLeastTwoTenantIDs(t, deps.CloudDB)
	tenantIDOne := tenantIDs[0]
	tenantIDTwo := tenantIDs[1]

	tenantAdminTenantOneJWT := mustGenerateJWTForRequester(t, deps.AuthTokenManager, identity.Requester{
		RequesterUserId:   999,
		RequesterTenantId: &tenantIDOne,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	})

	tenantAdminTenantTwoJWT := mustGenerateJWTForRequester(t, deps.AuthTokenManager, identity.Requester{
		RequesterUserId:   1000,
		RequesterTenantId: &tenantIDTwo,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	})

	gatewayNilTenantUnauthorized := uuid.NewString()
	sensorNilTenantUnauthorized := uuid.NewString()

	gatewayNilTenantSuperAdmin := uuid.NewString()
	sensorNilTenantSuperAdmin := uuid.NewString()

	gatewayTenantMismatch := uuid.NewString()
	sensorTenantMismatch := uuid.NewString()

	gatewayTenantMatch := uuid.NewString()
	sensorTenantMatch := uuid.NewString()

	tenantIDOneString := tenantIDOne.String()

	tests := []helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con jwt invalido",
			Method:    http.MethodGet,
			Path:      "/api/v1/sensor/" + uuid.NewString(),
			Header:    authHeader("invalid.jwt.token"),
			Body:      nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con sensorId invalido",
			Method:    http.MethodGet,
			Path:      "/api/v1/sensor/not-a-uuid",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: sensor.ErrInvalidSensorID.Error(),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Richiesta di un sensore che non esiste",
			Method:    http.MethodGet,
			Path:      "/api/v1/sensor/" + uuid.NewString(),
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayNilTenantUnauthorized, "Gateway Nil Tenant Unauthorized", nil),
				preSetupCreateSensor(sensorNilTenantUnauthorized, gatewayNilTenantUnauthorized, "Sensor Nil Tenant Unauthorized", 1400, sensor.HEART_RATE, sensor.Active),
			},
			Name:   "Richiesta di un sensore con gatewayId nil da parte di un utente non super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/sensor/" + sensorNilTenantUnauthorized,
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorExists(sensorNilTenantUnauthorized),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorNilTenantUnauthorized),
				postSetupDeleteByGateway(gatewayNilTenantUnauthorized),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayNilTenantSuperAdmin, "Gateway Nil Tenant Super Admin", nil),
				preSetupCreateSensor(sensorNilTenantSuperAdmin, gatewayNilTenantSuperAdmin, "Sensor Nil Tenant Super Admin", 1500, sensor.PULSE_OXIMETER, sensor.Active),
			},
			Name:   "Richiesta di un sensore con gatewayId nil da parte di un utente super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/sensor/" + sensorNilTenantSuperAdmin,
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensor_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetSensorResponseMatchesExpected(sensorNilTenantSuperAdmin, gatewayNilTenantSuperAdmin, "Sensor Nil Tenant Super Admin", 1500, sensor.PULSE_OXIMETER, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorNilTenantSuperAdmin),
				postSetupDeleteByGateway(gatewayNilTenantSuperAdmin),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayTenantMismatch, "Gateway Tenant Mismatch", &tenantIDOneString),
				preSetupCreateSensor(sensorTenantMismatch, gatewayTenantMismatch, "Sensor Tenant Mismatch", 1600, sensor.HEALTH_THERMOMETER, sensor.Active),
			},
			Name:   "Richiesta di un sensore con gatewayId non nil da parte di un utente non super admin con tenant diverso",
			Method: http.MethodGet,
			Path:   "/api/v1/sensor/" + sensorTenantMismatch,
			Header: authHeader(tenantAdminTenantTwoJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorExists(sensorTenantMismatch),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTenantMismatch),
				postSetupDeleteByGateway(gatewayTenantMismatch),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayTenantMatch, "Gateway Tenant Match", &tenantIDOneString),
				preSetupCreateSensor(sensorTenantMatch, gatewayTenantMatch, "Sensor Tenant Match", 1700, sensor.ENVIRONMENTAL_SENSING, sensor.Active),
			},
			Name:   "Richiesta di un sensore con gatewayId non nil da parte di un utente non super admin con tenant uguale",
			Method: http.MethodGet,
			Path:   "/api/v1/sensor/" + sensorTenantMatch,
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensor_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetSensorResponseMatchesExpected(sensorTenantMatch, gatewayTenantMatch, "Sensor Tenant Match", 1700, sensor.ENVIRONMENTAL_SENSING, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTenantMatch),
				postSetupDeleteByGateway(gatewayTenantMatch),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func mustLoadAtLeastTwoTenantIDs(t *testing.T, CloudDB clouddb.CloudDBConnection) []uuid.UUID {
	t.Helper()

	db := (*gorm.DB)(CloudDB)
	var tenantEntities []tenant.TenantEntity
	if err := db.Model(&tenant.TenantEntity{}).Order("id ASC").Limit(2).Find(&tenantEntities).Error; err != nil {
		t.Fatalf("failed to load tenants for integration setup: %v", err)
	}
	if len(tenantEntities) < 2 {
		t.Fatalf("expected at least 2 tenants in DB, got %d", len(tenantEntities))
	}

	tenantIDs := make([]uuid.UUID, 0, len(tenantEntities))
	for _, entity := range tenantEntities {
		tenantID, err := uuid.Parse(entity.ID)
		if err != nil {
			t.Fatalf("invalid tenant id in DB: %v", err)
		}
		tenantIDs = append(tenantIDs, tenantID)
	}

	return tenantIDs
}

func checkGetSensorResponseMatchesExpected(
	expectedSensorID string,
	expectedGatewayID string,
	expectedName string,
	expectedInterval int64,
	expectedProfile sensor.SensorProfile,
	expectedStatus sensor.SensorStatus,
) helper.IntegrationTestCheck {
	return func(w *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var resp getSensorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return false
		}

		interval := resp.SensorInterval
		if interval == 0 {
			interval = resp.DataInterval
		}

		if resp.SensorID != expectedSensorID {
			return false
		}
		if resp.GatewayID != expectedGatewayID {
			return false
		}
		if resp.SensorName != expectedName {
			return false
		}
		if interval != expectedInterval {
			return false
		}
		if resp.Profile != string(expectedProfile) {
			return false
		}
		if resp.Status != string(expectedStatus) {
			return false
		}

		return true
	}
}

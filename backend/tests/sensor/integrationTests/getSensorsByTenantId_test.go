package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/helper"

	"github.com/google/uuid"
)

type expectedTenantSensor struct {
	SensorID  string
	GatewayID string
	Name      string
	Interval  int64
	Profile   sensor.SensorProfile
	Status    sensor.SensorStatus
}

func TestGetSensorsByTenantIdIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT := mustGenerateJWTForRequester(t, deps.AuthTokenManager, identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	})

	tenantIDOne := uuid.MustParse(tenant1IdStr)
	tenantIDTwo := uuid.MustParse(tenant2IdStr)

	err := populateTenantDefaultData(deps.CloudDB)
	if err != nil {
		t.Fatalf("Impossibile popolare DB con dati di default: %v", err)
	}

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

	tenantOneString := tenantIDOne.String()
	tenantTwoString := tenantIDTwo.String()

	nonExistingTenantID := uuid.NewString()

	gatewayTenantOneA := uuid.NewString()
	gatewayTenantOneB := uuid.NewString()
	sensorTenantOneA := uuid.NewString()
	sensorTenantOneB := uuid.NewString()
	sensorTenantOneC := uuid.NewString()

	gatewayTenantTwo := uuid.NewString()
	sensorTenantTwo := uuid.NewString()

	expectedTenantOneSensors := []expectedTenantSensor{
		{SensorID: sensorTenantOneA, GatewayID: gatewayTenantOneA, Name: "Alpha Tenant One", Interval: 1000, Profile: sensor.HEART_RATE, Status: sensor.Active},
		{SensorID: sensorTenantOneB, GatewayID: gatewayTenantOneA, Name: "Beta Tenant One", Interval: 1100, Profile: sensor.ECG_CUSTOM, Status: sensor.Active},
		{SensorID: sensorTenantOneC, GatewayID: gatewayTenantOneB, Name: "Gamma Tenant One", Interval: 1200, Profile: sensor.PULSE_OXIMETER, Status: sensor.Inactive},
	}

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con jwt invalido",
			Method:    http.MethodGet,
			Path:      "/api/v1/tenant/" + tenantOneString + "/sensors?page=1&limit=10",
			Header:    authHeader("invalid.jwt.token"),
			Body:      nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con tenant_id invalido",
			Method:    http.MethodGet,
			Path:      "/api/v1/tenant/not-a-uuid/sensors?page=1&limit=10",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: tenant.ErrInvalidTenantID.Error(),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con i query params invalidi",
			Method:    http.MethodGet,
			Path:      "/api/v1/tenant/" + tenantOneString + "/sensors?page=0&limit=5",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid format",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con tenant non esistente",
			Method:    http.MethodGet,
			Path:      "/api/v1/tenant/" + nonExistingTenantID + "/sensors?page=1&limit=10",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensors\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetSensorsByTenantResponse(nil, 1, 10),
			},

			PostSetups: nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayTenantOneA, "Gateway Tenant One A", &tenantOneString),
				preSetupCreateGatewayWithTenant(gatewayTenantOneB, "Gateway Tenant One B", &tenantOneString),
				preSetupCreateGatewayWithTenant(gatewayTenantTwo, "Gateway Tenant Two", &tenantTwoString),
				preSetupCreateSensor(sensorTenantOneA, gatewayTenantOneA, "Alpha Tenant One", 1000, sensor.HEART_RATE, sensor.Active),
				preSetupCreateSensor(sensorTenantOneB, gatewayTenantOneA, "Beta Tenant One", 1100, sensor.ECG_CUSTOM, sensor.Active),
				preSetupCreateSensor(sensorTenantOneC, gatewayTenantOneB, "Gamma Tenant One", 1200, sensor.PULSE_OXIMETER, sensor.Inactive),
				preSetupCreateSensor(sensorTenantTwo, gatewayTenantTwo, "Tenant Two Sensor", 1300, sensor.HEALTH_THERMOMETER, sensor.Active),
			},
			Name:   "Richiesta di sensori da parte di un super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantOneString + "/sensors?page=1&limit=10",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensors\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetSensorsByTenantResponse(expectedTenantOneSensors, 1, 10),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTenantOneA),
				postSetupDeleteSensor(sensorTenantOneB),
				postSetupDeleteSensor(sensorTenantOneC),
				postSetupDeleteSensor(sensorTenantTwo),
				postSetupDeleteByGateway(gatewayTenantOneA),
				postSetupDeleteByGateway(gatewayTenantOneB),
				postSetupDeleteByGateway(gatewayTenantTwo),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayTenantOneA, "Gateway Tenant One A", &tenantOneString),
				preSetupCreateGatewayWithTenant(gatewayTenantOneB, "Gateway Tenant One B", &tenantOneString),
				preSetupCreateSensor(sensorTenantOneA, gatewayTenantOneA, "Alpha Tenant One", 1000, sensor.HEART_RATE, sensor.Active),
				preSetupCreateSensor(sensorTenantOneB, gatewayTenantOneA, "Beta Tenant One", 1100, sensor.ECG_CUSTOM, sensor.Active),
				preSetupCreateSensor(sensorTenantOneC, gatewayTenantOneB, "Gamma Tenant One", 1200, sensor.PULSE_OXIMETER, sensor.Inactive),
			},
			Name:   "Richiesta di sensori da parte di un utente non super admin con tenant diverso",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantOneString + "/sensors?page=1&limit=10",
			Header: authHeader(tenantAdminTenantTwoJWT),
			Body:   nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: identity.ErrUnauthorizedAccess.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorExists(sensorTenantOneA),
				checkSensorExists(sensorTenantOneB),
				checkSensorExists(sensorTenantOneC),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTenantOneA),
				postSetupDeleteSensor(sensorTenantOneB),
				postSetupDeleteSensor(sensorTenantOneC),
				postSetupDeleteByGateway(gatewayTenantOneA),
				postSetupDeleteByGateway(gatewayTenantOneB),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayTenantOneA, "Gateway Tenant One A", &tenantOneString),
				preSetupCreateGatewayWithTenant(gatewayTenantOneB, "Gateway Tenant One B", &tenantOneString),
				preSetupCreateSensor(sensorTenantOneA, gatewayTenantOneA, "Alpha Tenant One", 1000, sensor.HEART_RATE, sensor.Active),
				preSetupCreateSensor(sensorTenantOneB, gatewayTenantOneA, "Beta Tenant One", 1100, sensor.ECG_CUSTOM, sensor.Active),
				preSetupCreateSensor(sensorTenantOneC, gatewayTenantOneB, "Gamma Tenant One", 1200, sensor.PULSE_OXIMETER, sensor.Inactive),
			},
			Name:   "Richiesta di sensori da parte di un utente non super admin con tenant uguale",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantOneString + "/sensors?page=1&limit=10",
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensors\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetSensorsByTenantResponse(expectedTenantOneSensors, 1, 10),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTenantOneA),
				postSetupDeleteSensor(sensorTenantOneB),
				postSetupDeleteSensor(sensorTenantOneC),
				postSetupDeleteByGateway(gatewayTenantOneA),
				postSetupDeleteByGateway(gatewayTenantOneB),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func checkGetSensorsByTenantResponse(
	expectedAllSensors []expectedTenantSensor,
	page int,
	limit int,
) helper.IntegrationTestCheck {
	return func(w *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var resp getSensorsByGatewayResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return false
		}

		sorted := make([]expectedTenantSensor, len(expectedAllSensors))
		copy(sorted, expectedAllSensors)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})

		offset := (page - 1) * limit
		expectedPageSensors := []expectedTenantSensor{}
		if offset < len(sorted) {
			end := offset + limit
			if end > len(sorted) {
				end = len(sorted)
			}
			expectedPageSensors = sorted[offset:end]
		}

		if resp.Total != uint(len(sorted)) {
			return false
		}
		if resp.Count != uint(len(expectedPageSensors)) {
			return false
		}
		if len(resp.Sensors) != len(expectedPageSensors) {
			return false
		}

		for i := range expectedPageSensors {
			received := resp.Sensors[i]
			expected := expectedPageSensors[i]

			interval := received.SensorInterval
			if interval == 0 {
				interval = received.DataInterval
			}

			if received.SensorID != expected.SensorID {
				return false
			}
			if received.GatewayID != expected.GatewayID {
				return false
			}
			if received.SensorName != expected.Name {
				return false
			}
			if interval != expected.Interval {
				return false
			}
			if received.Profile != string(expected.Profile) {
				return false
			}
			if received.Status != string(expected.Status) {
				return false
			}
		}

		return true
	}
}

package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	clouddb "backend/internal/infra/database/cloud_db/connection"
	sensordb "backend/internal/infra/database/sensor_db"
	natsutils "backend/internal/infra/nats"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/tests/helper"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type getSensorsByGatewayResponse struct {
	Count   uint                    `json:"count"`
	Total   uint                    `json:"total"`
	Sensors []getSensorsByGatewayEl `json:"sensors"`
}

type getSensorsByGatewayEl struct {
	GatewayID      string `json:"gateway_id"`
	Profile        string `json:"profile"`
	SensorID       string `json:"sensor_id"`
	DataInterval   int64  `json:"data_interval"`
	SensorInterval int64  `json:"sensor_interval"`
	SensorName     string `json:"sensor_name"`
	Status         string `json:"status"`
}

type expectedGatewaySensor struct {
	SensorID string
	Name     string
	Interval int64
	Profile  sensor.SensorProfile
	Status   sensor.SensorStatus
}

func TestGetSensorsByGatewayIdIntegration(t *testing.T) {
	router, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx, jwtManager, ctx := helper.Setup(t)

	superAdminJWT := mustGenerateJWTForRequester(t, jwtManager, identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	})

	tenantIDs := mustLoadAtLeastTwoTenantIDs(t, cloudDB)
	tenantIDOne := tenantIDs[0]
	tenantIDTwo := tenantIDs[1]
	tenantIDOneString := tenantIDOne.String()

	tenantAdminTenantOneJWT := mustGenerateJWTForRequester(t, jwtManager, identity.Requester{
		RequesterUserId:   999,
		RequesterTenantId: &tenantIDOne,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	})

	tenantAdminTenantTwoJWT := mustGenerateJWTForRequester(t, jwtManager, identity.Requester{
		RequesterUserId:   1000,
		RequesterTenantId: &tenantIDTwo,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	})

	gatewayNilTenantUnauthorized := uuid.NewString()

	gatewayNilTenantSuperAdmin := uuid.NewString()
	sensorNilTenantSuperAdminA := uuid.NewString()
	sensorNilTenantSuperAdminB := uuid.NewString()
	sensorNilTenantSuperAdminC := uuid.NewString()

	gatewayTenantMismatch := uuid.NewString()
	sensorTenantMismatch := uuid.NewString()

	gatewayTenantMatch := uuid.NewString()
	sensorTenantMatchA := uuid.NewString()
	sensorTenantMatchB := uuid.NewString()

	expectedNilTenantSuperAdminSensors := []expectedGatewaySensor{
		{SensorID: sensorNilTenantSuperAdminA, Name: "Alpha Nil Tenant", Interval: 1000, Profile: sensor.HEART_RATE, Status: sensor.Active},
		{SensorID: sensorNilTenantSuperAdminB, Name: "Beta Nil Tenant", Interval: 1100, Profile: sensor.ECG_CUSTOM, Status: sensor.Active},
		{SensorID: sensorNilTenantSuperAdminC, Name: "Gamma Nil Tenant", Interval: 1200, Profile: sensor.PULSE_OXIMETER, Status: sensor.Inactive},
	}

	expectedTenantMatchSensors := []expectedGatewaySensor{
		{SensorID: sensorTenantMatchA, Name: "Alpha Tenant Match", Interval: 1300, Profile: sensor.HEALTH_THERMOMETER, Status: sensor.Active},
		{SensorID: sensorTenantMatchB, Name: "Beta Tenant Match", Interval: 1400, Profile: sensor.ENVIRONMENTAL_SENSING, Status: sensor.Active},
	}

	tests := []helper.TestCase{
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con jwt invalido",
			Method:    http.MethodGet,
			Path:      "/api/v1/gateway/" + uuid.NewString() + "/sensors?page=1&limit=10",
			Header:    authHeader("invalid.jwt.token"),
			Body:      nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con gateway_id invalido",
			Method:    http.MethodGet,
			Path:      "/api/v1/gateway/not-a-uuid/sensors?page=1&limit=10",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "ID gateway non valido",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con i query params invalidi",
			Method:    http.MethodGet,
			Path:      "/api/v1/gateway/" + uuid.NewString() + "/sensors?page=0&limit=5",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid format",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio della richiesta con gateway non esistente",
			Method:    http.MethodGet,
			Path:      "/api/v1/gateway/" + uuid.NewString() + "/sensors?page=1&limit=10",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: "gateway non trovato",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewayNilTenantUnauthorized, "Gateway Nil Tenant Unauthorized", nil),
			},
			Name:   "Richiesta dei sensori di un gateway con tenantId nil da parte di un utente non super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + gatewayNilTenantUnauthorized + "/sensors?page=1&limit=10",
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: "gateway non trovato",
			ResponseChecks:   nil,

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteByGateway(gatewayNilTenantUnauthorized),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewayNilTenantSuperAdmin, "Gateway Nil Tenant Super Admin", nil),
				preSetupCreateSensor(sensorNilTenantSuperAdminA, gatewayNilTenantSuperAdmin, "Alpha Nil Tenant", 1000, sensor.HEART_RATE, sensor.Active),
				preSetupCreateSensor(sensorNilTenantSuperAdminB, gatewayNilTenantSuperAdmin, "Beta Nil Tenant", 1100, sensor.ECG_CUSTOM, sensor.Active),
				preSetupCreateSensor(sensorNilTenantSuperAdminC, gatewayNilTenantSuperAdmin, "Gamma Nil Tenant", 1200, sensor.PULSE_OXIMETER, sensor.Inactive),
			},
			Name:   "Richiesta di un sensore con gateway e tenant id nil da parte di un utente super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + gatewayNilTenantSuperAdmin + "/sensors?page=1&limit=10",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensors\"",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkGetSensorsByGatewayResponse(gatewayNilTenantSuperAdmin, expectedNilTenantSuperAdminSensors, 1, 10),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorNilTenantSuperAdminA),
				postSetupDeleteSensor(sensorNilTenantSuperAdminB),
				postSetupDeleteSensor(sensorNilTenantSuperAdminC),
				postSetupDeleteByGateway(gatewayNilTenantSuperAdmin),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewayTenantMismatch, "Gateway Tenant Mismatch", &tenantIDOneString),
				preSetupCreateSensor(sensorTenantMismatch, gatewayTenantMismatch, "Mismatch Sensor", 1250, sensor.HEART_RATE, sensor.Active),
			},
			Name:   "Richiesta di sensori di un gateway non nil da parte di un utente non super admin con tenant diverso",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + gatewayTenantMismatch + "/sensors?page=1&limit=10",
			Header: authHeader(tenantAdminTenantTwoJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: "gateway non trovato",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorExists(sensorTenantMismatch),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorTenantMismatch),
				postSetupDeleteByGateway(gatewayTenantMismatch),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewayTenantMatch, "Gateway Tenant Match", &tenantIDOneString),
				preSetupCreateSensor(sensorTenantMatchA, gatewayTenantMatch, "Alpha Tenant Match", 1300, sensor.HEALTH_THERMOMETER, sensor.Active),
				preSetupCreateSensor(sensorTenantMatchB, gatewayTenantMatch, "Beta Tenant Match", 1400, sensor.ENVIRONMENTAL_SENSING, sensor.Active),
			},
			Name:   "Richiesta di sensori da parte di un utente non super admin con tenant uguale",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + gatewayTenantMatch + "/sensors?page=1&limit=10",
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensors\"",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkGetSensorsByGatewayResponse(gatewayTenantMatch, expectedTenantMatchSensors, 1, 10),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorTenantMatchA),
				postSetupDeleteSensor(sensorTenantMatchB),
				postSetupDeleteByGateway(gatewayTenantMatch),
			},
		},
	}

	helper.RunTests(router, ctx, tests, t, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx)
}

func checkGetSensorsByGatewayResponse(
	expectedGatewayID string,
	expectedAllSensors []expectedGatewaySensor,
	page int,
	limit int,
) func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		w *httptest.ResponseRecorder,
		_ clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		var resp getSensorsByGatewayResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return false
		}

		sorted := make([]expectedGatewaySensor, len(expectedAllSensors))
		copy(sorted, expectedAllSensors)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})

		offset := (page - 1) * limit
		expectedPageSensors := []expectedGatewaySensor{}
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

			if received.GatewayID != expectedGatewayID {
				return false
			}
			if received.SensorID != expected.SensorID {
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

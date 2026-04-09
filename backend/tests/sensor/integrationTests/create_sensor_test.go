package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/gateway"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	"backend/internal/shared/identity"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

type createSensorRequest struct {
	DataInterval int64  `json:"data_interval"`
	GatewayID    string `json:"gateway_id"`
	Profile      string `json:"profile"`
	SensorName   string `json:"sensor_name"`
}

type createSensorResponse struct {
	GatewayID      string `json:"gateway_id"`
	Profile        string `json:"profile"`
	SensorID       string `json:"sensor_id"`
	DataInterval   int64  `json:"data_interval"`
	SensorInterval int64  `json:"sensor_interval"`
	SensorName     string `json:"sensor_name"`
	Status         string `json:"status"`
}

func TestCreateSensorIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := deps.AuthTokenManager.GenerateForRequester(identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	})
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantID := uuid.New()
	tenantAdminJWT, err := deps.AuthTokenManager.GenerateForRequester(identity.Requester{
		RequesterUserId:   999,
		RequesterTenantId: &tenantID,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	})
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	gatewayForUnauthorized := uuid.NewString()
	gatewayForNotFound := uuid.NewString()
	gatewayForTimeout := uuid.NewString()
	gatewayForFailedReply := uuid.NewString()
	gatewayForSuccess := uuid.NewString()

	var timeoutSub *nats.Subscription
	var failedReplySub *nats.Subscription
	var successSub *nats.Subscription
	var successCmd sensor.CreateSensorCmdEntity

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor",
			Header:    integration.AuthHeader("invalid.jwt.token"),
			Body: helper.MustJSONBody(t, createSensorRequest{
				DataInterval: 1200,
				GatewayID:    uuid.NewString(),
				Profile:      string(sensorProfile.HEART_RATE),
				SensorName:   "Invalid JWT Sensor",
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayForUnauthorized, "Gateway Unauthorized"),
			},
			Name:   "Creazione di un sensore da un utente non super admin",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body: helper.MustJSONBody(t, createSensorRequest{
				DataInterval: 1400,
				GatewayID:    gatewayForUnauthorized,
				Profile:      string(sensorProfile.ECG_CUSTOM),
				SensorName:   "Unauthorized Sensor",
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSensorForGateway(gatewayForUnauthorized),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteByGateway(gatewayForUnauthorized),
			},
		},
		{
			PreSetups: nil,
			Name:      "Creazione di un sensore con gateway non esistente",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor",
			Header:    integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createSensorRequest{
				DataInterval: 1500,
				GatewayID:    gatewayForNotFound,
				Profile:      string(sensorProfile.HEART_RATE),
				SensorName:   "Gateway Missing Sensor",
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSensorForGateway(gatewayForNotFound),
			},

			PostSetups: nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayForTimeout, "Gateway Timeout"),
				preSetupCommandResponseListener(&timeoutSub, false, sensor.CommandResponse{}, sensor.CREATE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Creazione sensore con timeout NATS e nessun inserimento DB",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createSensorRequest{
				DataInterval: 1550,
				GatewayID:    gatewayForTimeout,
				Profile:      string(sensorProfile.PULSE_OXIMETER),
				SensorName:   "Timeout Sensor",
			}),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSensorForGateway(gatewayForTimeout),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteByGateway(gatewayForTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayForFailedReply, "Gateway Failed Reply"),
				preSetupCommandResponseListener(&failedReplySub, true, sensor.CommandResponse{Success: false, Message: "nats create failed"}, sensor.CREATE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Creazione sensore con NATS success false e nessun inserimento DB",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createSensorRequest{
				DataInterval: 1580,
				GatewayID:    gatewayForFailedReply,
				Profile:      string(sensorProfile.HEALTH_THERMOMETER),
				SensorName:   "Failed Reply Sensor",
			}),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats create failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSensorForGateway(gatewayForFailedReply),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteByGateway(gatewayForFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayForSuccess, "Gateway Success"),
				preSetupCommandResponseListener(
					&successSub,
					true,
					sensor.CommandResponse{Success: true, Message: "ok"},
					sensor.CREATE_SENSOR_CMD_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			Name:   "Creazione sensore con reply NATS valida e confronto response/DB",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createSensorRequest{
				DataInterval: 1600,
				GatewayID:    gatewayForSuccess,
				Profile:      string(sensorProfile.ENVIRONMENTAL_SENSING),
				SensorName:   "Successful Sensor",
			}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensor_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkResponseMatchesDBAndCommand(t, &successCmd),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteByGateway(gatewayForSuccess),
				postSetupUnsubscribe(&successSub),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func checkResponseMatchesDBAndCommand(
	t *testing.T,
	cmd *sensor.CreateSensorCmdEntity,
) helper.IntegrationTestCheck {
	t.Helper()
	return func(w *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var resp createSensorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Errorf("errore unmarshaling: %v", err)
			return false
		}

		interval := resp.SensorInterval
		if interval == 0 {
			interval = resp.DataInterval
		}

		db := (*gorm.DB)(deps.CloudDB)
		var dbSensor sensor.SensorEntity
		if err := db.Where("id = ?", resp.SensorID).First(&dbSensor).Error; err != nil {
			t.Errorf("errore db get: %v", err)
			return false
		}

		if resp.SensorID != dbSensor.ID || resp.GatewayID != dbSensor.GatewayID {
			t.Errorf("sensor id o gateway id sbagliato")
			return false
		}

		if resp.Profile != dbSensor.Profile || resp.SensorName != dbSensor.Name || resp.Status != dbSensor.Status {
			t.Errorf("sensor profile o sensor name o sensor status sbagliato")
			return false
		}

		if interval != dbSensor.Interval {
			t.Errorf("intervallo sbagliato\nresp: %#v\ndb:%#v", resp, dbSensor)
			return false
		}

		if cmd.SensorId == "" || cmd.SensorId != dbSensor.ID || cmd.GatewayId != dbSensor.GatewayID {
			t.Errorf("comando sbagliato")
			return false
		}

		if cmd.Interval != dbSensor.Interval || cmd.Profile != dbSensor.Profile {
			t.Errorf("intervallo o profilo sbagliato")
			return false
		}

		return true
	}
}

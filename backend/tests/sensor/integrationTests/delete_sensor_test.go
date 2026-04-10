package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/infra/transport/http/dto"
	"backend/internal/sensor"
	sensorProfile "backend/internal/sensor/profile"
	"backend/internal/shared/identity"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

type deleteSensorResponse struct {
	GatewayID      string `json:"gateway_id"`
	Profile        string `json:"profile"`
	SensorID       string `json:"sensor_id"`
	DataInterval   int64  `json:"data_interval"`
	SensorInterval int64  `json:"sensor_interval"`
	SensorName     string `json:"sensor_name"`
	Status         string `json:"status"`
}

func TestDeleteSensorIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT := mustGenerateJWTForRequester(t, deps.AuthTokenManager, identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	})

	tenantID := uuid.New()
	tenantAdminJWT := mustGenerateJWTForRequester(t, deps.AuthTokenManager, identity.Requester{
		RequesterUserId:   999,
		RequesterTenantId: &tenantID,
		RequesterRole:     identity.ROLE_TENANT_ADMIN,
	})

	gatewayUnauthorized := uuid.NewString()
	sensorUnauthorized := uuid.NewString()

	sensorNotFound := uuid.NewString()

	gatewayTimeout := uuid.NewString()
	sensorTimeout := uuid.NewString()

	gatewayFailedReply := uuid.NewString()
	sensorFailedReply := uuid.NewString()

	gatewaySuccess := uuid.NewString()
	sensorSuccess := uuid.NewString()

	var timeoutSub *nats.Subscription
	var failedReplySub *nats.Subscription
	var successSub *nats.Subscription
	var successCmd sensor.DeleteSensorCmdEntity

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodDelete,
			Path:      "/api/v1/sensor/" + uuid.NewString(),
			Header:    integration.AuthHeader("invalid.jwt.token"),
			Body:      nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio richiesta con sensor_id non valido",
			Method:    http.MethodDelete,
			Path:      "/api/v1/sensor/not-a-uuid",
			Header:    integration.AuthHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: sensor.ErrInvalidSensorID.Error(),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayUnauthorized, "Gateway Unauthorized Delete"),
				preSetupCreateSensor(sensorUnauthorized, gatewayUnauthorized, "Sensor Unauthorized Delete", 1000, sensorProfile.HEART_RATE, sensor.Active),
			},
			Name:   "Eliminazione di un sensore da un utente non super admin",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorUnauthorized,
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorExists(sensorUnauthorized),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorUnauthorized),
				postSetupDeleteByGateway(gatewayUnauthorized),
			},
		},
		{
			PreSetups: nil,
			Name:      "Eliminazione di un sensore non esistente",
			Method:    http.MethodDelete,
			Path:      "/api/v1/sensor/" + sensorNotFound,
			Header:    integration.AuthHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorNotExists(sensorNotFound),
			},

			PostSetups: nil,
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayTimeout, "Gateway Timeout Delete"),
				preSetupCreateSensor(sensorTimeout, gatewayTimeout, "Sensor Timeout Delete", 1100, sensorProfile.PULSE_OXIMETER, sensor.Active),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, sensor.DELETE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Eliminazione sensore con timeout NATS e nessuna eliminazione DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorTimeout,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorExists(sensorTimeout),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTimeout),
				postSetupDeleteByGateway(gatewayTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayFailedReply, "Gateway Failed Reply Delete"),
				preSetupCreateSensor(sensorFailedReply, gatewayFailedReply, "Sensor Failed Reply Delete", 1200, sensorProfile.HEALTH_THERMOMETER, sensor.Active),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats delete failed"}, sensor.DELETE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Eliminazione sensore con NATS success false e nessuna eliminazione DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorFailedReply,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats delete failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorExists(sensorFailedReply),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorFailedReply),
				postSetupDeleteByGateway(gatewayFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewaySuccess, "Gateway Success Delete"),
				preSetupCreateSensor(sensorSuccess, gatewaySuccess, "Sensor Success Delete", 1300, sensorProfile.ENVIRONMENTAL_SENSING, sensor.Active),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					sensor.DELETE_SENSOR_CMD_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			Name:   "Eliminazione sensore con reply NATS valida e sensore rimosso da DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorSuccess,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensor_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkDeleteResponseAndDB(&successCmd, sensorSuccess, gatewaySuccess),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorSuccess),
				postSetupDeleteByGateway(gatewaySuccess),
				postSetupUnsubscribe(&successSub),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func checkDeleteResponseAndDB(
	cmd *sensor.DeleteSensorCmdEntity,
	expectedSensorID string,
	expectedGatewayID string,
) helper.IntegrationTestCheck {
	return func(w *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var resp deleteSensorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return false
		}

		if resp.SensorID != expectedSensorID || resp.GatewayID != expectedGatewayID {
			return false
		}

		if cmd.SensorId != expectedSensorID || cmd.GatewayId != expectedGatewayID {
			return false
		}

		db := (*gorm.DB)(deps.CloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("id = ?", expectedSensorID).Count(&count).Error
		return err == nil && count == 0
	}
}

package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/gateway"
	clouddb "backend/internal/infra/database/cloud_db/connection"
	sensordb "backend/internal/infra/database/sensor_db"
	natsutils "backend/internal/infra/nats"
	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/tests/helper"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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
	router, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx, jwtManager, ctx := helper.Setup(t)

	superAdminJWT, err := jwtManager.GenerateForRequester(identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	})
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantID := uuid.New()
	tenantAdminJWT, err := jwtManager.GenerateForRequester(identity.Requester{
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

	tests := []helper.TestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor",
			Header:    authHeader("invalid.jwt.token"),
			Body: mustJSONBody(t, createSensorRequest{
				DataInterval: 1200,
				GatewayID:    uuid.NewString(),
				Profile:      string(sensor.HEART_RATE),
				SensorName:   "Invalid JWT Sensor",
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayForUnauthorized, "Gateway Unauthorized"),
			},
			Name:   "Creazione di un sensore da un utente non super admin",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor",
			Header: authHeader(tenantAdminJWT),
			Body: mustJSONBody(t, createSensorRequest{
				DataInterval: 1400,
				GatewayID:    gatewayForUnauthorized,
				Profile:      string(sensor.ECG_CUSTOM),
				SensorName:   "Unauthorized Sensor",
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkNoSensorForGateway(gatewayForUnauthorized),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteByGateway(gatewayForUnauthorized),
			},
		},
		{
			PreSetups: nil,
			Name:      "Creazione di un sensore con gateway non esistente",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor",
			Header:    authHeader(superAdminJWT),
			Body: mustJSONBody(t, createSensorRequest{
				DataInterval: 1500,
				GatewayID:    gatewayForNotFound,
				Profile:      string(sensor.HEART_RATE),
				SensorName:   "Gateway Missing Sensor",
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkNoSensorForGateway(gatewayForNotFound),
			},

			PostSetups: nil,
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayForTimeout, "Gateway Timeout"),
				preSetupCommandResponseListener(&timeoutSub, false, sensor.CommandResponse{}, sensor.CREATE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Creazione sensore con timeout NATS e nessun inserimento DB",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor",
			Header: authHeader(superAdminJWT),
			Body: mustJSONBody(t, createSensorRequest{
				DataInterval: 1550,
				GatewayID:    gatewayForTimeout,
				Profile:      string(sensor.PULSE_OXIMETER),
				SensorName:   "Timeout Sensor",
			}),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkNoSensorForGateway(gatewayForTimeout),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteByGateway(gatewayForTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayForFailedReply, "Gateway Failed Reply"),
				preSetupCommandResponseListener(&failedReplySub, true, sensor.CommandResponse{Success: false, Message: "nats create failed"}, sensor.CREATE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Creazione sensore con NATS success false e nessun inserimento DB",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor",
			Header: authHeader(superAdminJWT),
			Body: mustJSONBody(t, createSensorRequest{
				DataInterval: 1580,
				GatewayID:    gatewayForFailedReply,
				Profile:      string(sensor.HEALTH_THERMOMETER),
				SensorName:   "Failed Reply Sensor",
			}),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats create failed",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkNoSensorForGateway(gatewayForFailedReply),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteByGateway(gatewayForFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
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
			Header: authHeader(superAdminJWT),
			Body: mustJSONBody(t, createSensorRequest{
				DataInterval: 1600,
				GatewayID:    gatewayForSuccess,
				Profile:      string(sensor.ENVIRONMENTAL_SENSING),
				SensorName:   "Successful Sensor",
			}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensor_id\"",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkResponseMatchesDBAndCommand(&successCmd),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteByGateway(gatewayForSuccess),
				postSetupUnsubscribe(&successSub),
			},
		},
	}

	helper.RunTests(router, ctx, tests, t, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx)
}

func checkResponseMatchesDBAndCommand(
	cmd *sensor.CreateSensorCmdEntity,
) func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		w *httptest.ResponseRecorder,
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		var resp createSensorResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			return false
		}

		interval := resp.SensorInterval
		if interval == 0 {
			interval = resp.DataInterval
		}

		db := (*gorm.DB)(cloudDB)
		var dbSensor sensor.SensorEntity
		if err := db.Where("id = ?", resp.SensorID).First(&dbSensor).Error; err != nil {
			return false
		}

		if resp.SensorID != dbSensor.ID || resp.GatewayID != dbSensor.GatewayID {
			return false
		}

		if resp.Profile != dbSensor.Profile || resp.SensorName != dbSensor.Name || resp.Status != dbSensor.Status {
			return false
		}

		if interval != dbSensor.Interval {
			return false
		}

		if cmd.SensorId == "" || cmd.SensorId != dbSensor.ID || cmd.GatewayId != dbSensor.GatewayID {
			return false
		}

		if cmd.Interval != dbSensor.Interval || cmd.Profile != dbSensor.Profile {
			return false
		}

		return true
	}
}

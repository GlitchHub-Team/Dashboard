package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/tests/helper"

	clouddb "backend/internal/infra/database/cloud_db/connection"
	sensordb "backend/internal/infra/database/sensor_db"
	natsutils "backend/internal/infra/nats"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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
	router, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx, jwtManager, ctx := helper.Setup(t)

	superAdminJWT := mustGenerateJWTForRequester(t, jwtManager, identity.Requester{
		RequesterUserId: 1,
		RequesterRole:   identity.ROLE_SUPER_ADMIN,
	})

	tenantID := uuid.New()
	tenantAdminJWT := mustGenerateJWTForRequester(t, jwtManager, identity.Requester{
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

	tests := []helper.TestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodDelete,
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
			Name:      "Invio richiesta con sensor_id non valido",
			Method:    http.MethodDelete,
			Path:      "/api/v1/sensor/not-a-uuid",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: sensor.ErrInvalidSensorID.Error(),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayUnauthorized, "Gateway Unauthorized Delete"),
				preSetupCreateSensor(sensorUnauthorized, gatewayUnauthorized, "Sensor Unauthorized Delete", 1000, sensor.HEART_RATE, sensor.Active),
			},
			Name:   "Eliminazione di un sensore da un utente non super admin",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorUnauthorized,
			Header: authHeader(tenantAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorExists(sensorUnauthorized),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorUnauthorized),
				postSetupDeleteByGateway(gatewayUnauthorized),
			},
		},
		{
			PreSetups: nil,
			Name:      "Eliminazione di un sensore non esistente",
			Method:    http.MethodDelete,
			Path:      "/api/v1/sensor/" + sensorNotFound,
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorNotExists(sensorNotFound),
			},

			PostSetups: nil,
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayTimeout, "Gateway Timeout Delete"),
				preSetupCreateSensor(sensorTimeout, gatewayTimeout, "Sensor Timeout Delete", 1100, sensor.PULSE_OXIMETER, sensor.Active),
				preSetupCommandResponseListener(&timeoutSub, false, sensor.CommandResponse{}, sensor.DELETE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Eliminazione sensore con timeout NATS e nessuna eliminazione DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorTimeout,
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorExists(sensorTimeout),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorTimeout),
				postSetupDeleteByGateway(gatewayTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayFailedReply, "Gateway Failed Reply Delete"),
				preSetupCreateSensor(sensorFailedReply, gatewayFailedReply, "Sensor Failed Reply Delete", 1200, sensor.HEALTH_THERMOMETER, sensor.Active),
				preSetupCommandResponseListener(&failedReplySub, true, sensor.CommandResponse{Success: false, Message: "nats delete failed"}, sensor.DELETE_SENSOR_CMD_SUBJECT),
			},
			Name:   "Eliminazione sensore con NATS success false e nessuna eliminazione DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorFailedReply,
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats delete failed",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorExists(sensorFailedReply),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorFailedReply),
				postSetupDeleteByGateway(gatewayFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewaySuccess, "Gateway Success Delete"),
				preSetupCreateSensor(sensorSuccess, gatewaySuccess, "Sensor Success Delete", 1300, sensor.ENVIRONMENTAL_SENSING, sensor.Active),
				preSetupCommandResponseListener(
					&successSub,
					true,
					sensor.CommandResponse{Success: true, Message: "ok"},
					sensor.DELETE_SENSOR_CMD_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			Name:   "Eliminazione sensore con reply NATS valida e sensore rimosso da DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/sensor/" + sensorSuccess,
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"sensor_id\"",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkDeleteResponseAndDB(&successCmd, sensorSuccess, gatewaySuccess),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorSuccess),
				postSetupDeleteByGateway(gatewaySuccess),
				postSetupUnsubscribe(&successSub),
			},
		},
	}

	helper.RunTests(router, ctx, tests, t, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx)
}

func checkDeleteResponseAndDB(
	cmd *sensor.DeleteSensorCmdEntity,
	expectedSensorID string,
	expectedGatewayID string,
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

		db := (*gorm.DB)(cloudDB)
		var count int64
		err := db.Model(&sensor.SensorEntity{}).Where("id = ?", expectedSensorID).Count(&count).Error
		return err == nil && count == 0
	}
}

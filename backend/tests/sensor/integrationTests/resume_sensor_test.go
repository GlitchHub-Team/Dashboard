package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	"gorm.io/gorm"
)

func TestResumeSensorIntegration(t *testing.T) {
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

	sensorNotFound := uuid.NewString()

	gatewayUnauthorizedNil := uuid.NewString()
	sensorUnauthorizedNil := uuid.NewString()

	gatewayUnauthorizedMismatch := uuid.NewString()
	sensorUnauthorizedMismatch := uuid.NewString()

	gatewayAlreadyActive := uuid.NewString()
	sensorAlreadyActive := uuid.NewString()

	gatewayTimeout := uuid.NewString()
	sensorTimeout := uuid.NewString()

	gatewayFailedReply := uuid.NewString()
	sensorFailedReply := uuid.NewString()

	gatewaySuperAdminNil := uuid.NewString()
	sensorSuperAdminNil := uuid.NewString()

	gatewaySuccess := uuid.NewString()
	sensorSuccess := uuid.NewString()

	var timeoutSub *nats.Subscription
	var failedReplySub *nats.Subscription
	var superAdminNilSub *nats.Subscription
	var successSub *nats.Subscription
	var successCmd sensor.ResumeSensorCmdEntity

	tests := []helper.TestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor/" + uuid.NewString() + "/resume",
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
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor/not-a-uuid/resume",
			Header:    authHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: sensor.ErrInvalidSensorID.Error(),
			ResponseChecks:   nil,

			PostSetups: nil,
		},
		{
			PreSetups: nil,
			Name:      "Invio richiesta per sensore non esistente",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor/" + sensorNotFound + "/resume",
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
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedNil, "Gateway Unauthorized Nil Resume", nil),
				preSetupCreateSensor(sensorUnauthorizedNil, gatewayUnauthorizedNil, "Sensor Unauthorized Nil Resume", 1000, sensor.HEART_RATE, sensor.Inactive),
			},
			Name:   "Resume di un sensore da un utente non super admin associato ad un gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedNil + "/resume",
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorUnauthorizedNil, sensor.Inactive),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorUnauthorizedNil),
				postSetupDeleteByGateway(gatewayUnauthorizedNil),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedMismatch, "Gateway Unauthorized Mismatch Resume", &tenantIDOneString),
				preSetupCreateSensor(sensorUnauthorizedMismatch, gatewayUnauthorizedMismatch, "Sensor Unauthorized Mismatch Resume", 1100, sensor.ECG_CUSTOM, sensor.Inactive),
			},
			Name:   "Resume di un sensore da un utente non super admin con tenantId diverso da quello del gateway",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedMismatch + "/resume",
			Header: authHeader(tenantAdminTenantTwoJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorUnauthorizedMismatch, sensor.Inactive),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorUnauthorizedMismatch),
				postSetupDeleteByGateway(gatewayUnauthorizedMismatch),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayAlreadyActive, "Gateway Already Active"),
				preSetupCreateSensor(sensorAlreadyActive, gatewayAlreadyActive, "Sensor Already Active", 1200, sensor.HEALTH_THERMOMETER, sensor.Active),
			},
			Name:   "Resume sensore già attivo",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorAlreadyActive + "/resume",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: sensor.ErrSensorNotInactive.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorAlreadyActive, sensor.Active),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorAlreadyActive),
				postSetupDeleteByGateway(gatewayAlreadyActive),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayTimeout, "Gateway Timeout Resume"),
				preSetupCreateSensor(sensorTimeout, gatewayTimeout, "Sensor Timeout Resume", 1300, sensor.PULSE_OXIMETER, sensor.Inactive),
				preSetupCommandResponseListener(&timeoutSub, false, sensor.CommandResponse{}, sensor.RESUME_SENSOR_CMD_SUBJECT),
			},
			Name:   "Resume sensore valida ma request NATS in timeout",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorTimeout + "/resume",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorTimeout, sensor.Inactive),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorTimeout),
				postSetupDeleteByGateway(gatewayTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayFailedReply, "Gateway Failed Resume"),
				preSetupCreateSensor(sensorFailedReply, gatewayFailedReply, "Sensor Failed Resume", 1400, sensor.ENVIRONMENTAL_SENSING, sensor.Inactive),
				preSetupCommandResponseListener(&failedReplySub, true, sensor.CommandResponse{Success: false, Message: "nats resume failed"}, sensor.RESUME_SENSOR_CMD_SUBJECT),
			},
			Name:   "Resume sensore valida ma reply NATS con success false",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorFailedReply + "/resume",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats resume failed",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorFailedReply, sensor.Inactive),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorFailedReply),
				postSetupDeleteByGateway(gatewayFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewaySuperAdminNil, "Gateway Super Admin Nil Resume", nil),
				preSetupCreateSensor(sensorSuperAdminNil, gatewaySuperAdminNil, "Sensor Super Admin Nil Resume", 1500, sensor.HEART_RATE, sensor.Inactive),
				preSetupCommandResponseListener(&superAdminNilSub, true, sensor.CommandResponse{Success: true, Message: "ok"}, sensor.RESUME_SENSOR_CMD_SUBJECT),
			},
			Name:   "Resume sensore super admin di sensore con gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorSuperAdminNil + "/resume",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorSuperAdminNil, sensor.Active),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorSuperAdminNil),
				postSetupDeleteByGateway(gatewaySuperAdminNil),
				postSetupUnsubscribe(&superAdminNilSub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewaySuccess, "Gateway Success Resume", &tenantIDOneString),
				preSetupCreateSensor(sensorSuccess, gatewaySuccess, "Sensor Success Resume", 1600, sensor.ECG_CUSTOM, sensor.Inactive),
				preSetupCommandResponseListener(
					&successSub,
					true,
					sensor.CommandResponse{Success: true, Message: "ok"},
					sensor.RESUME_SENSOR_CMD_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			Name:   "Resume sensore valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorSuccess + "/resume",
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkResumeSuccessAndCommand(&successCmd, sensorSuccess, gatewaySuccess),
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

func checkResumeSuccessAndCommand(
	cmd *sensor.ResumeSensorCmdEntity,
	expectedSensorID string,
	expectedGatewayID string,
) func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool {
	return func(
		_ *httptest.ResponseRecorder,
		cloudDB clouddb.CloudDBConnection,
		_ sensordb.SensorDBConnection,
		_ *nats.Conn,
		_ natsutils.NatsTestConnection,
		_ jetstream.JetStream,
		_ jetstream.JetStream,
	) bool {
		if cmd.SensorId != expectedSensorID || cmd.GatewayId != expectedGatewayID {
			return false
		}

		db := (*gorm.DB)(cloudDB)
		var entity sensor.SensorEntity
		if err := db.Where("id = ?", expectedSensorID).First(&entity).Error; err != nil {
			return false
		}

		return entity.Status == string(sensor.Active)
	}
}

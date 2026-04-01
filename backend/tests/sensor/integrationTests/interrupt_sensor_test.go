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

func TestInterruptSensorIntegration(t *testing.T) {
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

	gatewayAlreadyInterrupted := uuid.NewString()
	sensorAlreadyInterrupted := uuid.NewString()

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
	var successCmd sensor.InterruptSensorCmdEntity

	tests := []helper.TestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor/" + uuid.NewString() + "/interrupt",
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
			Path:      "/api/v1/sensor/not-a-uuid/interrupt",
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
			Path:      "/api/v1/sensor/" + sensorNotFound + "/interrupt",
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
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedNil, "Gateway Unauthorized Nil", nil),
				preSetupCreateSensor(sensorUnauthorizedNil, gatewayUnauthorizedNil, "Sensor Unauthorized Nil", 1000, sensor.HEART_RATE, sensor.Active),
			},
			Name:   "Interruzione di un sensore da un utente non super admin associato ad un gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedNil + "/interrupt",
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorUnauthorizedNil, sensor.Active),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorUnauthorizedNil),
				postSetupDeleteByGateway(gatewayUnauthorizedNil),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedMismatch, "Gateway Unauthorized Mismatch", &tenantIDOneString),
				preSetupCreateSensor(sensorUnauthorizedMismatch, gatewayUnauthorizedMismatch, "Sensor Unauthorized Mismatch", 1100, sensor.ECG_CUSTOM, sensor.Active),
			},
			Name:   "Interruzione di un sensore da un utente non super admin con tenantId diverso da quello del gateway",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedMismatch + "/interrupt",
			Header: authHeader(tenantAdminTenantTwoJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorUnauthorizedMismatch, sensor.Active),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorUnauthorizedMismatch),
				postSetupDeleteByGateway(gatewayUnauthorizedMismatch),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayAlreadyInterrupted, "Gateway Already Interrupted"),
				preSetupCreateSensor(sensorAlreadyInterrupted, gatewayAlreadyInterrupted, "Sensor Already Interrupted", 1200, sensor.HEALTH_THERMOMETER, sensor.Inactive),
			},
			Name:   "Interruzione sensore già interrotto",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorAlreadyInterrupted + "/interrupt",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: sensor.ErrSensorNotActive.Error(),
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorAlreadyInterrupted, sensor.Inactive),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorAlreadyInterrupted),
				postSetupDeleteByGateway(gatewayAlreadyInterrupted),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayTimeout, "Gateway Timeout Interrupt"),
				preSetupCreateSensor(sensorTimeout, gatewayTimeout, "Sensor Timeout Interrupt", 1300, sensor.PULSE_OXIMETER, sensor.Active),
				preSetupCommandResponseListener(&timeoutSub, false, sensor.CommandResponse{}, sensor.INTERRUPT_SENSOR_CMD_SUBJECT),
			},
			Name:   "Interruzione sensore valida ma request NATS in timeout",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorTimeout + "/interrupt",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorTimeout, sensor.Active),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorTimeout),
				postSetupDeleteByGateway(gatewayTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGateway(gatewayFailedReply, "Gateway Failed Interrupt"),
				preSetupCreateSensor(sensorFailedReply, gatewayFailedReply, "Sensor Failed Interrupt", 1400, sensor.ENVIRONMENTAL_SENSING, sensor.Active),
				preSetupCommandResponseListener(&failedReplySub, true, sensor.CommandResponse{Success: false, Message: "nats interrupt failed"}, sensor.INTERRUPT_SENSOR_CMD_SUBJECT),
			},
			Name:   "Interruzione sensore valida ma reply NATS con success false",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorFailedReply + "/interrupt",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats interrupt failed",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorFailedReply, sensor.Active),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorFailedReply),
				postSetupDeleteByGateway(gatewayFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewaySuperAdminNil, "Gateway Super Admin Nil", nil),
				preSetupCreateSensor(sensorSuperAdminNil, gatewaySuperAdminNil, "Sensor Super Admin Nil", 1500, sensor.HEART_RATE, sensor.Active),
				preSetupCommandResponseListener(&superAdminNilSub, true, sensor.CommandResponse{Success: true, Message: "ok"}, sensor.INTERRUPT_SENSOR_CMD_SUBJECT),
			},
			Name:   "Interruzione sensore super admin di sensore con gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorSuperAdminNil + "/interrupt",
			Header: authHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkSensorStatus(sensorSuperAdminNil, sensor.Inactive),
			},

			PostSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream){
				postSetupDeleteSensor(sensorSuperAdminNil),
				postSetupDeleteByGateway(gatewaySuperAdminNil),
				postSetupUnsubscribe(&superAdminNilSub),
			},
		},
		{
			PreSetups: []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				preSetupCreateGatewayWithTenant(gatewaySuccess, "Gateway Success Interrupt", &tenantIDOneString),
				preSetupCreateSensor(sensorSuccess, gatewaySuccess, "Sensor Success Interrupt", 1600, sensor.ECG_CUSTOM, sensor.Active),
				preSetupCommandResponseListener(
					&successSub,
					true,
					sensor.CommandResponse{Success: true, Message: "ok"},
					sensor.INTERRUPT_SENSOR_CMD_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			Name:   "Interruzione sensore valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorSuccess + "/interrupt",
			Header: authHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool{
				checkInterruptSuccessAndCommand(&successCmd, sensorSuccess, gatewaySuccess),
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

func checkInterruptSuccessAndCommand(
	cmd *sensor.InterruptSensorCmdEntity,
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

		return entity.Status == string(sensor.Inactive)
	}
}

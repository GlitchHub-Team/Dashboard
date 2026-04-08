package sensor_integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/sensor"
	"backend/internal/shared/identity"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

func TestInterruptSensorIntegration(t *testing.T) {
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

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor/" + uuid.NewString() + "/interrupt",
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
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor/not-a-uuid/interrupt",
			Header:    integration.AuthHeader(superAdminJWT),
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
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedNil, "Gateway Unauthorized Nil", nil),
				preSetupCreateSensor(sensorUnauthorizedNil, gatewayUnauthorizedNil, "Sensor Unauthorized Nil", 1000, sensorProfile.HEART_RATE, sensor.Active),
			},
			Name:   "Interruzione di un sensore da un utente non super admin associato ad un gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedNil + "/interrupt",
			Header: integration.AuthHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorUnauthorizedNil, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorUnauthorizedNil),
				postSetupDeleteByGateway(gatewayUnauthorizedNil),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedMismatch, "Gateway Unauthorized Mismatch", &tenant1IdStr),
				preSetupCreateSensor(sensorUnauthorizedMismatch, gatewayUnauthorizedMismatch, "Sensor Unauthorized Mismatch", 1100, sensorProfile.ECG_CUSTOM, sensor.Active),
			},
			Name:   "Interruzione di un sensore da un utente non super admin con tenantId diverso da quello del gateway",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedMismatch + "/interrupt",
			Header: integration.AuthHeader(tenantAdminTenantTwoJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorUnauthorizedMismatch, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorUnauthorizedMismatch),
				postSetupDeleteByGateway(gatewayUnauthorizedMismatch),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayAlreadyInterrupted, "Gateway Already Interrupted"),
				preSetupCreateSensor(sensorAlreadyInterrupted, gatewayAlreadyInterrupted, "Sensor Already Interrupted", 1200, sensorProfile.HEALTH_THERMOMETER, sensor.Inactive),
			},
			Name:   "Interruzione sensore già interrotto",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorAlreadyInterrupted + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: sensor.ErrSensorNotActive.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorAlreadyInterrupted, sensor.Inactive),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorAlreadyInterrupted),
				postSetupDeleteByGateway(gatewayAlreadyInterrupted),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayTimeout, "Gateway Timeout Interrupt"),
				preSetupCreateSensor(sensorTimeout, gatewayTimeout, "Sensor Timeout Interrupt", 1300, sensorProfile.PULSE_OXIMETER, sensor.Active),
				preSetupCommandResponseListener(&timeoutSub, false, sensor.CommandResponse{}, sensor.INTERRUPT_SENSOR_CMD_SUBJECT),
			},
			Name:   "Interruzione sensore valida ma request NATS in timeout",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorTimeout + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorTimeout, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTimeout),
				postSetupDeleteByGateway(gatewayTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayFailedReply, "Gateway Failed Interrupt"),
				preSetupCreateSensor(sensorFailedReply, gatewayFailedReply, "Sensor Failed Interrupt", 1400, sensorProfile.ENVIRONMENTAL_SENSING, sensor.Active),
				preSetupCommandResponseListener(&failedReplySub, true, sensor.CommandResponse{Success: false, Message: "nats interrupt failed"}, sensor.INTERRUPT_SENSOR_CMD_SUBJECT),
			},
			Name:   "Interruzione sensore valida ma reply NATS con success false",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorFailedReply + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats interrupt failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorFailedReply, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorFailedReply),
				postSetupDeleteByGateway(gatewayFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewaySuperAdminNil, "Gateway Super Admin Nil", nil),
				preSetupCreateSensor(sensorSuperAdminNil, gatewaySuperAdminNil, "Sensor Super Admin Nil", 1500, sensorProfile.HEART_RATE, sensor.Active),
				preSetupCommandResponseListener(&superAdminNilSub, true, sensor.CommandResponse{Success: true, Message: "ok"}, sensor.INTERRUPT_SENSOR_CMD_SUBJECT),
			},
			Name:   "Interruzione sensore super admin di sensore con gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorSuperAdminNil + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorSuperAdminNil, sensor.Inactive),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorSuperAdminNil),
				postSetupDeleteByGateway(gatewaySuperAdminNil),
				postSetupUnsubscribe(&superAdminNilSub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewaySuccess, "Gateway Success Interrupt", &tenant1IdStr),
				preSetupCreateSensor(sensorSuccess, gatewaySuccess, "Sensor Success Interrupt", 1600, sensorProfile.ECG_CUSTOM, sensor.Active),
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
			Header: integration.AuthHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkInterruptSuccessAndCommand(&successCmd, sensorSuccess, gatewaySuccess),
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

func checkInterruptSuccessAndCommand(
	cmd *sensor.InterruptSensorCmdEntity,
	expectedSensorID string,
	expectedGatewayID string,
) helper.IntegrationTestCheck {
	return func(r *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		if cmd.SensorId != expectedSensorID || cmd.GatewayId != expectedGatewayID {
			return false
		}

		db := (*gorm.DB)(deps.CloudDB)
		var entity sensor.SensorEntity
		if err := db.Where("id = ?", expectedSensorID).First(&entity).Error; err != nil {
			return false
		}

		return entity.Status == string(sensor.Inactive)
	}
}

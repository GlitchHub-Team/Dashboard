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

func TestResumeSensorIntegration(t *testing.T) {
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

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: nil,
			Name:      "Invio da parte di utente con jwt non valido",
			Method:    http.MethodPost,
			Path:      "/api/v1/sensor/" + uuid.NewString() + "/resume",
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
			Path:      "/api/v1/sensor/not-a-uuid/resume",
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
			Path:      "/api/v1/sensor/" + sensorNotFound + "/resume",
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
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedNil, "Gateway Unauthorized Nil Resume", nil),
				preSetupCreateSensor(sensorUnauthorizedNil, gatewayUnauthorizedNil, "Sensor Unauthorized Nil Resume", 1000, sensorProfile.HEART_RATE, sensor.Inactive),
			},
			Name:   "Resume di un sensore da un utente non super admin associato ad un gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedNil + "/resume",
			Header: integration.AuthHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorUnauthorizedNil, sensor.Inactive),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorUnauthorizedNil),
				postSetupDeleteByGateway(gatewayUnauthorizedNil),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewayUnauthorizedMismatch, "Gateway Unauthorized Mismatch Resume", &tenant1IdStr),
				preSetupCreateSensor(sensorUnauthorizedMismatch, gatewayUnauthorizedMismatch, "Sensor Unauthorized Mismatch Resume", 1100, sensorProfile.ECG_CUSTOM, sensor.Inactive),
			},
			Name:   "Resume di un sensore da un utente non super admin con tenantId diverso da quello del gateway",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorUnauthorizedMismatch + "/resume",
			Header: integration.AuthHeader(tenantAdminTenantTwoJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: sensor.ErrSensorNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorUnauthorizedMismatch, sensor.Inactive),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorUnauthorizedMismatch),
				postSetupDeleteByGateway(gatewayUnauthorizedMismatch),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayAlreadyActive, "Gateway Already Active"),
				preSetupCreateSensor(sensorAlreadyActive, gatewayAlreadyActive, "Sensor Already Active", 1200, sensorProfile.HEALTH_THERMOMETER, sensor.Active),
			},
			Name:   "Resume sensore già attivo",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorAlreadyActive + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: sensor.ErrSensorNotInactive.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorAlreadyActive, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorAlreadyActive),
				postSetupDeleteByGateway(gatewayAlreadyActive),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayTimeout, "Gateway Timeout Resume"),
				preSetupCreateSensor(sensorTimeout, gatewayTimeout, "Sensor Timeout Resume", 1300, sensorProfile.PULSE_OXIMETER, sensor.Inactive),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, sensor.RESUME_SENSOR_CMD_SUBJECT),
			},
			Name:   "Resume sensore valida ma request NATS in timeout",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorTimeout + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorTimeout, sensor.Inactive),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorTimeout),
				postSetupDeleteByGateway(gatewayTimeout),
				postSetupUnsubscribe(&timeoutSub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGateway(gatewayFailedReply, "Gateway Failed Resume"),
				preSetupCreateSensor(sensorFailedReply, gatewayFailedReply, "Sensor Failed Resume", 1400, sensorProfile.ENVIRONMENTAL_SENSING, sensor.Inactive),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats resume failed"}, sensor.RESUME_SENSOR_CMD_SUBJECT),
			},
			Name:   "Resume sensore valida ma reply NATS con success false",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorFailedReply + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats resume failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorFailedReply, sensor.Inactive),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorFailedReply),
				postSetupDeleteByGateway(gatewayFailedReply),
				postSetupUnsubscribe(&failedReplySub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewaySuperAdminNil, "Gateway Super Admin Nil Resume", nil),
				preSetupCreateSensor(sensorSuperAdminNil, gatewaySuperAdminNil, "Sensor Super Admin Nil Resume", 1500, sensorProfile.HEART_RATE, sensor.Inactive),
				preSetupCommandResponseListener(&superAdminNilSub, true, dto.CommandResponse{Success: true, Message: "ok"}, sensor.RESUME_SENSOR_CMD_SUBJECT),
			},
			Name:   "Resume sensore super admin di sensore con gateway con tenantId nil",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorSuperAdminNil + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSensorStatus(sensorSuperAdminNil, sensor.Active),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSensor(sensorSuperAdminNil),
				postSetupDeleteByGateway(gatewaySuperAdminNil),
				postSetupUnsubscribe(&superAdminNilSub),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithTenant(gatewaySuccess, "Gateway Success Resume", &tenant1IdStr),
				preSetupCreateSensor(sensorSuccess, gatewaySuccess, "Sensor Success Resume", 1600, sensorProfile.ECG_CUSTOM, sensor.Inactive),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					sensor.RESUME_SENSOR_CMD_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			Name:   "Resume sensore valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/sensor/" + sensorSuccess + "/resume",
			Header: integration.AuthHeader(tenantAdminTenantOneJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkResumeSuccessAndCommand(&successCmd, sensorSuccess, gatewaySuccess),
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

func checkResumeSuccessAndCommand(
	cmd *sensor.ResumeSensorCmdEntity,
	expectedSensorID string,
	expectedGatewayID string,
) helper.IntegrationTestCheck {
	return func(
		r *httptest.ResponseRecorder,
		deps helper.IntegrationTestDeps,
	) bool {
		if cmd.SensorId != expectedSensorID || cmd.GatewayId != expectedGatewayID {
			return false
		}

		db := (*gorm.DB)(deps.CloudDB)
		var entity sensor.SensorEntity
		if err := db.Where("id = ?", expectedSensorID).First(&entity).Error; err != nil {
			return false
		}

		return entity.Status == string(sensor.Active)
	}
}

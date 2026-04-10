package gateway_integrationtests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"backend/internal/gateway"
	"backend/internal/infra/transport/http/dto"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type createGatewayRequest struct {
	GatewayName string `json:"name"`
	Interval    int64  `json:"interval"`
}

type createGatewayResponse struct {
	GatewayID        string  `json:"gateway_id"`
	GatewayName      string  `json:"name"`
	TenantID         string  `json:"tenant_id"`
	Status           string  `json:"status"`
	Interval         int64   `json:"interval"`
	PublicIdentifier *string `json:"public_identifier"`
}

type createGatewayCommandPayload struct {
	GatewayID string `json:"gatewayId"`
	Interval  int64  `json:"interval"`
}

func TestCreateGatewayIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, 1)
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantID := uuid.New()
	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenantID, 999)
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	nameInvalidJWT := "gw-invalid-jwt-" + uuid.NewString()
	nameInvalidInterval := "gw-invalid-interval-" + uuid.NewString()
	nameNotSuperAdmin := "gw-not-super-admin-" + uuid.NewString()
	nameNatsTimeout := "gw-nats-timeout-" + uuid.NewString()
	nameNatsFailed := "gw-nats-failed-" + uuid.NewString()
	nameNatsMalformed := "gw-nats-malformed-" + uuid.NewString()
	nameSuccess := "gw-success-" + uuid.NewString()

	var timeoutSub *nats.Subscription
	var failedReplySub *nats.Subscription
	var malformedReplySub *nats.Subscription
	var successSub *nats.Subscription
	var successCmd createGatewayCommandPayload

	tests := []*helper.IntegrationTestCase{
		{
			Name:   "Utente senza JWT valido, status 401",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway",
			Header: integration.AuthHeader("invalid.jwt.token"),
			Body: helper.MustJSONBody(t, createGatewayRequest{
				GatewayName: nameInvalidJWT,
				Interval:    5000,
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByName(nameInvalidJWT),
			},

			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
		{
			Name:   "Interval <= 0, status 400",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createGatewayRequest{
				GatewayName: nameInvalidInterval,
				Interval:    0,
			}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid format",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByName(nameInvalidInterval),
			},

			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
		{
			Name:       "Body del json non valido, status 400",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       strings.NewReader("{\"name\":"),
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "Utente non super admin, status 401",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body: helper.MustJSONBody(t, createGatewayRequest{
				GatewayName: nameNotSuperAdmin,
				Interval:    5000,
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "cannot access resource",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByName(nameNotSuperAdmin),
			},

			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
		{
			Name:   "Errore ricezione risposta NATS, status 500 e nessun insert DB",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createGatewayRequest{
				GatewayName: nameNatsTimeout,
				Interval:    5100,
			}),
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.CREATE_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupComposite(
					postSetupDeleteGatewayByName(nameNatsTimeout),
					postSetupUnsubscribe(&timeoutSub),
				),
			},

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByName(nameNatsTimeout),
			},
		},
		{
			Name:   "Response con Success false, status 500 e nessun insert DB",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createGatewayRequest{
				GatewayName: nameNatsFailed,
				Interval:    5200,
			}),
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCommandResponseListener(
					&failedReplySub,
					true,
					dto.CommandResponse{Success: false, Message: "nats create failed"},
					gateway.CREATE_GATEWAY_COMMAND_SUBJECT,
				),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupComposite(
					postSetupDeleteGatewayByName(nameNatsFailed),
					postSetupUnsubscribe(&failedReplySub),
				),
			},

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats create failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByName(nameNatsFailed),
			},
		},
		{
			Name:   "Response json malformata, status 500 e nessun insert DB",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createGatewayRequest{
				GatewayName: nameNatsMalformed,
				Interval:    5300,
			}),
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupRawCommandResponseListener(
					&malformedReplySub,
					gateway.CREATE_GATEWAY_COMMAND_SUBJECT,
					[]byte("{not-json"),
				),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupComposite(
					postSetupDeleteGatewayByName(nameNatsMalformed),
					postSetupUnsubscribe(&malformedReplySub),
				),
			},

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByName(nameNatsMalformed),
			},
		},
		{
			Name:   "Creazione corretta da super admin con insert DB",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, createGatewayRequest{
				GatewayName: nameSuccess,
				Interval:    5400,
			}),
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.CREATE_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupComposite(
					postSetupDeleteGatewayByName(nameSuccess),
					postSetupUnsubscribe(&successSub),
				),
			},

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"gateway_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkCreateGatewayResponseAndDB(t, &successCmd, nameSuccess, 5400),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

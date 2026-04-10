package gateway_integrationtests

import (
	"encoding/json"
	"net/http"
	"testing"

	"backend/internal/gateway"
	"backend/internal/infra/transport/http/dto"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

func TestDeleteGatewayIntegration(t *testing.T) {
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

	gatewayUnauthorized := uuid.NewString()
	gatewayNotFound := uuid.NewString()
	gatewayTimeout := uuid.NewString()
	gatewayFailedReply := uuid.NewString()
	gatewayMalformed := uuid.NewString()
	gatewaySuccess := uuid.NewString()

	var timeoutSub *nats.Subscription
	var failedReplySub *nats.Subscription
	var malformedSub *nats.Subscription
	var successSub *nats.Subscription
	var successCmd gatewayCommandPayload

	tests := []*helper.IntegrationTestCase{
		{
			Name:       "Invio da parte di utente con jwt non valido",
			Method:     http.MethodDelete,
			Path:       "/api/v1/gateway/" + uuid.NewString(),
			Header:     integration.AuthHeader("invalid.jwt.token"),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "Invio richiesta con gateway_id non valido",
			Method:     http.MethodDelete,
			Path:       "/api/v1/gateway/not-a-uuid",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrInvalidGatewayID.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "Eliminazione di un gateway da utente non super admin",
			Method: http.MethodDelete,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized,
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Delete", 5000, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(gatewayUnauthorized),
			},

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayExistsByID(gatewayUnauthorized),
			},
		},
		{
			Name:       "Eliminazione di un gateway non esistente",
			Method:     http.MethodDelete,
			Path:       "/api/v1/gateway/" + gatewayNotFound,
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByID(gatewayNotFound),
			},
		},
		{
			Name:   "Eliminazione gateway con timeout NATS e nessuna eliminazione DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/gateway/" + gatewayTimeout,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewayTimeout, "Gateway Timeout Delete", 5000, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.DELETE_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(2,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayTimeout),
					postSetupUnsubscribe(&timeoutSub),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayExistsByID(gatewayTimeout),
			},
		},
		{
			Name:   "Eliminazione gateway con reply NATS success false e nessuna eliminazione DB",
			Method: http.MethodDelete,
			Path:   "/api/v1/gateway/" + gatewayFailedReply,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewayFailedReply, "Gateway Failed Delete", 5000, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats delete failed"}, gateway.DELETE_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(2,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayFailedReply),
					postSetupUnsubscribe(&failedReplySub),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats delete failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayExistsByID(gatewayFailedReply),
			},
		},
		{
			Name:   "Eliminazione gateway con risposta json NATS malformata",
			Method: http.MethodDelete,
			Path:   "/api/v1/gateway/" + gatewayMalformed,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewayMalformed, "Gateway Malformed Delete", 5000, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupRawCommandResponseListener(&malformedSub, gateway.DELETE_GATEWAY_COMMAND_SUBJECT, []byte("{not-json")),
			},
			PostSetups: postSetupsWithFinal(2,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayMalformed),
					postSetupUnsubscribe(&malformedSub),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayExistsByID(gatewayMalformed),
			},
		},
		{
			Name:   "Eliminazione gateway con request/reply NATS corretta",
			Method: http.MethodDelete,
			Path:   "/api/v1/gateway/" + gatewaySuccess,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Delete", 5000, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.DELETE_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: postSetupsWithFinal(2,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewaySuccess),
					postSetupUnsubscribe(&successSub),
				),
			),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"gateway_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkDeleteGatewayResponseAndCommand(&successCmd, gatewaySuccess, "Gateway Success Delete"),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

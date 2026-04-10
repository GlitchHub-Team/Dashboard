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

func TestResetGatewayIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, 1)
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantIDOne := uuid.New()
	tenantIDTwo := uuid.New()
	tenantOneIDStr := tenantIDOne.String()

	tenantAdminOneJWT, err := helper.NewTenantAdminJWT(deps, tenantIDOne, 999)
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT (one): %v", err)
	}
	tenantAdminTwoJWT, err := helper.NewTenantAdminJWT(deps, tenantIDTwo, 1000)
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT (two): %v", err)
	}

	gatewayNotFound := uuid.NewString()
	gatewayUnauthorized := uuid.NewString()
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
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/reset",
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
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/not-a-uuid/reset",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrInvalidGatewayID.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "Reset gateway non esistente",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + gatewayNotFound + "/reset",
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
			Name:   "Reset da tenant admin con tenant diverso",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized + "/reset",
			Header: integration.AuthHeader(tenantAdminTwoJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Reset", 8000, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayUnauthorized), integration.PostSetupDeleteTenant(t, tenantIDOne))),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayUnauthorized, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 8000),
			},
		},
		{
			Name:   "Reset valida ma timeout NATS",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayTimeout + "/reset",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayTimeout, "Gateway Timeout Reset", 8100, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.RESET_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayTimeout), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&timeoutSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayTimeout, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 8100),
			},
		},
		{
			Name:   "Reset valida ma reply NATS success false",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayFailedReply + "/reset",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayFailedReply, "Gateway Failed Reset", 8200, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats reset failed"}, gateway.RESET_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayFailedReply), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&failedReplySub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats reset failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayFailedReply, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 8200),
			},
		},
		{
			Name:   "Reset valida ma risposta json NATS malformata",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayMalformed + "/reset",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayMalformed, "Gateway Malformed Reset", 8300, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupRawCommandResponseListener(&malformedSub, gateway.RESET_GATEWAY_COMMAND_SUBJECT, []byte("{not-json")),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayMalformed), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&malformedSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayMalformed, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 8300),
			},
		},
		{
			Name:   "Reset valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewaySuccess + "/reset",
			Header: integration.AuthHeader(tenantAdminOneJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Reset", 8400, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.RESET_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewaySuccess), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&successSub))),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "Reset del gateway eseguito correttamente",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayCommandAndState(&successCmd, gatewaySuccess, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, gateway.DEFAULT_INTERVAL_LIMIT.Milliseconds()),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

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

func TestResumeGatewayIntegration(t *testing.T) {
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
	gatewayNotInactive := uuid.NewString()
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
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/resume",
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
			Path:       "/api/v1/gateway/not-a-uuid/resume",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrInvalidGatewayID.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "Resume gateway non esistente",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + gatewayNotFound + "/resume",
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
			Name:   "Resume da tenant admin con tenant diverso",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized + "/resume",
			Header: integration.AuthHeader(tenantAdminTwoJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Resume", 5000, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayUnauthorized), integration.PostSetupDeleteTenant(t, tenantIDOne))),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayUnauthorized, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, 5000),
			},
		},
		{
			Name:   "Resume di gateway non inattivo (caso extra)",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayNotInactive + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayNotInactive, "Gateway Not Inactive Resume", 5100, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayNotInactive), integration.PostSetupDeleteTenant(t, tenantIDOne))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: gateway.ErrGatewayNotInactive.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayNotInactive, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 5100),
			},
		},
		{
			Name:   "Resume valida ma timeout NATS",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayTimeout + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayTimeout, "Gateway Timeout Resume", 5200, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.RESUME_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayTimeout), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&timeoutSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayTimeout, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, 5200),
			},
		},
		{
			Name:   "Resume valida ma reply NATS success false",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayFailedReply + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayFailedReply, "Gateway Failed Resume", 5300, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats resume failed"}, gateway.RESUME_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayFailedReply), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&failedReplySub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats resume failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayFailedReply, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, 5300),
			},
		},
		{
			Name:   "Resume valida ma risposta json NATS malformata",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayMalformed + "/resume",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayMalformed, "Gateway Malformed Resume", 5400, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, nil),
				preSetupRawCommandResponseListener(&malformedSub, gateway.RESUME_GATEWAY_COMMAND_SUBJECT, []byte("{not-json")),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayMalformed), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&malformedSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayMalformed, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, 5400),
			},
		},
		{
			Name:   "Resume valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewaySuccess + "/resume",
			Header: integration.AuthHeader(tenantAdminOneJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Resume", 5500, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.RESUME_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewaySuccess), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&successSub))),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "ripreso correttamente",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayCommandAndState(&successCmd, gatewaySuccess, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 5500),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

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

func TestInterruptGatewayIntegration(t *testing.T) {
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
	gatewayNotActive := uuid.NewString()
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
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/interrupt",
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
			Path:       "/api/v1/gateway/not-a-uuid/interrupt",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrInvalidGatewayID.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "Interrupt gateway non esistente",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + gatewayNotFound + "/interrupt",
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
			Name:   "Interrupt da tenant admin con tenant diverso",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized + "/interrupt",
			Header: integration.AuthHeader(tenantAdminTwoJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Interrupt", 5000, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayUnauthorized), integration.PostSetupDeleteTenant(t, tenantIDOne))),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayUnauthorized, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 5000),
			},
		},
		{
			Name:   "Interrupt di gateway non attivo (caso extra)",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayNotActive + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayNotActive, "Gateway Not Active Interrupt", 5100, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayNotActive), integration.PostSetupDeleteTenant(t, tenantIDOne))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: gateway.ErrGatewayNotActive.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayNotActive, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, 5100),
			},
		},
		{
			Name:   "Interrupt valida ma timeout NATS",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayTimeout + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayTimeout, "Gateway Timeout Interrupt", 5200, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.INTERRUPT_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayTimeout), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&timeoutSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayTimeout, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 5200),
			},
		},
		{
			Name:   "Interrupt valida ma reply NATS success false",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayFailedReply + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayFailedReply, "Gateway Failed Interrupt", 5300, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats interrupt failed"}, gateway.INTERRUPT_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayFailedReply), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&failedReplySub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats interrupt failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayFailedReply, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 5300),
			},
		},
		{
			Name:   "Interrupt valida ma risposta json NATS malformata",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayMalformed + "/interrupt",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayMalformed, "Gateway Malformed Interrupt", 5400, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupRawCommandResponseListener(&malformedSub, gateway.INTERRUPT_GATEWAY_COMMAND_SUBJECT, []byte("{not-json")),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayMalformed), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&malformedSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayMalformed, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 5400),
			},
		},
		{
			Name:   "Interrupt valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewaySuccess + "/interrupt",
			Header: integration.AuthHeader(tenantAdminOneJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Interrupt", 5500, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.INTERRUPT_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewaySuccess), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&successSub))),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "interrotto correttamente",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayCommandAndState(&successCmd, gatewaySuccess, gateway.GATEWAY_STATUS_INACTIVE, &tenantOneIDStr, 5500),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

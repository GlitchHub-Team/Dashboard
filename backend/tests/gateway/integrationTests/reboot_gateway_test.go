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

func TestRebootGatewayIntegration(t *testing.T) {
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
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/reboot",
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
			Path:       "/api/v1/gateway/not-a-uuid/reboot",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrInvalidGatewayID.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "Reboot gateway non esistente",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + gatewayNotFound + "/reboot",
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
			Name:   "Reboot da tenant admin con tenant diverso",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized + "/reboot",
			Header: integration.AuthHeader(tenantAdminTwoJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Reboot", 9000, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayUnauthorized), integration.PostSetupDeleteTenant(t, tenantIDOne))),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayUnauthorized, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 9000),
			},
		},
		{
			Name:   "Reboot valida ma timeout NATS",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayTimeout + "/reboot",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayTimeout, "Gateway Timeout Reboot", 9100, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.REBOOT_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayTimeout), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&timeoutSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayTimeout, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 9100),
			},
		},
		{
			Name:   "Reboot valida ma reply NATS success false",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayFailedReply + "/reboot",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayFailedReply, "Gateway Failed Reboot", 9200, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats reboot failed"}, gateway.REBOOT_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayFailedReply), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&failedReplySub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats reboot failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayFailedReply, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 9200),
			},
		},
		{
			Name:   "Reboot valida ma risposta json NATS malformata",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayMalformed + "/reboot",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewayMalformed, "Gateway Malformed Reboot", 9300, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupRawCommandResponseListener(&malformedSub, gateway.REBOOT_GATEWAY_COMMAND_SUBJECT, []byte("{not-json")),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayMalformed), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&malformedSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayMalformed, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 9300),
			},
		},
		{
			Name:   "Reboot valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewaySuccess + "/reboot",
			Header: integration.AuthHeader(tenantAdminOneJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantIDOne, true),
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Reboot", 9400, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, nil),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.REBOOT_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewaySuccess), integration.PostSetupDeleteTenant(t, tenantIDOne), postSetupUnsubscribe(&successSub))),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "Reboot del gateway eseguito correttamente",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayCommandAndState(&successCmd, gatewaySuccess, gateway.GATEWAY_STATUS_ACTIVE, &tenantOneIDStr, 9400),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

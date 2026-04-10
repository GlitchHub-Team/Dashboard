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

func TestDecommissionGatewayIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, 1)
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantAdminTenantID := uuid.New()
	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenantAdminTenantID, 999)
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	tenantUUID := uuid.New()
	tenantID := tenantUUID.String()
	gatewayUnauthorized := uuid.NewString()
	gatewayNotFound := uuid.NewString()
	gatewayNotCommissioned := uuid.NewString()
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
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/decommission",
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
			Path:       "/api/v1/gateway/not-a-uuid/decommission",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrInvalidGatewayID.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "Decommission da utente non super admin",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized + "/decommission",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantUUID, true),
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Decommission", 5000, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayUnauthorized), integration.PostSetupDeleteTenant(t, tenantUUID))),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayUnauthorized, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, 5000),
			},
		},
		{
			Name:       "Decommission gateway non esistente",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + gatewayNotFound + "/decommission",
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
			Name:   "Decommission di gateway non commissionato (caso extra)",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayNotCommissioned + "/decommission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewayNotCommissioned, "Gateway Not Commissioned", 5100, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(gatewayNotCommissioned),
			},

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: gateway.ErrGatewayNotCommissioned.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayNotCommissioned, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, 5100),
			},
		},
		{
			Name:   "Decommission valida ma timeout NATS",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayTimeout + "/decommission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantUUID, true),
				preSetupCreateGatewayWithState(gatewayTimeout, "Gateway Timeout Decommission", 5200, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, nil),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.DECOMMISSION_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayTimeout), integration.PostSetupDeleteTenant(t, tenantUUID), postSetupUnsubscribe(&timeoutSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayTimeout, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, 5200),
			},
		},
		{
			Name:   "Decommission valida ma reply NATS success false",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayFailedReply + "/decommission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantUUID, true),
				preSetupCreateGatewayWithState(gatewayFailedReply, "Gateway Failed Decommission", 5300, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, nil),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats decommission failed"}, gateway.DECOMMISSION_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayFailedReply), integration.PostSetupDeleteTenant(t, tenantUUID), postSetupUnsubscribe(&failedReplySub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats decommission failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayFailedReply, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, 5300),
			},
		},
		{
			Name:   "Decommission valida ma risposta json NATS malformata",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayMalformed + "/decommission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantUUID, true),
				preSetupCreateGatewayWithState(gatewayMalformed, "Gateway Malformed Decommission", 5400, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, nil),
				preSetupRawCommandResponseListener(&malformedSub, gateway.DECOMMISSION_GATEWAY_COMMAND_SUBJECT, []byte("{not-json")),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewayMalformed), integration.PostSetupDeleteTenant(t, tenantUUID), postSetupUnsubscribe(&malformedSub))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayMalformed, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, 5400),
			},
		},
		{
			Name:   "Decommission valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewaySuccess + "/decommission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantUUID, true),
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Decommission", 5500, gateway.GATEWAY_STATUS_ACTIVE, &tenantID, nil),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.DECOMMISSION_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: postSetupsWithFinal(3, postSetupComposite(postSetupDeleteGatewayByID(gatewaySuccess), integration.PostSetupDeleteTenant(t, tenantUUID), postSetupUnsubscribe(&successSub))),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"gateway_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayCommandAndState(&successCmd, gatewaySuccess, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, 5500),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

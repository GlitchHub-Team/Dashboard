package gateway_integrationtests

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"backend/internal/gateway"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/tenant"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

type commissionGatewayRequest struct {
	TenantID        string `json:"tenant_id"`
	CommissionToken string `json:"commission_token"`
}

func TestCommissionGatewayIntegration(t *testing.T) {
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

	commissionToken := "commission-token-test"
	validTenantID := uuid.New()
	otherTenantUUID := uuid.New()
	otherTenantID := otherTenantUUID.String()

	gatewayUnauthorized := uuid.NewString()
	gatewayNotFound := uuid.NewString()
	gatewayAlreadyCommissioned := uuid.NewString()
	gatewayTenantNotFound := uuid.NewString()
	gatewayTimeout := uuid.NewString()
	gatewayFailedReply := uuid.NewString()
	gatewayMalformed := uuid.NewString()
	gatewaySuccess := uuid.NewString()

	var timeoutSub *nats.Subscription
	var failedReplySub *nats.Subscription
	var malformedSub *nats.Subscription
	var successSub *nats.Subscription
	var successCmd commissionGatewayCommandPayload

	tests := []*helper.IntegrationTestCase{
		{
			Name:       "Invio da parte di utente con jwt non valido",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/commission",
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
			Path:       "/api/v1/gateway/not-a-uuid/commission",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrInvalidGatewayID.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "Body JSON malformato",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/commission",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       strings.NewReader("{\"tenant_id\":"),
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "tenant_id non valido",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + uuid.NewString() + "/commission",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       helper.MustJSONBody(t, commissionGatewayRequest{TenantID: "not-a-uuid", CommissionToken: commissionToken}),
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid format",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "Commission da utente non super admin",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized + "/commission",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Commission", 6000, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(gatewayUnauthorized),
			},

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayUnauthorized, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, 6000),
			},
		},
		{
			Name:       "Commission gateway non esistente",
			Method:     http.MethodPost,
			Path:       "/api/v1/gateway/" + gatewayNotFound + "/commission",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByID(gatewayNotFound),
			},
		},
		{
			Name:   "Commission gateway già commissionato (caso extra)",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayAlreadyCommissioned + "/commission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(otherTenantUUID, true),
				preSetupCreateGatewayWithState(gatewayAlreadyCommissioned, "Gateway Already Commissioned", 6100, gateway.GATEWAY_STATUS_ACTIVE, &otherTenantID, nil),
			},
			PostSetups: postSetupsWithFinal(2, postSetupComposite(postSetupDeleteGatewayByID(gatewayAlreadyCommissioned), integration.PostSetupDeleteTenant(t, otherTenantUUID))),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: gateway.ErrGatewayAlreadyCommissioned.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayAlreadyCommissioned, gateway.GATEWAY_STATUS_ACTIVE, &otherTenantID, 6100),
			},
		},
		{
			Name:   "Commission con tenant non trovato (caso extra)",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayTenantNotFound + "/commission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(gatewayTenantNotFound, "Gateway Tenant Missing", 6200, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(gatewayTenantNotFound),
			},

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: tenant.ErrTenantNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayTenantNotFound, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, 6200),
			},
		},
		{
			Name:   "Commission valida ma timeout NATS",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayTimeout + "/commission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(validTenantID, true),
				preSetupCreateGatewayWithState(gatewayTimeout, "Gateway Timeout Commission", 6300, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupCommandResponseListener(&timeoutSub, false, dto.CommandResponse{}, gateway.COMMISSION_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayTimeout),
					integration.PostSetupDeleteTenant(t, validTenantID),
					postSetupUnsubscribe(&timeoutSub),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayTimeout, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, 6300),
			},
		},
		{
			Name:   "Commission valida ma reply NATS success false",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayFailedReply + "/commission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(validTenantID, true),
				preSetupCreateGatewayWithState(gatewayFailedReply, "Gateway Failed Commission", 6400, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupCommandResponseListener(&failedReplySub, true, dto.CommandResponse{Success: false, Message: "nats commission failed"}, gateway.COMMISSION_GATEWAY_COMMAND_SUBJECT),
			},
			PostSetups: postSetupsWithFinal(3,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayFailedReply),
					integration.PostSetupDeleteTenant(t, validTenantID),
					postSetupUnsubscribe(&failedReplySub),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "nats commission failed",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayFailedReply, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, 6400),
			},
		},
		{
			Name:   "Commission valida ma risposta json NATS malformata",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewayMalformed + "/commission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(validTenantID, true),
				preSetupCreateGatewayWithState(gatewayMalformed, "Gateway Malformed Commission", 6500, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupRawCommandResponseListener(&malformedSub, gateway.COMMISSION_GATEWAY_COMMAND_SUBJECT, []byte("{not-json")),
			},
			PostSetups: postSetupsWithFinal(3,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayMalformed),
					integration.PostSetupDeleteTenant(t, validTenantID),
					postSetupUnsubscribe(&malformedSub),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "invalid NATS response",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(gatewayMalformed, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, 6500),
			},
		},
		{
			Name:   "Commission valida con request/reply NATS corretta",
			Method: http.MethodPost,
			Path:   "/api/v1/gateway/" + gatewaySuccess + "/commission",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, commissionGatewayRequest{TenantID: validTenantID.String(), CommissionToken: commissionToken}),
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(validTenantID, true),
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Commission", 6600, gateway.GATEWAY_STATUS_DECOMMISSIONED, nil, nil),
				preSetupCommandResponseListener(
					&successSub,
					true,
					dto.CommandResponse{Success: true, Message: "ok"},
					gateway.COMMISSION_GATEWAY_COMMAND_SUBJECT,
					func(msg *nats.Msg) {
						_ = json.Unmarshal(msg.Data, &successCmd)
					},
				),
			},
			PostSetups: postSetupsWithFinal(3,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewaySuccess),
					integration.PostSetupDeleteTenant(t, validTenantID),
					postSetupUnsubscribe(&successSub),
				),
			),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"gateway_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkCommissionResponseAndCommand(t, &successCmd, gatewaySuccess, "Gateway Success Commission", validTenantID.String(), commissionToken, 6600),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

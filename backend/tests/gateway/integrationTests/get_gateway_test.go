package gateway_integrationtests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/gateway"
	"backend/internal/shared/identity"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

type gatewayListRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type gatewayListResponse struct {
	Count   uint                `json:"count"`
	Total   uint                `json:"total"`
	Gateway gatewayHTTPResponse `json:"gateways"`
}

func checkGatewayResponseMatchesExpected(t *testing.T, expected gatewayHTTPResponse) helper.IntegrationTestCheck {
	t.Helper()

	return func(w *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var resp gatewayHTTPResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Errorf("failed to unmarshal gateway response: %v", err)
			return false
		}

		if resp.GatewayID != expected.GatewayID {
			t.Errorf("unexpected gateway_id: got=%s want=%s", resp.GatewayID, expected.GatewayID)
			return false
		}
		if resp.GatewayName != expected.GatewayName {
			t.Errorf("unexpected name: got=%s want=%s", resp.GatewayName, expected.GatewayName)
			return false
		}
		if resp.TenantID != expected.TenantID {
			t.Errorf("unexpected tenant_id: got=%s want=%s", resp.TenantID, expected.TenantID)
			return false
		}
		if resp.Status != expected.Status {
			t.Errorf("unexpected status: got=%s want=%s", resp.Status, expected.Status)
			return false
		}
		if resp.Interval != expected.Interval {
			t.Errorf("unexpected interval: got=%d want=%d", resp.Interval, expected.Interval)
			return false
		}
		if expected.PublicIdentifier == nil {
			if resp.PublicIdentifier != nil {
				t.Errorf("unexpected public_identifier: got=%v want=nil", resp.PublicIdentifier)
				return false
			}
			return true
		}
		if resp.PublicIdentifier == nil || *resp.PublicIdentifier != *expected.PublicIdentifier {
			t.Errorf("unexpected public_identifier: got=%v want=%v", resp.PublicIdentifier, expected.PublicIdentifier)
			return false
		}

		return true
	}
}

func checkGatewayListResponseMatchesExpected(
	t *testing.T,
	expectedPage uint,
	expectedTotal uint,
	expectedGateways []gatewayHTTPResponse,
) helper.IntegrationTestCheck {
	t.Helper()

	return func(w *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		var resp []struct {
			Count   uint                `json:"count"`
			Total   uint                `json:"total"`
			Gateways []gatewayHTTPResponse `json:"gateways"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Errorf("failed to unmarshal gateway list response: %v", err)
			return false
		}

		if len(resp) != len(expectedGateways) {
			t.Errorf("unexpected response length: got=%d want=%d", len(resp), len(expectedGateways))
			return false
		}

		expectedByID := make(map[string]gatewayHTTPResponse, len(expectedGateways))
		for _, expectedGateway := range expectedGateways {
			expectedByID[expectedGateway.GatewayID] = expectedGateway
		}

		for _, item := range resp {
			if item.Count != expectedPage {
				t.Errorf("unexpected count: got=%d want=%d", item.Count, expectedPage)
				return false
			}
			if item.Total != expectedTotal {
				t.Errorf("unexpected total: got=%d want=%d", item.Total, expectedTotal)
				return false
			}
			if len(item.Gateways) != 1 {
				t.Errorf("unexpected gateway list size: got=%d want=1", len(item.Gateways))
				return false
			}

			actual := item.Gateways[0]
			expected, ok := expectedByID[actual.GatewayID]
			if !ok {
				t.Errorf("unexpected gateway id in response: %s", actual.GatewayID)
				return false
			}
			if actual.GatewayName != expected.GatewayName || actual.TenantID != expected.TenantID || string(actual.Status) != expected.Status {
				t.Errorf("gateway payload mismatch: got=%+v want=%+v", actual, expected)
				return false
			}
			if actual.Interval != expected.Interval {
				t.Errorf("unexpected interval for %s: got=%d want=%d", actual.GatewayID, actual.Interval, expected.Interval)
				return false
			}
			if expected.PublicIdentifier == nil {
				if actual.PublicIdentifier != nil {
					t.Errorf("unexpected public_identifier for %s: got=%v want=nil", actual.GatewayID, actual.PublicIdentifier)
					return false
				}
			} else if actual.PublicIdentifier == nil || *actual.PublicIdentifier != *expected.PublicIdentifier {
				t.Errorf("unexpected public_identifier for %s: got=%v want=%v", actual.GatewayID, actual.PublicIdentifier, expected.PublicIdentifier)
				return false
			}

			delete(expectedByID, actual.GatewayID)
		}

		if len(expectedByID) != 0 {
			t.Errorf("missing gateways in response: %v", expectedByID)
			return false
		}

		return true
	}
}

func TestGatewayController_GetGateway(t *testing.T) {
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

	unauthorizedGatewayID := uuid.NewString()
	notFoundGatewayID := uuid.NewString()
	successGatewayID := uuid.NewString()

	unauthorizedGatewayName := "gateway-get-unauthorized-" + uuid.NewString()
	successGatewayName := "gateway-get-success-" + uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			Name:   "Fail: invalid gateway id",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/not-a-uuid",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid UUID",
		},
		{
			Name:   "Fail: gateway not found",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + notFoundGatewayID,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayNotExistsByID(notFoundGatewayID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(unauthorizedGatewayID, unauthorizedGatewayName, 5000, gateway.GATEWAY_STATUS_ACTIVE, nil, nil),
			},
			Name:   "Fail: tenant admin cannot read gateway",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + unauthorizedGatewayID,
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: identity.ErrUnauthorizedAccess.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(unauthorizedGatewayID, gateway.GATEWAY_STATUS_ACTIVE, nil, 5000),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(unauthorizedGatewayID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateGatewayWithState(successGatewayID, successGatewayName, 6000, gateway.GATEWAY_STATUS_ACTIVE, nil, nil),
			},
			Name:   "Success: super admin reads gateway",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + successGatewayID,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: successGatewayID,
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayResponseMatchesExpected(t, gatewayHTTPResponse{
					GatewayID:        successGatewayID,
					GatewayName:      successGatewayName,
					TenantID:         "",
					Status:           string(gateway.GATEWAY_STATUS_ACTIVE),
					Interval:         6000,
					PublicIdentifier: nil,
				}),
				checkGatewayState(successGatewayID, gateway.GATEWAY_STATUS_ACTIVE, nil, 6000),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(successGatewayID),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func TestGatewayController_GetAllGateways(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, 1)
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantID := uuid.New()
	tenantIDStr := tenantID.String()
	otherTenantID := uuid.New()
	otherTenantJWT, err := helper.NewTenantAdminJWT(deps, otherTenantID, 1000)
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	firstGatewayID := uuid.NewString()
	secondGatewayID := uuid.NewString()
	firstGatewayName := "gateway-list-a-" + uuid.NewString()
	secondGatewayName := "gateway-list-b-" + uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
				preSetupCreateGatewayWithState(firstGatewayID, firstGatewayName, 7000, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, nil),
				preSetupCreateGatewayWithState(secondGatewayID, secondGatewayName, 8000, gateway.GATEWAY_STATUS_DECOMMISSIONED, &tenantIDStr, nil),
			},
			Name:   "Success: super admin gets all gateways",
			Method: http.MethodGet,
			Path:   "/api/v1/gateways",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, gatewayListRequest{Page: 1, Limit: 10}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "gateways",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayListResponseMatchesExpected(t, 1, 2, []gatewayHTTPResponse{
					{GatewayID: firstGatewayID, GatewayName: firstGatewayName, TenantID: tenantIDStr, Status: string(gateway.GATEWAY_STATUS_ACTIVE), Interval: 7000},
					{GatewayID: secondGatewayID, GatewayName: secondGatewayName, TenantID: tenantIDStr, Status: string(gateway.GATEWAY_STATUS_DECOMMISSIONED), Interval: 8000},
				}),
				checkGatewayState(firstGatewayID, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, 7000),
				checkGatewayState(secondGatewayID, gateway.GATEWAY_STATUS_DECOMMISSIONED, &tenantIDStr, 8000),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(firstGatewayID),
				postSetupComposite(
					postSetupDeleteGatewayByID(secondGatewayID),
					integration.PostSetupDeleteTenant(t, tenantID),
				),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
				preSetupCreateGatewayWithState(firstGatewayID, firstGatewayName, 7000, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, nil),
			},
			Name:   "Fail: tenant admin cannot list all gateways",
			Method: http.MethodGet,
			Path:   "/api/v1/gateways",
			Header: integration.AuthHeader(otherTenantJWT),
			Body:   helper.MustJSONBody(t, gatewayListRequest{Page: 1, Limit: 10}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: identity.ErrUnauthorizedAccess.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(firstGatewayID, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, 7000),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(firstGatewayID),
				integration.PostSetupDeleteTenant(t, tenantID),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func TestGatewayController_GetGatewaysByTenant(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, 1)
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantID := uuid.New()
	tenantIDStr := tenantID.String()
	otherTenantID := uuid.New()
	otherTenantJWT, err := helper.NewTenantAdminJWT(deps, otherTenantID, 2000)
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}
	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenantID, 2001)
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	firstGatewayID := uuid.NewString()
	secondGatewayID := uuid.NewString()
	firstGatewayName := "gateway-by-tenant-a-" + uuid.NewString()
	secondGatewayName := "gateway-by-tenant-b-" + uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			Name:   "Fail: invalid tenant id",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/not-a-uuid/gateways?page=1&limit=10",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid UUID",
		},
		{
			Name:   "Fail: tenant not found",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + uuid.NewString() + "/gateways?page=1&limit=10",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: "tenant not found",
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
				preSetupCreateGatewayWithState(firstGatewayID, firstGatewayName, 9000, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, nil),
				preSetupCreateGatewayWithState(secondGatewayID, secondGatewayName, 10000, gateway.GATEWAY_STATUS_INACTIVE, &tenantIDStr, nil),
			},
			Name:   "Success: super admin gets gateways by tenant",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantIDStr + "/gateways?page=1&limit=10",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "gateways",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayListResponseMatchesExpected(t, 1, 2, []gatewayHTTPResponse{
					{GatewayID: firstGatewayID, GatewayName: firstGatewayName, TenantID: tenantIDStr, Status: string(gateway.GATEWAY_STATUS_ACTIVE), Interval: 9000},
					{GatewayID: secondGatewayID, GatewayName: secondGatewayName, TenantID: tenantIDStr, Status: string(gateway.GATEWAY_STATUS_INACTIVE), Interval: 10000},
				}),
				checkGatewayState(firstGatewayID, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, 9000),
				checkGatewayState(secondGatewayID, gateway.GATEWAY_STATUS_INACTIVE, &tenantIDStr, 10000),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(firstGatewayID),
				postSetupComposite(
					postSetupDeleteGatewayByID(secondGatewayID),
					integration.PostSetupDeleteTenant(t, tenantID),
				),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
				preSetupCreateGatewayWithState(firstGatewayID, firstGatewayName, 9000, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, nil),
			},
			Name:   "Fail: tenant admin from another tenant cannot list gateways",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantIDStr + "/gateways?page=1&limit=10",
			Header: integration.AuthHeader(otherTenantJWT),
			Body:   nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: identity.ErrUnauthorizedAccess.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayState(firstGatewayID, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, 9000),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(firstGatewayID),
				integration.PostSetupDeleteTenant(t, tenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
				preSetupCreateGatewayWithState(firstGatewayID, firstGatewayName, 9000, gateway.GATEWAY_STATUS_ACTIVE, &tenantIDStr, nil),
				preSetupCreateGatewayWithState(secondGatewayID, secondGatewayName, 10000, gateway.GATEWAY_STATUS_INACTIVE, &tenantIDStr, nil),
			},
			Name:   "Success: tenant admin gets gateways by tenant",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantIDStr + "/gateways?page=1&limit=10",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "gateways",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayListResponseMatchesExpected(t, 1, 2, []gatewayHTTPResponse{
					{GatewayID: firstGatewayID, GatewayName: firstGatewayName, TenantID: tenantIDStr, Status: string(gateway.GATEWAY_STATUS_ACTIVE), Interval: 9000},
					{GatewayID: secondGatewayID, GatewayName: secondGatewayName, TenantID: tenantIDStr, Status: string(gateway.GATEWAY_STATUS_INACTIVE), Interval: 10000},
				}),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteGatewayByID(firstGatewayID),
				postSetupComposite(
					postSetupDeleteGatewayByID(secondGatewayID),
					integration.PostSetupDeleteTenant(t, tenantID),
				),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}
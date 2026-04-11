package gateway_integrationtests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/gateway"
	"backend/internal/tenant"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gatewayListHTTPResponse struct {
	Count    uint                  `json:"count"`
	Total    uint                  `json:"total"`
	Gateways []gatewayHTTPResponse `json:"gateways"`
}

func TestGetGatewayIntegration(t *testing.T) {
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

	targetTenantID := uuid.NewString()
	gatewayUnauthorized := uuid.NewString()
	gatewaySuccess := uuid.NewString()
	gatewayNotFound := uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			Name:       "JWT non valido",
			Method:     http.MethodGet,
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
			Name:       "gateway_id non valido",
			Method:     http.MethodGet,
			Path:       "/api/v1/gateway/not-a-uuid",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid UUID length",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "utente non super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + gatewayUnauthorized,
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(targetTenantID, "Tenant GET Gateway Unauthorized", true),
				preSetupCreateGatewayWithState(gatewayUnauthorized, "Gateway Unauthorized Get", 5000, gateway.GATEWAY_STATUS_ACTIVE, &targetTenantID, nil),
			},
			PostSetups: postSetupsWithFinal(2,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayUnauthorized),
					postSetupDeleteTenantByID(targetTenantID),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: gateway.ErrUnauthorizedAccess.Error(),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGatewayExistsByID(gatewayUnauthorized),
			},
		},
		{
			Name:       "gateway non trovato",
			Method:     http.MethodGet,
			Path:       "/api/v1/gateway/" + gatewayNotFound,
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: gateway.ErrGatewayNotFound.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "success super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/gateway/" + gatewaySuccess,
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(targetTenantID, "Tenant GET Gateway Success", true),
				preSetupCreateGatewayWithState(gatewaySuccess, "Gateway Success Get", 6500, gateway.GATEWAY_STATUS_ACTIVE, &targetTenantID, nil),
			},
			PostSetups: postSetupsWithFinal(2,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewaySuccess),
					postSetupDeleteTenantByID(targetTenantID),
				),
			),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"gateway_id\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetGatewayResponse(gatewaySuccess, "Gateway Success Get", targetTenantID, gateway.GATEWAY_STATUS_ACTIVE, 6500),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func TestGetAllGatewaysIntegration(t *testing.T) {
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

	tenantForDataID := uuid.NewString()
	gatewayA := uuid.NewString()
	gatewayB := uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			Name:       "JWT non valido",
			Method:     http.MethodGet,
			Path:       "/api/v1/gateways",
			Header:     integration.AuthHeader("invalid.jwt.token"),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "paginazione non valida",
			Method:     http.MethodGet,
			Path:       "/api/v1/gateways?page=0&limit=5",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid format",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "utente non super admin",
			Method:     http.MethodGet,
			Path:       "/api/v1/gateways?page=1&limit=5",
			Header:     integration.AuthHeader(tenantAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: gateway.ErrUnauthorizedAccess.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "success super admin",
			Method: http.MethodGet,
			Path:   "/api/v1/gateways?page=1&limit=10",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantForDataID, "Tenant GET All", true),
				preSetupCreateGatewayWithState(gatewayA, "Gateway List A", 1000, gateway.GATEWAY_STATUS_ACTIVE, &tenantForDataID, nil),
				preSetupCreateGatewayWithState(gatewayB, "Gateway List B", 2000, gateway.GATEWAY_STATUS_INACTIVE, &tenantForDataID, nil),
			},
			PostSetups: postSetupsWithFinal(3,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayA),
					postSetupDeleteGatewayByID(gatewayB),
					postSetupDeleteTenantByID(tenantForDataID),
				),
			),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"gateways\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetAllGatewaysResponse([]string{gatewayA, gatewayB}),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func TestGetGatewaysByTenantIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, 1)
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantTargetID := uuid.NewString()
	tenantOtherID := uuid.NewString()

	tenantAdminTargetJWT, err := helper.NewTenantAdminJWT(deps, uuid.MustParse(tenantTargetID), 100)
	if err != nil {
		t.Fatalf("failed to generate target tenant admin JWT: %v", err)
	}

	tenantAdminOtherJWT, err := helper.NewTenantAdminJWT(deps, uuid.MustParse(tenantOtherID), 101)
	if err != nil {
		t.Fatalf("failed to generate other tenant admin JWT: %v", err)
	}

	gatewayTargetA := uuid.NewString()
	gatewayTargetB := uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			Name:       "JWT non valido",
			Method:     http.MethodGet,
			Path:       "/api/v1/tenant/" + uuid.NewString() + "/gateways",
			Header:     integration.AuthHeader("invalid.jwt.token"),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: "",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "tenant_id non valido",
			Method:     http.MethodGet,
			Path:       "/api/v1/tenant/not-a-uuid/gateways",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "invalid UUID length",
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:       "tenant non trovato",
			Method:     http.MethodGet,
			Path:       "/api/v1/tenant/" + uuid.NewString() + "/gateways",
			Header:     integration.AuthHeader(superAdminJWT),
			Body:       nil,
			PreSetups:  []helper.IntegrationTestPreSetup{},
			PostSetups: []helper.IntegrationTestPostSetup{},

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: tenant.ErrTenantNotFound.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "tenant admin su tenant differente",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantTargetID + "/gateways",
			Header: integration.AuthHeader(tenantAdminOtherJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantTargetID, "Tenant Target Unauthorized", false),
				preSetupCreateTenant(tenantOtherID, "Tenant Other Unauthorized", false),
			},
			PostSetups: postSetupsWithFinal(2,
				postSetupComposite(
					postSetupDeleteTenantByID(tenantTargetID),
					postSetupDeleteTenantByID(tenantOtherID),
				),
			),

			WantStatusCode:   http.StatusInternalServerError,
			WantResponseBody: gateway.ErrUnauthorizedAccess.Error(),
			ResponseChecks:   []helper.IntegrationTestCheck{},
		},
		{
			Name:   "success tenant admin stesso tenant",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/" + tenantTargetID + "/gateways?page=1&limit=10",
			Header: integration.AuthHeader(tenantAdminTargetJWT),
			Body:   nil,
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantTargetID, "Tenant Target Success", false),
				preSetupCreateGatewayWithState(gatewayTargetA, "Gateway Tenant A", 1200, gateway.GATEWAY_STATUS_ACTIVE, &tenantTargetID, nil),
				preSetupCreateGatewayWithState(gatewayTargetB, "Gateway Tenant B", 2200, gateway.GATEWAY_STATUS_INACTIVE, &tenantTargetID, nil),
			},
			PostSetups: postSetupsWithFinal(3,
				postSetupComposite(
					postSetupDeleteGatewayByID(gatewayTargetA),
					postSetupDeleteGatewayByID(gatewayTargetB),
					postSetupDeleteTenantByID(tenantTargetID),
				),
			),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"gateways\"",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkGetAllGatewaysResponse([]string{gatewayTargetA, gatewayTargetB}),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

func preSetupCreateTenant(tenantID string, name string, canImpersonate bool) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		entity := tenant.TenantEntity{ID: tenantID, Name: name, CanImpersonate: canImpersonate}
		return db.Create(&entity).Error == nil
	}
}

func postSetupDeleteTenantByID(id string) helper.IntegrationTestPostSetup {
	return func(deps helper.IntegrationTestDeps) {
		db := (*gorm.DB)(deps.CloudDB)
		_ = db.Where("id = ?", id).Delete(&tenant.TenantEntity{}).Error
	}
}

func checkGetAllGatewaysResponse(expectedGatewayIDs []string) helper.IntegrationTestCheck {
	return func(r *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		_ = deps

		var resp gatewayListHTTPResponse
		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			return false
		}

		found := make(map[string]bool)
		for _, g := range resp.Gateways {
			found[g.GatewayID] = true
		}

		for _, expectedID := range expectedGatewayIDs {
			if !found[expectedID] {
				return false
			}
		}

		if resp.Total < uint(len(expectedGatewayIDs)) {
			return false
		}
		if resp.Count < uint(len(expectedGatewayIDs)) {
			return false
		}

		return true
	}
}

func checkGetGatewayResponse(
	expectedGatewayID string,
	expectedName string,
	expectedTenantID string,
	expectedStatus gateway.GatewayStatus,
	expectedInterval int64,
) helper.IntegrationTestCheck {
	return func(r *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
		_ = deps

		var resp gatewayHTTPResponse
		if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
			return false
		}

		if resp.GatewayID != expectedGatewayID {
			return false
		}
		if resp.GatewayName != expectedName {
			return false
		}
		if resp.TenantID != expectedTenantID {
			return false
		}
		if resp.Status != string(expectedStatus) {
			return false
		}
		if resp.Interval != expectedInterval {
			return false
		}

		return true
	}
}

package tenant_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestGetTenantListIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, uuid.New(), uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	listTenantID1 := uuid.New()
	listTenantID2 := uuid.New()
	listTenantName1 := "tenant-list-a-" + uuid.NewString()
	listTenantName2 := "tenant-list-b-" + uuid.NewString()

	listTenantID3 := uuid.New()
	listTenantID4 := uuid.New()
	listTenantName3 := "tenant-list-c-" + uuid.NewString()
	listTenantName4 := "tenant-list-d-" + uuid.NewString()

	listTenantID5 := uuid.New()
	listTenantID6 := uuid.New()
	listTenantName5 := "tenant-list-e-" + uuid.NewString()
	listTenantName6 := "tenant-list-f-" + uuid.NewString()

	invalidQueryTenantID := uuid.New()
	invalidQueryTenantName := "tenant-list-invalid-query-" + uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(listTenantID1, listTenantName1, true),
				PreSetupCreateTenantWithName(listTenantID2, listTenantName2, false),
			},
			Name:   "Success: super admin gets tenant list",
			Method: http.MethodGet,
			Path:   "/api/v1/tenants?page=1&limit=10",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "tenants",
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckResponseBodyContains(t, listTenantName1, listTenantName2),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, listTenantID1),
				integration.PostSetupDeleteTenant(t, listTenantID2),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(listTenantID3, listTenantName3, true),
				PreSetupCreateTenantWithName(listTenantID4, listTenantName4, true),
			},
			Name:   "Success: pagination page=2 limit=1",
			Method: http.MethodGet,
			Path:   "/api/v1/tenants?page=2&limit=1",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "\"count\":1",
			ResponseChecks:   []helper.IntegrationTestCheck{},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, listTenantID3),
				integration.PostSetupDeleteTenant(t, listTenantID4),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(listTenantID5, listTenantName5, true),
				PreSetupCreateTenantWithName(listTenantID6, listTenantName6, true),
			},
			Name:   "Fail: unauthorized, no JWT",
			Method: http.MethodGet,
			Path:   "/api/v1/tenants?page=1&limit=10",
			Header: http.Header{},
			Body:   nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, listTenantID5),
				CheckTenantExistsByID(t, listTenantID6),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, listTenantID5),
				integration.PostSetupDeleteTenant(t, listTenantID6),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(invalidQueryTenantID, invalidQueryTenantName, true),
			},
			Name:   "Fail: invalid query pagination",
			Method: http.MethodGet,
			Path:   "/api/v1/tenants?page=0&limit=10",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, invalidQueryTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, invalidQueryTenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Fail: tenant admin cannot list tenants",
			Method:    http.MethodGet,
			Path:      "/api/v1/tenants?page=1&limit=10",
			Header:    integration.AuthHeader(tenantAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(identity.ErrUnauthorizedAccess),
			ResponseChecks:   []helper.IntegrationTestCheck{},
			PostSetups:       []helper.IntegrationTestPostSetup{},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

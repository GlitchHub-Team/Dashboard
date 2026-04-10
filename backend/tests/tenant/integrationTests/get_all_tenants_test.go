package tenant_integration_test

import (
	"net/http"
	"testing"

	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestGetAllTenantsIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	tenantID1 := uuid.New()
	tenantID2 := uuid.New()
	tenantName1 := "tenant-all-a-" + uuid.NewString()
	tenantName2 := "tenant-all-b-" + uuid.NewString()

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(tenantID1, tenantName1, true),
				PreSetupCreateTenantWithName(tenantID2, tenantName2, false),
			},
			Name:   "Success: endpoint is public and returns tenants",
			Method: http.MethodGet,
			Path:   "/api/v1/all_tenants",
			Header: http.Header{},
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: "tenants",
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckResponseBodyContains(t, tenantName1, tenantName2),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, tenantID1),
				integration.PostSetupDeleteTenant(t, tenantID2),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

package tenant_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestGetTenantIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	otherTenantID := uuid.New()
	otherTenantAdminJWT, err := helper.NewTenantAdminJWT(deps, otherTenantID, uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	successTenantID := uuid.New()
	successTenantName := "tenant-get-success-" + uuid.NewString()
	noJWTTenantID := uuid.New()
	unauthorizedTenantID := uuid.New()
	notFoundTenantID := uuid.New()

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(successTenantID, successTenantName, true),
			},
			Name:   "Success: get tenant with valid URI and auth",
			Method: http.MethodGet,
			Path:   tenantPath(successTenantID),
			Header: integration.AuthHeader(superAdminJWT),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: successTenantID.String(),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckResponseBodyContains(t, successTenantName),
				CheckTenantExistsByID(t, successTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, successTenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(noJWTTenantID, "tenant-get-no-jwt-"+uuid.NewString(), true),
			},
			Name:   "Fail: unauthorized, no JWT",
			Method: http.MethodGet,
			Path:   tenantPath(noJWTTenantID),
			Header: http.Header{},

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, noJWTTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, noJWTTenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:   "Fail: invalid UUID in URI",
			Method: http.MethodGet,
			Path:   "/api/v1/tenant/not-a-uuid",
			Header: integration.AuthHeader(superAdminJWT),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:   "Fail: tenant not found",
			Method: http.MethodGet,
			Path:   tenantPath(notFoundTenantID),
			Header: integration.AuthHeader(superAdminJWT),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckNoTenantByID(t, notFoundTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(unauthorizedTenantID, "tenant-get-unauthorized-"+uuid.NewString(), true),
			},
			Name:   "Fail: tenant admin from other tenant cannot get",
			Method: http.MethodGet,
			Path:   tenantPath(unauthorizedTenantID),
			Header: integration.AuthHeader(otherTenantAdminJWT),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(identity.ErrUnauthorizedAccess),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, unauthorizedTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, unauthorizedTenantID),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

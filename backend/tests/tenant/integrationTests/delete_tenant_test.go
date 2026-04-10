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

func TestDeleteTenantIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, uuid.New(), uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	successTenantID := uuid.New()
	noJWTTenantID := uuid.New()
	unauthorizedTenantID := uuid.New()
	invalidURITenantID := uuid.New()
	notFoundTenantID := uuid.New()

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(successTenantID, "tenant-delete-success-"+uuid.NewString(), true),
			},
			Name:   "Success: super admin deletes tenant",
			Method: http.MethodDelete,
			Path:   tenantPath(successTenantID),
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusOK,
			WantResponseBody: successTenantID.String(),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckNoTenantByID(t, successTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, successTenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(noJWTTenantID, "tenant-delete-no-jwt-"+uuid.NewString(), true),
			},
			Name:   "Fail: unauthorized, no JWT",
			Method: http.MethodDelete,
			Path:   tenantPath(noJWTTenantID),
			Header: http.Header{},
			Body:   nil,

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
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(unauthorizedTenantID, "tenant-delete-unauthorized-"+uuid.NewString(), true),
			},
			Name:   "Fail: tenant admin cannot delete tenant",
			Method: http.MethodDelete,
			Path:   tenantPath(unauthorizedTenantID),
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(identity.ErrUnauthorizedAccess),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, unauthorizedTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, unauthorizedTenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(invalidURITenantID, "tenant-delete-invalid-uri-"+uuid.NewString(), true),
			},
			Name:   "Fail: invalid tenant_id in URI",
			Method: http.MethodDelete,
			Path:   "/api/v1/tenant/not-a-uuid",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, invalidURITenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, invalidURITenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Fail: tenant not found",
			Method:    http.MethodDelete,
			Path:      tenantPath(notFoundTenantID),
			Header:    integration.AuthHeader(superAdminJWT),
			Body:      nil,

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckNoTenantByID(t, notFoundTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

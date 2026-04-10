package tenant_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
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
	missingBodyTenantID := uuid.New()
	unauthorizedTenantID := uuid.New()
	notFoundTenantID := uuid.New()

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(successTenantID, successTenantName, true),
			},
			Name:   "Success: get tenant with valid body and auth",
			Method: http.MethodGet,
			Path:   tenantPath(successTenantID),
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, tenant.GetTenantDTO{
				TenantIdField: dto.TenantIdField{TenantId: successTenantID.String()},
			}),

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
			Body: helper.MustJSONBody(t, tenant.GetTenantDTO{
				TenantIdField: dto.TenantIdField{TenantId: noJWTTenantID.String()},
			}),

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
				PreSetupCreateTenantWithName(missingBodyTenantID, "tenant-get-missing-body-"+uuid.NewString(), true),
			},
			Name:   "Fail: missing JSON body",
			Method: http.MethodGet,
			Path:   tenantPath(missingBodyTenantID),
			Header: integration.AuthHeader(superAdminJWT),
			Body:   nil,

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, missingBodyTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, missingBodyTenantID),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Fail: tenant not found",
			Method:    http.MethodGet,
			Path:      tenantPath(notFoundTenantID),
			Header:    integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, tenant.GetTenantDTO{
				TenantIdField: dto.TenantIdField{TenantId: notFoundTenantID.String()},
			}),

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
			Body: helper.MustJSONBody(t, tenant.GetTenantDTO{
				TenantIdField: dto.TenantIdField{TenantId: unauthorizedTenantID.String()},
			}),

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

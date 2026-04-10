package tenant_integration_test

import (
	"bytes"
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

func TestCreateTenantIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, uuid.New(), uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	successTenantName := "tenant-create-success-" + uuid.NewString()
	unauthorizedTenantName := "tenant-create-unauthorized-" + uuid.NewString()
	missingJWTTenantName := "tenant-create-missing-jwt-" + uuid.NewString()
	duplicateTenantName := "tenant-create-duplicate-" + uuid.NewString()
	duplicateTenantID := uuid.New()

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{nil},
			Name:      "Success: super admin creates tenant",
			Method:    http.MethodPost,
			Path:      "/api/v1/tenant",
			Header:    integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, tenant.CreateTenantDTO{
				TenantNameField: dto.TenantNameField{TenantName: successTenantName},
				CanImpersonate:  true,
			}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: successTenantName,
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantInsertedByName(t, successTenantName, true),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenantByName(t, successTenantName),
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Fail: unauthorized, no JWT",
			Method:    http.MethodPost,
			Path:      "/api/v1/tenant",
			Header:    http.Header{},
			Body: helper.MustJSONBody(t, tenant.CreateTenantDTO{
				TenantNameField: dto.TenantNameField{TenantName: missingJWTTenantName},
				CanImpersonate:  true,
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckNoTenantByName(t, missingJWTTenantName),
			},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Fail: tenant admin cannot create tenant",
			Method:    http.MethodPost,
			Path:      "/api/v1/tenant",
			Header:    integration.AuthHeader(tenantAdminJWT),
			Body: helper.MustJSONBody(t, tenant.CreateTenantDTO{
				TenantNameField: dto.TenantNameField{TenantName: unauthorizedTenantName},
				CanImpersonate:  true,
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(identity.ErrUnauthorizedAccess),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckNoTenantByName(t, unauthorizedTenantName),
			},
			PostSetups: []helper.IntegrationTestPostSetup{},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Fail: invalid json body",
			Method:    http.MethodPost,
			Path:      "/api/v1/tenant",
			Header:    integration.AuthHeader(superAdminJWT),
			Body:      bytes.NewReader([]byte("{}")),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks:   []helper.IntegrationTestCheck{},
			PostSetups:       []helper.IntegrationTestPostSetup{},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupCreateTenantWithName(duplicateTenantID, duplicateTenantName, true),
			},
			Name:   "Fail: duplicate tenant name",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, tenant.CreateTenantDTO{
				TenantNameField: dto.TenantNameField{TenantName: duplicateTenantName},
				CanImpersonate:  false,
			}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantAlreadyExists),
			ResponseChecks: []helper.IntegrationTestCheck{
				CheckTenantExistsByID(t, duplicateTenantID),
			},
			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, duplicateTenantID),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

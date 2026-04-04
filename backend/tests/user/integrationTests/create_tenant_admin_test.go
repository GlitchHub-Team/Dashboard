package user_integration_test

import (
	"bytes"
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestCreateTenantAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// create a fresh tenant UUID v4 for tests to satisfy binding uuid4
	tenantID := uuid.New()
	tenant2ID := uuid.New()
	missingTenantId := uuid.New()

	PostSetupDeleteTenant := func(tenantId uuid.UUID) helper.IntegrationTestPostSetup {
		return integration.PostSetupDeleteTenant(t, tenantId)
	}

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenantID, uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	tenantAdminJWT_OtherTenant, err := helper.NewTenantAdminJWT(deps, tenant2ID, uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	tenantUserJWT, err := helper.NewTenantUserJWT(deps, tenantID, uint(5))
	if err != nil {
		t.Fatalf("failed to generate tenant user JWT: %v", err)
	}

	existingEmail := "tenant1@user.com"

	validEmail := "newadmin@tenant.test"
	validUsername := "New Tenant User"

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
			},

			Name:   "Success: insert new user in tenant schema",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body: helper.MustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: validEmail},
				UsernameField: dto.UsernameField{Username: validUsername},
			}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: validEmail, // Cerco email utente nel body
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckTenantMemberInserted(validEmail, tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
			},
		},

		// Super admin should be denied when tenant CanImpersonate == false
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, false),
			},
			Name:   "Fail: super admin cannot access tenant when CanImpersonate=false",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "impersonation@t.test"}, UsernameField: dto.UsernameField{Username: "Imp"}}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenantMember("impersonation@t.test", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
			},
			Name:   "Fail: role not allowed (tenant user) cannot create tenant admin",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: integration.AuthHeader(tenantUserJWT),
			Body:   helper.MustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "unauth@t.test"}, UsernameField: dto.UsernameField{Username: "Nope"}}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenantMember("unauth@t.test", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
			},
			Name:   "Fail: Unauthorized access, no JWT",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: http.Header{},
			Body:   helper.MustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "baduser@t.test"}, UsernameField: dto.UsernameField{Username: "Bad User"}}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenantMember("baduser@t.test", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
			},
			Name:   "Fail: URI binding fail",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/invalid-uuid/tenant_admin",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "x@t.test"}, UsernameField: dto.UsernameField{Username: "X"}}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenant("invalid-uuid"),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
			},
			Name:   "Fail: JSON binding fail",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   bytes.NewReader([]byte("{}")),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenantMember("", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: nil,
			Name:      "Fail: Tenant not found",
			Method:    http.MethodPost,
			// use a valid uuid4 that does NOT exist in tenants table
			Path:   "/api/v1/tenant/" + uuid.New().String() + "/tenant_admin",
			Header: integration.AuthHeader(superAdminJWT),
			Body:   helper.MustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "nt@t.test"}, UsernameField: dto.UsernameField{Username: "NT"}}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenantMember("nt@t.test", missingTenantId.String()),
			},

			PostSetups: nil,
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
			},
			Name:   "Fail: tenant mismatch (obfuscated)",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: integration.AuthHeader(tenantAdminJWT_OtherTenant),
			Body:   helper.MustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "mismatch@t.test"}, UsernameField: dto.UsernameField{Username: "MM"}}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenantMember("mismatch@t.test", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, true),
				integration.PreSetupAddTenantAdmin(t, nil, user.TenantMemberEntity{
					TenantId: tenantID.String(),
					Email:    existingEmail,
					Name:     "Pre-existing",
					Role:     string(identity.ROLE_TENANT_ADMIN), // NOTA: irrilevante ma richiesto da check constraint
				}, false),
			},
			Name:   "Fail: User already exists",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body:   helper.MustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: existingEmail}, UsernameField: dto.UsernameField{Username: "Duplicate"}}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: helper.ErrJsonString(user.ErrUserAlreadyExists),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckTenantMemberInserted(existingEmail, tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteTenant(tenantID),
				nil,
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

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

	"github.com/google/uuid"
)

func TestCreateTenantAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// create a fresh tenant UUID v4 for tests to satisfy binding uuid4
	tenantID := uuid.New()
	missingTenantId := uuid.New()

	postSetupDeleteTenant := func(tenantId uuid.UUID) helper.IntegrationTestPostSetup { return postSetupDeleteTenant(t, tenantId) }

	superAdminJWT, err := deps.AuthTokenManager.GenerateForRequester(identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN})
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantAdminJWT, err := deps.AuthTokenManager.GenerateForRequester(identity.Requester{RequesterUserId: 1, RequesterTenantId: &tenantID, RequesterRole: identity.ROLE_TENANT_ADMIN})
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	existingEmail := "tenant1@user.com"

	validEmail := "newadmin@tenant.test"
	validUsername := "New Tenant User"

	tests := []helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantID),
			},

			Name:   "Success: insert new user in tenant schema",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: authHeader(tenantAdminJWT),
			Body: mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: validEmail},
				UsernameField: dto.UsernameField{Username: validUsername},
			}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: validEmail, // Cerco email utente nel body
			ResponseChecks: []helper.IntegrationTestCheck{
				checkTenantMemberInserted(validEmail, tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantID),
			},
			Name:   "Fail: Unauthorized access, no JWT",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: http.Header{},
			Body:   mustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "baduser@t.test"}, UsernameField: dto.UsernameField{Username: "Bad User"}}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoTenantMember("baduser@t.test", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantID),
			},
			Name:   "Binding URI fallito ritorna errore di validazione",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/invalid-uuid/tenant_admin",
			Header: authHeader(superAdminJWT),
			Body:   mustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "x@t.test"}, UsernameField: dto.UsernameField{Username: "X"}}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoTenant("invalid-uuid"),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantID),
			},
			Name:   "Binding JSON fallito ritorna errore di validazione",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: authHeader(superAdminJWT),
			Body:   bytes.NewReader([]byte("{}")),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoTenantMember("", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: nil,
			Name:      "Tenant non trovato ritorna 404",
			Method:    http.MethodPost,
			// use a valid uuid4 that does NOT exist in tenants table
			Path:   "/api/v1/tenant/" + uuid.New().String() + "/tenant_admin",
			Header: authHeader(superAdminJWT),
			Body:   mustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "nt@t.test"}, UsernameField: dto.UsernameField{Username: "NT"}}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoTenantMember("nt@t.test", missingTenantId.String()),
			},

			PostSetups: nil,
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantID),
			},
			Name:   "Tenant mismatch ritorna 404 e non inserisce",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: func() http.Header {
				tenant2ID := uuid.New()
				jwt, _ := deps.AuthTokenManager.GenerateForRequester(identity.Requester{RequesterUserId: 1, RequesterTenantId: &tenant2ID, RequesterRole: identity.ROLE_TENANT_ADMIN})
				return authHeader(jwt)
			}(),
			Body: mustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: "mismatch@t.test"}, UsernameField: dto.UsernameField{Username: "MM"}}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoTenantMember("mismatch@t.test", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteTenant(tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantID),
				preSetupAddTenantMember(tenantID, &user.TenantMemberEntity{
					Email: existingEmail,
					Name:  "Pre-existing",
					Role:  string(identity.ROLE_TENANT_USER), // NOTA: irrilevante ma richiesto da check constraint
				}),
			},
			Name:   "Utente esistente ritorna 400 e non duplica",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_admin",
			Header: authHeader(tenantAdminJWT),
			Body:   mustJSONBody(t, user.CreateUserBodyDTO{EmailField: dto.EmailField{Email: existingEmail}, UsernameField: dto.UsernameField{Username: "Duplicate"}}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: helper.ErrJsonString(user.ErrUserAlreadyExists),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkTenantMemberInserted(existingEmail, tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteTenant(tenantID),
				nil,
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

package user_integration_test

import (
	"bytes"
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/tenant"

	// "backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"

	"github.com/google/uuid"
	// "github.com/google/uuid"
)

func TestCreateSuperAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// create a fresh tenant UUID v4 for tests to satisfy binding uuid4
	tenantID := uuid.New()
	// missingTenantId := uuid.New()

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenantID, uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}

	tenantUserJWT, err := helper.NewTenantUserJWT(deps, tenantID, uint(5))
	if err != nil {
		t.Fatalf("failed to generate tenant user JWT: %v", err)
	}

	existingEmail := "existing.superadmin@m31.com"

	validEmail := "superadmin@m31.com"
	validUsername := "Super Admin"

	tests := []*helper.IntegrationTestCase{
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				nil,
			},

			Name:   "Success: insert new user",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: authHeader(superAdminJWT),
			Body: mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: validEmail},
				UsernameField: dto.UsernameField{Username: validUsername},
			}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: validEmail, // Cerco email utente nel body
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSuperAdminInserted(validEmail),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteSuperAdmin(validEmail),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupCreateTenant(tenantID, false),
			},

			Name:   "Fail: super admin denied when tenant CanImpersonate=false",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_user",
			Header: authHeader(superAdminJWT),
			Body: mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "impersonation@t.test"},
				UsernameField: dto.UsernameField{Username: "Imp"},
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoTenantMember("impersonation@t.test", tenantID.String()),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteTenant(t, tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				nil,
			},

			Name:   "Fail: role not allowed (tenant admin) cannot create super admin",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: authHeader(tenantAdminJWT),
			Body: mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "unauthorized@t.test"},
				UsernameField: dto.UsernameField{Username: "Nope"},
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSuperAdmin("unauthorized@t.test"),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				nil,
			},
		},
		{
			PreSetups: []helper.IntegrationTestPreSetup{
				nil,
			},

			Name:   "Fail: role not allowed (tenant user) cannot create super admin",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: authHeader(tenantUserJWT),
			Body: mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "unauthorized@t.test"},
				UsernameField: dto.UsernameField{Username: "Nope"},
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSuperAdmin("unauthorized@t.test"),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				nil,
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				nil,
			},
			Name:   "Fail: Unauthorized access, no JWT",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: http.Header{},
			Body: mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "baduser@t.test"},
				UsernameField: dto.UsernameField{Username: "Bad User"},
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSuperAdmin("baduser@t.test"),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteSuperAdmin(validEmail),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Binding JSON fallito ritorna errore di validazione",
			Method:    http.MethodPost,
			Path:      "/api/v1/super_admin",
			Header:    authHeader(superAdminJWT),
			Body:      bytes.NewReader([]byte("{}")),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSuperAdmin(""),
			},

			PostSetups: []helper.IntegrationTestPostSetup{},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				PreSetupAddSuperAdmin(t, nil, user.SuperAdminEntity{
					Email: existingEmail,
					Name:  "Pre-existing",
				}, false),
			},
			Name:   "Fail: User already exists",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: authHeader(superAdminJWT),
			Body: mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: existingEmail},
				UsernameField: dto.UsernameField{Username: "Duplicate"},
			}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: helper.ErrJsonString(user.ErrUserAlreadyExists),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSuperAdminInserted(existingEmail),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				PostSetupDeleteSuperAdmin(existingEmail),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

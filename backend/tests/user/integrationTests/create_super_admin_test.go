package user_integration_test

import (
	"bytes"
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/tenant"

	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestCreateSuperAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// create a fresh tenant UUID v4 for tests to satisfy binding uuid
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
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: validEmail},
				UsernameField: dto.UsernameField{Username: validUsername},
			}),

			WantStatusCode:   http.StatusOK,
			WantResponseBody: validEmail, // Cerco email utente nel body
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckSuperAdminInserted(validEmail),
				integration.CheckCountSuperAdminConfirmAccountTokens(t, 1),
				integration.CheckSMTPMessageForToken(t, "confirm_account", true),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteSuperAdmin(validEmail),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupCreateTenant(tenantID, false),
			},

			Name:   "Fail: super admin denied when tenant CanImpersonate=false",
			Method: http.MethodPost,
			Path:   "/api/v1/tenant/" + tenantID.String() + "/tenant_user",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "impersonation@t.test"},
				UsernameField: dto.UsernameField{Username: "Imp"},
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoTenantMember("impersonation@t.test", tenantID.String()),
				integration.CheckSMTPMessageForToken(t, "confirm_account", false),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteTenant(t, tenantID),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				nil,
			},

			Name:   "Fail: role not allowed (tenant admin) cannot create super admin",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: integration.AuthHeader(tenantAdminJWT),
			Body: helper.MustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "unauthorized@t.test"},
				UsernameField: dto.UsernameField{Username: "Nope"},
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoSuperAdmin("unauthorized@t.test"),
				integration.CheckSMTPMessageForToken(t, "confirm_account", false),
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
			Header: integration.AuthHeader(tenantUserJWT),
			Body: helper.MustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "unauthorized@t.test"},
				UsernameField: dto.UsernameField{Username: "Nope"},
			}),

			WantStatusCode:   http.StatusNotFound,
			WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoSuperAdmin("unauthorized@t.test"),
				integration.CheckSMTPMessageForToken(t, "confirm_account", false),
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
			Body: helper.MustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: "baduser@t.test"},
				UsernameField: dto.UsernameField{Username: "Bad User"},
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoSuperAdmin("baduser@t.test"),
				integration.CheckSMTPMessageForToken(t, "confirm_account", false),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteSuperAdmin(validEmail),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{},
			Name:      "Binding JSON fallito ritorna errore di validazione",
			Method:    http.MethodPost,
			Path:      "/api/v1/super_admin",
			Header:    integration.AuthHeader(superAdminJWT),
			Body:      bytes.NewReader([]byte("{}")),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckNoSuperAdmin(""),
				integration.CheckSMTPMessageForToken(t, "confirm_account", false),
			},

			PostSetups: []helper.IntegrationTestPostSetup{},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				integration.PreSetupAddSuperAdmin(t, nil, user.SuperAdminEntity{
					Email: existingEmail,
					Name:  "Pre-existing",
				}, false),
			},
			Name:   "Fail: User already exists",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: integration.AuthHeader(superAdminJWT),
			Body: helper.MustJSONBody(t, user.CreateUserBodyDTO{
				EmailField:    dto.EmailField{Email: existingEmail},
				UsernameField: dto.UsernameField{Username: "Duplicate"},
			}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: helper.ErrJsonString(user.ErrUserAlreadyExists),
			ResponseChecks: []helper.IntegrationTestCheck{
				integration.CheckSuperAdminInserted(existingEmail),
				integration.CheckSMTPMessageForToken(t, "confirm_account", false),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				integration.PostSetupDeleteSuperAdmin(existingEmail),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

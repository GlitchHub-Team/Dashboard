package user_integration_test

import (
	"bytes"
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/infra/transport/http/dto"
	"backend/internal/shared/identity"

	// "backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"

	// "github.com/google/uuid"
)

func TestCreateSuperAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// create a fresh tenant UUID v4 for tests to satisfy binding uuid4
	// tenantID := uuid.New()
	// missingTenantId := uuid.New()

	superAdminJWT, err := deps.AuthTokenManager.GenerateForRequester(identity.Requester{RequesterUserId: 1, RequesterRole: identity.ROLE_SUPER_ADMIN})
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}

	// tenantAdminJWT, err := deps.AuthTokenManager.GenerateForRequester(identity.Requester{RequesterUserId: 1, RequesterTenantId: &tenantID, RequesterRole: identity.ROLE_TENANT_ADMIN})
	// if err != nil {
	// 	t.Fatalf("failed to generate tenant admin JWT: %v", err)
	// }

	existingEmail := "existing.superadmin@m31.com"

	validEmail := "superadmin@m31.com"
	validUsername := "Super Admin"

	tests := []helper.IntegrationTestCase{
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
				postSetupDeleteSuperAdmin(validEmail),
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
			Body:   mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField: dto.EmailField{Email: "baduser@t.test"}, 
				UsernameField: dto.UsernameField{Username: "Bad User"},
			}),

			WantStatusCode:   http.StatusUnauthorized,
			WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSuperAdmin("baduser@t.test"),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSuperAdmin(validEmail),
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
			},
			Name:   "Binding JSON fallito ritorna errore di validazione",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: authHeader(superAdminJWT),
			Body:   bytes.NewReader([]byte("{}")),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: "error",
			ResponseChecks: []helper.IntegrationTestCheck{
				checkNoSuperAdmin(""),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
			},
		},

		{
			PreSetups: []helper.IntegrationTestPreSetup{
				preSetupAddSuperAdmin(&user.SuperAdminEntity{
					Email: existingEmail,
					Name:  "Pre-existing",
				}),
			},
			Name:   "Utente esistente ritorna 400 e non duplica",
			Method: http.MethodPost,
			Path:   "/api/v1/super_admin",
			Header: authHeader(superAdminJWT),
			Body:   mustJSONBody(t, user.CreateUserBodyDTO{
				EmailField: dto.EmailField{Email: existingEmail}, 
				UsernameField: dto.UsernameField{Username: "Duplicate"},
			}),

			WantStatusCode:   http.StatusBadRequest,
			WantResponseBody: helper.ErrJsonString(user.ErrUserAlreadyExists),
			ResponseChecks: []helper.IntegrationTestCheck{
				checkSuperAdminInserted(existingEmail),
			},

			PostSetups: []helper.IntegrationTestPostSetup{
				postSetupDeleteSuperAdmin(existingEmail),
			},
		},
	}

	helper.RunIntegrationTests(t, tests, deps)
}

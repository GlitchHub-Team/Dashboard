package user_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"
	"backend/internal/user"
	"backend/tests/helper"

	"github.com/google/uuid"
)

func TestGetSuperAdminsIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, _ := helper.NewSuperAdminJWT(deps, uint(1))
	tenantUserJWT, _ := helper.NewTenantUserJWT(deps, uuid.New(), uint(3))

	existingEmail1 := "superadmin1@domain.test"
	existingEmail2 := "superadmin2@domain.test"

	superAdmin1Entity := user.SuperAdminEntity{Email: existingEmail1, Name: "S1"}
	superAdmin2Entity := user.SuperAdminEntity{Email: existingEmail2, Name: "S2"}

	tests := make([]*helper.IntegrationTestCase, 0)

	// Success default pagination
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, nil, superAdmin1Entity, false),
			PreSetupAddSuperAdmin(t, nil, superAdmin2Entity, false),
		},
		Name:   "Success: default pagination",
		Method: http.MethodGet,
		Path:   "/api/v1/super_admins",
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "\"count\":2",
		ResponseChecks: []helper.IntegrationTestCheck{
			checkSuperAdminInserted(existingEmail1),
			checkSuperAdminInserted(existingEmail2),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdmin(existingEmail1),
			PostSetupDeleteSuperAdmin(existingEmail2),
		},
	})

	// Success custom pagination page=2&limit=1
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, nil, superAdmin1Entity, false),
			PreSetupAddSuperAdmin(t, nil, superAdmin2Entity, false),
		},
		Name:   "Success: custom pagination",
		Method: http.MethodGet,
		Path:   "/api/v1/super_admins?page=2&limit=1",
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "\"count\":1",
		ResponseChecks:   []helper.IntegrationTestCheck{checkSuperAdminInserted(existingEmail1), checkSuperAdminInserted(existingEmail2)},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdmin(existingEmail1),
			PostSetupDeleteSuperAdmin(existingEmail2),
		},
	})

	// Unauthorized no JWT
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, nil, superAdmin1Entity, false),
			PreSetupAddSuperAdmin(t, nil, superAdmin2Entity, false),
		},
		Name:   "Fail: Unauthorized, no JWT",
		Method: http.MethodGet,
		Path:   "/api/v1/super_admins",
		Header: http.Header{},
		Body:   nil,

		WantStatusCode:   http.StatusUnauthorized,
		WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
		ResponseChecks:   []helper.IntegrationTestCheck{checkSuperAdminInserted(existingEmail1), checkSuperAdminInserted(existingEmail2)},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdmin(existingEmail1),
			PostSetupDeleteSuperAdmin(existingEmail2),
		},
	})

	// Query invalid page=-1
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, nil, superAdmin1Entity, false),
		},
		Name:   "Fail: Query binding invalid",
		Method: http.MethodGet,
		Path:   "/api/v1/super_admins?page=-1",
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks:   []helper.IntegrationTestCheck{checkSuperAdminInserted(existingEmail1)},
		PostSetups:       []helper.IntegrationTestPostSetup{PostSetupDeleteSuperAdmin(existingEmail1)},
	})

	// Unauthorized roles: tenant user should get 401
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, nil, superAdmin1Entity, false),
		},
		Name:   "Fail: tenant user cannot list superadmins",
		Method: http.MethodGet,
		Path:   "/api/v1/super_admins",
		Header: authHeader(tenantUserJWT),
		Body:   nil,

		WantStatusCode:   http.StatusUnauthorized,
		WantResponseBody: helper.ErrJsonString(identity.ErrUnauthorizedAccess),
		ResponseChecks:   []helper.IntegrationTestCheck{checkSuperAdminInserted(existingEmail1)},
		PostSetups:       []helper.IntegrationTestPostSetup{PostSetupDeleteSuperAdmin(existingEmail1)},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

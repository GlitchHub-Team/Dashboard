package user_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/user"
	"backend/tests/helper"

	"github.com/google/uuid"
)

func TestDeleteSuperAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}
	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, uuid.New(), uint(2))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}
	tenantUserJWT, err := helper.NewTenantUserJWT(deps, uuid.New(), uint(3))
	if err != nil {
		t.Fatalf("failed to generate tenant user JWT: %v", err)
	}

	existingEmail1 := "superadmin1@domain.test"
	existingEmail2 := "superadmin2@domain.test"

	existingSuperAdmin1Entity := user.SuperAdminEntity{
		Email: existingEmail1,
		Name:  "Existing Super Admin",
	}

	existingSuperAdmin2Entity := user.SuperAdminEntity{
		Email: existingEmail2,
		Name:  "Existing Super Admin",
	}

	tests := make([]*helper.IntegrationTestCase, 0)

	// Success
	var tcSuccess helper.IntegrationTestCase
	tcSuccess = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, &tcSuccess, existingSuperAdmin1Entity, true),
			PreSetupAddSuperAdmin(t, &tcSuccess, existingSuperAdmin2Entity, true),
		},
		Name:   "Success: delete existing super admin",
		Method: http.MethodDelete,
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: existingEmail2,
		ResponseChecks: []helper.IntegrationTestCheck{
			checkNoSuperAdmin(existingEmail2),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdmin(existingEmail1),
			PostSetupDeleteSuperAdmin(existingEmail2),
		},
	}
	tests = append(tests, &tcSuccess)

	// Unauthorized no JWT
	var tcNoJwt helper.IntegrationTestCase
	tcNoJwt = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, &tcNoJwt, existingSuperAdmin1Entity, true),
		},
		Name:   "Fail: Unauthorized access, no JWT",
		Method: http.MethodDelete,
		Header: http.Header{},
		Body:   nil,

		WantStatusCode:   http.StatusUnauthorized,
		WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkSuperAdminInserted(existingEmail1),
		},
		PostSetups: []helper.IntegrationTestPostSetup{PostSetupDeleteSuperAdmin(existingEmail1)},
	}
	tests = append(tests, &tcNoJwt)

	// URI invalid (non-numeric)
	tcInvalidUri := helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "Fail: URI binding invalid",
		Method:    http.MethodDelete,
		Path:      "/api/v1/super_admin/invalid-id",
		Header:    authHeader(superAdminJWT),
		Body:      nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks: []helper.IntegrationTestCheck{
			checkNoSuperAdmin(existingEmail1),
		},
		PostSetups: []helper.IntegrationTestPostSetup{},
	}
	tests = append(tests, &tcInvalidUri)

	// Unauthorized roles (tenant user) should be obfuscated as not found
	var tcUnauthTenantUser helper.IntegrationTestCase
	tcUnauthTenantUser = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, &tcUnauthTenantUser, existingSuperAdmin1Entity, true),
		},
		Name:   "Fail: tenant roles cannot delete super admin",
		Method: http.MethodDelete,
		Header: authHeader(tenantUserJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkSuperAdminInserted(existingEmail1),
		},
		PostSetups: []helper.IntegrationTestPostSetup{PostSetupDeleteSuperAdmin(existingEmail1)},
	}
	tests = append(tests, &tcUnauthTenantUser)

	// Unauthorized roles (tenant admin) should be obfuscated as not found
	var tcUnauthTenantAdmin helper.IntegrationTestCase
	tcUnauthTenantAdmin = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, &tcUnauthTenantAdmin, existingSuperAdmin1Entity, true),
		},
		Name:   "Fail: tenant roles cannot delete super admin",
		Method: http.MethodDelete,
		Header: authHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkSuperAdminInserted(existingEmail1),
		},
		PostSetups: []helper.IntegrationTestPostSetup{PostSetupDeleteSuperAdmin(existingEmail1)},
	}
	tests = append(tests, &tcUnauthTenantAdmin)

	// User not found
	tcUserNotFound := helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdmin(t, &tcNoJwt, existingSuperAdmin1Entity, true),
			PreSetupAddSuperAdmin(t, &tcNoJwt, existingSuperAdmin2Entity, true),
		},
		Name:   "Fail: user not found",
		Method: http.MethodDelete,
		Path:   "/api/v1/super_admin/999999",
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdmin(existingEmail1),
			PostSetupDeleteSuperAdmin(existingEmail2),
		},
	}

	tests = append(tests, &tcUserNotFound)

	helper.RunIntegrationTests(t, tests, deps)
}

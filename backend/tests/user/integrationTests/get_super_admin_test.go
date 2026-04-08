package user_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestGetSuperAdminIntegration(t *testing.T) {
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

	existingEmail1 := "getsuper1@domain.test"
	existingEmail2 := "getsuper2@domain.test"

	existingSuperAdmin1Entity := user.SuperAdminEntity{
		Email: existingEmail1,
		Name:  "Super One",
	}

	existingSuperAdmin2Entity := user.SuperAdminEntity{
		Email: existingEmail2,
		Name:  "Super Two",
	}

	tests := make([]*helper.IntegrationTestCase, 0)

	// Success: get existing super admin
	var tcSuccess helper.IntegrationTestCase
	tcSuccess = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{integration.PreSetupAddSuperAdmin(t, &tcSuccess, existingSuperAdmin1Entity, true)},
		Name:      "Success: get existing super admin",
		Method:    http.MethodGet,
		Header:    integration.AuthHeader(superAdminJWT),
		Body:      nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: existingEmail1,
		ResponseChecks:   []helper.IntegrationTestCheck{integration.CheckSuperAdminInserted(existingEmail1)},
		PostSetups:       []helper.IntegrationTestPostSetup{integration.PostSetupDeleteSuperAdmin(existingEmail1)},
	}
	tests = append(tests, &tcSuccess)

	// Unauthorized: no JWT
	var tcNoJwt helper.IntegrationTestCase
	tcNoJwt = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{integration.PreSetupAddSuperAdmin(t, &tcNoJwt, existingSuperAdmin1Entity, true)},
		Name:      "Fail: Unauthorized access, no JWT",
		Method:    http.MethodGet,
		Header:    http.Header{},
		Body:      nil,

		WantStatusCode:   http.StatusUnauthorized,
		WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
		ResponseChecks:   []helper.IntegrationTestCheck{integration.CheckSuperAdminInserted(existingEmail1)},
		PostSetups:       []helper.IntegrationTestPostSetup{integration.PostSetupDeleteSuperAdmin(existingEmail1)},
	}
	tests = append(tests, &tcNoJwt)

	// URI invalid
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{integration.PreSetupAddSuperAdmin(t, nil, existingSuperAdmin1Entity, false)},
		Name:      "Fail: URI binding invalid",
		Method:    http.MethodGet,
		Path:      "/api/v1/super_admin/invalid-id",
		Header:    integration.AuthHeader(superAdminJWT),
		Body:      nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks:   []helper.IntegrationTestCheck{integration.CheckSuperAdminInserted(existingEmail1)},
		PostSetups:       []helper.IntegrationTestPostSetup{integration.PostSetupDeleteSuperAdmin(existingEmail1)},
	})

	// Unauthorized roles (tenant user/admin) should be obfuscated as not found
	var tcUnauthTenantUser helper.IntegrationTestCase
	tcUnauthTenantUser = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{integration.PreSetupAddSuperAdmin(t, &tcUnauthTenantUser, existingSuperAdmin1Entity, true)},
		Name:      "Fail: tenant user cannot get super admin",
		Method:    http.MethodGet,
		Header:    integration.AuthHeader(tenantUserJWT),
		Body:      nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{integration.CheckSuperAdminInserted(existingEmail1)},
		PostSetups:       []helper.IntegrationTestPostSetup{integration.PostSetupDeleteSuperAdmin(existingEmail1)},
	}
	tests = append(tests, &tcUnauthTenantUser)

	var tcUnauthTenantAdmin helper.IntegrationTestCase
	tcUnauthTenantAdmin = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{integration.PreSetupAddSuperAdmin(t, &tcUnauthTenantAdmin, existingSuperAdmin2Entity, true)},
		Name:      "Fail: tenant admin cannot get super admin",
		Method:    http.MethodGet,
		Header:    integration.AuthHeader(tenantAdminJWT),
		Body:      nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{integration.CheckSuperAdminInserted(existingEmail2)},
		PostSetups:       []helper.IntegrationTestPostSetup{integration.PostSetupDeleteSuperAdmin(existingEmail2)},
	}
	tests = append(tests, &tcUnauthTenantAdmin)

	// User not found
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: nil,
		Name:      "Fail: user not found",
		Method:    http.MethodGet,
		Path:      "/api/v1/super_admin/999999",
		Header:    integration.AuthHeader(superAdminJWT),
		Body:      nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{integration.CheckNoSuperAdmin("doesnotexist@t.test")},
		PostSetups:       nil,
	})

	helper.RunIntegrationTests(t, tests, deps)
}

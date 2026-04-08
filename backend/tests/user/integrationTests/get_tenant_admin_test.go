package user_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"

	"github.com/google/uuid"
)

func TestGetTenantAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	tenant1Id := uuid.New()
	tenant2Id := uuid.New()

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("failed to generate super admin JWT: %v", err)
	}
	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenant1Id, uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}
	wrongTenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenant2Id, uint(1))
	if err != nil {
		t.Fatalf("failed to generate other tenant admin JWT: %v", err)
	}
	tenantUserJWT, err := helper.NewTenantUserJWT(deps, tenant1Id, uint(5))
	if err != nil {
		t.Fatalf("failed to generate tenant user JWT: %v", err)
	}

	existingEmail1 := "getadmin1@domain.test"
	existingEmail2 := "getadmin2@domain.test"

	existingTenantAdmin1Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail1,
		Name:      "Admin One",
		Confirmed: true,
		Role:      "tenant_admin",
	}

	existingTenantAdmin2Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail2,
		Name:      "Admin Two",
		Confirmed: true,
		Role:      "tenant_admin",
	}

	tests := make([]*helper.IntegrationTestCase, 0)

	// Success: get existing tenant admin
	var tcSuccess helper.IntegrationTestCase
	tcSuccess = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcSuccess, existingTenantAdmin1Entity, true),
		},
		Name:   "Success: get existing tenant admin",
		Method: http.MethodGet,
		Header: authHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: existingEmail1,
		ResponseChecks: []helper.IntegrationTestCheck{
			checkTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			PostSetupDeleteTenantMember(tenant1Id, existingEmail1),
		},
	}
	tests = append(tests, &tcSuccess)

	// Unauthorized: no JWT
	var tcNoJwt helper.IntegrationTestCase
	tcNoJwt = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcNoJwt, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: Unauthorized access, no JWT",
		Method: http.MethodGet,
		Header: http.Header{},
		Body:   nil,

		WantStatusCode:   http.StatusUnauthorized,
		WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcNoJwt)

	// URI invalid
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, nil, existingTenantAdmin1Entity, false),
		},
		Name:   "Fail: URI binding invalid",
		Method: http.MethodGet,
		Path:   "/api/v1/tenant/invalid-uuid/tenant_admin/123",
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks:   []helper.IntegrationTestCheck{checkTenantMemberInserted(existingEmail1, tenant1Id.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	})

	// Tenant not found
	var tcTenantNotFound helper.IntegrationTestCase
	tcTenantNotFound = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcTenantNotFound, existingTenantAdmin1Entity, false),
		},
		Name:   "Fail: tenant not found",
		Method: http.MethodGet,
		Header: authHeader(superAdminJWT),
		Body:   nil,
		Path:   "/api/v1/tenant/" + uuid.New().String() + "/tenant_admin/1",

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{checkTenantMemberInserted(existingEmail1, tenant1Id.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcTenantNotFound)

	// Unauthorized access: tenant user trying to get admin
	var tcUnauthorizedUser helper.IntegrationTestCase
	tcUnauthorizedUser = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcUnauthorizedUser, existingTenantAdmin1Entity, true),
			PreSetupAddTenantAdmin(t, &tcUnauthorizedUser, existingTenantAdmin2Entity, true),
		},
		Name:   "Fail: tenant user cannot get tenant admin",
		Method: http.MethodGet,
		Header: authHeader(tenantUserJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{checkTenantMemberInserted(existingEmail2, tenant1Id.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
			nil,
		},
	}
	tests = append(tests, &tcUnauthorizedUser)

	// Unauthorized tenant admin from other tenant
	var tcWrongTenantAdmin helper.IntegrationTestCase
	tcWrongTenantAdmin = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcWrongTenantAdmin, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: tenant admin from other tenant cannot get",
		Method: http.MethodGet,
		Header: authHeader(wrongTenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{checkTenantMemberInserted(existingEmail1, tenant1Id.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcWrongTenantAdmin)

	// Super admin denied when CanImpersonate=false
	var tcSuperDenied helper.IntegrationTestCase
	tcSuperDenied = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, false),
			PreSetupAddTenantAdmin(t, &tcSuperDenied, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: super admin denied when CanImpersonate=false",
		Method: http.MethodGet,
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{checkTenantMemberInserted(existingEmail1, tenant1Id.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcSuperDenied)

	// User not found
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
		},
		Name:   "Fail: user not found (tenant exists but user id does not)",
		Method: http.MethodGet,
		Path:   "/api/v1/tenant/" + tenant1Id.String() + "/tenant_admin/999999",
		Header: authHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{checkNoTenantMember("doesnotexist@t.test", tenant1Id.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

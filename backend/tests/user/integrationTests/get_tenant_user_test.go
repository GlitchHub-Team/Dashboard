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

func TestGetTenantUserIntegration(t *testing.T) {
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
	tenantUserJWT, err := helper.NewTenantUserJWT(deps, tenant1Id, uint(1))
	if err != nil {
		t.Fatalf("failed to generate tenant user JWT: %v", err)
	}

	existingEmail1 := "getuser1@domain.test"
	existingEmail2 := "getuser2@domain.test"

	existingTenantUser1Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail1,
		Name:      "User One",
		Confirmed: true,
		Role:      "tenant_user",
	}

	existingTenantUser2Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail2,
		Name:      "User Two",
		Confirmed: true,
		Role:      "tenant_user",
	}

	tests := make([]*helper.IntegrationTestCase, 0)

	// Success: get existing user (tenant admin)
	var tcSuccessTenantAdmin helper.IntegrationTestCase
	tcSuccessTenantAdmin = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantUser(t, &tcSuccessTenantAdmin, existingTenantUser1Entity, true),
		},
		Name:   "Success: get existing tenant user (tenant admin)",
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
	tests = append(tests, &tcSuccessTenantAdmin)

	// Success: get self (tenant user)
	var tcSuccessTenantUser helper.IntegrationTestCase
	tcSuccessTenantUser = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantUser(t, &tcSuccessTenantUser, existingTenantUser1Entity, true),
		},
		Name:   "Success: get existing tenant user",
		Method: http.MethodGet,
		Header: authHeader(tenantUserJWT),
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
	tests = append(tests, &tcSuccessTenantUser)

	// Unauthorized: no JWT
	var tcNoJwt helper.IntegrationTestCase
	tcNoJwt = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantUser(t, &tcNoJwt, existingTenantUser1Entity, true),
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
			PreSetupAddTenantUser(t, nil, existingTenantUser1Entity, false),
		},
		Name:   "Fail: URI binding invalid",
		Method: http.MethodGet,
		Path:   "/api/v1/tenant/invalid-uuid/tenant_user/123",
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks: []helper.IntegrationTestCheck{
			checkTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
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
			PreSetupAddTenantUser(t, &tcTenantNotFound, existingTenantUser1Entity, false),
		},
		Name:   "Fail: tenant not found",
		Method: http.MethodGet,
		Header: authHeader(superAdminJWT),
		Body:   nil,
		Path:   "/api/v1/tenant/" + uuid.New().String() + "/tenant_user/1",

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{checkTenantMemberInserted(existingEmail1, tenant1Id.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcTenantNotFound)

	// Unauthorized access: tenant user trying to get another user
	var tcUnauthorizedUser helper.IntegrationTestCase
	tcUnauthorizedUser = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantUser(t, &tcUnauthorizedUser, existingTenantUser1Entity, true),
			PreSetupAddTenantUser(t, &tcUnauthorizedUser, existingTenantUser2Entity, true),
		},
		Name:   "Fail: tenant user cannot get other user",
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
			PreSetupAddTenantUser(t, &tcWrongTenantAdmin, existingTenantUser1Entity, true),
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
			PreSetupAddTenantUser(t, &tcSuperDenied, existingTenantUser1Entity, true),
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
		Name:   "Fail: user not found",
		Method: http.MethodGet,
		Path:   "/api/v1/tenant/" + tenant1Id.String() + "/tenant_user/999999",
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

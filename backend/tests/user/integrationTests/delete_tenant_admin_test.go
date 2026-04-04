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

func TestDeleteTenantAdminIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	tenant1Id := uuid.New()
	tenant2Id := uuid.New()
	inexistentTenantId := uuid.New()

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
		t.Fatalf("failed to generate tenant admin JWT: %v", err)
	}
	tenantUserJWT, err := helper.NewTenantUserJWT(deps, tenant1Id, uint(5))
	if err != nil {
		t.Fatalf("failed to generate tenant user JWT: %v", err)
	}

	existingEmail1 := "tenantadmin1@domain.test"
	existingEmail2 := "tenantadmin2@domain.test"

	existingTenantAdmin1Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail1,
		Name:      "Existing Super Admin",
		Confirmed: true,
		Role:      "tenant_admin",
	}

	existingTenantAdmin2Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail2,
		Name:      "Existing Super Admin",
		Confirmed: true,
		Role:      "tenant_admin",
	}

	tests := make([]*helper.IntegrationTestCase, 0)

	// Success
	var tcSuccess helper.IntegrationTestCase
	tcSuccess = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcSuccess, existingTenantAdmin1Entity, true),
			PreSetupAddTenantAdmin(t, &tcSuccess, existingTenantAdmin2Entity, true),
		},
		Name:   "Success: delete existing tenant admin",
		Method: http.MethodDelete,
		Header: authHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: existingEmail2,
		ResponseChecks: []helper.IntegrationTestCheck{
			checkNoTenantMember(existingEmail2, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			PostSetupDeleteTenantMember(tenant1Id, existingEmail1),
			PostSetupDeleteTenantMember(tenant1Id, existingEmail2),
		},
	}
	tests = append(tests, &tcSuccess)

	// Unauthorized no JWT
	var tcNoJwt helper.IntegrationTestCase
	tcNoJwt = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcNoJwt, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: Unauthorized access, no JWT",
		Method: http.MethodDelete,
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
		Method: http.MethodDelete,
		Path:   "/api/v1/tenant/invalid-uuid/tenant_admin/123",
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
			PreSetupAddTenantAdmin(t, &tcTenantNotFound, existingTenantAdmin1Entity, false),
			// PreSetupAddTenantAdmin(t, &tcTenantNotFound, tenantAdminEntity_inexistentTenant, true),
		},
		Name:   "Fail: tenant not found",
		Method: http.MethodDelete,
		Header: authHeader(tenantAdminJWT),
		Path:   "/api/v1/tenant/" + inexistentTenantId.String() + "/tenant_admin/123",
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
			// nil,
		},
	}
	tests = append(tests, &tcTenantNotFound)

	// Unauthorized role (tenant user)
	var tcTenantUser helper.IntegrationTestCase
	tcTenantUser = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcTenantUser, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: unauthorized role (tenant user) cannot delete",
		Method: http.MethodDelete,
		Header: authHeader(tenantUserJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcTenantUser)

	// Unauthorized tenant admin (wrong tenant id)
	var tcWrongTenantId helper.IntegrationTestCase
	tcWrongTenantId = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
			PreSetupAddTenantAdmin(t, &tcWrongTenantId, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: unauthorized role (tenant user) cannot delete",
		Method: http.MethodDelete,
		Header: authHeader(wrongTenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcWrongTenantId)

	// Super admin denied when CanImpersonate=false
	var tcSuperAdmin_NoImpersonate helper.IntegrationTestCase
	tcSuperAdmin_NoImpersonate = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, false),
			PreSetupAddTenantAdmin(t, &tcSuperAdmin_NoImpersonate, existingTenantAdmin1Entity, true),
			PreSetupAddTenantAdmin(t, &tcSuperAdmin_NoImpersonate, existingTenantAdmin2Entity, true),
		},
		Name:   "Fail: super admin denied when CanImpersonate=false",
		Method: http.MethodDelete,
		Header: authHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
			nil,
			nil,
		},
	}
	tests = append(tests, &tcSuperAdmin_NoImpersonate)

	// User not found
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			preSetupCreateTenant(tenant1Id, true),
		},
		Name:   "Fail: user not found",
		Method: http.MethodDelete,
		Path:   "/api/v1/tenant/" + tenant1Id.String() + "/tenant_admin/999999",
		Header: authHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			checkNoTenantMember("doesnotexist@t.test", tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			postSetupDeleteTenant(t, tenant1Id),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

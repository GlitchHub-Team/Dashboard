package user_integration_test

import (
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestDeleteTenantUserIntegration(t *testing.T) {
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

	existingEmail1 := "tenantuser1@domain.test"
	existingEmail2 := "tenantuser2@domain.test"

	existingTenantAdmin1Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail1,
		Name:      "Existing Super Admin",
		Confirmed: true,
		Role:      "tenant_user",
	}

	existingTenantAdmin2Entity := user.TenantMemberEntity{
		TenantId:  tenant1Id.String(),
		Email:     existingEmail2,
		Name:      "Existing Super Admin",
		Confirmed: true,
		Role:      "tenant_user",
	}

	tests := make([]*helper.IntegrationTestCase, 0)

	// Success
	var tcSuccess helper.IntegrationTestCase
	tcSuccess = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, true),
			integration.PreSetupAddTenantUser(t, &tcSuccess, existingTenantAdmin1Entity, true),
			integration.PreSetupAddTenantUser(t, &tcSuccess, existingTenantAdmin2Entity, true),
		},
		Name:   "Success: delete existing tenant admin",
		Method: http.MethodDelete,
		Header: integration.AuthHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: existingEmail2,
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckNoTenantMember(existingEmail2, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
			integration.PostSetupDeleteTenantMember(tenant1Id, existingEmail1),
			integration.PostSetupDeleteTenantMember(tenant1Id, existingEmail2),
		},
	}
	tests = append(tests, &tcSuccess)

	// Unauthorized no JWT
	var tcNoJwt helper.IntegrationTestCase
	tcNoJwt = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, true),
			integration.PreSetupAddTenantUser(t, &tcNoJwt, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: Unauthorized access, no JWT",
		Method: http.MethodDelete,
		Header: http.Header{},
		Body:   nil,

		WantStatusCode:   http.StatusUnauthorized,
		WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcNoJwt)

	// URI invalid
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, true),
			integration.PreSetupAddTenantUser(t, nil, existingTenantAdmin1Entity, false),
		},
		Name:   "Fail: URI binding invalid",
		Method: http.MethodDelete,
		Path:   "/api/v1/tenant/invalid-uuid/tenant_user/123",
		Header: integration.AuthHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	})

	// Tenant not found
	var tcTenantNotFound helper.IntegrationTestCase
	tcTenantNotFound = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, true),
			integration.PreSetupAddTenantUser(t, &tcTenantNotFound, existingTenantAdmin1Entity, false),
		},
		Name:   "Fail: tenant not found",
		Method: http.MethodDelete,
		Header: integration.AuthHeader(tenantAdminJWT),
		Path:   "/api/v1/tenant/" + inexistentTenantId.String() + "/tenant_user/123",
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcTenantNotFound)

	// Unauthorized role (tenant user)
	var tcTenantUser helper.IntegrationTestCase
	tcTenantUser = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, true),
			integration.PreSetupAddTenantUser(t, &tcTenantUser, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: unauthorized role (tenant user) cannot delete",
		Method: http.MethodDelete,
		Header: integration.AuthHeader(tenantUserJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcTenantUser)

	// Unauthorized tenant admin (wrong tenant id)
	var tcWrongTenantId helper.IntegrationTestCase
	tcWrongTenantId = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, true),
			integration.PreSetupAddTenantUser(t, &tcWrongTenantId, existingTenantAdmin1Entity, true),
		},
		Name:   "Fail: unauthorized role (tenant user) cannot delete",
		Method: http.MethodDelete,
		Header: integration.AuthHeader(wrongTenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
			nil,
		},
	}
	tests = append(tests, &tcWrongTenantId)

	// Super admin denied when CanImpersonate=false
	var tcSuperAdmin_NoImpersonate helper.IntegrationTestCase
	tcSuperAdmin_NoImpersonate = helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, false),
			integration.PreSetupAddTenantUser(t, &tcSuperAdmin_NoImpersonate, existingTenantAdmin1Entity, true),
			integration.PreSetupAddTenantUser(t, &tcSuperAdmin_NoImpersonate, existingTenantAdmin2Entity, true),
		},
		Name:   "Fail: super admin denied when CanImpersonate=false",
		Method: http.MethodDelete,
		Header: integration.AuthHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
			nil,
			nil,
		},
	}
	tests = append(tests, &tcSuperAdmin_NoImpersonate)

	// User not found
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenant1Id, true),
		},
		Name:   "Fail: user not found",
		Method: http.MethodDelete,
		Path:   "/api/v1/tenant/" + tenant1Id.String() + "/tenant_user/999999",
		Header: integration.AuthHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckNoTenantMember("doesnotexist@t.test", tenant1Id.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenant1Id),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

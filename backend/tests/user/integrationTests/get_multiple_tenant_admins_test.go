package user_integration_test

import (
	"fmt"
	"net/http"
	"testing"

	transportHttp "backend/internal/infra/transport/http"
	"backend/internal/shared/identity"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"github.com/google/uuid"
)

func TestGetTenantAdminListIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	tests := make([]*helper.IntegrationTestCase, 0)

	tenantId := uuid.New()
	otherTenantId := uuid.New()
	nonexistentTenantId := uuid.New()

	existingEmail1 := "tenantadmin1@t.test"
	existingEmail2 := "tenantadmin2@t.test"

	tenantAdmin1Entity := user.TenantMemberEntity{
		TenantId: tenantId.String(),
		Email:    existingEmail1,
		Name:     "A1",
		Role:     string(identity.ROLE_TENANT_ADMIN),
	}
	tenantAdmin2Entity := user.TenantMemberEntity{
		TenantId: tenantId.String(),
		Email:    existingEmail2,
		Name:     "A2",
		Role:     string(identity.ROLE_TENANT_ADMIN),
	}

	superAdminJWT, err := helper.NewSuperAdminJWT(deps, uint(1))
	if err != nil {
		t.Fatalf("Cannot create super admin JWT: %v", err)
		return
	}

	tenantAdminJWT, err := helper.NewTenantAdminJWT(deps, tenantId, uint(1))
	if err != nil {
		t.Fatalf("Cannot create tenant admin JWT: %v", err)
		return
	}

	otherAdminJWT, err := helper.NewTenantAdminJWT(deps, otherTenantId, uint(1))
	if err != nil {
		t.Fatalf("Cannot create tenant admin JWT: %v", err)
		return
	}

	tenantPath := fmt.Sprintf("/api/v1/tenant/%v/tenant_admins", tenantId.String())
	inexistentTenantPath := fmt.Sprintf("/api/v1/tenant/%v/tenant_admins", nonexistentTenantId.String())

	// Success default pagination
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenantId, true),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin1Entity, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin2Entity, false),
		},
		Name:   "Success: default pagination",
		Method: http.MethodGet,
		Path:   tenantPath,
		Header: integration.AuthHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "\"count\":2",
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenantId.String()),
			integration.CheckTenantMemberInserted(existingEmail2, tenantId.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{integration.PostSetupDeleteTenant(t, tenantId), nil, nil},
	})

	// Success custom pagination page=2&limit=1
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenantId, true),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin1Entity, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin2Entity, false),
		},
		Name:   "Success: custom pagination",
		Method: http.MethodGet,
		Path:   tenantPath + "?page=2&limit=1",
		Header: integration.AuthHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "\"count\":1",
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantMemberInserted(existingEmail1, tenantId.String()),
			integration.CheckTenantMemberInserted(existingEmail2, tenantId.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{integration.PostSetupDeleteTenant(t, tenantId), nil, nil},
	})

	// Success: super admin (CanImpersonate=false)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenantId, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin1Entity, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin2Entity, false),
		},
		Name:   "(Super Admin) Success: canImpersonate=false",
		Method: http.MethodGet,
		Path:   tenantPath + "?limit=100",
		Header: integration.AuthHeader(superAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "\"count\":2",
		ResponseChecks:   []helper.IntegrationTestCheck{},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenantId),
			nil,
			nil,
		},
	})

	// Unauthorized no JWT
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenantId, true),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin1Entity, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin2Entity, false),
		},
		Name:   "Fail: Unauthorized, no JWT",
		Method: http.MethodGet,
		Path:   tenantPath,
		Header: http.Header{},
		Body:   nil,

		WantStatusCode:   http.StatusUnauthorized,
		WantResponseBody: helper.ErrJsonString(transportHttp.ErrMissingIdentity),
		ResponseChecks:   []helper.IntegrationTestCheck{},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenantId),
			nil,
			nil,
		},
	})

	// URI invalid
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenantId, true),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin1Entity, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin2Entity, false),
		},
		Name:   "Fail: URI binding invalid",
		Method: http.MethodGet,
		Path:   "/api/v1/tenant/invalid-uuid/tenant_admins",
		Header: integration.AuthHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckNoTenant(tenantId.String()),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenantId),
			nil,
			nil,
		},
	})

	// Query invalid page=-1
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenantId, true),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin1Entity, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin2Entity, false),
		},
		Name:   "Fail: Query binding invalid",
		Method: http.MethodGet,
		Path:   tenantPath + "?page=-1",
		Header: integration.AuthHeader(tenantAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks:   []helper.IntegrationTestCheck{integration.CheckNoTenant(tenantId.String())},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenantId),
			nil,
			nil,
		},
	})

	// Tenant not found
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: nil,
		Name:      "Fail: tenant not found",
		Method:    http.MethodGet,
		Path:      inexistentTenantPath,
		Header:    integration.AuthHeader(tenantAdminJWT),
		Body:      nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckNoTenant(nonexistentTenantId.String()),
		},
		PostSetups: nil,
	})

	// Unauthorized access: other tenant admin
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			integration.PreSetupCreateTenant(tenantId, true),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin1Entity, false),
			integration.PreSetupAddTenantAdmin(t, nil, tenantAdmin2Entity, false),
			integration.PreSetupCreateTenant(otherTenantId, true),
		},
		Name:   "Fail (tenant admin): unauth access",
		Method: http.MethodGet,
		Path:   tenantPath,
		Header: integration.AuthHeader(otherAdminJWT),
		Body:   nil,

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(tenant.ErrTenantNotFound),
		ResponseChecks:   []helper.IntegrationTestCheck{},
		PostSetups: []helper.IntegrationTestPostSetup{
			integration.PostSetupDeleteTenant(t, tenantId),
			nil,
			nil,
			integration.PostSetupDeleteTenant(t, otherTenantId),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

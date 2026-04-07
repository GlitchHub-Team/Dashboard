package auth_integration_test

import (
	"net/http"
	"testing"

	"backend/internal/auth"
	"backend/internal/shared/identity"

	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"backend/internal/infra/transport/http/dto"

	"github.com/google/uuid"
)

func TestChangePasswordIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	tenantId := uuid.New()
	tenantIdStr := tenantId.String()

	oldPassword := "old-pa$$w0rd"
	newPassword := "new-pa$$w0rd"
	wrongOldPassword := "wrong-old-pass"
	shortNew := "short"

	emptyString := ""

	// tenant member
	tenantUserEntity := user.TenantMemberEntity{
		Email:     "",  // NOTA: verrà impostato dopo
		Password:  nil, // NOTA: verrà impostato dopo
		Name:      "Tenant Pwd User",
		Confirmed: true,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	// unconfirmed tenant

	unconfirmedTenantUserEntity := user.TenantMemberEntity{
		Email:     "", // NOTA: verrà impostato dopo
		Password:  &emptyString,
		Name:      "Unconfirmed Tenant",
		Confirmed: false,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	// super admin
	superAdminEmail := "super-admin@example.com"
	superAdminEntity := user.SuperAdminEntity{
		Email:     superAdminEmail,
		Name:      "Super Pwd",
		Password:  nil,
		Confirmed: true,
	}

	unconfirmedSuperAdminEmail := "unconfirmed-super-admin@example.com"
	unconfirmedSuperAdminEntity := user.SuperAdminEntity{
		Email:     unconfirmedSuperAdminEmail,
		Name:      "Unconfirmed Super",
		Password:  &emptyString,
		Confirmed: false,
	}

	// create tenant (common)
	integration.PreSetupCreateTenant(tenantId, true)(deps)

	// create isolated users for each case to avoid state bleed between test cases
	hashedOld, err := deps.SecretHasher.HashSecret(oldPassword)
	if err != nil {
		t.Fatalf("Cannot hash password for test: %v", err)
	}

	correctBody := auth.ChangePasswordBodyDTO{
		ChangePasswordFields: dto.ChangePasswordFields{
			OldPassword: oldPassword,
			NewPassword: newPassword,
		},
	}

	shortPasswordBody := auth.ChangePasswordBodyDTO{
		ChangePasswordFields: dto.ChangePasswordFields{
			OldPassword: oldPassword,
			NewPassword: shortNew,
		},
	}

	wrongPasswordBody := auth.ChangePasswordBodyDTO{
		ChangePasswordFields: dto.ChangePasswordFields{
			OldPassword: wrongOldPassword,
			NewPassword: newPassword,
		},
	}

	basePath := "/api/v1/auth/change_password"

	tests := []*helper.IntegrationTestCase{}

	// Success (Tenant Member) ----------------------------------------------------------------------------------------

	// Entity
	tenantUserEntity_Success := tenantUserEntity
	tenantUserEntity_Success.Email = "success@example.com"
	tenantUserEntity_Success.Password = &hashedOld
	preSetup, tenantUserId_Success := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity_Success)
	preSetup(deps)
	defer integration.PostSetupDeleteTenantMember(tenantId, tenantUserEntity_Success.Email)

	// Requester
	tenantUserRequester_Success := identity.Requester{
		RequesterUserId:   *tenantUserId_Success,
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}
	tenantUserJwt_Success, err := deps.AuthTokenManager.GenerateForRequester(tenantUserRequester_Success)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member)  Success",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(tenantUserJwt_Success),
		Body:      helper.MustJSONBody(t, correctBody),

		WantStatusCode: http.StatusOK,
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberPassword(t, tenantIdStr, *tenantUserId_Success, newPassword, true),
		},
	})

	// Missing JWT (Tenant Member): 401 --------------------------------------------------------------------------------
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member)  Fail: missing JWT",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, correctBody),

		WantStatusCode: http.StatusUnauthorized,
		ResponseChecks: []helper.IntegrationTestCheck{},
	})

	// Invalid binding (new password too short): 400, unchanged ------------------------------------------------------

	// Entity
	tenantUserEntity_Invalid := tenantUserEntity
	tenantUserEntity_Invalid.Email = "invalid-binding@example.com"
	tenantUserEntity_Invalid.Password = &hashedOld
	preSetup, tenantUserId_Invalid := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity_Invalid)
	preSetup(deps)
	defer integration.PostSetupDeleteTenantMember(tenantId, tenantUserEntity_Invalid.Email)

	// Requester
	tenantUserRequester_Invalid := identity.Requester{
		RequesterUserId:   *tenantUserId_Invalid,
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}
	tenantUserJwt_Invalid, err := deps.AuthTokenManager.GenerateForRequester(tenantUserRequester_Invalid)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member) Fail: invalid binding (short new password)",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(tenantUserJwt_Invalid),
		Body:      helper.MustJSONBody(t, shortPasswordBody),

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberPassword(t, tenantIdStr, *tenantUserId_Invalid, oldPassword, true),
			CheckTenantMemberPassword(t, tenantIdStr, *tenantUserId_Invalid, newPassword, false),
		},
	})

	// Wrong credentials (tenant): old password mismatch -> 404, unchanged ------------------------------------------------

	// Entity
	tenantUserEntity_WrongCreds := tenantUserEntity
	tenantUserEntity_WrongCreds.Email = "wrong-creds@example.com"
	tenantUserEntity_WrongCreds.Password = &hashedOld
	preSetup, tenantUserId_WrongCreds := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity_WrongCreds)
	preSetup(deps)
	defer integration.PostSetupDeleteTenantMember(tenantId, tenantUserEntity_WrongCreds.Email)

	// Requester
	tenantUserRequester_WrongCreds := identity.Requester{
		RequesterUserId:   *tenantUserId_WrongCreds,
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}
	tenantUserJwt_WrongCreds, err := deps.AuthTokenManager.GenerateForRequester(tenantUserRequester_WrongCreds)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member) Fail: wrong old password",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(tenantUserJwt_WrongCreds),
		Body:      helper.MustJSONBody(t, wrongPasswordBody),

		WantStatusCode: http.StatusNotFound,
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberPassword(t, tenantIdStr, *tenantUserId_WrongCreds, oldPassword, true),
		},
	})

	// Unconfirmed requester (tenant): expect 404 and empty password ------------------------------------------------------
	// Entity
	tenantUserEntity_Unconfirmed := unconfirmedTenantUserEntity
	tenantUserEntity_Unconfirmed.Email = "unconfirmed@example.com"
	tenantUserEntity_Unconfirmed.Password = nil
	preSetup, tenantUserId_Unconfirmed := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity_Unconfirmed)
	preSetup(deps)
	defer integration.PostSetupDeleteTenantMember(tenantId, tenantUserEntity_Unconfirmed.Email)

	// Requester
	tenantUserRequester_Unconfirmed := identity.Requester{
		RequesterUserId:   *tenantUserId_Unconfirmed,
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}
	tenantUserJwt_Unconfirmed, err := deps.AuthTokenManager.GenerateForRequester(tenantUserRequester_Unconfirmed)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member) Fail: requester unconfirmed",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(tenantUserJwt_Unconfirmed),
		Body:      helper.MustJSONBody(t, correctBody),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: "account not confirmed",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberPassword(t, tenantIdStr, *tenantUserId_Unconfirmed, "", true),
			CheckTenantMemberPassword(t, tenantIdStr, *tenantUserId_Unconfirmed, newPassword, false),
		},
	})

	// Casi Super Admin ===================================================================================================

	// Success (super admin) ---------------------------------------------------------------------------------------------------
	// Entity
	superAdminEntity_Success := superAdminEntity
	superAdminEntity_Success.Email = "success@m31.com"
	superAdminEntity_Success.Password = &hashedOld
	preSetup, superAdminId_Success := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity_Success)
	preSetup(deps)
	defer integration.PostSetupDeleteSuperAdmin(superAdminEntity_Success.Email)

	// Requester
	superAdminRequester_Success := identity.Requester{
		RequesterUserId:   *superAdminId_Success,
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}
	superAdminJwt_Success, err := deps.AuthTokenManager.GenerateForRequester(superAdminRequester_Success)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Success",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(superAdminJwt_Success),
		Body:      helper.MustJSONBody(t, correctBody),

		WantStatusCode: http.StatusOK,
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminPassword(t, *superAdminId_Success, newPassword, true),
		},
	})

	// Missing JWT (Super Admin) -> 401, unchanged -----------------------------------------------------------------------------
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Fail: missing JWT",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, correctBody),

		WantStatusCode: http.StatusUnauthorized,
		ResponseChecks: []helper.IntegrationTestCheck{},
	})

	// Invalid binding (short new password) (Super Admin) ------------------------------------------------------------------
	// Entity
	superAdminEntity_Invalid := superAdminEntity
	superAdminEntity_Invalid.Email = "invalid@m31.com"
	superAdminEntity_Invalid.Password = &hashedOld
	preSetup, superAdminId_Invalid := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity_Invalid)
	preSetup(deps)
	defer integration.PostSetupDeleteSuperAdmin(superAdminEntity_Success.Email)

	// Requester
	superAdminRequester_Invalid := identity.Requester{
		RequesterUserId:   *superAdminId_Invalid,
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}
	superAdminJwt_Invalid, err := deps.AuthTokenManager.GenerateForRequester(superAdminRequester_Invalid)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Fail: invalid binding (short new password)",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(superAdminJwt_Invalid),
		Body:      helper.MustJSONBody(t, shortPasswordBody),

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminPassword(t, *superAdminId_Invalid, oldPassword, true),
		},
	})

	// Wrong credentials (Super Admin) -> 404 -----------------------------------------------------------------------------
	// Entity
	superAdminEntity_WrongCreds := superAdminEntity
	superAdminEntity_WrongCreds.Email = "wrong-creds@m31.com"
	superAdminEntity_WrongCreds.Password = &hashedOld
	preSetup, superAdminId_WrongCreds := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity_WrongCreds)
	preSetup(deps)
	defer integration.PostSetupDeleteSuperAdmin(superAdminEntity_Success.Email)

	// Requester
	superAdminRequester_WrongCreds := identity.Requester{
		RequesterUserId:   *superAdminId_WrongCreds,
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}
	superAdminJwt_WrongCreds, err := deps.AuthTokenManager.GenerateForRequester(superAdminRequester_WrongCreds)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Fail: wrong old password",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(superAdminJwt_WrongCreds),
		Body:      helper.MustJSONBody(t, wrongPasswordBody),

		WantStatusCode: http.StatusNotFound,
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminPassword(t, *superAdminId_WrongCreds, oldPassword, true),
		},
	})

	// Unconfirmed super admin requester -> 401 and empty password -------------------------------------------------------
	// Entity
	superAdminEntity_Unconfirmed := unconfirmedSuperAdminEntity
	superAdminEntity_Unconfirmed.Email = "unconfirmed@m31.com"
	superAdminEntity_Unconfirmed.Password = nil
	preSetup, superAdminId_Unconfirmed := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity_Unconfirmed)
	preSetup(deps)
	defer integration.PostSetupDeleteSuperAdmin(superAdminEntity_Success.Email)

	// Requester
	superAdminRequester_Unconfirmed := identity.Requester{
		RequesterUserId:   *superAdminId_Unconfirmed,
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}
	superAdminJwt_Unconfirmed, err := deps.AuthTokenManager.GenerateForRequester(superAdminRequester_Unconfirmed)
	if err != nil {
		t.Fatalf("cannot create jwt: %v", err)
	}

	// Case
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Fail: requester unconfirmed",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    integration.AuthHeader(superAdminJwt_Unconfirmed),
		Body:      helper.MustJSONBody(t, correctBody),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: "account not confirmed",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminPassword(t, *superAdminId_Unconfirmed, "", true),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

package auth_integration_test

import (
	"bytes"
	"net/http"

	// "net/http/httptest"
	"testing"

	"backend/internal/auth"
	"backend/internal/shared/identity"

	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"backend/internal/infra/transport/http/dto"

	"github.com/google/uuid"
)

func TestLoginUserIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// common values
	tenantId := uuid.New()
	tenantIdStr := tenantId.String()

	confirmedTenantUserEmail := "confirmed@domain.test"
	unconfirmedTenantUserEmail := "unconfirmed@domain.test"
	superAdminEmail := "super-admin@m31.com"
	correctPassword := "P@ssw0rd"
	wrongPassword := "wrongPassword!"

	hashedPassword, err := deps.SecretHasher.HashSecret(correctPassword)
	if err != nil {
		t.Fatalf("Cannot hash test password: %v", err)
	}

	confirmedTenantUserEntity := user.TenantMemberEntity{
		Email:     confirmedTenantUserEmail,
		Password:  &hashedPassword,
		Name:      "Confirmed Tenant User",
		Confirmed: true,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantId.String(),
	}

	unconfirmedTenantUserEntity := user.TenantMemberEntity{
		Email:     unconfirmedTenantUserEmail,
		Password:  &hashedPassword,
		Name:      "Unconfirmed Tenant User",
		Confirmed: false,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantId.String(),
	}

	superAdminEntity := user.SuperAdminEntity{
		Email:     superAdminEmail,
		Name:      "Super Admin",
		Password:  &hashedPassword,
		Confirmed: true,
	}

	// NOTA: Queste funzioni vengono inserite all'inizio perché il login non cambia lo stato
	integration.PreSetupCreateTenant(tenantId, true)(deps)
	preSetup, confirmedTenantUserId := integration.PreSetupAddTenantMember_ReturnUserId(t, &confirmedTenantUserEntity)
	preSetup(deps)

	preSetup, _ = integration.PreSetupAddTenantMember_ReturnUserId(t, &unconfirmedTenantUserEntity)
	preSetup(deps)

	preSetup, superAdminId := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity)
	preSetup(deps)

	// Request body
	tenantUserBody := auth.LoginUserDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &tenantIdStr,
		},
		EmailField: dto.EmailField{
			Email: confirmedTenantUserEmail,
		},
		PasswordField: dto.PasswordField{
			Password: correctPassword,
		},
	}

	tenantUserBody_NoEmail := auth.LoginUserDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &tenantIdStr,
		},
		PasswordField: dto.PasswordField{
			Password: correctPassword,
		},
	}

	tenantUserBody_WrongPassword := auth.LoginUserDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &tenantIdStr,
		},
		EmailField: dto.EmailField{
			Email: confirmedTenantUserEmail,
		},
		PasswordField: dto.PasswordField{
			Password: wrongPassword,
		},
	}

	unconfirmedTenantUserBody := auth.LoginUserDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &tenantIdStr,
		},
		EmailField: dto.EmailField{
			Email: unconfirmedTenantUserEmail,
		},
		PasswordField: dto.PasswordField{
			Password: correctPassword,
		},
	}

	superAdminBody := auth.LoginUserDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: nil,
		},
		EmailField: dto.EmailField{
			Email: superAdminEmail,
		},
		PasswordField: dto.PasswordField{
			Password: correctPassword,
		},
	}

	// Requester
	expectedTenantUserRequester := identity.Requester{
		RequesterUserId:   *confirmedTenantUserId,
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	expectedSuperAdminRequester := identity.Requester{
		RequesterUserId:   *superAdminId,
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}

	tests := []*helper.IntegrationTestCase{}

	// Success: valid credentials (tenant user)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant User) Success: valid credentials",
		Method:    http.MethodPost,
		Path:      "/api/v1/auth/login",
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, tenantUserBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "\"jwt\":",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckValidJWTInResponse(t, expectedTenantUserRequester),
		},
		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Success: valid credentials (tenant user)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Success: valid credentials",
		Method:    http.MethodPost,
		Path:      "/api/v1/auth/login",
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, superAdminBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "\"jwt\":",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckValidJWTInResponse(t, expectedSuperAdminRequester),
		},
		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: binding JSON invalid
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "Fail: binding JSON",
		Method:    http.MethodPost,
		Path:      "/api/v1/auth/login",
		Header:    http.Header{},
		Body:      bytes.NewBufferString("not-a-json"),

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks:   nil,
		PostSetups:       []helper.IntegrationTestPostSetup{},
	})

	// Fail: account not confirmed
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "Fail: account not confirmed",
		Method:    http.MethodPost,
		Path:      "/api/v1/auth/login",
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, unconfirmedTenantUserBody),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrAccountNotConfirmed),
		ResponseChecks:   nil,
		PostSetups:       []helper.IntegrationTestPostSetup{},
	})

	// Fail: wrong credentials (email missing)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "Fail: wrong credentials - email missing",
		Method:    http.MethodPost,
		Path:      "/api/v1/auth/login",
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, tenantUserBody_NoEmail),

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "invalid format",
		ResponseChecks:   nil,
		PostSetups:       []helper.IntegrationTestPostSetup{},
	})

	// Fail: wrong password for existing email
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "Fail: wrong password",
		Method:    http.MethodPost,
		Path:      "/api/v1/auth/login",
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, tenantUserBody_WrongPassword),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrWrongCredentials),
		ResponseChecks:   nil,
		PostSetups:       []helper.IntegrationTestPostSetup{},
	})

	helper.RunIntegrationTests(t, tests, deps)

	// Post Setup globale
	integration.PostSetupDeleteTenant(t, tenantId)(deps)
}

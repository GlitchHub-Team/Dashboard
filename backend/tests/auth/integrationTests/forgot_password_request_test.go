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

func TestRequestForgotPasswordTokenIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// common values
	tenantId := uuid.New()
	tenantIdStr := tenantId.String()

	// otherTenantId := uuid.New()
	// otherTenantIdStr := otherTenantId.String()

	invalidTenantStr := "invalid-uuid"

	// Stato tabella tenant_members
	tenantUserEmail := "tu1@example.com"
	tenantUserEmail_NotFound := "tu1-not.found@example.com"
	tenantUserEmail_NotConfirmed := "tu1-not-confirmed@example.com"

	userPw := "pw123"

	tenantUserEntity := user.TenantMemberEntity{
		Email:     tenantUserEmail,
		Password:  &userPw,
		Name:      "Tenant User",
		Confirmed: true,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	tenantUserEntity_NotConfirmed := user.TenantMemberEntity{
		Email:     tenantUserEmail_NotConfirmed,
		Password:  &userPw,
		Name:      "Unconfirmed Tenant User",
		Confirmed: false,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	superAdminEmail := "superadmin@m31.com"
	superAdminEmail_NotFound := "not-found@m31.com"
	superAdminEmail_NotConfirmed := "not-confirmed-123@m31.com"

	superAdminEntity := user.SuperAdminEntity{
		Email:     superAdminEmail,
		Name:      "Super Admin",
		Password:  &userPw,
		Confirmed: true,
	}

	superAdminEntity_NotConfirmed := user.SuperAdminEntity{
		Email:     superAdminEmail_NotConfirmed,
		Name:      "Super Admin",
		Password:  &userPw,
		Confirmed: false,
	}

	// NOTA: Queste funzioni vengono inserite all'inizio perché non si cambia lo stato degli utenti creati

	integration.PreSetupCreateTenant(tenantId, true)(deps)
	defer integration.PostSetupDeleteTenant(t, tenantId)(deps)

	// Aggiungi tenant user
	preSetup, tenantUserId := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity)
	preSetup(deps)

	preSetup, tenantUserId_NotConfirmed := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity_NotConfirmed)
	preSetup(deps)

	// Aggiungi super admin
	preSetup, superAdminId := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity)
	preSetup(deps)

	preSetup, superAdminId_NotConfirmed := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity_NotConfirmed)
	preSetup(deps)

	// Request body
	tenantMemberBody := auth.RequestForgotPasswordBodyDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &tenantIdStr,
		},
		EmailField: dto.EmailField{
			Email: tenantUserEmail,
		},
	}

	tenantMemberBody_UserNotFound := auth.RequestForgotPasswordBodyDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &tenantIdStr,
		},
		EmailField: dto.EmailField{
			Email: tenantUserEmail_NotFound,
		},
	}

	tenantMemberBody_NotConfirmed := auth.RequestForgotPasswordBodyDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &tenantIdStr,
		},
		EmailField: dto.EmailField{
			Email: tenantUserEmail_NotConfirmed,
		},
	}

	tenantMemberBody_Invalid := auth.RequestForgotPasswordBodyDTO{
		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
			TenantId: &invalidTenantStr,
		},
		EmailField: dto.EmailField{
			Email: tenantUserEmail,
		},
	}

	superAdminBody := auth.RequestForgotPasswordBodyDTO{
		EmailField: dto.EmailField{
			Email: superAdminEmail,
		},
	}

	superAdminBody_NotFound := auth.RequestForgotPasswordBodyDTO{
		EmailField: dto.EmailField{
			Email: superAdminEmail_NotFound,
		},
	}

	superAdminBody_NotConfirmed := auth.RequestForgotPasswordBodyDTO{
		EmailField: dto.EmailField{
			Email: superAdminEmail_NotConfirmed,
		},
	}

	tests := []*helper.IntegrationTestCase{}

	basePath := "/api/v1/auth/forgot_password/request"

	// Success (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			nil,
		},
		Name:   "(Tenant Member) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantForgotPasswordTokenExistsForUser(t, tenantIdStr, *tenantUserId),
			integration.CheckSMTPMessageForToken(t, "forgot_password", true),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantForgotPasswordTokensForUser(t, tenantIdStr, *tenantUserId),
		},
	})

	// Success (super admin)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			nil,
		},
		Name:   "(Tenant Member) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminForgotPasswordTokenExistsForUser(t, tenantIdStr, *tenantUserId),
			integration.CheckSMTPMessageForToken(t, "forgot_password", true),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminForgotPasswordTokensForUser(t, tenantIdStr, *tenantUserId),
		},
	})

	// Fail: Binding JSON fallisce
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "Fail: JSON binding fail",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, tenantMemberBody_Invalid),

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckNoTenantForgotPasswordTokenForUser(t, tenantIdStr, *tenantUserId),
			integration.CheckSMTPMessageForToken(t, "forgot_password", false),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Tenant member not found
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member) Fail: user not found",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, tenantMemberBody_UserNotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckNoTenantForgotPasswordTokenForUser(t, tenantIdStr, *tenantUserId),
			integration.CheckSMTPMessageForToken(t, "forgot_password", false),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Super admin not found
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Fail: user not found",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, superAdminBody_NotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(user.ErrUserNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckNoSuperAdminForgotPasswordTokenForUser(t, *superAdminId),
			integration.CheckSMTPMessageForToken(t, "forgot_password", false),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Tenant member not confirmed
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member) Fail: user not confirmed",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, tenantMemberBody_NotConfirmed),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrAccountNotConfirmed),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckNoTenantForgotPasswordTokenForUser(t, tenantIdStr, *tenantUserId_NotConfirmed),
			integration.CheckSMTPMessageForToken(t, "forgot_password", false),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Super admin not confirmed
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Fail: user not confirmed",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, superAdminBody_NotConfirmed),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrAccountNotConfirmed),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckNoSuperAdminForgotPasswordTokenForUser(t, *superAdminId_NotConfirmed),
			integration.CheckSMTPMessageForToken(t, "forgot_password", false),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

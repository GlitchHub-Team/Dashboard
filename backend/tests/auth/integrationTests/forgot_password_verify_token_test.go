package auth_integration_test

import (
	"net/http"
	"testing"
	"time"

	"backend/internal/auth"
	"backend/internal/shared/identity"

	"backend/internal/user"
	"backend/tests/helper"
	"backend/tests/helper/integration"

	"backend/internal/infra/transport/http/dto"

	"github.com/google/uuid"
)

func TestVerifyForgotPasswordTokenIntegration(t *testing.T) {
	deps := helper.SetupIntegrationTest(t)

	// common values
	tenantId := uuid.New()
	tenantIdStr := tenantId.String()

	correctToken, _, err := deps.SecurityTokenGenerator.GenerateToken()
	t.Logf("correctToken: %v", correctToken)
	if err != nil {
		t.Fatalf("cannot create random security token: %v", err)
		return
	}
	wrongToken := "wrong-token-123"

	// Stato tabella tenant_members
	tenantUserEmail := "tu1@example.com"
	userPw := "pw123"
	tenantUserEntity := user.TenantMemberEntity{
		Email:     tenantUserEmail,
		Password:  &userPw,
		Name:      "Tenant User",
		Confirmed: true,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	superAdminEmail := "superadmin@m31.com"
	superAdminEntity := user.SuperAdminEntity{
		Email:     superAdminEmail,
		Name:      "Super Admin",
		Password:  &userPw,
		Confirmed: true,
	}

	// NOTA: Queste funzioni vengono inserite all'inizio perché non si cambia lo stato degli utenti creati

	integration.PreSetupCreateTenant(tenantId, true)(deps)
	defer integration.PostSetupDeleteTenant(t, tenantId)(deps)

	// Aggiungi tenant user
	preSetup, tenantUserId := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity)
	preSetup(deps)

	// Aggiungi super admin
	preSetup, superAdminId := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity)
	preSetup(deps)

	// Entities -----

	tenantPasswordTokenEntity := auth.TenantPasswordTokenEntity{
		Token:     correctToken,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId,
		ExpiresAt: deps.SecurityTokenGenerator.ExpiryFromNow(),
	}

	tenantPasswordTokenEntity_Expired := auth.TenantPasswordTokenEntity{
		Token:     correctToken,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}

	superAdminPasswordTokenEntity := auth.SuperAdminPasswordTokenEntity{
		Token:     correctToken,
		UserId:    *superAdminId,
		ExpiresAt: deps.SecurityTokenGenerator.ExpiryFromNow(),
	}

	superAdminPasswordTokenEntity_Expired := auth.SuperAdminPasswordTokenEntity{
		Token:     correctToken,
		UserId:    *superAdminId,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}

	// Request body
	tenantMemberBody := auth.VerifyForgotPasswordTokenBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &tenantIdStr,
			},
		},
	}

	superAdminBody := auth.VerifyForgotPasswordTokenBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken,
		},
	}

	tenantMemberBody_NotFound := auth.VerifyForgotPasswordTokenBodyDTO{
		TokenFields: dto.TokenFields{
			Token: wrongToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &tenantIdStr,
			},
		},
	}

	superAdminBody_NotFound := auth.VerifyForgotPasswordTokenBodyDTO{
		TokenFields: dto.TokenFields{
			Token: wrongToken,
		},
	}

	tests := []*helper.IntegrationTestCase{}

	basePath := "/api/v1/auth/forgot_password/verify_token"
	// Token exists (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantForgotPasswordToken(t, tenantPasswordTokenEntity),
		},
		Name:   "(Tenant Member) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantForgotPasswordTokenExists(t, tenantIdStr, correctToken),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantForgotPasswordToken(t, tenantIdStr, correctToken),
		},
	})

	// Token exists (super admin)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdminForgotPasswordToken(t, superAdminPasswordTokenEntity),
		},
		Name:   "(Super Admin) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminForgotPasswordTokenExists(t, correctToken),
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminForgotPasswordToken(t, correctToken),
		},
	})

	// Fail: Token not found (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member) Fail: token not found",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, tenantMemberBody_NotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrTokenNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckNoTenantForgotPasswordToken(t, tenantIdStr, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Token not found (super admin)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Super Admin) Fail: token not found",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, superAdminBody_NotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrTokenNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckNoSuperAdminForgotPasswordToken(t, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Token expired (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantForgotPasswordToken(t, tenantPasswordTokenEntity_Expired),
		},
		Name:   "(Tenant Member) Fail: token expired",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody_NotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrTokenNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantForgotPasswordTokenExists(t, tenantIdStr, correctToken),
			CheckTenantForgotPasswordTokenExpired(t, tenantIdStr, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantForgotPasswordToken(t, tenantIdStr, correctToken),
		},
	})

	// Fail: Token expired (super admin)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdminForgotPasswordToken(t, superAdminPasswordTokenEntity_Expired),
		},
		Name:   "(Super Admin) Fail: token expired",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody_NotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrTokenNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminForgotPasswordTokenExists(t, correctToken),
			CheckSuperAdminForgotPasswordTokenExpired(t, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminForgotPasswordToken(t, correctToken),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

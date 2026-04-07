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

func TestConfirmForgotPasswordIntegration(t *testing.T) {
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

	correctToken_NotConfirmed, _, err := deps.SecurityTokenGenerator.GenerateToken()
	t.Logf("correctToken_Confirmed: %v", correctToken_NotConfirmed)
	if err != nil {
		t.Fatalf("cannot create random security token: %v", err)
		return
	}
	wrongToken := "wrong-token-123"

	// Stato tabella tenant_members
	tenantUserEmail := "tu1@example.com"
	confirmedTenantUserEmail := "confirmed-tu1@example.com"

	oldPassword := "old-pa$$w0rd"
	newPassword := "new-pa$$w0rd"

	tenantUserEntity := user.TenantMemberEntity{
		Email:     tenantUserEmail,
		Password:  &oldPassword,
		Name:      "Tenant User",
		Confirmed: true,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	tenantUserEntity_NotConfirmed := user.TenantMemberEntity{
		Email:     confirmedTenantUserEmail,
		Password:  &oldPassword,
		Name:      "Confirmed Tenant User",
		Confirmed: false,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	superAdminEmail := "superadmin@m31.com"
	confirmedSuperAdminEmail := "confirmed-superadmin@m31.com"

	superAdminEntity := user.SuperAdminEntity{
		Email:     superAdminEmail,
		Name:      "Super Admin",
		Password:  &oldPassword,
		Confirmed: true,
	}

	superAdminEntity_NotConfirmed := user.SuperAdminEntity{
		Email:     confirmedSuperAdminEmail,
		Name:      "Confirmed Super Admin",
		Password:  &oldPassword,
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

	// Entities -----

	tenantConfirmTokenEntity := auth.TenantPasswordTokenEntity{
		Token:     correctToken,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId,
		ExpiresAt: time.Now().Add(1000 * time.Hour),
	}

	tenantConfirmTokenEntity_Expired := auth.TenantPasswordTokenEntity{
		Token:     correctToken,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}

	tenantConfirmTokenEntity_NotConfirmed := auth.TenantPasswordTokenEntity{
		Token:     correctToken_NotConfirmed,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId_NotConfirmed,
		ExpiresAt: time.Now().Add(1000 * time.Hour),
	}

	superAdminConfirmTokenEntity := auth.SuperAdminPasswordTokenEntity{
		Token:     correctToken,
		UserId:    *superAdminId,
		ExpiresAt: time.Now().Add(1000 * time.Hour),
	}

	superAdminConfirmTokenEntity_Expired := auth.SuperAdminPasswordTokenEntity{
		Token:     correctToken,
		UserId:    *superAdminId,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}

	superAdminConfirmTokenEntity_NotConfirmed := auth.SuperAdminPasswordTokenEntity{
		Token:     correctToken_NotConfirmed,
		UserId:    *superAdminId_NotConfirmed,
		ExpiresAt: time.Now().Add(1000 * time.Hour),
	}

	// Request body
	tenantMemberBody := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &tenantIdStr,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	tenantMemberBody_NotFound := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: wrongToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &tenantIdStr,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	tenantMemberBody_NotConfirmed := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken_NotConfirmed,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &tenantIdStr,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	superAdminBody := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken,
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	superAdminBody_NotFound := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: wrongToken,
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	superAdminBody_NotConfirmed := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken_NotConfirmed,
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	invalidUuid := "invalid-uuid"
	invalidBody := auth.ConfirmForgotPasswordBodyDTO{
		TokenFields: dto.TokenFields{
			Token: wrongToken,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &invalidUuid,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	tests := []*helper.IntegrationTestCase{}

	basePath := "/api/v1/auth/forgot_password"

	// Successo (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantForgotPasswordToken(t, tenantConfirmTokenEntity),
		},
		Name:   "(Tenant Member) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberConfirmed(t, tenantIdStr, *tenantUserId, true), // controlla utente confermato
			CheckNoTenantForgotPasswordToken(t, tenantIdStr, correctToken),  // controlla eliminazione token
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantForgotPasswordToken(t, tenantIdStr, correctToken),
		},
	})

	// Successo (super admin)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdminForgotPasswordToken(t, superAdminConfirmTokenEntity),
		},
		Name:   "(Super Admin) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminConfirmed(t, *superAdminId, true),      // controlla utente confermato
			CheckNoSuperAdminForgotPasswordToken(t, correctToken), // controlla eliminazione token
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminForgotPasswordToken(t, correctToken),
		},
	})

	// Fail: Binding JSON invalido
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{},
		Name:      "(Tenant Member) Fail: JSON binding fail",
		Method:    http.MethodPost,
		Path:      basePath,
		Header:    http.Header{},
		Body:      helper.MustJSONBody(t, invalidBody),

		WantStatusCode:   http.StatusBadRequest,
		WantResponseBody: "error",
		ResponseChecks:   []helper.IntegrationTestCheck{
			// CheckNoTenantForgotPasswordToken(t, tenantIdStr, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
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
			CheckTenantMemberConfirmed(t, tenantIdStr, *tenantUserId, false),
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
			CheckSuperAdminConfirmed(t, *superAdminId, false),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Token expired (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantForgotPasswordToken(t, tenantConfirmTokenEntity_Expired),
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
			PreSetupAddSuperAdminForgotPasswordToken(t, superAdminConfirmTokenEntity_Expired),
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

	// Fail: tenant member not confirmed
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantForgotPasswordToken(t, tenantConfirmTokenEntity_NotConfirmed),
			PreSetupAddTenantForgotPasswordToken(t, tenantConfirmTokenEntity),
		},
		Name:   "(Tenant Member) Fail: account not confirmed",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody_NotConfirmed),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrAccountNotConfirmed),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberConfirmed(t, tenantIdStr, *tenantUserId, true),
			CheckTenantForgotPasswordTokenExists(t, tenantIdStr, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantForgotPasswordToken(t, tenantIdStr, correctToken_NotConfirmed),
			PostSetupDeleteTenantForgotPasswordToken(t, tenantIdStr, correctToken),
		},
	})

	// Fail: super admin not confirmed
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdminForgotPasswordToken(t, superAdminConfirmTokenEntity_NotConfirmed),
			PreSetupAddSuperAdminForgotPasswordToken(t, superAdminConfirmTokenEntity),
		},
		Name:   "(Super Admin) Fail: account not confirmed",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody_NotConfirmed),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrAccountNotConfirmed),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminConfirmed(t, *tenantUserId, true),
			CheckSuperAdminForgotPasswordTokenExists(t, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminForgotPasswordToken(t, correctToken_NotConfirmed),
			PostSetupDeleteSuperAdminForgotPasswordToken(t, correctToken),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

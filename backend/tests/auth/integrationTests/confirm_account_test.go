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

func TestConfirmAccountIntegration(t *testing.T) {
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

	correctToken_Confirmed, _, err := deps.SecurityTokenGenerator.GenerateToken()
	t.Logf("correctToken_Confirmed: %v", correctToken_Confirmed)
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
		Confirmed: false,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	tenantUserEntity_Confirmed := user.TenantMemberEntity{
		Email:     confirmedTenantUserEmail,
		Password:  &oldPassword,
		Name:      "Confirmed Tenant User",
		Confirmed: true,
		Role:      string(identity.ROLE_TENANT_USER),
		TenantId:  tenantIdStr,
	}

	superAdminEmail := "superadmin@m31.com"
	confirmedSuperAdminEmail := "confirmed-superadmin@m31.com"

	superAdminEntity := user.SuperAdminEntity{
		Email:     superAdminEmail,
		Name:      "Super Admin",
		Password:  &oldPassword,
		Confirmed: false,
	}

	superAdminEntity_Confirmed := user.SuperAdminEntity{
		Email:     confirmedSuperAdminEmail,
		Name:      "Confirmed Super Admin",
		Password:  &oldPassword,
		Confirmed: true,
	}

	// NOTA: Queste funzioni vengono inserite all'inizio perché non si cambia lo stato degli utenti creati

	integration.PreSetupCreateTenant(tenantId, true)(deps)
	defer integration.PostSetupDeleteTenant(t, tenantId)(deps)

	// Aggiungi tenant user
	preSetup, tenantUserId := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity)
	preSetup(deps)

	preSetup, tenantUserId_Confirmed := integration.PreSetupAddTenantMember_ReturnUserId(t, &tenantUserEntity_Confirmed)
	preSetup(deps)

	// Aggiungi super admin
	preSetup, superAdminId := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity)
	preSetup(deps)

	preSetup, superAdminId_Confirmed := integration.PreSetupAddSuperAdmin_ReturnUserId(t, &superAdminEntity_Confirmed)
	preSetup(deps)

	// Entities -----

	tenantConfirmTokenEntity := auth.TenantConfirmTokenEntity{
		Token:     correctToken,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId,
		ExpiresAt: deps.SecurityTokenGenerator.ExpiryFromNow(),
	}

	tenantConfirmTokenEntity_Expired := auth.TenantConfirmTokenEntity{
		Token:     correctToken,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}

	tenantConfirmTokenEntity_Confirmed := auth.TenantConfirmTokenEntity{
		Token:     correctToken_Confirmed,
		TenantId:  tenantIdStr,
		UserId:    *tenantUserId_Confirmed,
		ExpiresAt: deps.SecurityTokenGenerator.ExpiryFromNow(),
	}

	superAdminConfirmTokenEntity := auth.SuperAdminConfirmTokenEntity{
		Token:     correctToken,
		UserId:    *superAdminId,
		ExpiresAt: deps.SecurityTokenGenerator.ExpiryFromNow(),
	}

	superAdminConfirmTokenEntity_Expired := auth.SuperAdminConfirmTokenEntity{
		Token:     correctToken,
		UserId:    *superAdminId,
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}

	superAdminConfirmTokenEntity_Confirmed := auth.SuperAdminConfirmTokenEntity{
		Token:     correctToken_Confirmed,
		UserId:    *superAdminId_Confirmed,
		ExpiresAt: deps.SecurityTokenGenerator.ExpiryFromNow(),
	}

	// Request body
	tenantMemberBody := auth.ConfirmUserAccountBodyDTO{
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

	tenantMemberBody_NotFound := auth.ConfirmUserAccountBodyDTO{
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

	tenantMemberBody_Confirmed := auth.ConfirmUserAccountBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken_Confirmed,
			TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
				TenantId: &tenantIdStr,
			},
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	superAdminBody := auth.ConfirmUserAccountBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken,
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	superAdminBody_NotFound := auth.ConfirmUserAccountBodyDTO{
		TokenFields: dto.TokenFields{
			Token: wrongToken,
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	superAdminBody_Confirmed := auth.ConfirmUserAccountBodyDTO{
		TokenFields: dto.TokenFields{
			Token: correctToken_Confirmed,
		},
		NewPasswordField: dto.NewPasswordField{
			NewPassword: newPassword,
		},
	}

	invalidUuid := "invalid-uuid"
	invalidBody := auth.ConfirmUserAccountBodyDTO{
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

	// Requester

	expectedTenantUserRequester := identity.Requester{
		RequesterUserId:   *tenantUserId,
		RequesterTenantId: &tenantId,
		RequesterRole:     identity.ROLE_TENANT_USER,
	}

	expectedSuperAdminRequester := identity.Requester{
		RequesterUserId:   *superAdminId,
		RequesterTenantId: nil,
		RequesterRole:     identity.ROLE_SUPER_ADMIN,
	}

	tests := []*helper.IntegrationTestCase{}

	basePath := "/api/v1/auth/confirm_account"

	// Successo (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantConfirmAccountToken(t, tenantConfirmTokenEntity),
		},
		Name:   "(Tenant Member) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberConfirmed(t, tenantIdStr, *tenantUserId, true),            // controlla utente confermato
			integration.CheckNoTenantConfirmAccountToken(t, tenantIdStr, correctToken), // controlla eliminazione token
			CheckValidJWTInResponse(t, expectedTenantUserRequester),                    // controlla JWT corretto
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantConfirmAccountToken(t, tenantIdStr, correctToken),
		},
	})

	// Successo (super admin)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdminConfirmAccountToken(t, superAdminConfirmTokenEntity),
		},
		Name:   "(Super Admin) Success",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody),

		WantStatusCode:   http.StatusOK,
		WantResponseBody: "",
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminConfirmed(t, *superAdminId, true),                  // controlla utente confermato
			integration.CheckNoSuperAdminConfirmAccountToken(t, correctToken), // controlla eliminazione token
			CheckValidJWTInResponse(t, expectedSuperAdminRequester),           // controlla JWT corretto
		},
		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminConfirmAccountToken(t, correctToken),
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
			// integration.CheckNoTenantConfirmAccountToken(t, tenantIdStr, correctToken),
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
			integration.CheckNoSuperAdminConfirmAccountToken(t, correctToken),
			CheckSuperAdminConfirmed(t, *tenantUserId, false),
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
			integration.CheckNoSuperAdminConfirmAccountToken(t, correctToken),
			CheckSuperAdminConfirmed(t, *superAdminId, false),
		},

		PostSetups: []helper.IntegrationTestPostSetup{},
	})

	// Fail: Token expired (tenant member)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantConfirmAccountToken(t, tenantConfirmTokenEntity_Expired),
		},
		Name:   "(Tenant Member) Fail: token expired",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody_NotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrTokenNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckTenantConfirmAccountTokenExists(t, tenantIdStr, correctToken),
			integration.CheckTenantConfirmAccountTokenExpired(t, tenantIdStr, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantConfirmAccountToken(t, tenantIdStr, correctToken),
		},
	})

	// Fail: Token expired (super admin)
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdminConfirmAccountToken(t, superAdminConfirmTokenEntity_Expired),
		},
		Name:   "(Super Admin) Fail: token expired",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody_NotFound),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrTokenNotFound),
		ResponseChecks: []helper.IntegrationTestCheck{
			integration.CheckSuperAdminConfirmAccountTokenExists(t, correctToken),
			integration.CheckSuperAdminConfirmAccountTokenExpired(t, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminConfirmAccountToken(t, correctToken),
		},
	})

	// Fail: tenant member already confirmed
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddTenantConfirmAccountToken(t, tenantConfirmTokenEntity_Confirmed),
			PreSetupAddTenantConfirmAccountToken(t, tenantConfirmTokenEntity),
		},
		Name:   "(Tenant Member) Fail: account already confirmed",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, tenantMemberBody_Confirmed),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrAccountAlreadyConfirmed),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckTenantMemberConfirmed(t, tenantIdStr, *tenantUserId, true),
			integration.CheckTenantConfirmAccountTokenExists(t, tenantIdStr, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteTenantConfirmAccountToken(t, tenantIdStr, correctToken_Confirmed),
			PostSetupDeleteTenantConfirmAccountToken(t, tenantIdStr, correctToken),
		},
	})

	// Fail: super admin already confirmed
	tests = append(tests, &helper.IntegrationTestCase{
		PreSetups: []helper.IntegrationTestPreSetup{
			PreSetupAddSuperAdminConfirmAccountToken(t, superAdminConfirmTokenEntity_Confirmed),
			PreSetupAddSuperAdminConfirmAccountToken(t, superAdminConfirmTokenEntity),
		},
		Name:   "(Super Admin) Fail: account already confirmed",
		Method: http.MethodPost,
		Path:   basePath,
		Header: http.Header{},
		Body:   helper.MustJSONBody(t, superAdminBody_Confirmed),

		WantStatusCode:   http.StatusNotFound,
		WantResponseBody: helper.ErrJsonString(auth.ErrAccountAlreadyConfirmed),
		ResponseChecks: []helper.IntegrationTestCheck{
			CheckSuperAdminConfirmed(t, *tenantUserId, true),
			integration.CheckSuperAdminConfirmAccountTokenExists(t, correctToken),
		},

		PostSetups: []helper.IntegrationTestPostSetup{
			PostSetupDeleteSuperAdminConfirmAccountToken(t, correctToken_Confirmed),
			PostSetupDeleteSuperAdminConfirmAccountToken(t, correctToken),
		},
	})

	helper.RunIntegrationTests(t, tests, deps)
}

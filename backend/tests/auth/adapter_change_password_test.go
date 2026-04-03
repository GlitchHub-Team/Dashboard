package auth_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"backend/internal/auth"
	"backend/internal/shared/identity"
	"backend/internal/user"
	"backend/tests/auth/mocks"
	cryptoMocks "backend/tests/shared/crypto/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

type mockSetupFunc_changePasswordTokenPgAdapter func(
	hasher *cryptoMocks.MockSecretHasher,
	tokenGenerator *cryptoMocks.MockSecurityTokenGenerator,
	tenantMemberRepo *mocks.MockTenantPasswordTokenRepository,
	superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
) *gomock.Call

func setupMockSteps_ChangePasswordTokenPgAdapter(
	t *testing.T,
	setupSteps []mockSetupFunc_changePasswordTokenPgAdapter,
) (
	mockHasher *cryptoMocks.MockSecretHasher,
	mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator,
	mockTenantRepo *mocks.MockTenantPasswordTokenRepository,
	mockSuperAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
) {
	ctrl := gomock.NewController(t)

	mockHasher = cryptoMocks.NewMockSecretHasher(ctrl)
	mockTokenGenerator = cryptoMocks.NewMockSecurityTokenGenerator(ctrl)
	mockTenantRepo = mocks.NewMockTenantPasswordTokenRepository(ctrl)
	mockSuperAdminRepo = mocks.NewMockSuperAdminPasswordTokenRepository(ctrl)

	var expectedCalls []any
	for _, step := range setupSteps {
		if call := step(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo); call != nil {
			expectedCalls = append(expectedCalls, call)
		}
	}

	if len(expectedCalls) > 0 {
		gomock.InOrder(expectedCalls...)
	}

	return
}

func TestChangePasswordTokenPgAdapter_NewForgotPasswordToken(t *testing.T) {
	type testCase struct {
		name          string
		inputUser     user.User
		setupSteps    []mockSetupFunc_changePasswordTokenPgAdapter
		expectedToken string
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------
	tenantMemberUuid := uuid.New()
	tenantMemberUser := user.User{
		Id:           uint(1),
		Name:         "username",
		Email:        "info@example.com",
		PasswordHash: new(string),
		Role:         identity.ROLE_TENANT_ADMIN,
		TenantId:     &tenantMemberUuid,
		Confirmed:    true,
	}

	superAdminUser := user.User{
		Id:           uint(1),
		Name:         "username",
		Email:        "info@example.com",
		PasswordHash: new(string),
		Role:         identity.ROLE_SUPER_ADMIN,
		TenantId:     nil,
		Confirmed:    true,
	}

	invalidUser := user.User{
		Role: "",
	}

	mockTenantMemberEntity := gomock.AssignableToTypeOf(&auth.TenantPasswordTokenEntity{})
	mockSuperAdminEntity := gomock.AssignableToTypeOf(&auth.SuperAdminPasswordTokenEntity{})

	expectedRawToken := "raw-token"
	expectedHashedToken := "hashed-token"

	// Step 1: generate token
	step1GenerateTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tokenGenerator.EXPECT().
			GenerateToken().
			Return(expectedRawToken, expectedHashedToken, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1GenerateTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tokenGenerator.EXPECT().
			GenerateToken().
			Return("", "", errMockStep1).
			Times(1)
	}

	// Step 2: Save token
	errMockStep2 := errors.New("unexpected error in step 2")

	// Tenant Member
	step2SaveTokenOk_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveToken(mockTenantMemberEntity).
			Return(nil).
			Times(1)
	}

	step2SaveTokenError_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveToken(mockTenantMemberEntity).
			Return(errMockStep2).
			Times(1)
	}

	step2SaveTokenNeverCalled_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveToken(gomock.Any()).
			Times(0)
	}

	// Super Admin
	step2SaveTokenOk_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveToken(mockSuperAdminEntity).
			Return(nil).
			Times(1)
	}

	step2SaveTokenError_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveToken(mockSuperAdminEntity).
			Return(errMockStep2).
			Times(1)
	}

	step2SaveTokenNeverCalled_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveToken(gomock.Any()).
			Times(0)
	}

	cases := []testCase{
		// COMMON ===================================================
		// Step 1
		{
			name:      "Fail (step 1): unexpected error",
			inputUser: user.User{}, // Irrilevante
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1GenerateTokenError,
			},
			expectedToken: "",
			expectedError: errMockStep1,
		},
		// Step 2
		{
			name:      "Fail (step 2): unknown role",
			inputUser: invalidUser,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1GenerateTokenOk,
				step2SaveTokenNeverCalled_Tenant,
				step2SaveTokenNeverCalled_SuperAdmin,
			},
			expectedToken: "",
			expectedError: identity.ErrUnknownRole,
		},

		// SUPER ADMIN ==============================================
		// Success
		{
			name:      "(Super Admin) Success",
			inputUser: superAdminUser,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1GenerateTokenOk,
				step2SaveTokenOk_SuperAdmin,
				step2SaveTokenNeverCalled_Tenant,
			},
			expectedToken: expectedRawToken,
			expectedError: nil,
		},

		// Step 2
		{
			name:      "(Super Admin) Fail (step 2): unexpected error",
			inputUser: superAdminUser,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1GenerateTokenOk,
				step2SaveTokenError_SuperAdmin,
				step2SaveTokenNeverCalled_Tenant,
			},
			expectedToken: "",
			expectedError: errMockStep2,
		},

		// TENANT MEMBER ==============================================
		// Success
		{
			name:      "(Tenant Member) Success",
			inputUser: tenantMemberUser,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1GenerateTokenOk,
				step2SaveTokenOk_Tenant,
				step2SaveTokenNeverCalled_SuperAdmin,
			},
			expectedToken: expectedRawToken,
			expectedError: nil,
		},

		// Step 2
		{
			name:      "(Tenant Member) Fail (step 2): unexpected error",
			inputUser: tenantMemberUser,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1GenerateTokenOk,
				step2SaveTokenError_Tenant,
				step2SaveTokenNeverCalled_SuperAdmin,
			},
			expectedToken: "",
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ChangePasswordTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewChangePasswordTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			rawToken, err := adapter.NewForgotPasswordToken(tc.inputUser)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if tc.expectedToken != rawToken {
				t.Errorf("want token %v, got %v", tc.expectedToken, rawToken)
			}
		})
	}
}

func TestChangePasswordTokenPgAdapter_DeleteForgotPasswordToken(t *testing.T) {
	type testCase struct {
		name          string
		inputToken    auth.ForgotPasswordToken
		setupSteps    []mockSetupFunc_changePasswordTokenPgAdapter
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------

	targetHashedToken := "hashed-token"
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetUserId := uint(1)
	targetExpiryDate := time.Now()

	tenantMemberToken := auth.ForgotPasswordToken{
		HashedToken: targetHashedToken,
		TenantId:    &targetTenantId,
		UserId:      targetUserId,
		ExpiryDate:  targetExpiryDate,
	}

	superAdminToken := auth.ForgotPasswordToken{
		HashedToken: targetHashedToken,
		TenantId:    nil,
		UserId:      targetUserId,
		ExpiryDate:  targetExpiryDate,
	}

	expectedTenantMemberEntity := &auth.TenantPasswordTokenEntity{
		Token:     targetHashedToken,
		TenantId:  &targetTenantIdStr,
		UserId:    targetUserId,
		ExpiresAt: targetExpiryDate,
	}

	expectedSuperAdminEntity := &auth.SuperAdminPasswordTokenEntity{
		Token:     targetHashedToken,
		UserId:    targetUserId,
		ExpiresAt: targetExpiryDate,
	}

	// Step: Save token
	errMock := errors.New("unexpected error in step 2")

	// Tenant Member ----------------------
	step1DeleteTokenOk_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteToken(expectedTenantMemberEntity).
			Return(nil).
			Times(1)
	}

	stepDeleteTokenError_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteToken(expectedTenantMemberEntity).
			Return(errMock).
			Times(1)
	}

	stepDeleteTokenNeverCalled_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteToken(gomock.Any()).
			Times(0)
	}

	// Super Admin
	stepDeleteTokenOk_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			DeleteToken(expectedSuperAdminEntity).
			Return(nil).
			Times(1)
	}

	stepDeleteTokenError_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			DeleteToken(expectedSuperAdminEntity).
			Return(errMock).
			Times(1)
	}

	stepDeleteTokenNeverCalled_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			DeleteToken(gomock.Any()).
			Times(0)
	}

	cases := []testCase{
		// TENANT MEMBER ==============================================
		// Success
		{
			name:       "(Tenant Member) Success",
			inputToken: tenantMemberToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1DeleteTokenOk_Tenant,
				stepDeleteTokenNeverCalled_SuperAdmin,
			},
			expectedError: nil,
		},

		// Step 1
		{
			name:       "(Tenant Member) Fail (step 1): unexpected error",
			inputToken: tenantMemberToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				stepDeleteTokenError_Tenant,
				stepDeleteTokenNeverCalled_SuperAdmin,
			},
			expectedError: errMock,
		},

		// SUPER ADMIN ==============================================
		// Success
		{
			name:       "(Super Admin) Success",
			inputToken: superAdminToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				stepDeleteTokenOk_SuperAdmin,
				stepDeleteTokenNeverCalled_Tenant,
			},
			expectedError: nil,
		},

		// Step 1
		{
			name:       "(Super Admin) Fail (step 1): unexpected error",
			inputToken: superAdminToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				stepDeleteTokenError_SuperAdmin,
				stepDeleteTokenNeverCalled_Tenant,
			},
			expectedError: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ChangePasswordTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewChangePasswordTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			err := adapter.DeleteForgotPasswordToken(tc.inputToken)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}
		})
	}
}

func TestChangePasswordTokenPgAdapter_GetTenantMemberByForgotPasswordToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTenantId    uuid.UUID
		inputTokenString string
		setupSteps       []mockSetupFunc_changePasswordTokenPgAdapter
		expectedUser     user.User
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetExpiryDate := time.Now()
	targetToken := "token"
	expectedHashedToken := "hashed-token-123"

	targetUserId := uint(1)
	targetUserName := "name"
	targetUserEmail := "email@email.com"
	targetPasswordHash := "hash"
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetConfirmed := true
	targetRole := identity.ROLE_TENANT_USER
	// tenantMemberToken := auth.ForgotPasswordToken{
	// 	HashedToken: targetToken,
	// 	TenantId:    &targetTenantId,
	// 	UserId:      targetUserId,
	// 	ExpiryDate:  targetExpiryDate,
	// }

	expectedTenantMemberEntity := &auth.TenantPasswordTokenEntity{
		Token:    targetToken,
		TenantId: &targetTenantIdStr,
		UserId:   targetUserId,
		TenantMember: user.TenantMemberEntity{
			ID:        targetUserId,
			Name:      targetUserName,
			Email:     targetUserEmail,
			Password:  &targetPasswordHash,
			Confirmed: targetConfirmed,
			Role:      string(targetRole),
			TenantId:  targetTenantId.String(),
		},
		ExpiresAt: targetExpiryDate,
	}

	expectedUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &targetPasswordHash,
		TenantId:     &targetTenantId,
		Role:         targetRole,
		Confirmed:    targetConfirmed,
	}

	// Step 1: hash token
	step1HashTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return(expectedHashedToken, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")

	step1HashTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return("", errMockStep1).
			Times(1)
	}

	// Step 2: Save token
	errMockStep2 := errors.New("unexpected error in step 2")

	// Get token
	step2GetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTokenWithUser(targetTenantId.String(), expectedHashedToken).
			Return(expectedTenantMemberEntity, nil).
			Times(1)
	}

	step2GetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTokenWithUser(targetTenantId.String(), expectedHashedToken).
			Return(&auth.TenantPasswordTokenEntity{}, errMockStep2).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		// Step 1
		{
			name:             "Fail (step 1): unexpected error",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:             "Fail (step 2): unexpected error",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ChangePasswordTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewChangePasswordTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			userFound, err := adapter.GetTenantMemberByForgotPasswordToken(tc.inputTenantId, tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, userFound) {
				t.Errorf("want %#v, got user %#v", tc.expectedUser, userFound)
			}
		})
	}
}

func TestChangePasswordTokenPgAdapter_GetSuperAdminByForgotPasswordToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTokenString string
		setupSteps       []mockSetupFunc_changePasswordTokenPgAdapter
		expectedUser     user.User
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetExpiryDate := time.Now()
	targetToken := "token"
	expectedHashedToken := "hashed-token-123"

	targetUserId := uint(1)
	targetUserName := "name"
	targetUserEmail := "email@email.com"
	targetPasswordHash := "hash"
	targetConfirmed := true
	targetRole := identity.ROLE_SUPER_ADMIN

	expectedTenantMemberEntity := &auth.SuperAdminPasswordTokenEntity{
		Token:  targetToken,
		UserId: targetUserId,
		SuperAdmin: user.SuperAdminEntity{
			ID:        targetUserId,
			Name:      targetUserName,
			Email:     targetUserEmail,
			Password:  &targetPasswordHash,
			Confirmed: targetConfirmed,
		},
		ExpiresAt: targetExpiryDate,
	}

	expectedUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &targetPasswordHash,
		TenantId:     nil,
		Role:         targetRole,
		Confirmed:    targetConfirmed,
	}

	// Step 1: hash token
	step1HashTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return(expectedHashedToken, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")

	step1HashTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return("", errMockStep1).
			Times(1)
	}

	// Step 2: Save token
	errMockStep2 := errors.New("unexpected error in step 2")

	// Get token
	step2GetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetTokenWithUser(expectedHashedToken).
			Return(expectedTenantMemberEntity, nil).
			Times(1)
	}

	step2GetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetTokenWithUser(expectedHashedToken).
			Return(&auth.SuperAdminPasswordTokenEntity{}, errMockStep2).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		// Step 1
		{
			name:             "Fail (step 1): unexpected error",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:             "Fail (step 2): unexpected error",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ChangePasswordTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewChangePasswordTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			userFound, err := adapter.GetSuperAdminByForgotPasswordToken(tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, userFound) {
				t.Errorf("want %#v, got user %#v", tc.expectedUser, userFound)
			}
		})
	}
}

func TestChangePasswordTokenPgAdapter_GetTenantForgotPasswordToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTenantId    uuid.UUID
		inputTokenString string
		setupSteps       []mockSetupFunc_changePasswordTokenPgAdapter
		expectedToken    auth.ForgotPasswordToken
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetToken := "token"
	expectedHashedToken := "hashed-token-123"

	targetUserId := uint(1)
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetExpiry := time.Now()

	expectedTenantMemberToken := &auth.TenantPasswordTokenEntity{
		Token:     expectedHashedToken,
		TenantId:  &targetTenantIdStr,
		UserId:    targetUserId,
		ExpiresAt: targetExpiry,
	}

	expectedDomainToken := auth.ForgotPasswordToken{
		HashedToken: expectedHashedToken,
		TenantId:    &targetTenantId,
		UserId:      targetUserId,
		ExpiryDate:  targetExpiry,
	}

	// Step 1: hash token
	step1HashTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return(expectedHashedToken, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")

	step1HashTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return("", errMockStep1).
			Times(1)
	}

	// Step 2: Save token
	errMockStep2 := errors.New("unexpected error in step 2")

	// Get token
	step2GetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetToken(targetTenantId.String(), expectedHashedToken).
			Return(expectedTenantMemberToken, nil).
			Times(1)
	}

	step2GetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetToken(targetTenantId.String(), expectedHashedToken).
			Return(&auth.TenantPasswordTokenEntity{}, errMockStep2).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenOk,
			},
			expectedToken: expectedDomainToken,
			expectedError: nil,
		},

		// Step 1
		{
			name:             "Fail (step 1): unexpected error",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenError,
			},
			expectedToken: auth.ForgotPasswordToken{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:             "Fail (step 2): unexpected error",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenError,
			},
			expectedToken: auth.ForgotPasswordToken{},
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ChangePasswordTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewChangePasswordTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			token, err := adapter.GetTenantForgotPasswordToken(tc.inputTenantId, tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedToken, token) {
				t.Errorf("want %#v, got token %#v", tc.expectedToken, token)
			}
		})
	}
}

func TestChangePasswordTokenPgAdapter_GetSuperAdminForgotPasswordToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTokenString string
		setupSteps       []mockSetupFunc_changePasswordTokenPgAdapter
		expectedUser     auth.ForgotPasswordToken
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetExpiryDate := time.Now()
	targetToken := "token"
	expectedHashedToken := "hashed-token-123"

	targetUserId := uint(1)

	expectedTenantMemberEntity := &auth.SuperAdminPasswordTokenEntity{
		Token:     expectedHashedToken,
		UserId:    targetUserId,
		ExpiresAt: targetExpiryDate,
	}

	expectedDomainToken := auth.ForgotPasswordToken{
		HashedToken: expectedHashedToken,
		TenantId:    nil,
		UserId:      targetUserId,
		ExpiryDate:  targetExpiryDate,
	}
	// Step 1: hash token
	step1HashTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return(expectedHashedToken, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")

	step1HashTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return hasher.EXPECT().
			HashSecret(targetToken).
			Return("", errMockStep1).
			Times(1)
	}

	// Step 2: Save token
	errMockStep2 := errors.New("unexpected error in step 2")

	// Get token
	step2GetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetToken(expectedHashedToken).
			Return(expectedTenantMemberEntity, nil).
			Times(1)
	}

	step2GetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantPasswordTokenRepository, superAdminRepo *mocks.MockSuperAdminPasswordTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetToken(expectedHashedToken).
			Return(&auth.SuperAdminPasswordTokenEntity{}, errMockStep2).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenOk,
			},
			expectedUser:  expectedDomainToken,
			expectedError: nil,
		},

		// Step 1
		{
			name:             "Fail (step 1): unexpected error",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenError,
			},
			expectedUser:  auth.ForgotPasswordToken{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:             "Fail (step 2): unexpected error",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_changePasswordTokenPgAdapter{
				step1HashTokenOk,
				step2GetTokenError,
			},
			expectedUser:  auth.ForgotPasswordToken{},
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ChangePasswordTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewChangePasswordTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			tokenObj, err := adapter.GetSuperAdminForgotPasswordToken(tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, tokenObj) {
				t.Errorf("want %#v, got user %#v", tc.expectedUser, tokenObj)
			}
		})
	}
}

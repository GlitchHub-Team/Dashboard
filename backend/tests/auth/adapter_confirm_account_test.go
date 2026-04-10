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

type mockSetupFunc_confirmAccountTokenPgAdapter func(
	hasher *cryptoMocks.MockSecretHasher,
	tokenGenerator *cryptoMocks.MockSecurityTokenGenerator,
	tenantMemberRepo *mocks.MockTenantConfirmTokenRepository,
	superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
) *gomock.Call

func setupMockSteps_ConfirmAccountTokenPgAdapter(
	t *testing.T,
	setupSteps []mockSetupFunc_confirmAccountTokenPgAdapter,
) (
	mockHasher *cryptoMocks.MockSecretHasher,
	mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator,
	mockTenantRepo *mocks.MockTenantConfirmTokenRepository,
	mockSuperAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
) {
	ctrl := gomock.NewController(t)

	mockHasher = cryptoMocks.NewMockSecretHasher(ctrl)
	mockTokenGenerator = cryptoMocks.NewMockSecurityTokenGenerator(ctrl)
	mockTenantRepo = mocks.NewMockTenantConfirmTokenRepository(ctrl)
	mockSuperAdminRepo = mocks.NewMockSuperAdminConfirmTokenRepository(ctrl)

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

func TestConfirmAccountTokenPgAdapter_NewConfirmAccountToken(t *testing.T) {
	type testCase struct {
		name          string
		inputUser     user.User
		setupSteps    []mockSetupFunc_confirmAccountTokenPgAdapter
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

	mockTenantMemberEntity := gomock.AssignableToTypeOf(&auth.TenantConfirmTokenEntity{})
	mockSuperAdminEntity := gomock.AssignableToTypeOf(&auth.SuperAdminConfirmTokenEntity{})

	expectedRawToken := "raw-token"
	expectedHashedToken := "hashed-token"

	expectedExpiry := time.Now().Add(time.Hour)

	// Step 1: generate token
	step1GenerateTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tokenGenerator.EXPECT().
			GenerateToken().
			Return(expectedRawToken, expectedHashedToken, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1GenerateTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tokenGenerator.EXPECT().
			GenerateToken().
			Return("", "", errMockStep1).
			Times(1)
	}

	// Step 1.1: generate expiry
	step1_1GenerateExpiryOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tokenGenerator.EXPECT().
			ExpiryFromNow().
			Return(expectedExpiry).
			Times(1)
	}

	// Step 2: Save token
	errMockStep2 := errors.New("unexpected error in step 2")

	// Tenant Member
	step2SaveTokenOk_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveToken(mockTenantMemberEntity).
			Return(nil).
			Times(1)
	}

	step2SaveTokenError_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveToken(mockTenantMemberEntity).
			Return(errMockStep2).
			Times(1)
	}

	step2SaveTokenNeverCalled_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			SaveToken(gomock.Any()).
			Times(0)
	}

	// Super Admin
	step2SaveTokenOk_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveToken(mockSuperAdminEntity).
			Return(nil).
			Times(1)
	}

	step2SaveTokenError_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			SaveToken(mockSuperAdminEntity).
			Return(errMockStep2).
			Times(1)
	}

	step2SaveTokenNeverCalled_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
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
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				step1GenerateTokenError,
			},
			expectedToken: "",
			expectedError: errMockStep1,
		},
		// Step 2
		{
			name:      "Fail (step 2): unknown role",
			inputUser: invalidUser,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				step1GenerateTokenOk,
				step1_1GenerateExpiryOk,
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
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				step1GenerateTokenOk,
				step1_1GenerateExpiryOk,
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
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				step1GenerateTokenOk,
				step1_1GenerateExpiryOk,
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
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				step1GenerateTokenOk,
				step1_1GenerateExpiryOk,
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
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				step1GenerateTokenOk,
				step1_1GenerateExpiryOk,
				step2SaveTokenError_Tenant,
				step2SaveTokenNeverCalled_SuperAdmin,
			},
			expectedToken: "",
			expectedError: errMockStep2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ConfirmAccountTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewConfirmAccountTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			rawToken, err := adapter.NewConfirmAccountToken(tc.inputUser)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if tc.expectedToken != rawToken {
				t.Errorf("want token %v, got %v", tc.expectedToken, rawToken)
			}
		})
	}
}

func TestConfirmAccountTokenPgAdapter_DeleteConfirmAccountToken(t *testing.T) {
	type testCase struct {
		name          string
		inputToken    auth.ConfirmAccountToken
		setupSteps    []mockSetupFunc_confirmAccountTokenPgAdapter
		expectedError error
	}

	// Input ----------------------------------------------------------------------------------

	targetHashedToken := "hashed-token"
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetUserId := uint(1)
	targetExpiryDate := time.Now()

	tenantMemberToken := auth.ConfirmAccountToken{
		Token:      targetHashedToken,
		TenantId:   &targetTenantId,
		UserId:     targetUserId,
		ExpiryDate: targetExpiryDate,
	}

	superAdminToken := auth.ConfirmAccountToken{
		Token:      targetHashedToken,
		TenantId:   nil,
		UserId:     targetUserId,
		ExpiryDate: targetExpiryDate,
	}

	expectedTenantMemberEntity := &auth.TenantConfirmTokenEntity{
		Token:     targetHashedToken,
		TenantId:  targetTenantIdStr,
		UserId:    targetUserId,
		ExpiresAt: targetExpiryDate,
	}

	expectedSuperAdminEntity := &auth.SuperAdminConfirmTokenEntity{
		Token:     targetHashedToken,
		UserId:    targetUserId,
		ExpiresAt: targetExpiryDate,
	}

	// Step: Save token
	errMock := errors.New("unexpected error in step 2")

	// Tenant Member ----------------------
	step1DeleteTokenOk_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteToken(expectedTenantMemberEntity).
			Return(nil).
			Times(1)
	}

	stepDeleteTokenError_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteToken(expectedTenantMemberEntity).
			Return(errMock).
			Times(1)
	}

	stepDeleteTokenNeverCalled_Tenant := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			DeleteToken(gomock.Any()).
			Times(0)
	}

	// Super Admin
	stepDeleteTokenOk_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			DeleteToken(expectedSuperAdminEntity).
			Return(nil).
			Times(1)
	}

	stepDeleteTokenError_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			DeleteToken(expectedSuperAdminEntity).
			Return(errMock).
			Times(1)
	}

	stepDeleteTokenNeverCalled_SuperAdmin := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
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
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				step1DeleteTokenOk_Tenant,
				stepDeleteTokenNeverCalled_SuperAdmin,
			},
			expectedError: nil,
		},

		// Step 1
		{
			name:       "(Tenant Member) Fail (step 1): unexpected error",
			inputToken: tenantMemberToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
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
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepDeleteTokenOk_SuperAdmin,
				stepDeleteTokenNeverCalled_Tenant,
			},
			expectedError: nil,
		},

		// Step 1
		{
			name:       "(Super Admin) Fail (step 1): unexpected error",
			inputToken: superAdminToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepDeleteTokenError_SuperAdmin,
				stepDeleteTokenNeverCalled_Tenant,
			},
			expectedError: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ConfirmAccountTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewConfirmAccountTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			err := adapter.DeleteConfirmAccountToken(tc.inputToken)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}
		})
	}
}

func TestConfirmAccountTokenPgAdapter_GetTenantMemberByConfirmAccountToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTenantId    uuid.UUID
		inputTokenString string
		setupSteps       []mockSetupFunc_confirmAccountTokenPgAdapter
		expectedUser     user.User
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetExpiryDate := time.Now()
	targetToken := "token"

	targetUserId := uint(1)
	targetUserName := "name"
	targetUserEmail := "email@email.com"
	targetPasswordHash := "hash"
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetConfirmed := true
	targetRole := identity.ROLE_TENANT_USER

	expectedTenantMemberEntity := &auth.TenantConfirmTokenEntity{
		Token:    targetToken,
		TenantId: targetTenantIdStr,
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

	// Save token
	errMockStep1 := errors.New("unexpected error in step 2")

	// Get token
	stepGetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTokenWithUser(targetTenantId.String(), targetToken).
			Return(expectedTenantMemberEntity, nil).
			Times(1)
	}

	stepGetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetTokenWithUser(targetTenantId.String(), targetToken).
			Return(&auth.TenantConfirmTokenEntity{}, errMockStep1).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		// Fail
		{
			name:             "Fail: unexpected error",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenError,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ConfirmAccountTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewConfirmAccountTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			userFound, err := adapter.GetTenantMemberByConfirmAccountToken(tc.inputTenantId, tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, userFound) {
				t.Errorf("want %#v, got user %#v", tc.expectedUser, userFound)
			}
		})
	}
}

func TestConfirmAccountTokenPgAdapter_GetSuperAdminByConfirmAccountToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTokenString string
		setupSteps       []mockSetupFunc_confirmAccountTokenPgAdapter
		expectedUser     user.User
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetExpiryDate := time.Now()
	targetToken := "token"

	targetUserId := uint(1)
	targetUserName := "name"
	targetUserEmail := "email@email.com"
	targetPasswordHash := "hash"
	targetConfirmed := true
	targetRole := identity.ROLE_SUPER_ADMIN

	expectedTenantMemberEntity := &auth.SuperAdminConfirmTokenEntity{
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

	// Step 2: Save token
	errMock := errors.New("unexpected error in step 2")

	// Get token
	stepGetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetTokenWithUser(targetToken).
			Return(expectedTenantMemberEntity, nil).
			Times(1)
	}

	stepGetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetTokenWithUser(targetToken).
			Return(&auth.SuperAdminConfirmTokenEntity{}, errMock).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenOk,
			},
			expectedUser:  expectedUser,
			expectedError: nil,
		},

		// Fail
		{
			name:             "Fail: unexpected error",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenError,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ConfirmAccountTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewConfirmAccountTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			userFound, err := adapter.GetSuperAdminByConfirmAccountToken(tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, userFound) {
				t.Errorf("want %#v, got user %#v", tc.expectedUser, userFound)
			}
		})
	}
}

func TestConfirmAccountTokenPgAdapter_GetTenantConfirmAccountToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTenantId    uuid.UUID
		inputTokenString string
		setupSteps       []mockSetupFunc_confirmAccountTokenPgAdapter
		expectedToken    auth.ConfirmAccountToken
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetToken := "token"

	targetUserId := uint(1)
	targetTenantId := uuid.New()
	targetTenantIdStr := targetTenantId.String()
	targetExpiry := time.Now()

	expectedTenantMemberToken := &auth.TenantConfirmTokenEntity{
		Token:     targetToken,
		TenantId:  targetTenantIdStr,
		UserId:    targetUserId,
		ExpiresAt: targetExpiry,
	}

	expectedDomainToken := auth.ConfirmAccountToken{
		Token:      targetToken,
		TenantId:   &targetTenantId,
		UserId:     targetUserId,
		ExpiryDate: targetExpiry,
	}

	// Step 2: Save token
	errMock := errors.New("unexpected error in step 2")

	// Get token
	stepGetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetToken(targetTenantId.String(), targetToken).
			Return(expectedTenantMemberToken, nil).
			Times(1)
	}

	stepGetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return tenantMemberRepo.EXPECT().
			GetToken(targetTenantId.String(), targetToken).
			Return(&auth.TenantConfirmTokenEntity{}, errMock).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenOk,
			},
			expectedToken: expectedDomainToken,
			expectedError: nil,
		},

		// Fail
		{
			name:             "Fail (step 2): unexpected error",
			inputTenantId:    targetTenantId,
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenError,
			},
			expectedToken: auth.ConfirmAccountToken{},
			expectedError: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ConfirmAccountTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewConfirmAccountTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			token, err := adapter.GetTenantConfirmAccountToken(tc.inputTenantId, tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedToken, token) {
				t.Errorf("want %#v, got token %#v", tc.expectedToken, token)
			}
		})
	}
}

func TestConfirmAccountTokenPgAdapter_GetSuperAdminConfirmAccountToken(t *testing.T) {
	type testCase struct {
		name             string
		inputTokenString string
		setupSteps       []mockSetupFunc_confirmAccountTokenPgAdapter
		expectedUser     auth.ConfirmAccountToken
		expectedError    error
	}

	// Input ----------------------------------------------------------------------------------
	targetExpiryDate := time.Now()
	targetToken := "token"

	targetUserId := uint(1)

	expectedTenantMemberEntity := &auth.SuperAdminConfirmTokenEntity{
		Token:     targetToken,
		UserId:    targetUserId,
		ExpiresAt: targetExpiryDate,
	}

	expectedDomainToken := auth.ConfirmAccountToken{
		Token:      targetToken,
		TenantId:   nil,
		UserId:     targetUserId,
		ExpiryDate: targetExpiryDate,
	}

	// Save token
	errMock := errors.New("unexpected error in step 2")

	// Get token
	stepGetTokenOk := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetToken(targetToken).
			Return(expectedTenantMemberEntity, nil).
			Times(1)
	}

	stepGetTokenError := func(
		hasher *cryptoMocks.MockSecretHasher, tokenGenerator *cryptoMocks.MockSecurityTokenGenerator, tenantMemberRepo *mocks.MockTenantConfirmTokenRepository, superAdminRepo *mocks.MockSuperAdminConfirmTokenRepository,
	) *gomock.Call {
		return superAdminRepo.EXPECT().
			GetToken(targetToken).
			Return(&auth.SuperAdminConfirmTokenEntity{}, errMock).
			Times(1)
	}

	cases := []testCase{
		// Success
		{
			name:             "Success",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenOk,
			},
			expectedUser:  expectedDomainToken,
			expectedError: nil,
		},

		// Step 1
		{
			name:             "Fail: unexpected error",
			inputTokenString: targetToken,
			setupSteps: []mockSetupFunc_confirmAccountTokenPgAdapter{
				stepGetTokenError,
			},
			expectedUser:  auth.ConfirmAccountToken{},
			expectedError: errMock,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo := setupMockSteps_ConfirmAccountTokenPgAdapter(t, tc.setupSteps)

			adapter := auth.NewConfirmAccountTokenPgAdapter(mockHasher, mockTokenGenerator, mockTenantRepo, mockSuperAdminRepo)

			tokenObj, err := adapter.GetSuperAdminConfirmAccountToken(tc.inputTokenString)

			if !errors.Is(err, tc.expectedError) {
				t.Errorf("want error %v, got error %v", tc.expectedError, err)
			}

			if !reflect.DeepEqual(tc.expectedUser, tokenObj) {
				t.Errorf("want %#v, got user %#v", tc.expectedUser, tokenObj)
			}
		})
	}
}

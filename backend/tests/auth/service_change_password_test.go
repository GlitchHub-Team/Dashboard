package auth

import (
	"errors"
	"testing"
	"time"

	"backend/internal/auth"
	"backend/internal/shared/identity"
	"backend/internal/user"
	"backend/tests/auth/mocks"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"

	cryptoMocks "backend/tests/shared/crypto/mocks"
	userMocks "backend/tests/user/mocks"
)

type mockSetupFunc_ChangePasswordService func(
	mockLogger *zap.Logger,
	mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator,
	mockSecretHasher *cryptoMocks.MockSecretHasher,

	mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort,
	mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort,
	mockGetUserPort *userMocks.MockGetUserPort,
	mockSaveUserPort *userMocks.MockSaveUserPort,
) *gomock.Call

func TestChangePasswordService_VerifyForgotPasswordToken(t *testing.T) {
	// type mockSetupFunc func(
	// 	mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator,
	// 	mockSecretHasher *cryptoMocks.MockSecretHasher,

	// 	mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort,
	// 	mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort,
	// 	mockGetUserPort *userMocks.MockGetUserPort,
	// 	mockSaveUserPort *userMocks.MockSaveUserPort,
	// ) *gomock.Call

	type testCase struct {
		name          string
		input         auth.VerifyForgotPasswordTokenCommand
		setupSteps    []mockSetupFunc_ChangePasswordService
		expectedError error
	}

	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(100)
	targetCorrectToken := "token123"
	expectedTokenHash := "hash"
	targetExpiryDate := time.Now().Add(time.Hour * 4)

	expectedTokenObj := auth.ForgotPasswordToken{
		Token: expectedTokenHash,
		TenantId:    &targetTenantId,
		ExpiryDate:  targetExpiryDate,
		UserId:      targetUserId,
	}

	expectedExpiredTokenObj := auth.ForgotPasswordToken{
		Token: expectedTokenHash,
		TenantId:    &targetTenantId,
		ExpiryDate:  time.Now().Add(time.Hour * -4),
		UserId:      targetUserId,
	}

	// Step 1: get token -------------------------------------------------------------------------------------

	// - Tenant Member
	stepGetTokenOk_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(expectedTokenObj, nil).
			Times(1)
	}

	stepGetTokenExpired_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(expectedExpiredTokenObj, auth.ErrTokenExpired).
			Times(1)
	}

	errMockGetToken := errors.New("unexpected database error")
	stepGetTokenError_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(auth.ForgotPasswordToken{
				ExpiryDate: time.Now().Add(1 * time.Hour),
			}, errMockGetToken).
			Times(1)
	}

	// - Super Admin
	stepGetTokenOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminForgotPasswordToken(targetCorrectToken).
			Return(expectedTokenObj, nil).
			Times(1)
	}

	stepGetTokenExpired_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminForgotPasswordToken(targetCorrectToken).
			Return(expectedExpiredTokenObj, auth.ErrTokenExpired).
			Times(1)
	}

	stepGetTokenError_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminForgotPasswordToken(targetCorrectToken).
			Return(auth.ForgotPasswordToken{
				ExpiryDate: time.Now().Add(1 * time.Hour),
			}, errMockGetToken).
			Times(1)
	}

	// --- Inputs ---
	baseInput_TenantMember := auth.VerifyForgotPasswordTokenCommand{
		TenantId: &targetTenantId,
		Token:    targetCorrectToken,
	}

	baseInput_SuperAdmin := auth.VerifyForgotPasswordTokenCommand{
		TenantId: nil,
		Token:    targetCorrectToken,
	}

	cases := []testCase{
		// TENANT MEMBER ------------------------------------------------------------------------------
		{
			name:          "(Tenant Member) Success",
			input:         baseInput_TenantMember,
			setupSteps:    []mockSetupFunc_ChangePasswordService{stepGetTokenOk_TenantMember},
			expectedError: nil,
		},
		{
			name:          "(Tenant Member) Fail: token expired",
			input:         baseInput_TenantMember,
			setupSteps:    []mockSetupFunc_ChangePasswordService{stepGetTokenExpired_TenantMember},
			expectedError: auth.ErrTokenExpired,
		},
		{
			name:          "(Tenant Member) Fail: unexpected error",
			input:         baseInput_TenantMember,
			setupSteps:    []mockSetupFunc_ChangePasswordService{stepGetTokenError_TenantMember},
			expectedError: errMockGetToken,
		},

		// SUPER ADMIN ---------------------------------------------------------------------------------
		{
			name:          "(Super Admin) Success",
			input:         baseInput_SuperAdmin,
			setupSteps:    []mockSetupFunc_ChangePasswordService{stepGetTokenOk_SuperAdmin},
			expectedError: nil,
		},
		{
			name:          "(Super Admin) Fail: token expired",
			input:         baseInput_SuperAdmin,
			setupSteps:    []mockSetupFunc_ChangePasswordService{stepGetTokenExpired_SuperAdmin},
			expectedError: auth.ErrTokenExpired,
		},
		{
			name:          "(Super Admin) Fail: unexpected error",
			input:         baseInput_SuperAdmin,
			setupSteps:    []mockSetupFunc_ChangePasswordService{stepGetTokenError_SuperAdmin},
			expectedError: errMockGetToken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockLogger := zaptest.NewLogger(t)
			mockTokenGen := cryptoMocks.NewMockSecurityTokenGenerator(mockController)
			mockHasher := cryptoMocks.NewMockSecretHasher(mockController)

			mockTokenPort := mocks.NewMockForgotPasswordTokenPort(mockController)
			mockEmailPort := mocks.NewMockSendForgotPasswordEmailPort(mockController)
			mockGetUserPort := userMocks.NewMockGetUserPort(mockController)
			mockSaveUserPort := userMocks.NewMockSaveUserPort(mockController)

			var expectedCalls []any

			for _, step := range tc.setupSteps {
				call := step(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			service := auth.NewChangePasswordService(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)

			err := service.VerifyForgotPasswordToken(tc.input)

			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestChangePasswordService_RequestForgotPassword(t *testing.T) {
	type testCase struct {
		name          string
		input         auth.RequestForgotPasswordCommand
		setupSteps    []mockSetupFunc_ChangePasswordService
		expectedError error
	}

	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(100)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := true
	targetPassword := "pw123"

	expectedUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &targetPassword,
		Role:         identity.ROLE_TENANT_USER, // Irrilevante
		TenantId:     &targetTenantId,
		Confirmed:    targetConfirmed,
	}

	expectedUnconfirmedUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &targetPassword,
		Role:         identity.ROLE_TENANT_USER, // Irrilevante
		TenantId:     &targetTenantId,
		Confirmed:    false,
	}

	expectedToken := "1234"

	// Step 1: get user -------------------------------------------------------------------------------------

	step1GetUserOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUserByEmail(&targetTenantId, targetUserEmail).
			Return(expectedUser, nil).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error in step 1")
	step1UserError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUserByEmail(&targetTenantId, targetUserEmail).
			Return(user.User{}, errMockStep1).
			Times(1)
	}

	step1UserNotConfirmed := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUserByEmail(&targetTenantId, targetUserEmail).
			Return(expectedUnconfirmedUser, nil).
			Times(1)
	}

	step1UserNotFound := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUserByEmail(&targetTenantId, targetUserEmail).
			Return(user.User{}, nil).
			Times(1)
	}

	// Step 2: crea token
	step2CreateTokenOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			NewForgotPasswordToken(expectedUser).
			Return(expectedToken, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error step 3")
	step2CreateTokenError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			NewForgotPasswordToken(expectedUser).
			Return("", errMockStep2).
			Times(1)
	}

	// Step 3: Invia mail
	step3SendMailOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSendEmailPort.EXPECT().
			SendForgotPasswordEmail(targetUserEmail, &targetTenantId, expectedToken).
			Return(nil).
			Times(1)
	}

	errMockStep3 := errors.New("unexpected error step 3")
	step3SendMailError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSendEmailPort.EXPECT().
			SendForgotPasswordEmail(targetUserEmail, &targetTenantId, expectedToken).
			Return(errMockStep3).
			Times(1)
	}

	// --- Inputs ---
	baseInput := auth.RequestForgotPasswordCommand{
		TenantId: &targetTenantId,
		Email:    targetUserEmail,
	}

	cases := []testCase{
		// Successo
		{
			name:  "Success",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserOk,
				step2CreateTokenOk,
				step3SendMailOk,
			},
			expectedError: nil,
		},

		// Step 1: get user
		{
			name:  "Fail (step 1): zero user",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1UserNotFound,
			},
			expectedError: user.ErrUserNotFound,
		},
		{
			name:  "Fail (step 1): user not confirmed",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1UserNotConfirmed,
			},
			expectedError: auth.ErrAccountNotConfirmed,
		},
		{
			name:  "Fail (step 1): error",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1UserError,
			},
			expectedError: errMockStep1,
		},

		// Step 2: crea token
		{
			name:  "Fail (step 2): error",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserOk,
				step2CreateTokenError,
			},
			expectedError: errMockStep2,
		},

		// Step 3: invia mail
		{
			name:  "Fail (step 3): error",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserOk,
				step2CreateTokenOk,
				step3SendMailError,
			},
			expectedError: errMockStep3,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockLogger := zaptest.NewLogger(t)
			mockTokenGen := cryptoMocks.NewMockSecurityTokenGenerator(mockController)
			mockHasher := cryptoMocks.NewMockSecretHasher(mockController)

			mockTokenPort := mocks.NewMockForgotPasswordTokenPort(mockController)
			mockEmailPort := mocks.NewMockSendForgotPasswordEmailPort(mockController)
			mockGetUserPort := userMocks.NewMockGetUserPort(mockController)
			mockSaveUserPort := userMocks.NewMockSaveUserPort(mockController)

			var expectedCalls []any

			for _, step := range tc.setupSteps {
				call := step(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			service := auth.NewChangePasswordService(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)

			err := service.RequestForgotPassword(tc.input)

			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestChangePasswordService_ConfirmForgotPassword(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(100)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := true
	targetExpiryDate := time.Now().Add(time.Hour * 12)

	targetNewPassword := "password"
	expectedNewPasswordHash := "hashed_password123123"

	// Tenant User non confermato all'inizio
	// NOTA: il ruolo non conta
	expectedTenantUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: nil,
		Role:         identity.ROLE_TENANT_USER, // NOTA: potrebbe anche essere Tenant Admin
		TenantId:     &targetTenantId,
		Confirmed:    targetConfirmed,
	}

	expectedUnconfirmedTenantUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: nil,
		Role:         identity.ROLE_TENANT_USER, // NOTA: potrebbe anche essere Tenant Admin
		TenantId:     &targetTenantId,
		Confirmed:    false,
	}


	// Tenant User confermato dopo SaveUser() (step 4)
	targetConfirmedTenantUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &expectedNewPasswordHash,
		Role:         identity.ROLE_TENANT_USER, // NOTA: potrebbe anche essere Tenant Admin
		TenantId:     &targetTenantId,
		Confirmed:    true,
	}

	// Super Admin non confermato all'inizio
	expectedSuperAdmin := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: nil,
		Role:         identity.ROLE_SUPER_ADMIN,
		TenantId:     nil,
		Confirmed:    targetConfirmed,
	}

	expectedUnconfirmedSuperAdmin := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: nil,
		Role:         identity.ROLE_SUPER_ADMIN,
		TenantId:     nil,
		Confirmed:    false,
	}

	// Super Admin confermato dopo SaveUser() (step 4)
	targetConfirmedSuperAdmin := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &expectedNewPasswordHash,
		Role:         identity.ROLE_SUPER_ADMIN,
		TenantId:     nil,
		Confirmed:    true,
	}

	// passwordHash := "hash-123"

	targetCorrectToken := "token123"
	expectedTokenHash := "hash"
	// targetWrongToken := "token456"

	expectedTokenObj := auth.ForgotPasswordToken{
		Token: expectedTokenHash,
		TenantId:    &targetTenantId,
		ExpiryDate:  targetExpiryDate,
		UserId:      targetUserId,
	}

	expectedExpiredTokenObj := auth.ForgotPasswordToken{
		Token: expectedTokenHash,
		TenantId:    &targetTenantId,
		ExpiryDate:  time.Now().Add(time.Hour * -4),
		UserId:      targetUserId,
	}

	// test case
	type testCase struct {
		name          string
		input         auth.ConfirmForgotPasswordCommand
		setupSteps    []mockSetupFunc_ChangePasswordService
		expectedUser  user.User
		expectedError error
	}

	// Step 1: get token -------------------------------------------------------------------------------------

	// - Tenant Member
	step1GetTokenOk_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(expectedTokenObj, nil).
			Times(1)
	}

	step1GetTokenExpired_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(expectedExpiredTokenObj, auth.ErrTokenExpired).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error 1")
	step1GetTokenError_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(auth.ForgotPasswordToken{}, errMockStep1).
			Times(1)
	}

	// - Super Admin
	step1GetTokenOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminForgotPasswordToken(targetCorrectToken).
			Return(expectedTokenObj, nil).
			Times(1)
	}

	step1GetTokenExpired_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminForgotPasswordToken(targetCorrectToken).
			Return(expectedExpiredTokenObj, auth.ErrTokenExpired).
			Times(1)
	}

	step1GetTokenError_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminForgotPasswordToken(targetCorrectToken).
			Return(auth.ForgotPasswordToken{}, errMockStep1).
			Times(1)
	}

	// Step 2: get user -------------------------------------------------------------------------------------

	// - Tenant Member
	step2GetUserOk_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantMemberByForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(expectedTenantUser, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error 2")
	step2GetUserError_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantMemberByForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2GetUserUnconfirmed_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetTenantMemberByForgotPasswordToken(targetTenantId, targetCorrectToken).
			Return(expectedUnconfirmedTenantUser, nil).
			Times(1)
	}


	// - Super Admin
	step2GetUserOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminByForgotPasswordToken(targetCorrectToken).
			Return(expectedSuperAdmin, nil).
			Times(1)
	}

	step2GetUserError_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminByForgotPasswordToken(targetCorrectToken).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2GetUserUnconfirmed_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			GetSuperAdminByForgotPasswordToken(targetCorrectToken).
			Return(expectedUnconfirmedSuperAdmin, nil).
			Times(1)
	}

	// Step 3: crea hash ------------------------------------------------------------
	errMockStep3 := errors.New("unexpected error in step 3")
	step3CreateHashOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			HashSecret(targetNewPassword).
			Return(expectedNewPasswordHash, nil).
			Times(1)
	}

	step3CreateHashError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			HashSecret(targetNewPassword).
			Return("", errMockStep3).
			Times(1)
	}

	// Step 4: imposta campi utente ------------------------------------------------------------
	step4SaveUserOk_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedTenantUser).
			Return(targetConfirmedTenantUser, nil).
			Times(1)
	}

	errMockStep4 := errors.New("unexpected error in step 4")
	step4SaveUserError_TenantMember := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedTenantUser).
			Return(user.User{}, errMockStep4).
			Times(1)
	}

	step4SaveUserOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedSuperAdmin).
			Return(targetConfirmedSuperAdmin, nil).
			Times(1)
	}

	step4SaveUserError_SuperAdmin := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedSuperAdmin).
			Return(user.User{}, errMockStep4).
			Times(1)
	}

	// step4SaveUserNeverCalled := func(
	// 	mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	// ) *gomock.Call {
	// 	return mockSaveUserPort.EXPECT().
	// 		SaveUser(gomock.Any()).
	// 		Times(0)
	// }

	// Step 5: elimina token -----------------------------------------------------------------------
	step5DeleteTokenOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			DeleteForgotPasswordToken(expectedTokenObj).
			Return(nil).
			Times(1)
	}

	err5MockStep := errors.New("unexpected error in step 5")

	step5DeleteTokenError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockForgotPasswordPort.EXPECT().
			DeleteForgotPasswordToken(expectedTokenObj).
			Return(err5MockStep).
			Times(1)
	}

	baseInput_TenantMember := auth.ConfirmForgotPasswordCommand{
		TenantId:    &targetTenantId,
		Token:       targetCorrectToken,
		NewPassword: targetNewPassword,
	}

	baseInput_SuperAdmin := auth.ConfirmForgotPasswordCommand{
		TenantId:    nil,
		Token:       targetCorrectToken,
		NewPassword: targetNewPassword,
	}

	cases := []testCase{
		// TENANT MEMBER ------------------------------------------------------------------------------
		// Successo
		{
			name:  "(Tenant Member) Success",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_TenantMember,
				step2GetUserOk_TenantMember,
				step3CreateHashOk,
				step4SaveUserOk_TenantMember,
				step5DeleteTokenOk,
			},
			expectedUser:  targetConfirmedTenantUser,
			expectedError: nil,
		},

		{
			name:  "(Tenant Member) Success -> cannot delete token",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_TenantMember,
				step2GetUserOk_TenantMember,
				step3CreateHashOk,
				step4SaveUserOk_TenantMember,
				step5DeleteTokenError,
			},
			expectedUser:  targetConfirmedTenantUser,
			expectedError: nil,
		},

		// Step 1
		{
			name:  "(Tenant Member) Fail (step 1): token expired",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenExpired_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenExpired,
		},

		{
			name:  "(Tenant Member) Fail (step 1): unexpected error",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenError_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:  "(Tenant Member) Fail (step 2): unconfirmed user",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_TenantMember,
				step2GetUserUnconfirmed_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrAccountNotConfirmed,
		},
		{
			name:  "(Tenant Member) Fail (step 2): unexpected error",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_TenantMember,
				step2GetUserError_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenNotFound,
		},

		// Step 3
		{
			name:  "(Tenant Member) Fail (step 3): unexpected error",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_TenantMember,
				step2GetUserOk_TenantMember,
				step3CreateHashError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep3,
		},

		// Step 4
		{
			name:  "(Tenant Member) Fail (step 4): unexpected error",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_TenantMember,
				step2GetUserOk_TenantMember,
				step3CreateHashOk,
				step4SaveUserError_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep4,
		},

		// SUPER ADMIN ---------------------------------------------------------------------------------
		// Successo
		{
			name:  "(Super Admin) Success",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_SuperAdmin,
				step2GetUserOk_SuperAdmin,
				step3CreateHashOk,
				step4SaveUserOk_SuperAdmin,
				step5DeleteTokenOk,
			},
			expectedUser:  targetConfirmedSuperAdmin,
			expectedError: nil,
		},

		{
			name:  "(Super Admin) Success -> cannot delete token",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_SuperAdmin,
				step2GetUserOk_SuperAdmin,
				step3CreateHashOk,
				step4SaveUserOk_SuperAdmin,
				step5DeleteTokenError,
			},
			expectedUser:  targetConfirmedSuperAdmin,
			expectedError: nil,
		},
		// Step 1
		{
			name:  "(Super Admin) Fail (step 1): token expired",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenExpired_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenExpired,
		},

		{
			name:  "(Super Admin) Fail (step 1): unexpected error",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenError_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:  "(Super Admin) Fail (step 2): unconfirmed user",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_SuperAdmin,
				step2GetUserUnconfirmed_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrAccountNotConfirmed,
		},
		{
			name:  "(Super Admin) Fail (step 2): unexpected error",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_SuperAdmin,
				step2GetUserError_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenNotFound,
		},

		// Step 3
		{
			name:  "(Super Admin) Fail (step 3): unexpected error",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_SuperAdmin,
				step2GetUserOk_SuperAdmin,
				step3CreateHashError,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep3,
		},

		// Step 4
		{
			name:  "(Super Admin) Fail (step 4): unexpected error",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetTokenOk_SuperAdmin,
				step2GetUserOk_SuperAdmin,
				step3CreateHashOk,
				step4SaveUserError_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockLogger := zaptest.NewLogger(t)
			mockTokenGen := cryptoMocks.NewMockSecurityTokenGenerator(mockController)
			mockHasher := cryptoMocks.NewMockSecretHasher(mockController)

			mockTokenPort := mocks.NewMockForgotPasswordTokenPort(mockController)
			mockEmailPort := mocks.NewMockSendForgotPasswordEmailPort(mockController)
			mockGetUserPort := userMocks.NewMockGetUserPort(mockController)
			mockSaveUserPort := userMocks.NewMockSaveUserPort(mockController)

			var expectedCalls []any

			for _, step := range tc.setupSteps {
				call := step(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			service := auth.NewChangePasswordService(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)

			// Esegui funzione in oggetto
			err := service.ConfirmForgotPassword(tc.input)

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

func TestChangePasswordService_ChangePassword(t *testing.T) {
	targetTenantId := uuid.New()
	targetUserId := uint(100)
	targetOldPassword := "old_password"
	targetNewPassword := "new_password"
	targetOldHash := "old_hash_123"
	targetNewHash := "new_hash_456"

	// Base user for successful retrieval
	expectedConfirmedUser := user.User{
		Id:           targetUserId,
		TenantId:     &targetTenantId,
		Confirmed:    true,
		PasswordHash: &targetOldHash,
	}

	// User state failing the account confirmation check
	expectedUnconfirmedUser := user.User{
		Id:           targetUserId,
		TenantId:     &targetTenantId,
		Confirmed:    false,
		PasswordHash: &targetOldHash,
	}

	// Expected state after domain entity update (SetPasswordHash)
	expectedUserAfterChange := user.User{
		Id:           targetUserId,
		TenantId:     &targetTenantId,
		Confirmed:    true,
		PasswordHash: &targetNewHash,
	}

	type testCase struct {
		name          string
		input         auth.ChangePasswordCommand
		setupSteps    []mockSetupFunc_ChangePasswordService
		expectedError error
	}

	// Step 1: Get user --------------------------------------------------------------------------------------
	step1GetUserOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(&targetTenantId, targetUserId).
			Return(expectedConfirmedUser, nil).
			Times(1)
	}

	step1GetUserUnconfirmed := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(&targetTenantId, targetUserId).
			Return(expectedUnconfirmedUser, nil).
			Times(1)
	}

	errMockStep1 := errors.New("database error getting user")
	step1GetUserError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(&targetTenantId, targetUserId).
			Return(user.User{}, errMockStep1).
			Times(1)
	}

	// Step 3: Check old password ----------------------------------------------------------------------------
	step3CompareHashOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			CompareHashAndSecret(targetOldHash, targetOldPassword).
			Return(nil).
			Times(1)
	}

	step3CompareHashError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			CompareHashAndSecret(targetOldHash, targetOldPassword).
			Return(errors.New("hash mismatch")).
			Times(1)
	}

	step3NeverCalled := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			CompareHashAndSecret(gomock.Any(), gomock.Any()).
			Times(0)
	}

	// Step 4: Generate new hash -----------------------------------------------------------------------------
	step4HashSecretOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			HashSecret(targetNewPassword).
			Return(targetNewHash, nil).
			Times(1)
	}

	errMockStep4 := errors.New("error hashing new password")
	step4HashSecretError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			HashSecret(targetNewPassword).
			Return("", errMockStep4).
			Times(1)
	}

	// Step 6: Save user -------------------------------------------------------------------------------------
	step6SaveUserOk := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(expectedUserAfterChange).
			Return(expectedUserAfterChange, nil).
			Times(1)
	}

	errMockStep6 := errors.New("database error saving user")
	step6SaveUserError := func(
		mockLogger *zap.Logger, mockTokenGenerator *cryptoMocks.MockSecurityTokenGenerator, mockSecretHasher *cryptoMocks.MockSecretHasher, mockForgotPasswordPort *mocks.MockForgotPasswordTokenPort, mockSendEmailPort *mocks.MockSendForgotPasswordEmailPort, mockGetUserPort *userMocks.MockGetUserPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(expectedUserAfterChange).
			Return(user.User{}, errMockStep6).
			Times(1)
	}

	// Il ruolo dell'utente non importa
	baseInput := auth.ChangePasswordCommand{
		// RequesterTenantId: &targetTenantId,
		// RequesterUserId:   targetUserId,
		Requester: identity.Requester{
			RequesterUserId:   targetUserId,
			RequesterTenantId: &targetTenantId,
			RequesterRole:     identity.ROLE_TENANT_USER,
		},
		OldPassword: targetOldPassword,
		NewPassword: targetNewPassword,
	}

	cases := []testCase{
		{
			name:  "Success",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserOk,
				step3CompareHashOk,
				step4HashSecretOk,
				step6SaveUserOk,
			},
			expectedError: nil,
		},
		{
			name:  "Fail (Step 1): Unexpected error getting user",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserError,
			},
			expectedError: errMockStep1,
		},
		{
			name:  "Fail (Step 2): Account not confirmed",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserUnconfirmed,
				step3NeverCalled,
			},
			expectedError: auth.ErrAccountNotConfirmed,
		},
		{
			name:  "Fail (Step 3): Wrong old credentials",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserOk,
				step3CompareHashError,
			},
			expectedError: auth.ErrWrongCredentials,
		},
		{
			name:  "Fail (Step 4): Error generating new hash",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserOk,
				step3CompareHashOk,
				step4HashSecretError,
			},
			expectedError: errMockStep4,
		},
		{
			name:  "Fail (Step 6): Unexpected error saving user",
			input: baseInput,
			setupSteps: []mockSetupFunc_ChangePasswordService{
				step1GetUserOk,
				step3CompareHashOk,
				step4HashSecretOk,
				step6SaveUserError,
			},
			expectedError: errMockStep6,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockLogger := zaptest.NewLogger(t)
			mockTokenGen := cryptoMocks.NewMockSecurityTokenGenerator(mockController)
			mockHasher := cryptoMocks.NewMockSecretHasher(mockController)

			mockTokenPort := mocks.NewMockForgotPasswordTokenPort(mockController)
			mockEmailPort := mocks.NewMockSendForgotPasswordEmailPort(mockController)
			mockGetUserPort := userMocks.NewMockGetUserPort(mockController)
			mockSaveUserPort := userMocks.NewMockSaveUserPort(mockController)

			var expectedCalls []any

			for _, step := range tc.setupSteps {
				call := step(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			service := auth.NewChangePasswordService(mockLogger, mockTokenGen, mockHasher, mockTokenPort, mockEmailPort, mockGetUserPort, mockSaveUserPort)

			// Esegui funzione in oggetto
			err := service.ChangePassword(tc.input)

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

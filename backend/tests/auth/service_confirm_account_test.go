package auth_test

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

func TestConfirmAccountService_ConfirmAccount(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(100)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := false
	targetExpiryDate := time.Now().Add(time.Hour * 4)

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

	passwordHash := "hash-123"

	// Tenant Member già confermato prima di iniziare procedura
	expectedAlreadyConfirmedTenantMember := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &passwordHash,
		Role:         identity.ROLE_TENANT_USER, // NOTA: potrebbe anche essere Tenant Admin
		TenantId:     &targetTenantId,
		Confirmed:    true,
	}

	// Super Admin già confermato prima di iniziare procedura
	expectedAlreadyConfirmedSuperAdmin := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: &passwordHash,
		Role:         identity.ROLE_SUPER_ADMIN, // NOTA: potrebbe anche essere Tenant Admin
		TenantId:     nil,
		Confirmed:    true,
	}

	targetCorrectToken := "token123"
	expectedTokenHash := "hash"
	// targetWrongToken := "token456"

	expectedTokenObj_TenantAdmin := auth.ConfirmAccountToken{
		Token:      expectedTokenHash,
		TenantId:   &targetTenantId,
		ExpiryDate: targetExpiryDate,
		UserId:     targetUserId,
	}

	expectedTokenObj_SuperAdmin := auth.ConfirmAccountToken{
		Token:      expectedTokenHash,
		TenantId:   nil,
		ExpiryDate: targetExpiryDate,
		UserId:     targetUserId,
	}

	type mockSetupFunc func(
		mockLogger *zap.Logger,
		mockSecretHasher *cryptoMocks.MockSecretHasher,
		mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort,
		mockSaveUserPort *userMocks.MockSaveUserPort,
		mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call

	// test case
	type testCase struct {
		name          string
		input         auth.ConfirmAccountCommand
		setupSteps    []mockSetupFunc
		expectedUser  user.User
		expectedError error
	}

	// Step 1: get token -------------------------------------------------------------------------------------

	// - Tenant Member
	step1GetTokenOk_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetTenantConfirmAccountToken(targetTenantId, targetCorrectToken).
			Return(expectedTokenObj_TenantAdmin, nil).
			Times(1)
	}

	step1GetTokenExpired_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetTenantConfirmAccountToken(targetTenantId, targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, auth.ErrTokenExpired).
			Times(1)
	}

	errMockStep1 := errors.New("unexpected error 1")
	step1GetTokenError_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetTenantConfirmAccountToken(targetTenantId, targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, errMockStep1).
			Times(1)
	}

	// - Super Admin
	step1GetTokenOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetSuperAdminConfirmAccountToken(targetCorrectToken).
			Return(expectedTokenObj_SuperAdmin, nil).
			Times(1)
	}

	step1GetTokenExpired_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetSuperAdminConfirmAccountToken(targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, auth.ErrTokenExpired).
			Times(1)
	}

	step1GetTokenError_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetSuperAdminConfirmAccountToken(targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, errMockStep1).
			Times(1)
	}

	// Step 2: get user -------------------------------------------------------------------------------------

	// - Tenant Member
	step2GetUserOk_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(&targetTenantId, targetUserId).
			Return(expectedTenantUser, nil).
			Times(1)
	}

	errMockStep2 := errors.New("unexpected error 2")
	step2GetUserError_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(&targetTenantId, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2GetConfirmedUser_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(&targetTenantId, targetUserId).
			Return(expectedAlreadyConfirmedTenantMember, errMockStep2).
			Times(1)
	}

	// - Super Admin
	step2GetUserOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(nil, targetUserId).
			Return(expectedSuperAdmin, nil).
			Times(1)
	}

	step2GetUserError_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(nil, targetUserId).
			Return(user.User{}, errMockStep2).
			Times(1)
	}

	step2GetConfirmedUser_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockGetUserPort.EXPECT().
			GetUser(nil, targetUserId).
			Return(expectedAlreadyConfirmedSuperAdmin, errMockStep2).
			Times(1)
	}

	// Step 3: crea hash ------------------------------------------------------------
	errMockStep3 := errors.New("unexpected error in step 3")
	step3CreateHashOk := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			HashSecret(targetNewPassword).
			Return(expectedNewPasswordHash, nil).
			Times(1)
	}

	step3CreateHashError := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockSecretHasher.EXPECT().
			HashSecret(targetNewPassword).
			Return("", errMockStep3).
			Times(1)
	}

	// Step 4: imposta campi utente ------------------------------------------------------------
	step4SaveUserOk_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedTenantUser).
			Return(targetConfirmedTenantUser, nil).
			Times(1)
	}

	errMockStep4 := errors.New("unexpected error in step 4")
	step4SaveUserError_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedTenantUser).
			Return(user.User{}, errMockStep4).
			Times(1)
	}

	step4SaveUserOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedSuperAdmin).
			Return(targetConfirmedSuperAdmin, nil).
			Times(1)
	}

	step4SaveUserError_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockSaveUserPort.EXPECT().
			SaveUser(targetConfirmedSuperAdmin).
			Return(user.User{}, errMockStep4).
			Times(1)
	}

	// Step 5: elimina token -----------------------------------------------------------------------
	step5DeleteTokenOk := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			DeleteConfirmAccountToken(gomock.AssignableToTypeOf(expectedTokenObj_TenantAdmin)).
			Return(nil).
			Times(1)
	}

	err5MockStep := errors.New("unexpected error in step 5")
	step5DeleteTokenError := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			DeleteConfirmAccountToken(gomock.AssignableToTypeOf(expectedTokenObj_TenantAdmin)).
			Return(err5MockStep).
			Times(1)
	}

	baseInput_TenantMember := auth.ConfirmAccountCommand{
		TenantId:    &targetTenantId,
		Token:       targetCorrectToken,
		NewPassword: targetNewPassword,
	}

	baseInput_SuperAdmin := auth.ConfirmAccountCommand{
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
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
				step1GetTokenExpired_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenExpired,
		},

		{
			name:  "(Tenant Member) Fail (step 1): unexpected error",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc{
				step1GetTokenError_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:  "(Tenant Member) Fail (step 2): user confirmed",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc{
				step1GetTokenOk_TenantMember,
				step2GetConfirmedUser_TenantMember,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrAccountAlreadyConfirmed,
		},
		{
			name:  "(Tenant Member) Fail (step 2): unexpected error",
			input: baseInput_TenantMember,
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
				step1GetTokenExpired_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrTokenExpired,
		},

		{
			name:  "(Super Admin) Fail (step 1): unexpected error",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc{
				step1GetTokenError_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: errMockStep1,
		},

		// Step 2
		{
			name:  "(Super Admin) Fail (step 2): user confirmed",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc{
				step1GetTokenOk_SuperAdmin,
				step2GetConfirmedUser_SuperAdmin,
			},
			expectedUser:  user.User{},
			expectedError: auth.ErrAccountAlreadyConfirmed,
		},
		{
			name:  "(Super Admin) Fail (step 2): unexpected error",
			input: baseInput_SuperAdmin,
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
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
			setupSteps: []mockSetupFunc{
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
			mockHasher := cryptoMocks.NewMockSecretHasher(mockController)
			mockTokenPort := mocks.NewMockConfirmAccountTokenPort(mockController)
			mockSaveUserPort := userMocks.NewMockSaveUserPort(mockController)
			mockGetUserPort := userMocks.NewMockGetUserPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockLogger, mockHasher, mockTokenPort, mockSaveUserPort, mockGetUserPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			service := auth.NewConfirmUserAccountService(mockLogger, mockHasher, mockTokenPort, mockSaveUserPort, mockGetUserPort)

			// Esegui funzione in oggetto
			loggedUser, err := service.ConfirmAccount(tc.input)

			// Assertions
			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
			if loggedUser != tc.expectedUser {
				t.Errorf("expected user %v, got %v", tc.expectedUser, loggedUser)
			}
		})
	}
}

func TestConfirmAccountService_VerifyConfirmAccountToken(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(100)
	targetCorrectToken := "token123"
	expectedTokenHash := "hash"
	targetExpiryDate := time.Now().Add(time.Hour * 4)

	expectedTokenObj := auth.ConfirmAccountToken{
		Token:      expectedTokenHash,
		TenantId:   &targetTenantId,
		ExpiryDate: targetExpiryDate,
		UserId:     targetUserId,
	}

	type mockSetupFunc func(
		mockLogger *zap.Logger,
		mockSecretHasher *cryptoMocks.MockSecretHasher,
		mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort,
		mockSaveUserPort *userMocks.MockSaveUserPort,
		mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call

	type testCase struct {
		name          string
		input         auth.VerifyConfirmAccountTokenCommand
		setupSteps    []mockSetupFunc
		expectedError error
	}

	// Step 1: get token -------------------------------------------------------------------------------------

	// - Tenant Member
	stepGetTokenOk_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetTenantConfirmAccountToken(targetTenantId, targetCorrectToken).
			Return(expectedTokenObj, nil).
			Times(1)
	}

	stepGetTokenExpired_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetTenantConfirmAccountToken(targetTenantId, targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, auth.ErrTokenExpired).
			Times(1)
	}

	errMockGetToken := errors.New("unexpected database error")
	stepGetTokenError_TenantMember := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetTenantConfirmAccountToken(targetTenantId, targetCorrectToken).
			Return(auth.ConfirmAccountToken{
				ExpiryDate: time.Now().Add(1 * time.Hour),
			}, errMockGetToken).
			Times(1)
	}

	// - Super Admin
	stepGetTokenOk_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetSuperAdminConfirmAccountToken(targetCorrectToken).
			Return(expectedTokenObj, nil).
			Times(1)
	}

	stepGetTokenExpired_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetSuperAdminConfirmAccountToken(targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, auth.ErrTokenExpired).
			Times(1)
	}

	stepGetTokenError_SuperAdmin := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort, mockGetUserPort *userMocks.MockGetUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetSuperAdminConfirmAccountToken(targetCorrectToken).
			Return(auth.ConfirmAccountToken{
				ExpiryDate: time.Now().Add(1 * time.Hour),
			}, errMockGetToken).
			Times(1)
	}

	// --- Inputs ---
	baseInput_TenantMember := auth.VerifyConfirmAccountTokenCommand{
		TenantId: &targetTenantId,
		Token:    targetCorrectToken,
	}

	baseInput_SuperAdmin := auth.VerifyConfirmAccountTokenCommand{
		TenantId: nil,
		Token:    targetCorrectToken,
	}

	cases := []testCase{
		// TENANT MEMBER ------------------------------------------------------------------------------
		{
			name:          "(Tenant Member) Success",
			input:         baseInput_TenantMember,
			setupSteps:    []mockSetupFunc{stepGetTokenOk_TenantMember},
			expectedError: nil,
		},
		{
			name:          "(Tenant Member) Fail: token expired",
			input:         baseInput_TenantMember,
			setupSteps:    []mockSetupFunc{stepGetTokenExpired_TenantMember},
			expectedError: auth.ErrTokenExpired,
		},
		{
			name:          "(Tenant Member) Fail: unexpected error",
			input:         baseInput_TenantMember,
			setupSteps:    []mockSetupFunc{stepGetTokenError_TenantMember},
			expectedError: errMockGetToken,
		},

		// SUPER ADMIN ---------------------------------------------------------------------------------
		{
			name:          "(Super Admin) Success",
			input:         baseInput_SuperAdmin,
			setupSteps:    []mockSetupFunc{stepGetTokenOk_SuperAdmin},
			expectedError: nil,
		},
		{
			name:          "(Super Admin) Fail: token expired",
			input:         baseInput_SuperAdmin,
			setupSteps:    []mockSetupFunc{stepGetTokenExpired_SuperAdmin},
			expectedError: auth.ErrTokenExpired,
		},
		{
			name:          "(Super Admin) Fail: unexpected error",
			input:         baseInput_SuperAdmin,
			setupSteps:    []mockSetupFunc{stepGetTokenError_SuperAdmin},
			expectedError: errMockGetToken,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)

			mockLogger := zaptest.NewLogger(t)
			mockHasher := cryptoMocks.NewMockSecretHasher(mockController)
			mockTokenPort := mocks.NewMockConfirmAccountTokenPort(mockController)
			mockSaveUserPort := userMocks.NewMockSaveUserPort(mockController)
			mockGetUserPort := userMocks.NewMockGetUserPort(mockController)

			var expectedCalls []any

			for _, step := range tc.setupSteps {
				call := step(mockLogger, mockHasher, mockTokenPort, mockSaveUserPort, mockGetUserPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			service := auth.NewConfirmUserAccountService(mockLogger, mockHasher, mockTokenPort, mockSaveUserPort, mockGetUserPort)

			err := service.VerifyConfirmAccountToken(tc.input)

			if err != tc.expectedError {
				t.Errorf("expected error %v, got %v", tc.expectedError, err)
			}
		})
	}
}

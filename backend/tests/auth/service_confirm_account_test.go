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

	cryptoMocks "backend/internal/tests/shared/crypto/mocks"
	userMocks "backend/internal/tests/user/mocks"
)

func TestConfirmAccount(t *testing.T) {
	// Dati test
	targetTenantId := uuid.New()
	targetUserId := uint(100)
	targetUserEmail := "test@example.com"
	targetUserName := "Test"
	targetConfirmed := false
	targetExpiryDate := time.Now().Add(time.Hour * 4)

	// NOTA: il ruolo non conta
	expectedUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: nil,
		TenantId:     &targetTenantId,
		Confirmed:    targetConfirmed,
	}

	expectedUnconfirmedUser := user.User{
		Id:           targetUserId,
		Name:         targetUserName,
		Email:        targetUserEmail,
		PasswordHash: nil,
		TenantId:     &targetTenantId,
		Confirmed:    false,
	}

	// expectedTenantAdmin := user.User{
	// 	Id:           targetUserId,
	// 	Name:         targetUserName,
	// 	Email:        targetUserEmail,
	// 	PasswordHash: &expectedHash,
	// 	Role:         identity.ROLE_TENANT_ADMIN,
	// 	TenantId:     &targetTenantId,
	// 	Confirmed:    targetConfirmed,
	// }

	// expectedSuperAdmin := user.User{
	// 	Id:           targetUserId,
	// 	Name:         targetUserName,
	// 	Email:        targetUserEmail,
	// 	PasswordHash: &expectedHash,
	// 	Role:         identity.ROLE_SUPER_ADMIN,
	// 	TenantId:     &targetTenantId,
	// 	Confirmed:    targetConfirmed,
	// }

	// expectedUnconfirmedUser := user.User{
	// 	Id:           targetUserId,
	// 	Name:         targetUserName,
	// 	Email:        targetUserEmail,
	// 	PasswordHash: nil,
	// 	Role:         identity.ROLE_TENANT_USER, // Non importa ruolo
	// 	TenantId:     &targetTenantId,
	// 	Confirmed:    false,
	// }
		
	targetCorrectToken := "token123"
	expectedTokenHash := "hash"
	targetWrongToken := "token456"

	expectedTokenObj := auth.ConfirmAccountToken{
		HashedToken: expectedTokenHash,
		TenantId: &targetTenantId,
		ExpiryDate: targetExpiryDate,
		UserId: targetUserId, 
	}

	type mockSetupFunc func(
		mockLogger *zap.Logger,
		mockSecretHasher *cryptoMocks.MockSecretHasher,
		mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort,
		mockSaveUserPort *userMocks.MockSaveUserPort,
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

	step1GetTokenOk := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetConfirmAccountToken(targetCorrectToken).
			Return(expectedTokenObj, nil).
			Times(1)
	}

	step1GetTokenExpired := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetConfirmAccountToken(targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, auth.ErrTokenExpired).
			Times(1)
	}

	step1MockError := errors.New("unexpected error 1")
	step1GetTokenError := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetConfirmAccountToken(targetCorrectToken).
			Return(auth.ConfirmAccountToken{}, step1MockError).
			Times(1)
	}

	// Step 1: get user -------------------------------------------------------------------------------------

	step2GetUserOk := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetUserByConfirmAccountToken(targetCorrectToken).
			Return(expectedUser, nil).
			Times(1)
	}

	step2MockError := errors.New("unexpected error 2")
	step2GetUserError := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {
		return mockConfirmTokenPort.EXPECT().
			GetUserByConfirmAccountToken(targetWrongToken).
			Return(user.User{}, step2MockError).
			Times(1)
	}

	step2GetConfirmedUser := func(
		mockLogger *zap.Logger, mockSecretHasher *cryptoMocks.MockSecretHasher, mockConfirmTokenPort *mocks.MockConfirmAccountTokenPort, mockSaveUserPort *userMocks.MockSaveUserPort,
	) *gomock.Call {

	}

	/*
		step1GetUserOk_CaseSuperAdmin := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockGetUserPort.EXPECT().
				GetSuperAdminByEmail(targetUserEmail).
				Return(expectedSuperAdmin, nil).
				Times(1)
		}

		step1GetUserOk_CaseTenantAdmin := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockGetUserPort.EXPECT().
				GetTenantAdminByEmail(targetTenantId, targetUserEmail).
				Return(expectedTenantAdmin, nil).
				Times(1)
		}

		step1GetUserOk_CaseTenantUser := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockGetUserPort.EXPECT().
				GetTenantUserByEmail(targetTenantId, targetUserEmail).
				Return(expectedTenantUser, nil).
				Times(1)
		}

		step1GetUserErr_NoRole := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockGetUserPort.EXPECT().
				GetTenantUserByEmail(gomock.Any(), gomock.Any()).
				// Return(expectedRolelessUser, identity.ErrUnknownRole).
				Times(0)
		}

		step1GetUserErr_NotConfirmed := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockGetUserPort.EXPECT().
				GetTenantUserByEmail(targetTenantId, targetUserEmail).
				Return(expectedUnconfirmedUser, nil).
				Times(1)
		}

		step2Ok := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockSecretHasher.EXPECT().
				CompareHashAndSecret(expectedHash, targetCorrectPassword).
				Return(nil).
				Times(1)
		}

		errMockStep2 := errors.New("wrong password")
		step2Fail := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockSecretHasher.EXPECT().
				CompareHashAndSecret(expectedHash, targetWrongPassword).
				Return(errMockStep2).
				Times(1)
		}

		step2NeverCalled := func(
			mockSecretHasher *cryptoMocks.MockSecretHasher, mockGetUserPort *userMocks.MockGetUserPort,
		) *gomock.Call {
			return mockSecretHasher.EXPECT().
				CompareHashAndSecret(gomock.Any(), gomock.Any()).
				Times(0)
		}
	*/

	input_tenantUser := auth.LoginUserCommand{
		TenantId: &targetTenantId,
		Email:    targetUserEmail,
		Password: targetCorrectPassword,
		Role:     identity.ROLE_TENANT_USER,
	}

	input_tenantAdmin := auth.LoginUserCommand{
		TenantId: &targetTenantId,
		Email:    targetUserEmail,
		Password: targetCorrectPassword,
		Role:     identity.ROLE_TENANT_ADMIN,
	}

	input_superAdmin := auth.LoginUserCommand{
		TenantId: &targetTenantId,
		Email:    targetUserEmail,
		Password: targetCorrectPassword,
		Role:     identity.ROLE_SUPER_ADMIN,
	}

	input_noRole := auth.LoginUserCommand{
		TenantId: &targetTenantId,
		Email:    targetUserEmail,
		Password: targetCorrectPassword,
	}

	input_wrongPassword := auth.LoginUserCommand{
		TenantId: &targetTenantId,
		Email:    targetUserEmail,
		Password: targetWrongPassword,
		Role:     identity.ROLE_TENANT_USER,
	}

	cases := []testCase{
		/*

			{
				name:  "Success: Login Tenant User",
				input: input_tenantUser,
				setupSteps: []mockSetupFunc{
					step1GetUserOk_CaseTenantUser,
					step2Ok,
				},
				expectedUser:  expectedTenantUser,
				expectedError: nil,
			},
			{
				name:  "Success: Login Tenant Admin",
				input: input_tenantAdmin,
				setupSteps: []mockSetupFunc{
					step1GetUserOk_CaseTenantAdmin,
					step2Ok,
				},
				expectedUser:  expectedTenantAdmin,
				expectedError: nil,
			},
			{
				name:  "Success: Login Super Admin",
				input: input_superAdmin,
				setupSteps: []mockSetupFunc{
					step1GetUserOk_CaseSuperAdmin,
					step2Ok,
				},
				expectedUser:  expectedSuperAdmin,
				expectedError: nil,
			},

			// Step 1
			{
				name:  "Fail: Unknown role",
				input: input_noRole,
				setupSteps: []mockSetupFunc{
					step1GetUserErr_NoRole,
					step2NeverCalled,
				},
				expectedUser:  user.User{},
				expectedError: identity.ErrUnknownRole,
			},
			{
				name:  "Fail: User not confirmed",
				input: input_tenantUser,
				setupSteps: []mockSetupFunc{
					step1GetUserErr_NotConfirmed,
					step2NeverCalled,
				},
				expectedUser:  user.User{},
				expectedError: auth.ErrAccountNotConfirmed,
			},

			// Step 2
			{
				name:  "Fail: Wrong password",
				input: input_wrongPassword,
				setupSteps: []mockSetupFunc{
					step1GetUserOk_CaseTenantUser,
					step2Fail,
				},
				expectedUser:  user.User{},
				expectedError: auth.ErrWrongCredentials,
			},
		*/
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// NOTA: il controller di gomock va inizializzato qua dentro!
			mockController := gomock.NewController(t)

			mockLogger := zaptest.NewLogger(t)
			mockHasher := cryptoMocks.NewMockSecretHasher(mockController)
			mockTokenPort := mocks.NewMockConfirmAccountTokenPort(mockController)
			mockSaveUserPort := userMocks.NewMockSaveUserPort(mockController)

			// Slice con chiamate da eseguire
			var expectedCalls []any // NOTA: Dovrebbe essere []*gomock.Call, però il compilatore non accetta

			// Collezione le chiamate per questo test case
			for _, step := range tc.setupSteps {
				call := step(mockLogger, mockHasher, mockTokenPort, mockSaveUserPort)
				if call != nil {
					expectedCalls = append(expectedCalls, call)
				}
			}

			// Richiedi ordine nelle chiamate
			if len(expectedCalls) > 0 {
				gomock.InOrder(expectedCalls...)
			}

			// Crea servizio con porte mock
			service := auth.NewAuthSessionService(mockHasher, mockSaveUserPort)

			// Esegui funzione in oggetto
			loggedUser, err := service.LoginUser(tc.input)

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
